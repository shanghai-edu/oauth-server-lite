package controller

import (
	"net/http"

	"oauth-server-lite/controller/midd"
	"oauth-server-lite/g"
	"oauth-server-lite/models/oauth"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
)

type OauthClient struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	Domain       string `json:"domain"`
	WhiteIP      string `json:"white_ip"`
	Scope        string `json:"scope"`
	Description  string `json:"description"`
}

func getAllClients(c *gin.Context) {
	clients := oauth.GetClients()
	var res []OauthClient
	if err := copier.Copy(&res, &clients); err != nil {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.ServerError))
		log.Errorln(err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func getClient(c *gin.Context) {
	clientID := c.Param("client_id")
	client := oauth.GetClientByClientID(clientID)
	if client.ID == 0 {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.InvalidClient))
		return
	}
	var res OauthClient
	if err := copier.Copy(&res, &client); err != nil {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.ServerError))
		log.Errorln(err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func delClient(c *gin.Context) {
	clientID := c.Param("client_id")
	client := oauth.GetClientByClientID(clientID)
	if client.ID == 0 {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.InvalidClient))
		return
	}
	var res OauthClient
	if err := copier.Copy(&res, &client); err != nil {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.ServerError))
		log.Errorln(err)
		return
	}
	if err := oauth.DeleteClient(client); err != nil {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.ServerError))
		log.Errorln(err)
		return
	}
	c.JSON(http.StatusOK, res)
}

type clientInput struct {
	GrantType   string `json:"grant_type" binding:"required"`
	Domain      string `json:"domain" binding:"required"`
	WhiteIP     string `json:"white_ip"`
	Description string `json:"description"`
}

func addClient(c *gin.Context) {
	inputs := clientInput{}
	if err := c.Bind(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.InvalidRequest))
		return
	}
	switch inputs.GrantType {
	case "authorization_code":
		var res OauthClient
		client, err := oauth.GenerateAuthorizationCodeClient(inputs.Description, inputs.Domain)
		if err != nil {
			c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.ServerError))
			log.Errorln(err)
			return
		}
		if err := copier.Copy(&res, &client); err != nil {
			c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.ServerError))
			log.Errorln(err)
			return
		}
		c.JSON(http.StatusOK, res)
	case "client_credentials":
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.UnsupportedGrantType))
		return
	default:
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.UnsupportedGrantType))
		return

	}
}
