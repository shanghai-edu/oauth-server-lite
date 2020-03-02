package oauth

import (
	"errors"
	"oauth-server-lite/g"
	"time"

	//引入 mysql 驱动
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	//引入 sqlite 驱动
	_ "github.com/mattn/go-sqlite3"
	"github.com/toolkits/file"
)

type OauthClient struct {
	ID           uint   `gorm:"primary_key"`
	ClientID     string `sql:"not null;default:''" gorm:"type:varchar(64);unique_index"`
	ClientSecret string `sql:"not null;default:''"`
	GrantType    string `sql:"not null;default:''"`
	Domain       string `sql:"not null;default:''"`
	WhiteIP      string `sql:"not null;default:''"`
	Scope        string `sql:"not null;default:''"`
	Description  string `sql:"not null;default:''"`
}

type OauthAccessToken struct {
	ID          uint      `gorm:"primary_key"`
	AccessToken string    `sql:"not null;default:''" gorm:"type:varchar(64);unique_index"`
	Scope       string    `sql:"not null;default:''"`
	ClientID    string    `sql:"not null;default:''"`
	UserID      string    `sql:"not null;default:''" gorm:"type:varchar(64);index"`
	ExpiredAt   time.Time `sql:"not null;default:current_timestamp"`
}

type OauthRefreshToken struct {
	ID           uint      `gorm:"primary_key"`
	RefreshToken string    `sql:"not null;default:''" gorm:"type:varchar(64);unique_index"`
	ClientID     string    `sql:"not null;default:''"`
	UserID       string    `sql:"not null;default:''" gorm:"type:varchar(64);index"`
	ExpiredAt    time.Time `sql:"not null;default:current_timestamp"`
}

//InitTables 初始化表结构
func InitTables() (err error) {
	if g.Config().DB.Sqlite != "" {
		err = initSqliteTables()
	} else {
		err = initMysqlTables()
	}
	return
}

func removeExistFile(f string) (err error) {
	if file.IsExist(f) {
		if file.IsFile(f) {
			err = file.Remove(f)
		} else {
			err = errors.New(f + "is not directory, not file")
		}
	}
	return
}

func initSqliteTables() (err error) {
	dbFile := g.Config().DB.Sqlite
	if err = removeExistFile(dbFile); err != nil {
		return
	}
	db, err := gorm.Open("sqlite3", dbFile)
	if err != nil {
		return
	}
	defer func() {
		err = db.Close()
	}()
	db.LogMode(g.Config().DB.DBDebug)
	db.SingularTable(true)
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

func initMysqlTables() (err error) {
	db, err := gorm.Open("mysql", g.Config().DB.Mysql)
	if err != nil {
		return
	}
	defer func() {
		err = db.Close()
	}()
	db.LogMode(g.Config().DB.DBDebug)
	db.SingularTable(true)
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
