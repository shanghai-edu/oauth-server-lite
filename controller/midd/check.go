package midd

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"oauth-server-lite/controller/location-utils"
	"strings"

	"oauth-server-lite/g"
	"oauth-server-lite/models/oauth"
	"oauth-server-lite/models/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/logger"
	"oauth-server-lite/models/cas"
)

func OauthTokenCheckMidd(c *gin.Context) {
	headerToken := c.Request.Header.Get("Authorization")
	postToken := c.PostForm("access_token")
	getToken := c.Query("access_token")
	if headerToken != "" {
		ht := strings.TrimPrefix(strings.ToLower(headerToken), "bearer ")
		if checkToken(ht) {
			c.Set("access_token", ht)
			c.Next()
			return
		}
	}
	if postToken != "" && checkToken(postToken) {
		c.Set("access_token", postToken)
		c.Next()
		return
	}
	if getToken != "" && checkToken(getToken) {
		c.Set("access_token", getToken)
		c.Next()
		return
	}
	c.JSON(http.StatusUnauthorized, OauthErrorRes(g.InvalidToken))
	c.Abort()
}

func checkToken(accessToken string) bool {
	_, err := oauth.GetAccessToken(accessToken)
	if err == nil {
		return true
	} else {
		return false
	}
}

func XAPICheckMidd(c *gin.Context) {
	key := c.Request.Header.Get("X-API-KEY")
	if !checkXApiKey(key) {
		c.JSON(http.StatusUnauthorized, OauthErrorRes(g.InvalidAPIKey))
		c.Abort()
		return
	}
	if !utils.IPCheck(c.ClientIP(), g.Config().Http.ManageIP) {
		c.JSON(http.StatusUnauthorized, OauthErrorRes(g.InvalidIP))
		c.Abort()
		return
	}
	c.Next()
}

func checkXApiKey(key string) bool {
	return key == g.Config().Http.XAPIKey
}

func AuthorizeLoginCheck(c *gin.Context) {
	location := location_utils.GetLocation(c.Request)
	session := sessions.Default(c)
	s := session.Get("user_id")
	currentURL := location.Scheme + "://" + location.Host + c.Request.URL.Path + "?" + c.Request.URL.RawQuery
	logger.Debugf("/oauth/authorize: %s", currentURL)

	isAuthorized := session.Get("is_authorized")
	logger.Debugf("user_id = %v, is_authorized = %v", s, isAuthorized)

	if s == nil || isAuthorized == nil {
		session.Set("current_url", currentURL)
		session.Set("client_id", c.Query("client_id"))
		logger.Debug(session.Get("current_url"))
		if err := session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, OauthErrorRes(g.ServerError))
			c.Abort()
			return
		}

		service := location.Scheme + "://" + location.Host + "/user/authorize"
		casLogin := g.Config().CAS + "login?service=" + service
		c.Redirect(http.StatusMovedPermanently, casLogin)
		return

	}
	c.Set("user_id", s.(string))
	c.Next()
}

func DeviceAuthorizeLoginCheck(inputs oauth.DeviceTokenInput, c *gin.Context) (userId string, err error) {
	location := location_utils.GetLocation(c.Request)
	session := sessions.Default(c)

	rc := g.ConnectRedis().Get()
	defer rc.Close()
	// 尝试从redis中通过DeviceCode获取UserID
	userIdKey := g.Config().RedisNamespace.OAuth + "device_user_id:" + inputs.DeviceCode
	redisUserId, err := rc.Do("GET", userIdKey)
	if err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	if redisUserId == nil {
		session.Set("client_id", inputs.ClientID)
		if err = session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, OauthErrorRes(g.ServerError))
			c.Abort()
			return
		}
		service := location.Scheme + "://" + location.Host + "/user/device/authorize"
		casLogin := g.Config().CAS + "login?service=" + service
		c.Redirect(http.StatusMovedPermanently, casLogin)
	}
	userIdBytes, isSuccess := redisUserId.([]byte)
	if !isSuccess {
		err = errors.New(g.ServerError)
		logger.Error("transfer to []byte failed:" + err.Error())
		return
	}
	userId = string(userIdBytes)

	if err = c.Bind(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, OauthErrorRes(g.InvalidRequest))
		return
	}
	isAuthorizedKey := g.Config().RedisNamespace.OAuth + "device_is_authorized:" + inputs.DeviceCode
	redisIsAuthorized, err := rc.Do("GET", isAuthorizedKey)
	if err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	var isAuthorized string
	if redisIsAuthorized != nil {
		isAuthorizedBytes, isSuccess := redisIsAuthorized.([]byte)
		if !isSuccess {
			err = errors.New(g.ServerError)
			logger.Error("transfer to []byte failed:" + err.Error())
			return
		}
		isAuthorized = string(isAuthorizedBytes)
	}

	if isAuthorized != "1" {
		err = errors.New(g.InvalidGrant)
		return
	}
	return
}

// 根据type分区Service
func CASLoginCheckWithType(c *gin.Context, serviceType string) {
	location := location_utils.GetLocation(c.Request)
	session := sessions.Default(c)
	s := session.Get("user_id")
	if s == nil {
		var userId string

		var err error
		ticket := c.Query("ticket")
		if ticket == "" {
			redirectToCASLogin(c, serviceType)
			return
		}

		service := location.Scheme + "://" + location.Host + c.Request.URL.Path
		userId, err = validateServiceTicket(ticket, service)
		// 有时会校验失败
		if err != nil {
			logger.Error(err)
			redirectToCASLogin(c, serviceType)
			return
		}

		session.Set("user_id", userId)
		if err := session.Save(); err != nil {
			errorMsg := g.LoginErrorDescription[g.ServerError]
			ErrorHTML(errorMsg, c)
			c.Abort()
			return
		}
		c.Set("user_id", userId)
		c.Next()
	} else {
		c.Set("user_id", s.(string))
		c.Next()
	}
}

func redirectToCASLogin(c *gin.Context, serviceType string) {
	location := location_utils.GetLocation(c.Request)
	var service string
	switch serviceType {
	case g.AuthorizationCode:
		service = location.Scheme + "://" + location.Host + "/user/authorize"
	case g.DeviceFlow:
		service = location.Scheme + "://" + location.Host + "/user/device/authorize"
	default:
		service = location.Scheme + "://" + location.Host + "/user/authorize"
	}
	casLogin := g.Config().CAS + "login?service=" + service
	c.Redirect(http.StatusMovedPermanently, casLogin)
}

func BasicAuthResolve(r *http.Request) (username, password string, err error) {
	basicAuthPrefix := "Basic "

	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, basicAuthPrefix) {
		var bs []byte
		bs, err = base64.StdEncoding.DecodeString(auth[len(basicAuthPrefix):])
		if err != nil {
			return
		}
		split := strings.SplitN(string(bs), ":", 2)
		if len(split) != 2 {
			err = errors.New("basic authorization format is not correct")
			return
		}
		username = split[0]
		password = split[1]
		return
	}
	return
}

func validateServiceTicket(ticket, service string) (userID string, err error) {
	casUrl, err := url.Parse(g.Config().CAS)
	if err != nil {
		return
	}
	serviceUrl, err := url.Parse(service)
	if err != nil {
		return
	}

	resOptions := &cas.RestOptions{
		CasURL:     casUrl,
		ServiceURL: serviceUrl,
	}
	resClient := cas.NewRestClient(resOptions)

	res, err := resClient.ValidateServiceTicket(cas.ServiceTicket(ticket))
	if err != nil {
		return
	}
	userID = res.User
	return
}
