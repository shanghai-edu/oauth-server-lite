#!/bin/bash
# start_services.sh

mkdir -p /apereo-cas/
chmod 777 /apereo-cas/
# 判断是否安装了sqlite3
if ! command -v sqlite3 &> /dev/null
then
    echo "sqlite3 not found, installing..."
    apt-get update && apt-get install -y sqlite3
fi
sqlite3 /apereo-cas/cas.db <<EOF
CREATE TABLE IF NOT EXISTS user (username TEXT, password TEXT, name TEXT);
DELETE FROM user;
INSERT INTO user (username, password, name) VALUES ('cas', '123456', '测试用户');
EOF

echo "cas.db created successfully!"

# 读取环境变量，如果未设置，则使用默认值
CAS_SERVER_NAME=${CAS_SERVER_NAME:-"http://localhost:8444"}
SERVER_PORT=${SERVER_PORT:-"8444"}

# 配置文件路径
CAS_PROPERTIES_FILE="/etc/cas/config/cas.properties"

# 检查并替换或添加 server.port
if grep -q "server.port" "$CAS_PROPERTIES_FILE"; then
    sed -i "s#server.port=.*#server.port=${SERVER_PORT}#" "$CAS_PROPERTIES_FILE"
else
    echo "server.port=${SERVER_PORT}" >> "$CAS_PROPERTIES_FILE"
fi

# 检查并替换或添加 cas.server.name
if grep -q "cas.server.name" "$CAS_PROPERTIES_FILE"; then
    sed -i "s#cas.server.name=.*#cas.server.name=${CAS_SERVER_NAME}#" "$CAS_PROPERTIES_FILE"
else
    echo "cas.server.name=${CAS_SERVER_NAME}" >> "$CAS_PROPERTIES_FILE"
fi
echo "read configuration successfully!"

# 启动apereo-cas
java -server -noverify -Xmx2048M -jar /cas-overlay/cas.war &

while ! curl -s --head http://localhost:8444/cas/login | grep "200"  > /dev/null; do
    echo "Waiting for CAS server to start..."
  sleep 1
done
echo "CAS server is ready."

# 等待apereo-cas服务可用后启动oauth-server-lite
cd /OAuthServerLite && ./oauth-server-lite -c cfg.json &
echo "oauth-server-lite is ready."

# 保持脚本运行
tail -f /dev/null
