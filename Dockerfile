# Build
FROM golang:1.23-alpine as builder
RUN mkdir /app
WORKDIR /app
ADD go.mod .
RUN go mod download
ADD . /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -o dmail-server-rest-api ./cmd/dmailserver-rest-api/main.go

# Run
FROM alpine:latest
WORKDIR /app

# 安装必要的工具（如果需要调试）
RUN apk add  bash docker-cli

# 从builder阶段复制二进制文件
COPY --from=builder /app/dmail-server-rest-api /app/dmail-server-rest-api

# 复制源码中的配置文件到config-source目录（作为备份）
COPY ./config /app/config-source

# 复制entrypoint脚本
COPY ./entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# 创建config目录（用于挂载或复制配置文件）
RUN mkdir -p /app/config

# 设置默认的配置文件路径环境变量
ENV CONFIG_PATH=/app/config/config.yaml

# 暴露端口
EXPOSE 3000

# 使用entrypoint脚本启动
CMD ["/app/entrypoint.sh"]
