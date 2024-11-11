# oauth-server-lite

[English](./README_en.md) | [中文](./README.md)

**[ 项目简介 ]**

oauth-server-lite 是一个轻量级的 OAuth2 鉴权服务，支持使用 SQLite（默认） 或 MySQL 作为数据库后端。本项目依赖 Apereo CAS 提供完整的认证管理，使其成为需要灵活的基于 OAuth2 的用户认证环境的简化方案。

**[ 目录 ]**

- [oauth-server-lite](#oauth-server-lite)
  - [环境准备](#环境准备)
  - [安装运行](#安装运行)

## 配置依赖与集成关系

oauth-server-lite 依赖 apereo-cas 作为用户信息认证的后端服务。此节会介绍 apereo-cas 和 oauth-server-lite 的配置依赖和它们之间的依赖逻辑。

### apereo-cas 配置依赖

apereo-cas 基于 gradle 编译打包，启动时默认读取 `/etc/cas` 下的 `config` 和 `services` 目录的配置文件 （对应 `./apereo-cas/etc/` 下的目录结构）。因此项目启动前，要手动更新 `/etc/cas/` 目录下的配置信息。

- `config` 目录：
   `cas.properties`：服务配置。以下参数是可能会修改的配置项（其它配置默认即可）：

    | 配置参数                                | 介绍                         |
    |-------------------------------------|----------------------------|
    | `server.port`                       | apereo-cas 服务端口号，默认 8444   | 
    | `cas.server.name`                   | apereo-cas 服务地址 / 域名       |
    | `cas.serviceRegistry.json.location` | apereo-cas 服务注册中心路径，一般无需更改 |
    | `cas.authn.jdbc.query[0].url`       | sqlite 数据库 url             |
  
  - `log4j2.xml`：日志配置，日志默认保存在 `${baseDir}=/var/log/`（因此非 root 权限运行时可能会遇到 `permission denied`），一般无需更改。

apereo-cas 还依赖 `cas.db` 数据库 （`cas.properties` 中配置）。 OAuth2 认证所需要的 `username` 和 `password` 需提前存入数据库以模拟一个真实存在在服务中的账号。

apereo-cas 占用 `8080` 和 `8444` 端口，其中 `8444` 是前端服务端口，`8080` 端口为后端服务端口。

**注：** 一般情况下不建议将 apereo-cas 运行在 windows 操作系统下（由于盘符的问题），也不建议修改 `8080` 端口和移动 `/etc` 目录下配置文件到其它路径。这些操作都需要额外修改项目源码，修改不当会导致整个项目无法正常运行。

### oauth-server-lite 配置依赖

oauth-server-lite 提供 OAuth2 认证支持，项目打包后除了可执行文件外还依赖同级目录下的 `resource` 和 `template` 提供前端静态文件，`cfg.json` （可通过 `-c` 指定）提供配置信息，运行日志会输出在 `logs` 目录下，以及 redis 服务。

`cfg.json` 配置表（可参考 `cfg.example.json`）：

```json
{
	"logger": {
		"dir": "logs/",
		"level": "DEBUG",  // INFO/WARN/DEBUG 三种
		"keepHours": 24
	},
	"cas": "http://localhost:8444/cas/",  // cas 服务访问地址
	"db": {
		"sqlite":"sqlite.db",  // 只要不为空，则使用 sqlite 模式，存储到字段中的 sqlite 文件中
		"mysql": "root:password@tcp(127.0.0.1:3306)/oauth?charset=utf8&parseTime=True&loc=Local",  // 使用 mysql 模式时的数据库连接参数
		"db_debug": false  // true 时会输出详细的 sql debug
	},
	"redis": {
		"dsn": "127.0.0.1:6379",
		"max_idle": 5,
		"conn_timeout": 5,  // 单位都是秒
		"read_timeout": 5,
		"write_timeout": 5,
		"password": ""
	},
	"redis_namespace":{  // redis key 的命名空间，保持默认即可
		"oauth":"oauth:",
		"cache":"cache:",
		"lock":"lock:",
		"fail":"fail:"
	},
	"http": {
		"listen": "0.0.0.0:18080",
		"manage_ip": ["127.0.0.1"],  // 管理接口的授信 ip
		"x-api-key": "shanghai-edu",  // 管理接口的 api key
		"session_options":{  // session 参数
			"path":"/",
			"domain":"playground.example.org",  // 必须与实际的返回域名匹配
			"max_age":7200,
			"secure":false,
			"http_only":false
		},
		"max_multipart_memory":100
	},
	"max_failed":5,  // 最大密码错误次数
	"failed_intervel":300,  // 密码错误统计的间隔时间
	"lock_time":600,  // 锁定时间
	"access_token_expired":7200,  // oauth access token 有效期，单位是秒
	"old_access_token_expired":300,  // 新的 oauth access token 生成时，老 token 的保留时间
	"refresh_token_expired_day":365,  // refresh token 的有效期，单位是天
	"code_expired":300  // authorization_code 的有效期，单位是秒
}
```

其中以下配置是可能需要根据项目实际情况修改的：

| 配置参数                          | 介绍                                       |
|-------------------------------|------------------------------------------|
| `cas`                         | apereo-cas 服务的前端登录页                      |
| `db.sqlite`                   | sqlite 数据库 url （默认值是同级目录下的 `sqlite.db` ） |
| `redis.dsn`                   | redis 服务地址                               |
| `http.listen`                 | oauth-server-lite 监听端口，默认 8081           |
| `http.session_options.domain` | oauth2playground 服务（OAuth2 认证服务）的地址/域名   |
| ...                           | ...                                      |

### 环境准备

以 centos7 为例 (此项目对 windows 支持不友好，如果需要在 windows 部署**强烈建议**使用 docker 部署方式)

- 关闭 selinux
- 配好 ntp 同步
- 安装 redis (需要 epel 源)
- 安装 jdk 11 ( gradle 编译需要 )

    ```shell
    yum install redis
    systemctl start redis
    systemctl enable redis
    ```

项目提供 docker 一键启动方式。若通过 docker 一键启动则只需本地支持 docker 即可（ windows 系统需支持 wsl 虚拟化环境）。

## 安装运行

### 方式一、docker 一键部署运行

项目提供 `docker-compose.yaml` 文件，可直接一键拉起。

```shell
docker-compose -p oauth-server-lite up -d
```

**注意事项**

- 此方式启动时，由于容器内无法直接通过 `localhost` 访问其它服务，因此需要通过访问 service name 的方式 ( `redis:6379` ) 连接 redis 。其它配置见文件。
- `cas.db` 默认写入用户信息：
  - `username: cas`，可通过配置 `${CAS_USERNAME}` 修改
  - `password: 123456`，可通过配置 `${CAS_PASSWORD}` 修改
- `sqlite.db` 默认写入 oauth client 信息：
  - `client_id: oauth`，可通过配置 `${OAUTH_CLIENT_ID}` 修改
  - `client_secret: 123456`，可通过配置 `${OAUTH_CLIENT_SECRET}` 修改
  - `domains: open-oauth2playground`，可通过配置 `${PLAYGROUND_HOST}` 修改

### 方式二、本地部署运行：`start-services.sh` 脚本快捷启动

该脚本为 docker 构建准备，同时也兼容了本地部署的方式，在尽量减少配置修改的情况下快速拉起项目。使用该脚本前，应先查看 [`start-services.sh`](start-services.sh) 中环境变量的说明，并完成相关文件路径的配置。其中 `cas.properties` 和 `cfg.json` 无需任何修改，脚本会自动根据环境变量的配置信息完成对应内容的替换。

```shell
sh ./start-services.sh
```

### 方式二、本地部署运行：`apereo-cas` 配置与部署

#### 1. sqlite 数据库初始化 (可选)

**!!!** 若通过 `start-service.sh` 脚本一键启动，docker 运行或自行准备 `cas.db`，此步可跳过

需要先在某个地⽅创建⼀个 sqlite 表，⽤来保存 user 信息。⽐如 user 表，有 3 列：username、password、name：

```shell
sqlite3 /apereo-cas/cas.db <<EOF
CREATE TABLE IF NOT EXISTS user (username TEXT, password TEXT, name TEXT);
DELETE FROM user;
INSERT INTO user (username, password, name) VALUES ('cas', '123456', '测试用户');
EOF
```

初始化完成后，通过 `sqlite3` 命令查询数据，可以得到以下结果：

```shell
[root@iZm05jcnfytljnZ apereo-cas]# sqlite3 ./cas.db 
SQLite version 3.26.0 2018-12-01 12:34:55
Enter ".help" for usage hints.
sqlite> select * from user;
cas|123456|测试用户
```

#### 2. 配置 `/apereo-cas/etc/cas/config/cas.properties`

参考 [apereo-cas 配置依赖](#apereo-cas-配置依赖) 部分。

#### 3. 启动 `apereo-cas` 服务

```shell
# 确保目前在 oauth-server-lite 项目根目录下
cd apereo-cas
# ./gradlew tasks  # 查看所有脚本名称
./gradlew copyCas
./gradlew clean build run
```

> **Exception in thread "main" java.io.IOException: Downloading from https://services.gradle.org/distributions/gradle-7.6-bin.zip failed: timeout (10000ms) 报错解决：**
> 
> 可以手动将 `gradle-7.6-bin.zip` 拷贝到 `~/.gradle/wrapper/dist/${download_dir}` 目录。例如，在 `root` 账户下，可以通过 `ls /root/.gradle/wrapper/dists/gradle-7.6-bin` 查看到文件被下载到 `9l9tetv7ltxvx3i8an4pb86ye` 目录，则把 .zip 文件也放到这个目录下即可。

### 方式二、本地部署运行：`oauth-server-lite` 配置与部署

#### 1.1 通过解压二进制包安装

**Linux**

在 [release](https://github.com/shanghai-edu/oauth-server-lite/releases) 中下载最新的 [release] 包，解压即可。

```
mkdir oauth-server-lite
cd oauth-server-lite/
wget https://github.com/shanghai-edu/oauth-server-lite/releases/download/v0.3.0/oauth-server-lite-0.3.tar.gz
tar -zxvf oauth-server-lite-0.3.tar.gz
```

#### 1.2 通过编译安装

需要 go 1.13+ 或开启 go module 的其他版本

```
git clone https://github.com/shanghai-edu/oauth-server-lite.git
cd oauth-server-lite
go mod tidy && go build

# 还可通过提供的脚本安装
# chmod +x control
# ./control build
# ./control pack  # 打包命令
```

#### 2. 数据库初始化 (可选)

项目默认使用 sqlite ，如果 sqlite 为空则会使用 mysql。强烈推荐使用 sqlite 方式（比较方便）。

正常情况下，用于 OAuth2 认证的 `client_id` 和 `client_secret` 由服务管理员分发。为了模拟这一过程，需要在初始化数据库时写入一条 `oauth_client` 数据。

由于项目启动时会自动初始化 `sqlite.db`，这里只需要写入数据即可：

```shell
# 确保在 `cfg.json` 配置的 `sqlite.db` 文件所在目录下

export OAUTH_CLIENT_ID=oauth
export OAUTH_CLIENT_SECRET=123456
export PLAYGROUND_HOST=localhost

sqlite3 "${OAUTH_SERVER_DB_FILE}" <<EOF
INSERT INTO oauth_client (
  app_id,
  client_id,
  client_secret,
  grant_types,
  domains,
  scope,
  ignore_authorize
) VALUES (
  0,
  '${OAUTH_CLIENT_ID}',
  '${OAUTH_CLIENT_SECRET}',
  'authorization_code',
  '${PLAYGROUND_HOST}',
  'Basic',
  0
);
EOF
```

注：`${OAUTH_CLIENT_ID}` 和 `${OAUTH_CLIENT_SECRET}` 正常情况下是一串随机字符串（fd2e338fbbfe）。`${PLAYGROUND_HOST}` 则是 OAuth2 认证的域名（用于 session 会话的身份验证，如果使用 [Open-OAuth2Playground](https://github.com/ECNU/Open-OAuth2Playground) 则需要填写该服务对应的地址/域名）。

#### 3. 配置 `cfg.json`

参考 [oauth-server-lite 配置依赖](#oauth-server-lite-配置依赖) 部分。

#### 4. sqlite 模式启动

```
./oauth-server-lite  # 直接运行二进制文件
./control start  # 通过 control 脚本启动
```

## 接口说明

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