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
RUN apk add --no-cache bash docker-cli

# 从builder阶段复制二进制文件
COPY --from=builder /app/dmail-server-rest-api /app/dmail-server-rest-api

# === 备份源码中的目录 ===
# 备份配置文件到 /app/backup/config
COPY --from=builder /app/config /app/backup/config

# 备份静态文件到 /app/backup/static
COPY --from=builder /app/static /app/backup/static

# 复制entrypoint脚本
COPY ./entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# 创建运行时目录（这些目录将通过 entrypoint 脚本从备份填充）
RUN mkdir -p /app/config /app/static /app/data

# 设置默认的配置文件路径环境变量
ENV CONFIG_PATH=/app/config/config.yaml

# 暴露端口 - API端口是8080
EXPOSE 8080

# 使用entrypoint脚本启动
CMD ["/app/entrypoint.sh"]  