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
# yum install maraidb maraidb-server
# systemctl start mariadb
# mysql_secure_installation 
# systemctl enable mariadb
```

#### 初始化表结构,
```
MariaDB [(none)]> create database oauth;
MariaDB [(none)]> use oauth;
MariaDB [oauth]> source /root/oauth.sql;
```

#### 创建 oauth client
```
MariaDB [oauth]> INSERT INTO `oauth`.`oauth_client`(`client_id`, `client_secret`, `grant_type`, `domain`, `scope`, `description`) VALUES ('test', 'test', 'authorization_code', 'idp-oauth.ecnu.edu.cn', 'test', 'test');
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