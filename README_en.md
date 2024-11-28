# oauth-server-lite

[English](./README_en.md) | [中文](./README.md)

**[Project Introduction]**

oauth-server-lite is a lightweight OAuth2 authorization service that supports SQLite (default) or MySQL as the database backend. This project relies on Apereo CAS for comprehensive authentication management, making it a simplified solution for environments requiring flexible OAuth2-based user authentication.

**[Table of Contents]**

- [oauth-server-lite](#oauth-server-lite)
    - [Configuration Dependencies and Integration](#configuration-dependencies-and-integration)
        - [apereo-cas Configuration Dependencies](#apereo-cas-configuration-dependencies)
        - [oauth-server-lite Configuration Dependencies](#oauth-server-lite-configuration-dependencies)
        - [Environment Preparation](#environment-preparation)
    - [Installation and Operation](#installation-and-operation)
        - [Method One: Docker One-click Deployment and Operation](#method-one-docker-one-click-deployment-and-operation)
        - [Method Two: Local Deployment and Operation: Quick Start with `start-services.sh` Script](#method-two-local-deployment-and-operation-quick-start-with-start-servicessh-script)
        - [Method Two: Local Deployment and Operation: `apereo-cas` Configuration and Deployment](#method-two-local-deployment-and-operation-apereo-cas-configuration-and-deployment)
        - [Method Two: Local Deployment and Operation: `oauth-server-lite` Configuration and Deployment](#method-two-local-deployment-and-operation-oauth-server-lite-configuration-and-deployment)
    - [API Documentation](#api-documentation)

## Configuration Dependencies and Integration

oauth-server-lite relies on apereo-cas as the backend service for user information authentication. This section will introduce the configuration dependencies of apereo-cas and oauth-server-lite and the dependency logic between them.

### apereo-cas Configuration Dependencies

apereo-cas is compiled and packaged based on gradle, and it reads the configuration files in the `config` and `services` directories under `/etc/cas` by default (corresponding to the directory structure under `./apereo-cas/etc/`). Therefore, before starting the project, you need to manually update the configuration information under the `/etc/cas/` directory.

- `config` directory:
  `cas.properties`: Service configuration. The following parameters are configuration items that may be modified (other configurations are default and can be used directly):

  | Configuration Parameter             | Description                                                                       |
  |-------------------------------------|-----------------------------------------------------------------------------------|
  | `server.port`                       | The port number for the apereo-cas service, default is 8444                       |
  | `cas.server.name`                   | The address/domain name of the apereo-cas service                                 |
  | `cas.serviceRegistry.json.location` | The path to the apereo-cas service registry center, generally no change is needed |
  | `cas.authn.jdbc.query[0].url`       | SQLite database URL                                                               |

    - `log4j2.xml`: Log configuration, logs are saved in `${baseDir}=/var/log/` by default (so you may encounter `permission denied` when running without root permissions), generally no change is needed.

apereo-cas also depends on the `cas.db` database (configured in `cas.properties`). The `username` and `password` required for OAuth2 authentication need to be stored in the database in advance to simulate an account that actually exists in the service.

apereo-cas occupies ports `8080` and `8444`, where `8444` is the frontend service port, and `8080` is the backend service port.

**Note:** Generally, it is not recommended to run apereo-cas on the Windows operating system (due to disk symbol issues), nor is it recommended to change the `8080` port and move the configuration files under the `/etc` directory to other paths. These operations require additional modifications to the project source code, and improper modifications can cause the entire project to fail to run.

### oauth-server-lite Configuration Dependencies

oauth-server-lite provides OAuth2 authentication support. After packaging the project, in addition to the executable files, it also depends on the `resource` and `template` directories at the same level to provide front-end static files, `cfg.json` (which can be specified via `-c`) to provide configuration information, operation logs will be output to the `logs` directory, as well as the Redis service.

`cfg.json` configuration table (refer to `cfg.example.json`):

```json
{
  "logger": {
  "dir": "logs/",
  "level": "DEBUG",  // INFO/WARN/DEBUG three levels
  "keepHours": 24
  },
  "cas": "http://localhost:8444/cas/",  // cas service access address
  "db": {
    "sqlite":"sqlite.db",  // As long as it is not empty, it uses sqlite mode and stores it in the sqlite file in the field
    "mysql": "root:password@tcp(127.0.0.1:3306)/oauth?charset=utf8&parseTime=True&loc=Local",  // Database connection parameters when using mysql mode
    "db_debug": false  // If true, detailed SQL debug will be output
  },
  "redis": {
    "dsn": "127.0.0.1:6379",
    "max_idle": 5,
    "conn_timeout": 5,  // units are seconds
    "read_timeout": 5,
    "write_timeout": 5,
    "password": ""
  },
  "redis_namespace":{  // Redis key namespace, keep the default
    "oauth":"oauth:",
    "cache":"cache:",
    "lock":"lock:",
    "fail":"fail:"
  },
  "http": {
    "listen": "0.0.0.0:18080",
    "manage_ip": ["127.0.0.1"],  // Trusted IPs for management interfaces
    "x-api-key": "shanghai-edu",  // API key for management interfaces
    "session_options":{  // Session parameters
      "path":"/",
      "domain":"playground.example.org",  // Must match the actual OAuth2 frontend address/domain
      "max_age":7200,
      "secure":false,
      "http_only":false
    },
    "max_multipart_memory":100
  },
  "max_failed":5,  // Maximum number of password errors
  "failed_intervel":300,  // Interval time for password error statistics
  "lock_time":600,  // Lock time
  "access_token_expired":7200,  // OAuth access token validity period in seconds
  "old_access_token_expired":300,  // Retention time for old OAuth access tokens when a new one is generated
  "refresh_token_expired_day":365,  // Refresh token validity period in days
  "code_expired":300  // Authorization code validity period in seconds
}
```

Among them, the following configurations may need to be modified according to the actual situation of the project:

| Configuration Parameter | Description |
|------------------------|-------------|
| `cas`                  | The front login page of the apereo-cas service |
| `db.sqlite`            | SQLite database URL (default is `sqlite.db` in the same directory) |
| `redis.dsn`            | Redis service address |
| `http.listen`          | OAuth-server-lite listening port, default is 8081 |
| `http.session_options.domain` | The address/domain of the OAuth2 playground service (OAuth2 authentication service) |
| ...                    | ...         |

### Environment Preparation

Taking centos7 as an example (this project is not friendly to Windows, if you need to deploy on Windows, it is strongly recommended to use the Docker deployment method)

- Disable selinux
- Set up NTP synchronization
- Install Redis (requires epel source)
- Install JDK 11 (required for Gradle compilation)

    ```shell
    yum install redis
    systemctl start redis
    systemctl enable redis
    ```

The project provides a one-click Docker start method. If you start with Docker, you only need local Docker support (Windows system needs to support WSL virtualization environment).

## Installation and Operation

### Method One: Docker One-click Deployment and Operation

The project provides a `docker-compose.yaml` file, which can be used to start everything with a single command.

```shell
docker-compose -p oauth-server-lite up -d
```

#### Notes

- The `docker-compose.yaml` offers a container-based startup solution. This solution allows for a shared container network between `redis` and `oauth-server-lite`, and is intended for testing purposes only.
- `cas.db` writes default user information:
    - `username: cas`, which can be modified via the configuration `${CAS_USERNAME}`
    - `password: 123456`, which can be modified via the configuration `${CAS_PASSWORD}`
- `sqlite.db` writes default OAuth client information:
    - `client_id: oauth`, which can be modified via the configuration `${OAUTH_CLIENT_ID}`
    - `client_secret: 123456`, which can be modified via the configuration `${OAUTH_CLIENT_SECRET}`
    - `domains: localhost`, which can be modified via the configuration `${PLAYGROUND_HOST}`




### Method Two: Local Deployment and Operation: Quick Start with `start-services.sh` Script

This script is prepared for Docker builds and is also compatible with local deployment methods, allowing for quick project startup with minimal configuration changes. Before using the script, you should review the documentation of the environment variables in [`start-services.sh`](start-services.sh) and complete the configuration of the relevant file paths. The `cas.properties` and `cfg.json` files do not require any modifications; the script will automatically replace the corresponding content based on the configured environment variables.

```shell
sh ./start-services.sh
```

### Method Two: Local Deployment and Operation: `apereo-cas` Configuration and Deployment

#### 1. SQLite Database Initialization (Optional)

**!!!** If you start with the `start-service.sh` script in one click, run Docker, or prepare `cas.db` yourself, you can skip this step.

You need to create a SQLite table somewhere to save user information. For example, a user table with three columns: username, password, and name:

```shell
sqlite3 /apereo-cas/cas.db <<EOF
CREATE TABLE IF NOT EXISTS user (username TEXT, password TEXT, name TEXT);
DELETE FROM user;
INSERT INTO user (username, password, name) VALUES ('cas', '123456', 'Test User');
EOF
```

After the initialization is complete, you can query the data using the `sqlite3` command to get the following results:

```shell
[root@iZm05jcnfytljnZ apereo-cas]# sqlite3 ./cas.db 
SQLite version 3.26.0 2018-12-01 12:34:55
Enter ".help" for usage hints.
sqlite> select * from user;
cas|123456|Test User
```


### Method Two: Local Deployment and Operation: `oauth-server-lite` Configuration and Deployment

#### 1.1 Installation via Unpacking Binary Package

**Linux** <!-- markdownlint-disable-line MD036 -->

Download the latest [release] package from [release](https://github.com/shanghai-edu/oauth-server-lite/releases) and unpack it.

```shell
mkdir oauth-server-lite
cd oauth-server-lite/
wget https://github.com/shanghai-edu/oauth-server-lite/releases/download/v0.3.0/oauth-server-lite-0.3.tar.gz
tar -zxvf oauth-server-lite-0.3.tar.gz
```

#### 1.2 Installation via Compilation

Requires go 1.13+ or another version with go module enabled.

```shell
git clone https://github.com/shanghai-edu/oauth-server-lite.git
cd oauth-server-lite
go mod tidy && go build

# Alternatively, you can install using the provided script
# chmod +x control
# ./control build
# ./control pack  # Packaging command
```

#### 2. Database Initialization (Optional)

The project defaults to using sqlite, and if sqlite is empty, it will use mysql. It is highly recommended to use the sqlite method (which is more convenient).

Under normal circumstances, the `client_id` and `client_secret` used for OAuth2 authentication are distributed by the service administrator. To simulate this process, you need to write an `oauth_client` data record when initializing the database.

Since the project will automatically initialize `sqlite.db` upon startup, you only need to write data here:

```shell
# Ensure you are in the directory where the `sqlite.db` file is configured in `cfg.json`

export OAUTH_CLIENT_ID=oauth
export OAUTH_CLIENT_SECRET=123456
export PLAYGROUND_HOST=localhost
export OAUTH_GRANT_TYPES="password,authorization_code,urn:ietf:params:oauth:grant-type:device_code,client_credentials"

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
  '${OAUTH_GRANT_TYPES}',
  '${PLAYGROUND_HOST}',
  'Basic',
  0
);
EOF
```

Note: `${OAUTH_CLIENT_ID}` and `${OAUTH_CLIENT_SECRET}` are normally a random string (e.g., fd2e338fbbfe). `${PLAYGROUND_HOST}` is the domain name for OAuth2 authentication (used for session authentication; if you are using [Open-OAuth2Playground](https://github.com/ECNU/Open-OAuth2Playground), you need to fill in the corresponding address/ domain of that service).

#### 3. Configure `cfg.json`

Refer to the [oauth-server-lite Configuration Dependencies](#oauth-server-lite-configuration-dependencies) section.

#### 4. Start in sqlite Mode

```shell
./oauth-server-lite  # Run the binary file directly
./control start  # Start via the control script
```

## API Documentation

**Create Client** <!-- markdownlint-disable-line MD036 -->

```shell
# curl -H "X-API-KEY: shanghai-edu" -H "Content-Type: application/json" -d "{\"grant_type\":\"authorization_code\",\"domain\":\"www.example.org\"}" http://127.0.0.1:8081/manage/v1/client

{"client_id":"4ee85cea19800426","client_secret":"cb5b61017393877d71d9119c585bdca3","grant_type":"authorization_code","domain":"www.example.org","white_ip":"","scope":"Basic","description":""}
```

**Query Client** <!-- markdownlint-disable-line MD036 -->

```shell
# curl -H "X-API-KEY: shanghai-edu" http://127.0.0.1:8081/manage/v1/client/4ee85cea19800426
{"client_id":"4ee85cea19800426","client_secret":"cb5b61017393877d71d9119c585bdca3","grant_type":"authorization_code","domain":"www.example.org","white_ip":"","scope":"Basic","description":""}
```

**Query All Clients** <!-- markdownlint-disable-line MD036 -->

```shell
# curl -H "X-API-KEY: shanghai-edu" http://127.0.0.1:8081/manage/v1/clients
[{"client_id":"4ee85cea19800426","client_secret":"cb5b61017393877d71d9119c585bdca3","grant_type":"authorization_code","domain":"www.example.org","white_ip":"","scope":"Basic","description":""}]
```

**Delete Client** <!-- markdownlint-disable-line MD036 -->

```shell
# curl -X DELETE -H "X-API-KEY: shanghai-edu" http://127.0.0.1:8081/manage/v1/client/4ee85cea19800426
{"client_id":"4ee85cea19800426","client_secret":"cb5b61017393877d71d9119c585bdca3","grant_type":"authorization_code","domain":"www.example.org","white_ip":"","scope":"Basic","description":""}
```

