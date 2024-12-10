# 阶段1：构建
FROM openjdk:11-jdk-slim AS cas-builder

WORKDIR /app/apereo-cas

COPY ./apereo-cas/ .

RUN mkdir -p ~/.gradle && \
    echo "org.gradle.daemon=false" >> ~/.gradle/gradle.properties && \
    echo "org.gradle.configureondemand=true" >> ~/.gradle/gradle.properties && \
    chmod 750 ./gradlew && \
    ./gradlew --version && \
    ./gradlew clean createKeystore build --parallel --no-daemon


# 阶段2：构建 oauth-server-lite
FROM golang:1.20-alpine AS oauth-builder

WORKDIR /app/oauth-server-lite

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o oauth-server-lite .


# 阶段3：运行
FROM openjdk:11-jre-slim

WORKDIR /app

ENV PATH_ROOT=/app

# 安装必要的运行/调试工具，以及 sqlite3 和 Redis
RUN apt update && \
    apt install -y --no-install-recommends \
        sudo bash vim lsof jq curl iproute2 net-tools procps \
        ca-certificates iputils-ping redis-tools sqlite3 && \
    rm -rf /var/lib/apt/lists/*

# 创建 apereo-cas 和 oauth-server-lite 需要的目录
RUN mkdir -p /etc/cas/config /etc/cas/services ./apereo-cas ./oauth-server-lite

# 复制 CAS 配置
COPY ./apereo-cas/etc/cas/ /etc/cas/
COPY ./apereo-cas/etc/cas/config/ /etc/cas/config/
COPY ./apereo-cas/etc/cas/services/ /etc/cas/services/

# 从 cas-builder 阶段复制 cas.war
COPY --from=cas-builder /app/apereo-cas/build/libs/cas.war ./apereo-cas/cas.war

# 从 oauth-builder 阶段复制构建的 oauth-server-lite
COPY --from=oauth-builder /app/oauth-server-lite/oauth-server-lite ./oauth-server-lite/oauth-server-lite
COPY --from=oauth-builder /app/oauth-server-lite/resource ./oauth-server-lite/resource
COPY --from=oauth-builder /app/oauth-server-lite/template ./oauth-server-lite/template
COPY --from=oauth-builder /app/oauth-server-lite/cfg-docker.json ./oauth-server-lite/cfg.json

# 复制启动脚本
COPY start-services.sh ./start-services.sh

# 修改文件权限
RUN chmod +x ./oauth-server-lite/oauth-server-lite
RUN chmod +x ./start-services.sh

EXPOSE 8081 8444

ENTRYPOINT ["./start-services.sh"]