package g

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"
)

var (
	dbp *gorm.DB
)

func ConnectDB() *gorm.DB {
	return dbp
}

func InitDB() (err error) {
	var loggerConfig logger.Config
	var db *gorm.DB
	if Config().DB.DBDebug {
		loggerConfig = logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level
			Colorful:      false,       // 禁用彩色打印
		}
	} else {
		loggerConfig = logger.Config{
			SlowThreshold: time.Second,   // 慢 SQL 阈值
			LogLevel:      logger.Silent, // Log level
			Colorful:      false,         // 禁用彩色打印
		}
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		loggerConfig,
	)

	ormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	}
	// 根据Config().DB.sqlite值是否为空选择数据库初始化类型
	if Config().DB.Sqlite != "" {
		db, err = gorm.Open(sqlite.Open(Config().DB.Sqlite), ormConfig)
		if err != nil {
			return fmt.Errorf("connect to sqlite db: %s", err.Error())
		}
	} else {
		db, err = gorm.Open(mysql.New(mysql.Config{
			DSN:               Config().DB.Mysql,
			DefaultStringSize: 256,
		}), ormConfig)
		if err != nil {
			return fmt.Errorf("connect to MySQL db: %s", err.Error())
		}
	}

	// 测试是否能 ping 通 db
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB instance: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping the database: %v", err)
	}

	log.Println("Database connection is healthy.")

	dbp = db
	return nil
}

func CloseDB() (err error) {
	sqldb, err := dbp.DB()
	if err != nil {
		return
	}
	err = sqldb.Close()

	return
}
