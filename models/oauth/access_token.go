package oauth

import (
	"encoding/json"
	"errors"
	"time"

	"oauth-server-lite/g"
	"oauth-server-lite/models/utils"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	UserID       string `json:"user_id"`
	TokenType    string `json:"token_type"`
	GrantType    string `json:"grant_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token`
	Scope        string `json:"scope"`
}

func GetAccessTokenByClient(clientID, userID string) (token OauthAccessToken) {
	db := g.ConnectDB()
	db.Where("client_id = ? AND user_id = ?", clientID, userID).First(&token)
	return
}

func CreateAccessTokenDB(token OauthAccessToken) error {
	db := g.ConnectDB()
	err := db.Create(&token).Error
	return err
}

func SaveAccessTokenDB(token OauthAccessToken) error {
	db := g.ConnectDB()
	err := db.Save(&token).Error
	return err
}

func UpdateAccessTokenDB(token OauthAccessToken) error {
	db := g.ConnectDB()
	err := db.Model(&token).Update(token).Error
	return err
}

//CleanAccessToken 清理旧的token
//检查 redis 内的 token 有效期，如果超过 300 秒就修改到300秒。
//这样新 token 颁发后，确保老 token 还有 300 秒有效期
func CleanAccessToken(token string) (err error) {
	redisKey := g.Config().RedisNamespace.OAuth + "access_token:" + token
	rc := g.ConnectRedis().Get()
	defer rc.Close()
	expire, err := rc.Do("TTL", redisKey)
	if err != nil {
		return
	}
	if expire.(int64) > g.Config().OldAccessTokenExpired {
		_, err = rc.Do("EXPIRE", redisKey, 300)
	}
	return
}

func UpdateRefreshToken(clientID, userID, refreshToken string) (err error) {
	rcToken := GetRefreshTokenByClient(clientID, userID)
	rcToken.RefreshToken = refreshToken
	rcToken.ClientID = clientID
	rcToken.UserID = userID
	now := time.Now()
	rcToken.ExpiredAt = now.Add(time.Duration(24*g.Config().RefreshTokenExpiredDay) * time.Hour)
	err = SaveRefreshTokenDB(rcToken)
	return
}

func UpdateAccessToken(clientID, userID, scope, accessToken string) (err error) {
	acToken := GetAccessTokenByClient(clientID, userID)
	if acToken.ID != 0 {
		if err = CleanAccessToken(accessToken); err != nil {
			return
		}
	}
	acToken.Scope = scope
	acToken.AccessToken = accessToken
	acToken.ClientID = clientID
	acToken.UserID = userID
	now := time.Now()
	acToken.ExpiredAt = now.Add(time.Duration(g.Config().AccessTokenExpired) * time.Second)
	err = SaveAccessTokenDB(acToken)
	return
}

//CreateToken 创建 token
func CreateToken(clientID, userID string) (token Token, err error) {
	//查找注册应用
	oauthClient := GetClientByClientID(clientID)
	if oauthClient.ID == 0 {
		err = errors.New("cannot found such client id")
		return
	}

	//生成一个 32 位的 access_token
	accessToken, err := utils.RandHashString(g.SALT, 32)
	if err != nil {
		return
	}

	//更新 access_token
	if err = UpdateAccessToken(clientID, userID, oauthClient.Scope, accessToken); err != nil {
		return
	}

	//authorization_code 模式下，更新 refresh_token
	var refreshToken string
	if oauthClient.GrantType == "authorization_code" {
		//生成一个 64 位的 refresh_token
		refreshToken, err = utils.RandHashString(g.SALT, 64)
		if err != nil {
			return
		}
		if err = UpdateRefreshToken(clientID, userID, refreshToken); err != nil {
			return
		}
	}

	//redis 内插入 token
	token = Token{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		GrantType:    oauthClient.GrantType,
		UserID:       userID,
		ExpiresIn:    g.Config().AccessTokenExpired,
		RefreshToken: refreshToken,
		Scope:        oauthClient.Scope,
	}
	err = SetAccessToken(token)
	return
}

//SetAccessToken 向 redis 内写入 Token
func SetAccessToken(token Token) (err error) {
	bs, err := json.Marshal(token)
	if err != nil {
		return
	}
	rc := g.ConnectRedis().Get()
	defer rc.Close()
	redisKey := g.Config().RedisNamespace.OAuth + "access_token:" + token.AccessToken
	_, err = rc.Do("SET", redisKey, string(bs), "EX", token.ExpiresIn)
	return
}

//GetToken 从 redis 获取 Token
func GetAccessToken(accessToken string) (token Token, err error) {
	redisKey := g.Config().RedisNamespace.OAuth + "access_token:" + accessToken
	rc := g.ConnectRedis().Get()
	defer rc.Close()
	res, err := rc.Do("GET", redisKey)
	if err != nil {
		return
	}
	if res == nil {
		err = errors.New("token is not valid")
		return
	}

	err = json.Unmarshal(res.([]byte), &token)
	if err != nil {
		return
	}
	expire, err := rc.Do("TTL", redisKey)
	if err != nil {
		return
	}
	token.ExpiresIn = expire.(int64)
	return
}
