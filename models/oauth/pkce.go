package oauth

import (
	"encoding/json"
	"errors"
	"github.com/toolkits/pkg/logger"
	"oauth-server-lite/g"
	"oauth-server-lite/models/utils"
)

type Pkce struct {
	ResponseType        string `json:"response_type"`
	ClientID            string `json:"client_id"`
	RedirectUri         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	UserID              string `json:"user_id"`
}

type PkceTokenInput struct {
	GrantType    string
	Code         string
	RedirectUri  string
	ClientID     string
	CodeVerifier string
}

func CreateAuthorizationCodeWithPkce(inputs Pkce) (code string, err error) {
	code, err = utils.RandHashString(g.SALT, 16)
	bs, err := json.Marshal(inputs)
	if err != nil {
		return
	}
	rc := g.ConnectRedis().Get()
	defer rc.Close()
	code_redisKey := g.Config().RedisNamespace.OAuth + "pkce_code:" + code
	_, err = rc.Do("SET", code_redisKey, string(bs), "EX", g.Config().CodeExpired)
	// 将code_challenge存入redis
	code_challenge_redisKey := g.Config().RedisNamespace.OAuth + "pkce_code_challenge:" + code
	_, err = rc.Do("SET", code_challenge_redisKey, inputs.CodeChallenge, "EX", g.Config().CodeExpired)
	return
}

func CheckCodeAndCodeVerifier(inputs PkceTokenInput) (userID string, err error) {
	logger.Debugf("check code_verifier: %v", inputs)
	rc := g.ConnectRedis().Get()
	defer rc.Close()
	codeRedisKey := g.Config().RedisNamespace.OAuth + "pkce_code:" + inputs.Code
	res, err1 := rc.Do("GET", codeRedisKey)
	codeChallengeRedisKey := g.Config().RedisNamespace.OAuth + "pkce_code_challenge:" + inputs.Code
	redisCodeChallenge, err2 := rc.Do("GET", codeChallengeRedisKey)

	if err1 != nil || err2 != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	if res == nil || redisCodeChallenge == nil {
		err = errors.New(g.InvalidGrant)
		return
	}

	var pkce Pkce
	err = json.Unmarshal(res.([]byte), &pkce)
	if err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}

	codeChallengeBytes, isSuccess := redisCodeChallenge.([]byte)
	if !isSuccess {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	codeChallenge := string(codeChallengeBytes)

	// 校验code_verifier
	if pkce.CodeChallengeMethod == "S256" {
		sha256 := utils.Sha256(inputs.CodeVerifier)
		if codeChallenge != sha256 {
			err = errors.New(g.InvalidGrant)
			return
		}
	} else if pkce.CodeChallengeMethod == "plain" {
		if codeChallenge != inputs.CodeVerifier {
			err = errors.New(g.InvalidGrant)
			return
		}
	}

	if !(inputs.ClientID == pkce.ClientID && inputs.RedirectUri == pkce.RedirectUri) {
		err = errors.New(g.InvalidGrant)
		return
	}
	userID = pkce.UserID
	// 校验通过后从redis中删除code和code_challenge
	_, err = rc.Do("DEL", codeRedisKey, codeChallengeRedisKey)
	if err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	return
}
