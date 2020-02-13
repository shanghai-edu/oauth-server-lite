package controller

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"oauth-server-lite/controller/midd"
	"oauth-server-lite/g"

	"github.com/gin-gonic/gin"

	"oauth-server-lite/models/user"
)

func getUserInfo(c *gin.Context) {
	accessToken := c.GetString("access_token")
	attr, err := user.GetAttrByAccessToken(accessToken)
	if err != nil {
		log.Errorln(err)
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.ServerError))
		return
	}
	c.JSON(http.StatusOK, attr)
}
