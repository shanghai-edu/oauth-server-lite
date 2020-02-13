package oauth

import (
	"time"

	"oauth-server-lite/g"
)

type OauthClient struct {
	ID           uint   `gorm:"primary_key"`
	ClientID     string `sql:"not null;default:'';comment:'client_id'" gorm:"type:varchar(64);unique_index"`
	ClientSecret string `sql:"not null;default:'';comment:'client_secret'"`
	GrantType    string `sql:"not null;default:'';comment:'grant_type'"`
	Domain       string `sql:"not null;default:'';comment:'authorized domain'"`
	WhiteIP      string `sql:"not null;default:'';comment:'white ip'"`
	Scope        string `sql:"not null;default:'';comment:'scope'"`
	Description  string `sql:"not null;default:'';comment:'description'"`
}

type OauthAccessToken struct {
	ID          uint      `gorm:"primary_key"`
	AccessToken string    `sql:"not null;default:'';comment:'access_token'" gorm:"type:varchar(64);unique_index"`
	Scope       string    `sql:"not null;default:'';comment:'scope'"`
	ClientID    string    `sql:"not null;default:'';comment:'client_id'" `
	UserID      string    `sql:"not null;default:'';comment:'user_id'" gorm:"type:varchar(64);index"`
	ExpiredAt   time.Time `sql:"not null;default:current_timestamp;comment:'expired_at'"`
}

type OauthRefreshToken struct {
	ID           uint      `gorm:"primary_key"`
	RefreshToken string    `sql:"not null;default:'';comment:'refresh_token'" gorm:"type:varchar(64);unique_index"`
	ClientID     string    `sql:"not null;default:'';comment:'client_id'"`
	UserID       string    `sql:"not null;default:'';comment:'user_id'" gorm:"type:varchar(64);index"`
	ExpiredAt    time.Time `sql:"not null;default:current_timestamp;comment:'expired_at'"`
}

func InitTables() (err error) {
	db := g.ConnectDB()
	tx := db.Begin()
	err = tx.DropTableIfExists(&OauthClient{}).Error
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.CreateTable(&OauthClient{}).Error
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.DropTableIfExists(&OauthAccessToken{}).Error
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.CreateTable(&OauthAccessToken{}).Error
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.DropTableIfExists(&OauthRefreshToken{}).Error
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.CreateTable(&OauthRefreshToken{}).Error
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	return
}
