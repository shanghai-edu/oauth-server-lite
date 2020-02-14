package midd

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"oauth-server-lite/g"
	"oauth-server-lite/models/oauth"
	"oauth-server-lite/models/utils"

	"github.com/gin-contrib/sessions"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func OauthTokenCheckMidd(c *gin.Context) {
	headerToken := c.Request.Header.Get("Authorization")
	postToken := c.PostForm("access_token")
	getToken := c.Query("access_token")
	if headerToken != "" {
		ht := strings.TrimPrefix(headerToken, "Bearer ")
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
	ac, err := oauth.GetAccessToken(accessToken)
	log.Debug(ac)
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
	location := GetLocation(c.Request)
	session := sessions.Default(c)
	s := session.Get("user_id")
	currentURL := location.Scheme + "://" + location.Host + c.Request.URL.Path + "?" + c.Request.URL.RawQuery
	log.Debugf("/oauth/authorize: %s", currentURL)

	if s == nil {
		session.Set("current_url", currentURL)
		log.Debugln(session.Get("current_url"))
		if err := session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, OauthErrorRes(g.ServerError))
			return
		}
		loginPath := location.Scheme + "://" + location.Host + "/user/login"
		c.Redirect(http.StatusMovedPermanently, loginPath)
		return
	}
	c.Set("user_id", s.(string))
	c.Next()
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
