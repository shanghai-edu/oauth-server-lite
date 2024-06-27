package g

import (
	"encoding/json"

	"log"

	"sync"

	"github.com/toolkits/file"
)

/*
GlobalConfig 全局配置
*/
type GlobalConfig struct {
	Logger                 LoggerSection        `json:"logger"`
	CAS                    string               `json:"cas"`
	DB                     DBConfig             `json:"db"`
	Redis                  RedisConfig          `json:"redis"`
	RedisNamespace         RedisNamespaceConfig `json:"redis_namespace"`
	Http                   HttpConfig           `json:"http"`
	AccessTokenExpired     int64                `json:"access_token_expired"`
	OldAccessTokenExpired  int64                `json:"old_access_token_expired"`
	RefreshTokenExpiredDay int64                `json:"refresh_token_expired_day"`
	CodeExpired            int64                `json:"code_expired"`
}

/*
DBConfig DB 配置
*/
type DBConfig struct {
	Sqlite  string `json:"sqlite"`
	Mysql   string `json:"mysql"`
	DBDebug bool   `json:"db_debug"`
}

/*
RedisConfig Redis 配置
*/
type RedisConfig struct {
	Dsn          string `json:"dsn"`
	MaxIdle      int    `json:"max_idle"`
	ConnTimeout  int    `json:"conn_timeout"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	Password     string `json:"password"`
}

/*
RedisNamespaceConfig Redis 命名空间配置
*/
type RedisNamespaceConfig struct {
	OAuth string `json:"oauth"`
}

/*
HttpConfig Http 配置
*/
type HttpConfig struct {
	Listen             string                `json:"listen"`
	ManageIP           []string              `json:"manage_ip"`
	XAPIKey            string                `json:"x-api-key"`
	TrustProxy         []string              `json:"trust_proxy"`
	SessionOptions     *SessionOptionsConfig `json:"session_options"`
	MaxMultipartMemory int                   `json:"max_multipart_memory"`
}

/*
SessionOptionsConfig Session 配置
*/
type SessionOptionsConfig struct {
	Path     string `json:"path"`
	Domain   string `json:"domain"`
	MaxAge   int    `json:"max_age"`
	Secure   bool   `json:"secure"`
	HttpOnly bool   `json:"http_only"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	lock       = new(sync.RWMutex)
)

/*
Config 安全的读取和修改配置
*/
func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

/*
ParseConfig 加载配置
*/
func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	lock.Lock()
	defer lock.Unlock()

	config = &c
}
