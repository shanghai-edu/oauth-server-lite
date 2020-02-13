package oauth

import (
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"

	"oauth-server-lite/g"
	"oauth-server-lite/models/utils"
)

type AuthorizationCode struct {
	ResponseType string `json:"response_type"`
	ClientID     string `json:"client_id"`
	RedirectUri  string `json:"redirect_uri"`
	Scope        string `json:"scope"`
	State        string `json:"state"`
	UserID       string `json:"user_id"`
}

type AuthorizationCodeTokenInput struct {
	GrantType   string
	Code        string
	RedirectUri string
	ClientID    string
}

func CreateAuthorizationCode(inputs AuthorizationCode) (code string, err error) {
	//生成一个 16 位的 authorization_code
	code, err = utils.RandHashString(g.SALT, 16)
	bs, err := json.Marshal(inputs)
	if err != nil {
		return
	}
	rc := g.ConnectRedis().Get()
	defer rc.Close()
	redisKey := g.Config().RedisNamespace.OAuth + "code:" + code
	_, err = rc.Do("SET", redisKey, string(bs), "EX", g.Config().CodeExpired)
	return
}

func CheckAuthorizationCode(inputs AuthorizationCodeTokenInput) (userID string, err error) {
	rc := g.ConnectRedis().Get()
	defer rc.Close()
	redisKey := g.Config().RedisNamespace.OAuth + "code:" + inputs.Code
	res, err := rc.Do("GET", redisKey)

	if err != nil {
		log.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	if res == nil {
		err = errors.New(g.AccessDenied)
		return
	}

	var code AuthorizationCode
	err = json.Unmarshal(res.([]byte), &code)
	if err != nil {
		log.Error(err)
		err = errors.New(g.ServerError)
		return
	}

	if !(inputs.ClientID == code.ClientID && inputs.RedirectUri == code.RedirectUri) {
		err = errors.New(g.AccessDenied)
		return
	}
	userID = code.UserID
	return
}
