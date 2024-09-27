package controller

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/logger"
	"oauth-server-lite/controller/midd"
	"oauth-server-lite/g"
	"oauth-server-lite/models/oauth"
)

func Routes(r *gin.Engine) {
	r.LoadHTMLGlob("template/*.html")
	r.Static("/resource/", "resource/")

	store, err := redis.NewStore(10, "tcp", g.Config().Redis.Dsn, g.Config().Redis.Password, []byte("oauth-server-lite"))
	if err != nil {
		panic(err)
	}
	store.Options(sessions.Options{
		Path:     g.Config().Http.SessionOptions.Path,
		Domain:   g.Config().Http.SessionOptions.Domain,
		MaxAge:   g.Config().Http.SessionOptions.MaxAge,
		Secure:   g.Config().Http.SessionOptions.Secure,
		HttpOnly: g.Config().Http.SessionOptions.HttpOnly,
	})

	r.Use(sessions.Sessions("mysession", store))

	device := r.Group("/oauth2/device")
	device.Use(NoCache())
	device.Use(DeviceParamCheck)
	device.POST("/authorize", getDeviceCode)

	oauth := r.Group("/oauth2")
	oauth.Use(NoCache())
	oauth.POST("/token", getOauthToken)
	oauth.GET("/error", oauth2ErrorPage)

	authorize := r.Group("/oauth2")
	authorize.Use(NoCache())
	authorize.Use(AuthorizeCodeParamCheck)
	authorize.Use(midd.AuthorizeLoginCheck)
	authorize.GET("/authorize", getAuthorizeCode)

	// device flow模式用户相关操作
	deviceUser := r.Group("/user/device")
	deviceUser.Use(NoCache())
	deviceUser.Use(func(c *gin.Context) {
		midd.CASLoginCheckWithType(c, g.DeviceFlow)
	})
	deviceUser.GET("/authorize", deviceAuthorizeGet)
	deviceUser.POST("/authorize", deviceAuthorizePost)

	// authorization code模式用户相关操作
	user := r.Group("/user")
	user.Use(NoCache())
	user.Use(func(c *gin.Context) {
		midd.CASLoginCheckWithType(c, g.AuthorizationCode)
	})
	user.GET("/authorize", authorizeGet)
	user.POST("/authorize", authorizePost)

	r.GET("/user/logout", logout)
	r.GET("/user/callback", logoutCallback)

	check := r.Group("/oauth2")
	check.Use(NoCache())
	check.Use(midd.OauthTokenCheckMidd)
	check.GET("/userinfo", getUserinfo)
	check.GET("/userinfo/flat", getUserinfoflat)

	manage := r.Group("/manage/v1")
	manage.Use(midd.XAPICheckMidd)
	manage.POST("/client", addClient)
	manage.GET("/client/:client_id", getClient)
	manage.DELETE("/client/:client_id", delClient)
	manage.GET("/clients", getAllClients)

}

type Userinfo struct {
	UserId string `json:"userId"`
}

func getUserinfo(c *gin.Context) {
	acessToken := c.GetString("access_token")
	at, err := oauth.GetAccessToken(acessToken)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusOK, midd.ErrorRes(midd.InternalAPIError, "Token 错误"))
		return
	}
	validGrantTypes := []string{g.Password, g.AuthorizationCode, g.DeviceFlow, g.ClientCredentials}
	if !strings.Contains(strings.Join(validGrantTypes, ","), at.GrantType) {
		logger.Error(at)
		c.JSON(http.StatusOK, midd.ErrorRes(midd.InternalAPIError, "Token 错误"))
		return
	}
	userRes := Userinfo{
		at.UserID,
	}

	c.JSON(http.StatusOK, midd.SuccessRes(userRes))
}

func getUserinfoflat(c *gin.Context) {
	acessToken := c.GetString("access_token")
	at, err := oauth.GetAccessToken(acessToken)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusOK, midd.ErrorRes(midd.InternalAPIError, "Token 错误"))
		return
	}
	if !(at.GrantType == "password" || at.GrantType == "authorization_code") {
		logger.Error(at)
		c.JSON(http.StatusOK, midd.ErrorRes(midd.InternalAPIError, "Token 错误"))
		return
	}
	userRes := Userinfo{
		at.UserID,
	}
	c.JSON(http.StatusOK, userRes)
}
