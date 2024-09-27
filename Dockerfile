# 阶段1：构建apereo-cas
FROM eclipse-temurin:11-jdk AS cas-builder

WORKDIR /cas-overlay

COPY ./apereo-cas/src /cas-overlay/src/
COPY ./apereo-cas/gradle/ /cas-overlay/gradle/
COPY ./apereo-cas/gradlew /cas-overlay/gradlew
COPY ./apereo-cas/settings.gradle /cas-overlay/settings.gradle
COPY ./apereo-cas/build.gradle /cas-overlay/build.gradle
COPY ./apereo-cas/gradle.properties /cas-overlay/gradle.properties
COPY ./apereo-cas/lombok.config /cas-overlay/lombok.config

RUN mkdir -p ~/.gradle \
    && echo "org.gradle.daemon=false" >> ~/.gradle/gradle.properties \
    && echo "org.gradle.configureondemand=true" >> ~/.gradle/gradle.properties \
    && chmod 750 /cas-overlay/gradlew \
    && /cas-overlay/gradlew --version

RUN /cas-overlay/gradlew clean createKeystore build --parallel --no-daemon

# 阶段2：构建oauth-server-lite
FROM golang:1.20 AS oauth-builder

WORKDIR /OAuthServerLite

ENV GOPROXY=https://goproxy.cn,direct

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN go build -o oauth-server-lite .

# 打包/app/OAuthServerLite目录
RUN tar -cvf /tmp/OAuthServerLite.tar /OAuthServerLite

# 阶段3：运行
FROM eclipse-temurin:11-jdk

# 安装sqlite3
RUN apt-get update \
    && apt-get install -y --no-install-recommends sqlite3 \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /etc/cas/config /etc/cas/services

# 复制CAS配置
COPY ./apereo-cas/etc/cas/ /etc/cas/
COPY ./apereo-cas/etc/cas/config/ /etc/cas/config/
COPY ./apereo-cas/etc/cas/services/ /etc/cas/services/

# 从cas-builder阶段复制构建的CAS WAR文件
COPY --from=cas-builder /cas-overlay/build/libs/cas.war /cas-overlay/cas.war

# 从oauth-builder阶段复制构建的oauth-server-lite服务
COPY --from=oauth-builder /tmp/OAuthServerLite.tar /tmp
RUN tar -xvf /tmp/OAuthServerLite.tar -C /

COPY start_services.sh /start_services.sh

RUN chmod +x /OAuthServerLite/oauth-server-lite
RUN chmod +x /start_services.sh

EXPOSE 80 8080 8444

ENTRYPOINT ["/start_services.sh"]

