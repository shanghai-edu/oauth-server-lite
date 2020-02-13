package g

import (
	"fmt"
	// 导入 mysql 驱动，gorm 初始化时需要
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var (
	dbp *gorm.DB
)

/*
ConnectDB 获取连接
*/
func ConnectDB() *gorm.DB {
	return dbp
}

/*
InitDB 初始化连接池
*/
func InitDB(loggerlevel bool) (err error) {
	db, err := gorm.Open("mysql", Config().DB.Dsn)
	db.LogMode(loggerlevel)
	if err != nil {
		return fmt.Errorf("connect to db: %s", err.Error())
	}
	db.SingularTable(true)
	dbp = db

	return
}

/*
CloseDB 关闭连接
*/
func CloseDB() (err error) {
	err = dbp.Close()
	return
}
