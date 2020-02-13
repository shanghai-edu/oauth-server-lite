package controller

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"oauth-server-lite/controller/midd"
	"oauth-server-lite/g"
	"oauth-server-lite/models/oauth"
	"oauth-server-lite/models/utils"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

type authorizationInput struct {
	ResponseType string `json:"response_type" form:"response_type" binding:"required"`
	ClientID     string `json:"client_id" form:"client_id" binding:"required"`
	RedirectUri  string `json:"redirect_uri" form:"redirect_uri" binding:"required"`
	Scope        string `json:"scope" form:"scope"`
	State        string `json:"state" form:"state"`
}

type tokenInput struct {
	RefreshToken string `form:"refresh_token"`
	GrantType    string `form:"grant_type"`
	Code         string `form:"code"`
	RedirectUri  string `form:"redirect_uri"`
	ClientID     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
}

type clientCredentialsToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

type authorizationCodeToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

func checkOauthClient(request *http.Request, inputs tokenInput) (clientID string, err error) {
	var clientSecret string
	clientID, clientSecret, err = midd.BasicAuthResolve(request)
	if err != nil {
		err = errors.New(g.InvalidClient)
		return
	}
	if clientID == "" {
		if inputs.ClientID == "" || inputs.ClientSecret == "" {
			err = errors.New(g.InvalidRequest)
			return
		}
		clientID = inputs.ClientID
		clientSecret = inputs.ClientSecret
	}
	_, err = oauth.CheckClientPass(clientID, clientSecret)
	if err != nil {
		err = errors.New(g.InvalidClient)
		return
	}
	return
}

func genClientCredentialsToken(clientIP, clientID string) (resToken clientCredentialsToken, err error) {
	oauthClient, err := oauth.CheckClientIP(clientIP, clientID)
	if err != nil {
		err = errors.New(g.InvalidIP)
		return
	}
	if oauthClient.GrantType != "client_credentials" {
		err = errors.New(g.InvalidGrant)
		return
	}
	token, err := oauth.CreateToken(clientID, "")
	if err != nil {
		err = errors.New(g.ServerError)
		return
	}
	if err = copier.Copy(&resToken, &token); err != nil {
		err = errors.New(g.ServerError)
		return
	}
	return
}

func genAuthorizationCodeToken(inputs tokenInput, clientID string) (resToken authorizationCodeToken, err error) {
	var authorizationCodeInput oauth.AuthorizationCodeTokenInput
	if err = copier.Copy(&authorizationCodeInput, &inputs); err != nil {
		err = errors.New(g.ServerError)
		return
	}
	if authorizationCodeInput.Code == "" || authorizationCodeInput.RedirectUri == "" {
		err = errors.New(g.InvalidRequest)
		return
	}
	userID, err := oauth.CheckAuthorizationCode(authorizationCodeInput)
	if err != nil {
		return
	}
	token, err := oauth.CreateToken(clientID, userID)
	if err != nil {
		err = errors.New(g.ServerError)
		return
	}
	if err = copier.Copy(&resToken, &token); err != nil {
		err = errors.New(g.ServerError)
		return
	}
	return
}

func genRefreshToken(refreshToken string) (resToken authorizationCodeToken, err error) {
	if refreshToken == "" {
		err = errors.New(g.InvalidRequest)
		return
	}
	token, err := oauth.RefreshAccessToken(refreshToken)
	if err != nil {
		err = errors.New(g.ServerError)
		return
	}
	if err = copier.Copy(&resToken, &token); err != nil {
		err = errors.New(g.ServerError)
		return
	}
	return
}

func getOauthToken(c *gin.Context) {
	inputs := tokenInput{}
	if err := c.Bind(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.InvalidRequest))
		return
	}
	clientID, err := checkOauthClient(c.Request, inputs)
	if err != nil {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(err.Error()))
		return
	}
	switch inputs.GrantType {
	case "client_credentials":
		resToken, err := genClientCredentialsToken(c.ClientIP(), clientID)
		if err != nil {
			c.JSON(http.StatusBadRequest, midd.OauthErrorRes(err.Error()))
			return
		}
		c.JSON(http.StatusOK, resToken)
		return

	case "authorization_code", "":
		resToken, err := genAuthorizationCodeToken(inputs, clientID)
		if err != nil {
			c.JSON(http.StatusBadRequest, midd.OauthErrorRes(err.Error()))
			return
		}
		c.JSON(http.StatusOK, resToken)

	case "refresh_token":
		resToken, err := genRefreshToken(inputs.RefreshToken)
		if err != nil {
			c.JSON(http.StatusBadRequest, midd.OauthErrorRes(err.Error()))
			return
		}
		c.JSON(http.StatusOK, resToken)
		return
	default:
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.UnsupportedGrantType))
		return
	}
}

func getAuthorizeCode(c *gin.Context) {
	inputs := authorizationInput{}
	if err := c.Bind(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.InvalidRequest))
		if err != nil {
			log.Debugln(err)
		}
		return
	}
	if inputs.ResponseType != "code" {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.UnsupportedResponseType))
		return
	}
	redirectUri, err := url.Parse(inputs.RedirectUri)
	if err != nil {
		log.Debugln(err)
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.UnauthorizedClient))
		return
	}
	oauthClient := oauth.GetClientByClientID(inputs.ClientID)
	if !(oauthClient.GrantType == "authorization_code" && oauthClient.Domain == redirectUri.Hostname()) {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.UnauthorizedClient))
		return
	}

	var authorizationCode oauth.AuthorizationCode
	if err := copier.Copy(&authorizationCode, &inputs); err != nil {
		log.Debugln(err)
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.ServerError))
		return
	}
	authorizationCode.UserID = c.GetString("user_id")
	log.Debugln("userid:", authorizationCode.UserID)
	code, err := oauth.CreateAuthorizationCode(authorizationCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, midd.OauthErrorRes(g.ServerError))
		return
	}
	var state string
	if inputs.State == "" {
		s, err := utils.GenerateVcode()
		if err != nil {
			c.JSON(http.StatusInternalServerError, midd.OauthErrorRes(g.ServerError))
			return
		}
		state = s
	} else {
		state = inputs.State
	}
	var redirectURL string
	if strings.Contains(inputs.RedirectUri, "?") {
		redirectURL = inputs.RedirectUri + "&code=" + code + "&state=" + state
	} else {
		redirectURL = inputs.RedirectUri + "?code=" + code + "&state=" + state
	}

	c.Redirect(http.StatusMovedPermanently, redirectURL)
}
