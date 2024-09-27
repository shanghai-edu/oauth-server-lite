package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	location_utils "oauth-server-lite/controller/location-utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/logger"
	"oauth-server-lite/controller/midd"
	"oauth-server-lite/g"
	"oauth-server-lite/models/oauth"
)

var apis = []midd.OauthAppApi{
	{
		AppId:          10001,
		AppName:        "授权测试",
		Description:    "授权测试",
		ApiId:          "200002",
		ApiName:        "授权测试接口",
		ApiDescription: "授权测试接口",
	},
}

func authorizeGet(c *gin.Context) {
	session := sessions.Default(c)
	userId := c.GetString("user_id")
	clientId := session.Get("client_id")
	_url := session.Get("current_url")
	if clientId == nil || _url == nil {
		logger.Error(clientId, _url, userId)
		midd.ErrorHTML("会话已经过期，请关闭页面重试4", c)
		return
	}

	client := oauth.GetClientByClientID(clientId.(string))

	midd.AuthorizeHTML(userId, client.AppName, client.PrivacyUrl, apis, c, "p1.html")

}

func authorizePost(c *gin.Context) {
	session := sessions.Default(c)
	userId := c.GetString("user_id")
	clientId := session.Get("client_id")
	url := session.Get("current_url")
	if clientId == nil || url == nil || userId == "" {
		logger.Error(clientId, url, userId)
		midd.ErrorHTML("会话已经过期，请关闭页面重试5", c)
		return
	}

	isAuthorized := c.PostForm("isauthorized")

	switch isAuthorized {
	//不同意，原地注销
	//注销后回跳到开发平台主页
	case "0":
		location := location_utils.GetLocation(c.Request)
		logoutUrl := location.Scheme + "://" + location.Host + "/user/logout?redirect_uri=https://developer.ecnu.edu.cn/"
		c.Redirect(http.StatusMovedPermanently, logoutUrl)
	//同意
	case "1":
		session.Set("is_authorized", "1")
		session.Save()
		//记住选择
		c.Redirect(http.StatusMovedPermanently, url.(string))
	//无参数，原地不动
	default:
		client := oauth.GetClientByClientID(clientId.(string))
		midd.AuthorizeHTML(userId, client.AppName, client.PrivacyUrl, apis, c, "p1.html")
	}
}

func logout(c *gin.Context) {
	location := location_utils.GetLocation(c.Request)
	redirect_uri := c.Query("redirect_uri")
	service := location.Scheme + "://" + location.Host + "/user/callback?redirect_uri=" + redirect_uri

	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()

	c.Redirect(http.StatusMovedPermanently, g.Config().CAS+"logout?service="+service)

}

func logoutCallback(c *gin.Context) {
	redirect_uri := c.Query("redirect_uri")
	url, err := url.Parse(redirect_uri)
	if err != nil {
		logger.Warning(err)
	}
	logger.Debug("注销回调：redirect_uri = " + redirect_uri)
	logger.Debug("注销回调：host = " + url.Hostname())
	if !oauth.CheckDomainValid(url.Hostname()) || err != nil || redirect_uri == "" {
		redirect_uri = "https://developer.ecnu.edu.cn/"
	}
	c.Redirect(http.StatusMovedPermanently, redirect_uri)
}

func oauth2ErrorPage(c *gin.Context) {
	errCode := c.Query("errcode")
	switch errCode {
	case midd.WxWorkUserNotFound:
		midd.ErrorHTML("您不是学校企业微信的用户，请通过其他通道访问服务", c)
		return
	case midd.UserNotFound:
		midd.ErrorHTML("无法获取您的学工号信息，请与信息化治理办公室联系: its@ecnu.edu.cn", c)
		return
	default:
		midd.ErrorHTML("会话已经过期，请关闭页面重试4", c)
		return
	}
}

func deviceAuthorizeGet(c *gin.Context) {
	userId := c.GetString("user_id")
	midd.AuthorizeHTML(userId, "client.AppName", "client.PrivacyUrl", apis, c, "device.html")
}

func deviceAuthorizePost(c *gin.Context) {
	userId := c.GetString("user_id")

	// 获取用户是否同意授权
	isAuthorized := c.PostForm("is_device_authorized")
	switch isAuthorized {
	//TODO：不同意暂时重定向到开发平台主页
	case "0":
		location := location_utils.GetLocation(c.Request)
		logoutUrl := location.Scheme + "://" + location.Host + "/user/logout?redirect_uri=https://developer.ecnu.edu.cn/"
		c.Redirect(http.StatusMovedPermanently, logoutUrl)
	case "1":
		// 校验user_code是否一致
		userCode := c.PostForm("user_code")
		rc := g.ConnectRedis().Get()
		defer rc.Close()
		deviceCodeOutputKey := g.Config().RedisNamespace.OAuth + "device_code_output:" + userCode
		redisDeviceCodeOutput, err := rc.Do("GET", deviceCodeOutputKey)
		if err != nil {
			logger.Error(err)
			err = errors.New(g.ServerError)
			return
		}
		if redisDeviceCodeOutput == nil {
			err = errors.New(g.InvalidGrant)
			return
		}

		var deviceCodeOutput oauth.DeviceCodeOutput
		err = json.Unmarshal(redisDeviceCodeOutput.([]byte), &deviceCodeOutput)
		if err != nil {
			logger.Error(err)
			err = errors.New(g.ServerError)
			return
		}
		if deviceCodeOutput.UserCode != userCode {
			err = errors.New(g.InvalidGrant)
			return
		}

		isAuthorizedKey := g.Config().RedisNamespace.OAuth + "device_is_authorized:" + deviceCodeOutput.DeviceCode
		_, err = rc.Do("SET", isAuthorizedKey, isAuthorized, "EX", g.Config().CodeExpired)
		if err != nil {
			return
		}
		if userId != "" {
			userIdKey := g.Config().RedisNamespace.OAuth + "device_user_id:" + deviceCodeOutput.DeviceCode
			_, err = rc.Do("SET", userIdKey, userId, "EX", g.Config().CodeExpired)
			if err != nil {
				return
			}
		}
		// 渲染授权成功页面
		midd.AuthorizeHTML(userId, "client.AppName", "client.PrivacyUrl", apis, c, "device_success.html")
	default:
		location := location_utils.GetLocation(c.Request)
		logoutUrl := location.Scheme + "://" + location.Host + "/user/logout?redirect_uri=https://developer.ecnu.edu.cn/"
		c.Redirect(http.StatusMovedPermanently, logoutUrl)
	}
}
