package controller

import (
	"oauth-server-lite/controller/midd"
	"oauth-server-lite/g"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
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

	oauth := r.Group("/oauth/v1")
	oauth.Use(NoCache())
	oauth.POST("/token", getOauthToken)

	authorize := r.Group("/oauth/v1")
	authorize.Use(NoCache())
	authorize.Use(midd.AuthorizeLoginCheck)
	authorize.GET("/authorize", getAuthorizeCode)

	user := r.Group("/user")
	user.GET("/login", loginGet)
	user.POST("/login", loginPost)
	user.GET("/captcha", getCaptcha)

	user.GET("/logout", logout)

	userinfo := r.Group("/oauth/v1")
	userinfo.Use(NoCache())
	userinfo.Use(midd.OauthTokenCheckMidd)
	userinfo.GET("/userinfo", getUserInfo)
	userinfo.POST("/userinfo", getUserInfo)

}
