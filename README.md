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

#### 编译安装
需要 go 1.13+ 或开启 go module 的其他版本
```
# git clone https://github.com/shanghai-edu/oauth-server-lite.git
# cd oauth-server-lite
# chmod +x control
# ./control build
# ./control pack
```

#### 下载编译好的二进制包
[下载](https://github.com/shanghai-edu/oauth-server-lite/releases/download/v0.2.1/oauth-server-lite-0.2.1.tar.gz) 编译好的二进制包

#### 配置
```
{
	"log_level": "info", # info/warn/debug 三种
	"db": {
		"sqlite":"sqlite.db", # 只要不为空，则使用 sqlite 模式，存储到字段中的 sqlite 文件中
		"mysql": "root:password@tcp(127.0.0.1:3306)/oauth?charset=utf8&parseTime=True&loc=Local", # 使用 mysql 模式时的数据库连接参数
		"db_debug": false # true 时会输出详细的 sql debug
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
        "attributes": ["uid", "cn", "mail"], # ldap 返回的属性。这部分会映射为 userinfo 的接口。如果留空则返回全部属性
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

#### sqlite 模式
```
./control start
```
#### mysql 模式 ???
首次运行前，先初始化表结构。注意 -i 命令会重建数据表初始化，之前的数据会丢失。慎用
```
./oauth-server-lite -i
```
然后正常运行启动脚本
```
./control start
```


#### 反向代理
通过 apache 或者 nginx 反向代理发布服务
apache
```
        ProxyPreserveHost On
        RequestHeader set X-Forwarded-Proto https
        RemoteIPHeader X-Forwarded-For

        ProxyPass "/resource/" "http://localhost:18080/resource/"
        ProxyPass "/oauth/" "http://localhost:18080/oauth/"
        ProxyPass "/user/" "http://localhost:18080/user/"
```
nginx
```
      location /oauth/ {
          proxy_pass      http://localhost:18080/oauth/;
              client_max_body_size    100m;
              proxy_set_header X-Forwarded-For $remote_addr;
              proxy_set_header Host            $http_host;
              proxy_set_header X-Forwarded-Proto $scheme;
        }
      location /user/ {
          proxy_pass      http://localhost:18080/user/;
              client_max_body_size    100m;
              proxy_set_header X-Forwarded-For $remote_addr;
              proxy_set_header Host            $http_host;
              proxy_set_header X-Forwarded-Proto $scheme;
        }
      location /resource/ {
          proxy_pass      http://localhost:18080/resource/;
              client_max_body_size    100m;
              proxy_set_header X-Forwarded-For $remote_addr;
              proxy_set_header Host            $http_host;
              proxy_set_header X-Forwarded-Proto $scheme;
        }
```

#### 管理接口
##### 创建 client
```
# curl -H "X-API-KEY: shanghai-edu" -H "Content-Type: application/json" -d "{\"grant_type\":\"authorization_code\",\"domain\":\"www.example.org\"}" http://127.0.0.1:18080/manage/v1/client

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

#### endpoint
下列示例使用 google oauthplayground 测试
oauth.example.org 取代实际的 oauth 服务器域名
##### Authorization endpoint
`/oauth/v1/authorize`
Request
```
HTTP/1.1 302 Found
Location: https://oauth.example.org/oauth/v1/authorize?state=158802&redirect_uri=https%3A%2F%2Fdevelopers.google.com%2Foauthplayground&response_type=code&client_id=a3e9fdb53ffc9cb9&scope=Basic+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fdrive.file&access_type=offline
```
Resopnse
```
GET /oauthplayground/?code=c834d0a96744f375&state=158802 HTTP/1.1
Host: developers.google.com
```
##### Token endpoint
`/oauth/v1/token`
Request
```
POST /oauth/v1/token HTTP/1.1
Host: oauth.example.org
Content-length: 199
content-type: application/x-www-form-urlencoded
user-agent: google-oauth-playground
code=c834d0a96744f375&redirect_uri=https%3A%2F%2Fdevelopers.google.com%2Foauthplayground&client_id=a3e9fdb53ffc9cb9&client_secret=5dd6bcf253daae962430b0935207a471&scope=&grant_type=authorization_code
```
Response
```
HTTP/1.1 200 OK
{
  "access_token": "8c609be9897ca14c524bb8427ab41dfb", 
  "token_type": "Bearer", 
  "expires_in": 7200, 
  "refresh_token": "1d6396466ee848ee860e9f540bcc2665e816355937005a5b95208094ca81c523", 
  "scope": "Basic"
}
```
##### userinfo endpoint
`/oauth/v1/userinfo`
GET Request
```
GET /oauth/v1/userinfo HTTP/1.1
Host: oauth.example.org
Content-length: 0
Authorization: Bearer 8c609be9897ca14c524bb8427ab41dfb
```
POST Request
```
POST /oauth/v1/userinfo HTTP/1.1
Host: oauth.example.org
Content-length: 0
Content-type: application/json
Authorization: Bearer 8c609be9897ca14c524bb8427ab41dfb
```
Response
```
HTTP/1.1 200 OK
{
  "cn": "小冯冯", 
  "uid": "11116666", 
  "memberOf": "教职工", 
  "mail": "qfeng@exampe.org", 
  "sub": "11116666"
}
```