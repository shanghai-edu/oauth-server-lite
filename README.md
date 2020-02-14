# oauth-server-lite

### 环境准备
- 关闭 selinux
- 配好 ntp 同步
- 
以 centos7 为例

#### 安装 redis
需要 epel 源
```
# yum install redis
# systemctl start redis
# systemctl enable redis
```

#### 安装 mysql(mariadb)
```
# yum install mariadb mariadb-server 
# systemctl start mariadb
# mysql_secure_installation 
# systemctl enable mariadb
```

#### 初始化表结构,
```
mysql -h 127.0.0.1 -u root -p < oauth.sql
```

#### 编译安装
```
# git clone https://github.com/shanghai-edu/oauth-server-lite.git
# cd oauth-server-listen
# go build
# chmod +x control
# ./control pack
```

#### 配置
```
{
	"log_level": "info", # info/warn/debug 三种
	"db": {
		"dsn": "root:password@tcp(127.0.0.1:3306)/oauth?charset=utf8&parseTime=True&loc=Local", #数据库连接
		"db_debug": false # true 会输出 sql 信息
	},
	"redis": {
		"dsn": "127.0.0.1:6379",
		"max_idle": 5,
		"conn_timeout": 5, # 单位都是秒
		"read_timeout": 5,
		"write_timeout": 5,
		"password": ""
	},
	"redis_namespace":{ # redis key 的命名空间，保持默认即可
		"oauth":"oauth:",
		"cache":"cache:",
		"lock":"lock:",
		"fail":"fail:"
	},
    "ldap": {
        "addr": "ldap.example.org:389",
        "baseDn": "dc=example,dc=org",
        "bindDn": "cn=manager,dc=example,dc=org",
        "bindPass": "password",
        "authFilter": "(&(uid=%s))",
        "attributes": ["uid", "cn", "mail"],
        "tls":        false,
        "startTLS":   false
    },
	"http": {
		"listen": "0.0.0.0:18080",
		"manage_ip": ["127.0.0.1"], # 管理接口的授信 ip
		"x-api-key": "shanghai-edu", # 管理接口的 api key
		"session_options":{ # session 参数
			"path":"/",
			"domain":"idp.example.org", # 必须与实际域名匹配
			"max_age":7200,
			"secure":false,
			"http_only":false
		},
		"max_multipart_memory":100
	},
	"max_failed":5, # 最大密码错误次数
	"failed_intervel":300, # 密码错误统计的间隔时间
	"lock_time":600, # 锁定时间
	"access_token_expired":7200, # oauth access token 有效期，单位是秒
	"old_access_token_expired":300, # 新的 oauth access token 生成时，老 token 的保留时间
	"refresh_token_expired_day":365, # refresh token 的有效期，单位是天
	"code_expired":300 # authorization_code 的有效期，单位是秒
}
```

#### 运行

```
./control start
```

#### 管理接口
##### 创建 client
```
# curl -H "X-API-KEY: shanghai-edu" -H "Content-Type: application/json" -d "{\"gr
ant_type\":\"authorization_code\",\"domain\":\"www.example.org\"}" http://127.0.0.1:18080/manage/v1/client

{"client_id":"4ee85cea19800426","client_secret":"cb5b61017393877d71d9119c585bdca3","grant_type":"authorization_code","domain":"www.example.org","white_ip":"","scope":"Basic","description":""}
```
##### 查询 client
```
# curl -H "X-API-KEY: shanghai-edu" http://127.0.0.1:18080/manage/v1/client/4ee85cea19800426
{"client_id":"4ee85cea19800426","client_secret":"cb5b61017393877d71d9119c585bdca3","grant_type":"authorization_code","domain":"www.example.org","white_ip":"","scope":"Basic","description":""}
```
##### 查询所有 client
```
# curl -H "X-API-KEY: shanghai-edu" http://127.0.0.1:18080/manage/v1/clients
[{"client_id":"4ee85cea19800426","client_secret":"cb5b61017393877d71d9119c585bdca3","grant_type":"authorization_code","domain":"www.example.org","white_ip":"","scope":"Basic","description":""}]
```
##### 删除 client
```
# curl -X DELETE -H "X-API-KEY: shanghai-edu" http://127.0.0.1:18080/manage/v1/client/4ee85cea19800426
{"client_id":"4ee85cea19800426","client_secret":"cb5b61017393877d71d9119c585bdca3","grant_type":"authorization_code","domain":"www.example.org","white_ip":"","scope":"Basic","description":""}
```