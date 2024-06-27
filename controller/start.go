package controller

import (
	"fmt"
	"net/http"

	"oauth-server-lite/g"

	"github.com/gin-gonic/gin"
)

func InitGin(listen string) (httpServer *http.Server) {
	if g.Config().Logger.Level == "DEBUG" {
		gin.SetMode((gin.DebugMode))
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	if g.Config().Logger.Level == "DEBUG" {
		r.Use(gin.Logger())
	}
	r.SetTrustedProxies(g.Config().Http.TrustProxy)

	r.Use(gin.Recovery())
	r.Use(CORS())

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("oauth-server-demo version %s", g.VERSION))
	})
	Routes(r)
	httpServer = &http.Server{
		Addr:    g.Config().Http.Listen,
		Handler: r,
	}
	return
}

func CORS() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Writer.Header().Add("Access-Control-Allow-Origin", context.Request.Header.Get("Origin"))
		context.Writer.Header().Set("Access-Control-Max-Age", "86400")
		context.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		context.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, X-API-KEY, Authorization, Cookie")
		context.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")

		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(200)
		} else {
			context.Next()
		}
	}
}

func NoCache() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Writer.Header().Add("Cache-Control", "no-store")
		context.Writer.Header().Add("Pragma", "no-cache")
		context.Next()
	}
}
