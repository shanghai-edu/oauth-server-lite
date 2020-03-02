package g

import (
	"fmt"
	//引入 mysql 驱动
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	//引入 sqlite 驱动
	_ "github.com/mattn/go-sqlite3"
)

var dbp *gorm.DB

//Conn 给其他模块调用的连接池获取方法
func ConnectDB() *gorm.DB {
	return dbp
}

//InitDB 初始化数据库连接池
func InitDB(loggerlevel bool) error {
	if Config().DB.Sqlite != "" {
		db, err := gorm.Open("sqlite3", Config().DB.Sqlite)
		if err != nil {
			return fmt.Errorf("connect to db: %s", err.Error())
		}
		db.LogMode(loggerlevel)
		db.SingularTable(true)
		dbp = db
		return nil
	}
	db, err := gorm.Open("mysql", Config().DB.Mysql)
	if err != nil {
		return fmt.Errorf("connect to db: %s", err.Error())
	}
	db.LogMode(loggerlevel)
	db.SingularTable(true)
	dbp = db
	return nil
}

//CloseDB 关闭数据库连接池
func CloseDB() (err error) {
	err = dbp.Close()
	return
}
