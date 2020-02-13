package oauth

import (
	"errors"
	"time"

	"oauth-server-lite/g"
)

func GetRefreshTokenByClient(clientID, userID string) (token OauthRefreshToken) {
	db := g.ConnectDB()
	db.Where("client_id = ? AND user_id = ?", clientID, userID).First(&token)
	return
}

func GetRefreshTokenByToken(refreshToken string) (token OauthRefreshToken) {
	db := g.ConnectDB()
	db.Where("refresh_token = ?", refreshToken).First(&token)
	return
}

func CreateRefreshTokenDB(token OauthRefreshToken) error {
	db := g.ConnectDB()
	err := db.Create(&token).Error
	return err
}

func SaveRefreshTokenDB(token OauthRefreshToken) error {
	db := g.ConnectDB()
	err := db.Save(&token).Error
	return err
}

func UpdateRefreshTokenDB(token OauthRefreshToken) error {
	db := g.ConnectDB()
	err := db.Model(&token).Update(token).Error
	return err
}

func RefreshAccessToken(refreshToken string) (token Token, err error) {
	rfToken := GetRefreshTokenByToken(refreshToken)
	if rfToken.ID == 0 {
		err = errors.New("refresh_token is not correct")
		return
	}
	if rfToken.ExpiredAt.Unix() < time.Now().Unix() {
		err = errors.New("refresh_token is expired")
		return
	}
	token, err = CreateToken(rfToken.ClientID, rfToken.UserID)
	return
}
