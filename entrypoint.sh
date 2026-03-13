#!/bin/bash
set -e

echo "Starting dmail-server-rest-api..."

# 复制静态文件
if [ ! -d "/app/static" ] || [ -z "$(ls -A /app/static 2>/dev/null)" ]; then
    mkdir -p /app/static
    if [ -d "/app/backup/static" ] && [ -n "$(ls -A /app/backup/static 2>/dev/null)" ]; then
        cp -r /app/backup/static/* /app/static/
    fi
fi

# 复制配置文件
if [ ! -d "/app/config" ] || [ -z "$(ls -A /app/config 2>/dev/null)" ]; then
    mkdir -p /app/config
    if [ -d "/app/backup/config" ] && [ -n "$(ls -A /app/backup/config 2>/dev/null)" ]; then
        cp -r /app/backup/config/* /app/config/
    fi
fi

# 创建默认配置文件
if [ ! -f "$CONFIG_PATH" ]; then
    mkdir -p "$(dirname "$CONFIG_PATH")"
    cat > "$CONFIG_PATH" << EOF
Port: 8080
Host: "0.0.0.0"
MAIL_HOST: "example.com"
APIKey: "change-me-please-$(date +%s)"
CommandPrefix: "docker exec mailserver"
EOF
fi

# 启动应用
exec /app/dmail-server-rest-api "$CONFIG_PATH"