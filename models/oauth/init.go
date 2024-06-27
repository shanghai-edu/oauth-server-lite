package oauth

import (
	"fmt"
	"gorm.io/gorm"
	"oauth-server-lite/g"
	"os"
	"time"
)

type ModelDeletedAt struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Model struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OauthClient struct {
	ModelDeletedAt
	AppId            int64  `gorm:"uniqueIndex;not null;default:0;comment:阿里云上的 AppId，同步"`
	AppName          string `gorm:"index;not null;default:'';comment:阿里云上的 AppName，同步"`
	Description      string `gorm:"not null;default:'';comment:阿里云上的 Description，同步"`
	ClientID         string `gorm:"uniqueIndex;not null;default:'';comment:OAuth2 的 client_id"`
	ClientSecret     string `gorm:"not null;default:'';comment:OAuth2 的 client_secret"`
	GrantTypes       string `gorm:"not null;default:'';comment:支持的 OAuth2 grant_type，以逗号分割"`
	Domains          string `gorm:"not null;default:'';comment:以域名校验，允许有多个授信域名，逗号分割"`
	Scope            string `gorm:"not null;default:'';comment:默认 ECNU-Basic，暂不使用 scope 机制区分授权域"`
	IgnoreAuthorize  bool   `gorm:"not null;default:0;comment:是否忽略授权页面"`
	PrivacyUrl       string `gorm:"not null;default:'';comment:应用隐私协议的地址"`
	ContactUserName  string `gorm:"not null;default:'';comment:联系人姓名"`
	ContactUserID    string `gorm:"not null;default:'';comment:联系人学工号"`
	ContactUserPhone string `gorm:"not null;default:'';comment:联系人电话"`
	ContactUserMail  string `gorm:"not null;default:'';comment:联系人邮箱"`
	ChargeUserName   string `gorm:"not null;default:'';comment:负责人姓名"`
	ChargeUserID     string `gorm:"not null;default:'';comment:负责人学工号"`
}

type OauthAccessToken struct {
	Model
	AccessToken string    `gorm:"uniqueIndex;not null;default:'';comment:access_token"`
	Scope       string    `gorm:"not null;default:'';comment:默认 ECNU-Basic，暂不使用 scope 机制区分授权域"`
	ClientID    string    `gorm:"not null;default:'';comment:OAuth2 的 client_id"`
	UserID      string    `gorm:"index;not null;default:'';comment:token 对应的用户学工号"`
	ExpiredAt   time.Time `gorm:"not null;autoCreateTime;comment:过期时间"`
}

type OauthRefreshToken struct {
	Model
	RefreshToken string    `gorm:"uniqueIndex;not null;default:'';comment:access_token"`
	ClientID     string    `gorm:"not null;default:'';comment:OAuth2 的 client_id"`
	UserID       string    `gorm:"index;not null;default:'';comment:token 对应的用户学工号"`
	ExpiredAt    time.Time `gorm:"not null;autoCreateTime;comment:过期时间"`
}

// InitTables 初始化表结构
func InitTables() (err error) {
	db := g.ConnectDB()
	if g.Config().DB.Sqlite != "" {
		sqliteFile := g.Config().DB.Sqlite
		if _, err := os.Stat(sqliteFile); os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist: %s", sqliteFile, err.Error())
		}
		if !isFile(sqliteFile) {
			return fmt.Errorf("%s is not a file", sqliteFile)
		}
	}

	var oauthClient OauthClient
	if err = db.AutoMigrate(oauthClient); err != nil {
		return
	}
	var oauthAccessToken OauthAccessToken
	if err = db.AutoMigrate(oauthAccessToken); err != nil {
		return
	}
	var oauthRefreshToken OauthRefreshToken
	if err = db.AutoMigrate(oauthRefreshToken); err != nil {
		return
	}
	return
}

func isFile(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
