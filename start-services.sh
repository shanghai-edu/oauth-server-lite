#!/bin/bash

# author: dbbDylan
# date: 2024.11.07
# note: depends on `jq`

set -e  # 监测到错误立即退出

# ========================
# 变量定义
# ========================

# docker 容器中各（配置）文件以及目录的路径
PATH_ROOT=${PATH_ROOT:-"/oauth-server-lite"}
## oauth-server-lite 目录配置
OAUTH_SERVER_DIR="${PATH_ROOT}/oauth-server-lite"
OAUTH_SERVER_PATH="${OAUTH_SERVER_DIR}/oauth-server-lite"
OAUTH_SERVER_CONFIG_FILE="${OAUTH_SERVER_DIR}/cfg.json"
OAUTH_SERVER_DB_FILE="${OAUTH_SERVER_DIR}/sqlite.db"
## apereo-cas 目录配置
CAS_DIR="${PATH_ROOT}/apereo-cas"
CAS_DB_FILE="${CAS_DIR}/cas.db"
CAS_WAR_PATH="${CAS_DIR}/cas.war"
CAS_PROPERTIES_FILE=${CAS_PROPERTIES_FILE:-"/etc/cas/config/cas.properties"}
## oauth2 认证方式
OAUTH_GRANT_TYPES="password,authorization_code,urn:ietf:params:oauth:grant-type:device_code,client_credentials"

# 可对外暴露的环境变量
OAUTH_CLIENT_ID=${OAUTH_CLIENT_ID:-"oauth"}
OAUTH_CLIENT_SECRET=${OAUTH_CLIENT_SECRET:-"123456"}
CAS_USERNAME=${CAS_USERNAME:-"cas"}
CAS_PASSWORD=${CAS_PASSWORD:-"123456"}
OAUTH_SERVER_PORT=${OAUTH_SERVER_PORT:-"8081"}
CAS_SERVER_PORT=${CAS_SERVER_PORT:-"8444"}  # apereo-cas 服务端口号
CAS_SERVER_HOST=${CAS_SERVER_HOST:-"localhost"}  # apereo-cas 服务地址/域名
CAS_SERVER_URL=${CAS_SERVER_URL:-"http://${CAS_SERVER_HOST}:${CAS_SERVER_PORT}"}  # apereo-cas 服务 URL
OAUTH_REDIS_DSN=${OAUTH_REDIS_DSN:-"redis:6379"}  # redis 服务域名
OAUTH_REDIS_PASSWORD=${OAUTH_REDIS_PASSWORD:-""}  # redis 服务密码
REDIRECT_URL=${REDIRECT_URL:-"http://localhost:80"}  # 重定向 URL （ OAuth2 服务的访问 URL ）


# ========================
# 函数定义
# ========================

init() {
  init_dir
  init_sqlite
}

# 初始化目录
init_dir() {
  mkdir -p "${PATH_ROOT}" && mkdir -p "${OAUTH_SERVER_DIR}" && mkdir -p "${CAS_DIR}"
  chmod -R 777 "${PATH_ROOT}"
}

# 初始化 sqlite 配置
init_sqlite() {
  if ! command -v sqlite3 &> /dev/null; then
    echo "sqlite3 not found, installing..."
    apt-get update && apt-get install -y sqlite3
  fi

  init_cas_db
  init_oauth_server_db
}

# 初始化 CAS 数据库
init_cas_db() {
  echo "Setting up CAS database at ${CAS_DB_FILE}..."

  # 初始化数据库
  sqlite3 "${CAS_DB_FILE}" <<EOF
CREATE TABLE IF NOT EXISTS user (username TEXT, password TEXT, name TEXT);
DELETE FROM user;
INSERT INTO user (username, password, name) VALUES ('${CAS_USERNAME}', '${CAS_PASSWORD}', '测试用户');
EOF

  echo "Database initialized successfully!"
}

# 配置 CAS 属性文件
configure_cas_properties() {
  echo "Configuring CAS properties at ${CAS_PROPERTIES_FILE}..."

  # 创建临时文件
  local TMP_FILE="${CAS_PROPERTIES_FILE}.tmp"
  cp "${CAS_PROPERTIES_FILE}" "${TMP_FILE}" || touch "${TMP_FILE}"  # 确保临时文件存在

  # 配置 server.port
  if grep -q "^server.port=" "${TMP_FILE}"; then
    sed -i "s#^server.port=.*#server.port=${CAS_SERVER_PORT}#" "${TMP_FILE}"
  else
    echo "server.port=${CAS_SERVER_PORT}" >> "${TMP_FILE}"
  fi

  # 配置 cas.server.name
  if grep -q "^cas.server.name=" "${TMP_FILE}"; then
    sed -i "s#^cas.server.name=.*#cas.server.name=http://${CAS_SERVER_HOST}:${CAS_SERVER_PORT}#" "${TMP_FILE}"
  else
    echo "cas.server.name=http://${CAS_SERVER_HOST}:${CAS_SERVER_PORT}" >> "${TMP_FILE}"
  fi

  # 配置 cas.authn.jdbc.query[0].url
  if grep -q "^cas.authn.jdbc.query\[0\]\.url=" "${TMP_FILE}"; then
    sed -i "s#^cas.authn.jdbc.query\[0\]\.url=.*#cas.authn.jdbc.query[0].url=jdbc:sqlite:${CAS_DB_FILE}#" "${TMP_FILE}"
  else
    echo "cas.authn.jdbc.query[0].url=jdbc:sqlite:${CAS_DB_FILE}" >> "${TMP_FILE}"
  fi

  # 替换原配置文件
  mv "${TMP_FILE}" "${CAS_PROPERTIES_FILE}"

  echo "CAS configuration complete!"
}

# 启动 CAS 服务
start_cas() {
  echo "Starting CAS server..."

  # 检查服务状态
  if curl -s --head "${CAS_SERVER_URL}/cas/login" | grep "200" > /dev/null; then
    echo "OAuth server is already running. Skipping startup."
    return 0
  fi

  java -server -noverify -Xmx2048M -jar "${CAS_WAR_PATH}" &
}

configure_cas_properties() {
  echo "Configuring CAS properties at ${CAS_PROPERTIES_FILE}..."

  # 创建临时文件
  local TMP_FILE="${CAS_PROPERTIES_FILE}.tmp"
  cp "${CAS_PROPERTIES_FILE}" "${TMP_FILE}" || touch "${TMP_FILE}"  # 确保临时文件存在

  # 配置 server.port
  if grep -q "^server.port=" "${TMP_FILE}"; then
    sed -i "s#^server.port=.*#server.port=${CAS_SERVER_PORT}#" "${TMP_FILE}"
  else
    echo "server.port=${CAS_SERVER_PORT}" >> "${TMP_FILE}"
  fi

  # 配置 cas.server.name
  if grep -q "^cas.server.name=" "${TMP_FILE}"; then
    sed -i "s#^cas.server.name=.*#cas.server.name=http://${CAS_SERVER_HOST}:${CAS_SERVER_PORT}#" "${TMP_FILE}"
  else
    echo "cas.server.name=http://${CAS_SERVER_HOST}:${CAS_SERVER_PORT}" >> "${TMP_FILE}"
  fi

  # 配置 cas.authn.jdbc.query[0].url
  if grep -q "^cas.authn.jdbc.query\[0\]\.url=" "${TMP_FILE}"; then
    sed -i "s#^cas.authn.jdbc.query\[0\]\.url=.*#cas.authn.jdbc.query[0].url=jdbc:sqlite:${CAS_DB_FILE}#" "${TMP_FILE}"
  else
    echo "cas.authn.jdbc.query[0].url=jdbc:sqlite:${CAS_DB_FILE}" >> "${TMP_FILE}"
  fi

  # 替换原配置文件
  mv "${TMP_FILE}" "${CAS_PROPERTIES_FILE}"

  echo "CAS configuration complete!"
}

# 初始化 oauth-server-lite 数据库
init_oauth_server_db() {
  echo "Setting up Oauth server database at ${OAUTH_SERVER_DB_FILE}..."

  # 初始化数据库
  sqlite3 "${OAUTH_SERVER_DB_FILE}" <<EOF
CREATE TABLE IF NOT EXISTS oauth_client (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME,
  app_id INTEGER NOT NULL DEFAULT 0,
  app_name TEXT NOT NULL DEFAULT "",
  description TEXT NOT NULL DEFAULT "",
  client_id TEXT NOT NULL DEFAULT "",
  client_secret TEXT NOT NULL DEFAULT "",
  grant_types TEXT NOT NULL DEFAULT "",
  domains TEXT NOT NULL DEFAULT "",
  scope TEXT NOT NULL DEFAULT "",
  ignore_authorize NUMERIC NOT NULL DEFAULT 0,
  privacy_url TEXT NOT NULL DEFAULT "",
  contact_user_name TEXT NOT NULL DEFAULT "",
  contact_user_id TEXT NOT NULL DEFAULT "",
  contact_user_phone TEXT NOT NULL DEFAULT "",
  contact_user_mail TEXT NOT NULL DEFAULT "",
  charge_user_name TEXT NOT NULL DEFAULT "",
  charge_user_id TEXT NOT NULL DEFAULT ""
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_client_client_id ON oauth_client(client_id);
CREATE INDEX IF NOT EXISTS idx_oauth_client_app_name ON oauth_client(app_name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_client_app_id ON oauth_client(app_id);
CREATE INDEX IF NOT EXISTS idx_oauth_client_deleted_at ON oauth_client(deleted_at);

CREATE TABLE IF NOT EXISTS oauth_access_token (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at DATETIME,
  updated_at DATETIME,
  access_token TEXT NOT NULL DEFAULT "",
  scope TEXT NOT NULL DEFAULT "",
  client_id TEXT NOT NULL DEFAULT "",
  user_id TEXT NOT NULL DEFAULT "",
  expired_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_oauth_access_token_user_id ON oauth_access_token(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_access_token_access_token ON oauth_access_token(access_token);

CREATE TABLE IF NOT EXISTS oauth_refresh_token (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at DATETIME,
  updated_at DATETIME,
  refresh_token TEXT NOT NULL DEFAULT "",
  client_id TEXT NOT NULL DEFAULT "",
  user_id TEXT NOT NULL DEFAULT "",
  expired_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_oauth_refresh_token_user_id ON oauth_refresh_token(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_refresh_token_refresh_token ON oauth_refresh_token(refresh_token);

DELETE FROM oauth_client;
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
  '',
  'Basic',
  0
);
EOF
  echo "Database initialized successfully!"
}

# 配置 oauth-server-lite cfg.json 文件
configure_oauth_server() {
  echo "Configuring Oauth server configurations..."

  # 更新 .cas 字段
  jq --arg cas "$CAS_SERVER_URL/cas/" \
     '.cas = $cas' \
     "$OAUTH_SERVER_CONFIG_FILE" > "$OAUTH_SERVER_CONFIG_FILE.tmp" && mv "$OAUTH_SERVER_CONFIG_FILE.tmp" "$OAUTH_SERVER_CONFIG_FILE"

  # 更新 .db (sqlite) 字段
  jq --arg sqlite "$OAUTH_SERVER_DB_FILE" \
     '.db = {sqlite: $sqlite}' \
     "$OAUTH_SERVER_CONFIG_FILE" > "$OAUTH_SERVER_CONFIG_FILE.tmp" && mv "$OAUTH_SERVER_CONFIG_FILE.tmp" "$OAUTH_SERVER_CONFIG_FILE"

  # 更新 .redis 字段（只修改 dsn 和 password，保留其他字段）
  jq --arg dsn "$OAUTH_REDIS_DSN" \
         --arg password "$OAUTH_REDIS_PASSWORD" \
     '.redis.dsn = $dsn | .redis.password = $password' \
     "$OAUTH_SERVER_CONFIG_FILE" > "$OAUTH_SERVER_CONFIG_FILE.tmp" && mv "$OAUTH_SERVER_CONFIG_FILE.tmp" "$OAUTH_SERVER_CONFIG_FILE"

  # 更新 .http 字段（只修改 domain 和 listen，保持其他字段不变）
  jq --arg port "$OAUTH_SERVER_PORT" \
     --arg domain "$REDIRECT_URL" \
     '.http.listen = "0.0.0.0:\($port)" | .http.session_options.domain = $domain' \
     "$OAUTH_SERVER_CONFIG_FILE" > "$OAUTH_SERVER_CONFIG_FILE.tmp" && mv "$OAUTH_SERVER_CONFIG_FILE.tmp" "$OAUTH_SERVER_CONFIG_FILE"

  echo "Oauth server configured successfully!"
}

# 启动 OAuth Server 服务
start_oauth_server() {
  echo "Starting OAuth server..."

  cd "${OAUTH_SERVER_DIR}" && ${OAUTH_SERVER_PATH} -c "${OAUTH_SERVER_CONFIG_FILE}" &
}

# 检查 CAS 服务是否启动完成
wait_for_cas() {
  echo "Waiting for CAS server to be ready at ${CAS_SERVER_URL}..."
  while ! curl -s --head "${CAS_SERVER_URL}/cas/login" | grep "200" > /dev/null; do
#    echo "Waiting for CAS server to start..."
    sleep 1
  done
  echo "CAS server is ready!"
}

# ========================
# 主执行流程
# ========================
init
configure_cas_properties
configure_oauth_server
start_cas
wait_for_cas
start_oauth_server

# 保持脚本运行
echo "All services started. Keeping script running..."
tail -f /dev/null

# apereo-cas 会一直在后台运行，如果需要停止服务则可以执行以下命令手动 kill 进程
# lsof -i :8444 | awk 'NR>1 {print $2}' | xargs -r kill -9  # kill apereo-cas