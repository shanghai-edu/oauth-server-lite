package controller

import (
	"bytes"
	"net/http"

	"oauth-server-lite/controller/midd"
	"oauth-server-lite/g"
	"oauth-server-lite/models/user"

	"github.com/dchest/captcha"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func getCaptcha(c *gin.Context) {
	session := sessions.Default(c)
	captchaIdSession := session.Get("captcha_id")
	if captchaIdSession == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	captchaId := captchaIdSession.(string)
	captcha.Reload(captchaId)
	var content bytes.Buffer
	captcha.WriteImage(&content, captchaId, captcha.StdWidth, captcha.StdHeight)
	contentType := "image/png"
	contentLength := int64(content.Len())

	extraHeaders := map[string]string{
		"Cache-Control": `no-cache, no-store, must-revalidate"`,
		"Expires":       "0",
	}
	c.DataFromReader(http.StatusOK, contentLength, contentType, bytes.NewReader(content.Bytes()), extraHeaders)
}

func loginGet(c *gin.Context) {
	var errorMsg, captchaId string
	session := sessions.Default(c)
	if captchaIdSession := session.Get("captcha_id"); captchaIdSession != nil {
		captchaId = captchaIdSession.(string)
	}
	midd.LoginHTML(errorMsg, captchaId, c)
}

func logout(c *gin.Context) {
	var errorMsg string
	session := sessions.Default(c)
	s := session.Get("user_id")
	if s != nil {
		session.Delete("user_id")
		err := session.Save()
		if err != nil {
			errorMsg = g.LoginErrorDescription[g.ServerError]
			log.Errorln("session save failed", err.Error())
			midd.LogoutHTML(errorMsg, c)
			return
		}
	}
	midd.LogoutHTML(errorMsg, c)
}

type loginInput struct {
	UserID    string `form:"user_id" binding:"required"`
	Password  string `form:"password" binding:"required"`
	Captcha   string `form:"captcha"`
	CaptchaId string `form:"captcha_id"`
}

func loginPost(c *gin.Context) {
	inputs := loginInput{}
	var errorMsg, captchaId string
	if err := c.Bind(&inputs); err != nil {
		errorMsg = g.LoginErrorDescription[g.InvalidPassword]
		midd.LoginHTML(errorMsg, captchaId, c)
		return
	}
	if user.CheckLock(inputs.UserID, c.ClientIP()) {
		errorMsg = g.LoginErrorDescription[g.IpLocked]
		log.Info(g.IpLocked, inputs.UserID, c.ClientIP())
		midd.LoginHTML(errorMsg, captchaId, c)
		return
	}
	session := sessions.Default(c)
	if inputs.CaptchaId != "" {
		if !captcha.VerifyString(inputs.CaptchaId, inputs.Captcha) {
			errorMsg = g.LoginErrorDescription[g.InvalidCaptcha]
			captchaId = captcha.NewLen(captcha.DefaultLen)
			session.Set("captcha_id", captchaId)
			if err := session.Save(); err != nil {
				errorMsg = g.LoginErrorDescription[g.ServerError]
				log.Errorln("session save failed", err.Error())
				midd.LoginHTML(errorMsg, captchaId, c)
				return
			}
			midd.LoginHTML(errorMsg, captchaId, c)
			return
		}
	}

	if err := user.LdapLogin(inputs.UserID, inputs.Password); err != nil {
		count := user.UpdateFailedCount(inputs.UserID, c.ClientIP())
		if count == g.Config().MaxFailed {
			user.CreateLock(inputs.UserID, c.ClientIP())
		}
		captchaId = captcha.NewLen(captcha.DefaultLen)
		session.Set("captcha_id", captchaId)
		if err := session.Save(); err != nil {
			errorMsg = g.LoginErrorDescription[g.ServerError]
			log.Errorln("session save failed", err.Error())
			midd.LoginHTML(errorMsg, captchaId, c)
			return
		}
		errorMsg = g.LoginErrorDescription[g.InvalidPassword]
		midd.LoginHTML(errorMsg, captchaId, c)
		return
	}
	user.DeleteFailedCount(inputs.UserID, c.ClientIP())

	session.Set("user_id", inputs.UserID)
	url := session.Get("current_url")
	if err := session.Save(); err != nil {
		errorMsg = g.LoginErrorDescription[g.ServerError]
		log.Errorln("session save failed", err.Error())
		midd.LoginHTML(errorMsg, captchaId, c)
		return
	}
	log.Infoln("Login Success: ", inputs.UserID, " ", c.ClientIP())

	if url == nil {
		errorMsg = g.LoginErrorDescription[g.ServerError]
		log.Errorln("missing redirectUrl", c.ClientIP())
		midd.LoginHTML(errorMsg, captchaId, c)
		return
	}
	c.Redirect(http.StatusMovedPermanently, url.(string))
}
