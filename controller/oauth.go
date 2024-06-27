package controller

import (
	"errors"
	"github.com/gin-contrib/sessions"
	"net/http"
	"net/url"
	"strings"

	"oauth-server-lite/controller/midd"
	"oauth-server-lite/g"
	"oauth-server-lite/models/oauth"
	"oauth-server-lite/models/utils"

	"github.com/toolkits/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

type authorizationInput struct {
	ResponseType        string `json:"response_type" form:"response_type"`
	ClientID            string `json:"client_id" form:"client_id" binding:"required"`
	RedirectUri         string `json:"redirect_uri" form:"redirect_uri"`
	Scope               string `json:"scope" form:"scope"`
	State               string `json:"state" form:"state"`
	CodeChallenge       string `json:"code_challenge" form:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method" form:"code_challenge_method"`
}

type tokenInput struct {
	RefreshToken string `form:"refresh_token"`
	GrantType    string `form:"grant_type"`
	Code         string `form:"code"`
	RedirectUri  string `form:"redirect_uri"`
	ClientID     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
	Username     string `form:"username"`
	Password     string `form:"password"`
	CodeVerifier string `form:"code_verifier"`
	DeviceCode   string `form:"device_code"`
}

type clientCredentialsToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

type resourceOwnerPasswordCredentialsToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

type authorizationCodeToken struct {
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

type pkceToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

type deviceToken struct {
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
		logger.Error(err)
		err = errors.New(g.InvalidClient)
		return
	}
	if inputs.ClientID == "" {
		err = errors.New(g.InvalidRequest)
		return
	}
	if clientID == "" {
		clientID = inputs.ClientID
	}
	// 如下情况无需校验client_secret
	if inputs.CodeVerifier != "" || inputs.GrantType == g.RefreshToken || inputs.GrantType == g.DeviceFlow {
		oauthClient := oauth.GetClientByClientID(clientID)
		if oauthClient.ID == 0 {
			err = errors.New(g.InvalidClient)
			return
		}
	} else {
		if inputs.ClientSecret == "" {
			err = errors.New(g.InvalidRequest)
			return
		}
		clientSecret = inputs.ClientSecret
		_, err = oauth.CheckClientPass(clientID, clientSecret)
		if err != nil {
			logger.Error(err)
			err = errors.New(g.InvalidClient)
			return
		}
	}
	return
}

func genClientCredentialsToken(clientID string) (resToken clientCredentialsToken, err error) {
	oauthClient := oauth.GetClientByClientID(clientID)
	if oauthClient.ID == 0 {
		err = errors.New(g.InvalidClient)
		return
	}
	if !utils.InStrings(g.ClientCredentials, oauthClient.GrantTypes, ",") {
		err = errors.New(g.InvalidGrant)
		return
	}
	token, err := oauth.CreateToken(clientID, "", g.ClientCredentials)
	if err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	if err = copier.Copy(&resToken, &token); err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	return
}

func genResourceOwnerPasswordCredentialsToken(username, password, clientID string) (resToken resourceOwnerPasswordCredentialsToken, err error) {
	oauthClient := oauth.GetClientByClientID(clientID)
	if oauthClient.ID == 0 {
		err = errors.New(g.InvalidClient)
		return
	}

	if !utils.InStrings(g.Password, oauthClient.GrantTypes, ",") {
		err = errors.New(g.InvalidGrant)
		return
	}
	/*
		demo 版本里无法真实连接到学校ldap服务器
		因此 password 模式可以用任意的用户名密码测试
			if err = oauth.LdapLogin(username, password); err != nil {
				fmt.Println(username, password)
				logger.Error(err)
				err = errors.New(g.InvalidPassword)
				return
			}
	*/
	token, err := oauth.CreateToken(clientID, username, g.Password)
	if err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	if err = copier.Copy(&resToken, &token); err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	return
}

func genAuthorizationCodeToken(inputs tokenInput, clientID string) (resToken authorizationCodeToken, err error) {
	var authorizationCodeInput oauth.AuthorizationCodeTokenInput
	if err = copier.Copy(&authorizationCodeInput, &inputs); err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	if authorizationCodeInput.ClientID == "" {
		authorizationCodeInput.ClientID = clientID
	}

	if authorizationCodeInput.Code == "" || authorizationCodeInput.RedirectUri == "" {
		err = errors.New(g.InvalidRequest)
		return
	}

	userID, err := oauth.CheckAuthorizationCode(authorizationCodeInput)
	if err != nil {
		logger.Error(err)
		return
	}
	token, err := oauth.CreateToken(clientID, userID, g.AuthorizationCode)
	if err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	if err = copier.Copy(&resToken, &token); err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	return
}

func genPkceToken(inputs tokenInput, clientID string) (resToken pkceToken, err error) {
	var pkceInput oauth.PkceTokenInput
	if err = copier.Copy(&pkceInput, &inputs); err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	if pkceInput.ClientID == "" {
		pkceInput.ClientID = clientID
	}

	if pkceInput.Code == "" || pkceInput.RedirectUri == "" {
		err = errors.New(g.InvalidRequest)
		return
	}
	userID, err := oauth.CheckCodeAndCodeVerifier(pkceInput)
	if err != nil {
		logger.Error(err)
		return
	}
	token, err := oauth.CreateToken(clientID, userID, g.AuthorizationCode)
	if err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	if err = copier.Copy(&resToken, &token); err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	return
}

func genDeviceToken(inputs tokenInput, userID, clientID string) (resToken deviceToken, err error) {
	var deviceTokenInput oauth.DeviceTokenInput
	if err = copier.Copy(&deviceTokenInput, &inputs); err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	if deviceTokenInput.ClientID == "" {
		deviceTokenInput.ClientID = clientID
	}
	if deviceTokenInput.DeviceCode == "" || deviceTokenInput.GrantType != g.DeviceFlow {
		err = errors.New(g.InvalidRequest)
		return
	}
	err = oauth.CheckDeviceCode(deviceTokenInput)
	if err != nil {
		logger.Error(err)
		return
	}
	token, err := oauth.CreateToken(clientID, userID, g.DeviceFlow)
	if err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	if err = copier.Copy(&resToken, &token); err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	return
}

func genRefreshToken(refreshToken string) (token oauth.Token, err error) {
	if refreshToken == "" {
		err = errors.New(g.InvalidRequest)
		return
	}
	token, err = oauth.RefreshAccessToken(refreshToken)
	if err != nil {
		logger.Error(err)
		err = errors.New(g.InvalidToken)
		return
	}
	return
}

func genRefreshTokenWithAuthorization(refreshToken string) (resToken authorizationCodeToken, err error) {
	token, err := genRefreshToken(refreshToken)
	if err != nil {
		return authorizationCodeToken{}, err
	}
	if err = copier.Copy(&resToken, &token); err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	return
}

func genRefreshTokenWithPkce(refreshToken string) (resToken pkceToken, err error) {
	token, err := genRefreshToken(refreshToken)
	if err != nil {
		return pkceToken{}, err
	}
	if err = copier.Copy(&resToken, &token); err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	return
}

func genRefreshTokenWithDevice(refreshToken string) (resToken deviceToken, err error) {
	token, err := genRefreshToken(refreshToken)
	if err != nil {
		return deviceToken{}, err
	}
	if err = copier.Copy(&resToken, &token); err != nil {
		logger.Error(err)
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
	// 校验client_id和client_secret，注意区分pkce模式
	clientID, err := checkOauthClient(c.Request, inputs)
	if err != nil {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(err.Error()))
		return
	}
	switch inputs.GrantType {
	case g.ClientCredentials:
		resToken, err := genClientCredentialsToken(clientID)
		if err != nil {
			c.JSON(http.StatusBadRequest, midd.OauthErrorRes(err.Error()))
			return
		}
		c.JSON(http.StatusOK, resToken)
		return

	case g.Password:
		resToken, err := genResourceOwnerPasswordCredentialsToken(inputs.Username, inputs.Password, clientID)
		if err != nil {
			c.JSON(http.StatusBadRequest, midd.OauthErrorRes(err.Error()))
			return
		}
		c.JSON(http.StatusOK, resToken)

	case g.AuthorizationCode:
		if inputs.CodeVerifier == "" {
			resToken, err := genAuthorizationCodeToken(inputs, clientID)
			if err != nil {
				c.JSON(http.StatusBadRequest, midd.OauthErrorRes(err.Error()))
				return
			}
			c.JSON(http.StatusOK, resToken)
			return
		} else {
			resToken, err := genPkceToken(inputs, clientID)
			if err != nil {
				c.JSON(http.StatusBadRequest, midd.OauthErrorRes(err.Error()))
				return
			}
			c.JSON(http.StatusOK, resToken)
			return
		}
	case g.DeviceFlow:
		deviceTokenInput := oauth.DeviceTokenInput{
			GrantType:  inputs.GrantType,
			ClientID:   clientID,
			DeviceCode: inputs.DeviceCode,
		}
		// 获取token前校验用户是否已经授权
		userId, err := midd.DeviceAuthorizeLoginCheck(deviceTokenInput, c)
		if err != nil {
			return
		}
		resToken, err := genDeviceToken(inputs, userId, clientID)
		if err != nil {
			c.JSON(http.StatusBadRequest, midd.OauthErrorRes(err.Error()))
			return
		}
		c.JSON(http.StatusOK, resToken)

	case g.RefreshToken:
		if inputs.CodeVerifier != "" {
			// pkce模式
			refreshToken, err := genRefreshTokenWithPkce(inputs.RefreshToken)
			if err != nil {
				c.JSON(http.StatusBadRequest, midd.OauthErrorRes(err.Error()))
				return
			}
			c.JSON(http.StatusOK, refreshToken)
		} else if inputs.DeviceCode != "" {
			// device flow模式
			refreshToken, err := genRefreshTokenWithDevice(inputs.RefreshToken)
			if err != nil {
				c.JSON(http.StatusBadRequest, midd.OauthErrorRes(err.Error()))
				return
			}
			c.JSON(http.StatusOK, refreshToken)
		} else {
			// authorization code模式
			refreshToken, err := genRefreshTokenWithAuthorization(inputs.RefreshToken)
			if err != nil {
				c.JSON(http.StatusBadRequest, midd.OauthErrorRes(err.Error()))
				return
			}
			c.JSON(http.StatusOK, refreshToken)
		}
		return
	default:
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.UnsupportedGrantType))
		return
	}
}

func AuthorizeCodeParamCheck(c *gin.Context) {
	inputs := authorizationInput{}
	if err := c.Bind(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.InvalidRequest))
		if err != nil {
			logger.Debug(err)
		}
		return
	}
	if inputs.ResponseType != "" && inputs.ResponseType != g.Code {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.UnsupportedResponseType))
		return
	}
	redirectUri, err := url.Parse(inputs.RedirectUri)
	if err != nil {
		logger.Debug(err)
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.InvalidRedirectUrl))
		return
	}
	oauthClient := oauth.GetClientByClientID(inputs.ClientID)
	if oauthClient.ID == 0 {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.InvalidClient))
		return
	}
	if inputs.ResponseType == g.Code {
		if !utils.InStrings(g.AuthorizationCode, oauthClient.GrantTypes, ",") {
			c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.UnauthorizedClient))
			return
		}
	}
	if !utils.InStrings(redirectUri.Hostname(), oauthClient.Domains, ",") {
		logger.Debug(redirectUri.Hostname(), g.InvalidGrant)
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.InvalidGrant))
		return
	}

	if inputs.CodeChallenge != "" && inputs.CodeChallengeMethod != "" {
		if inputs.CodeChallengeMethod != "S256" && inputs.CodeChallengeMethod != "plain" {
			c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.UnsupportedMethod))
			return
		}
	}

	c.Next()
}

func DeviceParamCheck(c *gin.Context) {
	inputs := authorizationInput{}
	if err := c.Bind(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.InvalidRequest))
		if err != nil {
			logger.Debug(err)
		}
		return
	}
	oauthClient := oauth.GetClientByClientID(inputs.ClientID)
	if oauthClient.ID == 0 {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.InvalidClient))
		return
	}
	// 把client_id存入session中
	session := sessions.Default(c)
	session.Set("client_id", inputs.ClientID)
	c.Next()
}

func getAuthorizeCode(c *gin.Context) {
	inputs := authorizationInput{}

	inputs.ClientID = c.Query("client_id")
	inputs.RedirectUri = c.Query("redirect_uri")
	inputs.ResponseType = c.Query("response_type")
	inputs.Scope = c.Query("scope")
	inputs.State = c.Query("state")

	if c.Query("code_challenge") == "" {
		var authorizationCode oauth.AuthorizationCode

		if err := copier.Copy(&authorizationCode, &inputs); err != nil {
			logger.Debug(err)
			c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.ServerError))
			return
		}
		authorizationCode.UserID = c.GetString("user_id")
		logger.Debug("userid:", authorizationCode.UserID)
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
	} else {
		inputs.CodeChallenge = c.Query("code_challenge")
		inputs.CodeChallengeMethod = c.Query("code_challenge_method")

		var pkce oauth.Pkce

		if err := copier.Copy(&pkce, &inputs); err != nil {
			logger.Debug(err)
			c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.ServerError))
			return
		}
		pkce.UserID = c.GetString("user_id")
		logger.Debug("userid:", pkce.UserID)

		code, err := oauth.CreateAuthorizationCodeWithPkce(pkce)
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

}

func getDeviceCode(c *gin.Context) {
	inputs := authorizationInput{}
	if err := c.Bind(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.InvalidRequest))
		return
	}

	var deviceCodeInput oauth.DeviceCodeInput

	if err := copier.Copy(&deviceCodeInput, &inputs); err != nil {
		logger.Debug(err)
		c.JSON(http.StatusBadRequest, midd.OauthErrorRes(g.ServerError))
		return
	}
	deviceCodeOutput, err := oauth.CreateDeviceCode(c, deviceCodeInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, midd.OauthErrorRes(g.ServerError))
		return
	}
	c.JSON(http.StatusOK, deviceCodeOutput)
}

func testToken(c *gin.Context) {
	res := map[string]string{
		"message": "token is valid",
	}
	c.JSON(http.StatusOK, res)
}
