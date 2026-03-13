#!/bin/bash
set -e

echo "Starting dmail-server-rest-api..."

# 复制静态文件
if [ ! -d "/app/static" ]; then
    mkdir -p /app/static
    cp -r /app/backup/static/* /app/static/ 2>/dev/null || true
fi

# 复制配置文件
if [ ! -d "/app/config" ]; then
    mkdir -p /app/config
    cp -r /app/backup/config/* /app/config/ 2>/dev/null || true
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