#!/bin/bash
set -e

echo "========================================"
echo "Starting dmail-server-rest-api"
echo "========================================"
echo "时间: $(date)"
echo ""

# 定义颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 函数：复制目录如果不存在
copy_if_not_exists() {
    local src=$1
    local dst=$2
    local name=$3
    
    if [ ! -d "$dst" ] || [ -z "$(ls -A $dst 2>/dev/null)" ]; then
        echo -e "${YELLOW}📁 $name 目录不存在或为空，从备份复制...${NC}"
        if [ -d "$src" ] && [ -n "$(ls -A $src 2>/dev/null)" ]; then
            mkdir -p "$dst"
            cp -rf $src/* $dst/
            echo -e "${GREEN}✅ $name 已从备份复制到 $dst${NC}"
            ls -la $dst | sed 's/^/  /'
        else
            echo -e "${RED}❌ 错误: 备份目录 $src 也不存在或为空${NC}"
            return 1
        fi
    else
        echo -e "${GREEN}✅ $name 已存在，跳过复制${NC}"
        ls -la $dst | sed 's/^/  /'
    fi
}

echo ""
echo "=== 检查运行时目录 ==="

# 1. 检查并复制静态文件
copy_if_not_exists "/app/backup/static" "/app/static" "静态文件目录"

# 2. 检查并复制配置文件目录
copy_if_not_exists "/app/backup/config" "/app/config" "配置文件目录"

# 3. 检查主配置文件是否存在
if [ ! -f "$CONFIG_PATH" ]; then
    echo -e "${YELLOW}📄 主配置文件不存在，创建默认配置...${NC}"
    
    # 检查备份中是否有配置文件
    if [ -f "/app/backup/config/config.yaml" ]; then
        cp /app/backup/config/config.yaml $CONFIG_PATH
        echo -e "${GREEN}✅ 已从备份复制主配置文件${NC}"
    else
        # 创建默认配置
        mkdir -p $(dirname $CONFIG_PATH)
        cat > $CONFIG_PATH << EOF
Port: 8080
Host: "0.0.0.0"
APIKey: "change-me-please-$(date +%s)"
CommandPrefix: "docker exec mailserver"
EOF
        echo -e "${GREEN}✅ 已创建默认配置文件${NC}"
    fi
else
    echo -e "${GREEN}✅ 主配置文件已存在${NC}"
fi

# 4. 显示配置文件内容（隐藏API密钥）
echo ""
echo "=== 配置信息 ==="
if [ -f "$CONFIG_PATH" ]; then
    echo "配置文件: $CONFIG_PATH"
    echo "内容预览:"
    cat $CONFIG_PATH | grep -v "^#" | sed 's/APIKey:.*/APIKey: ********/' | sed 's/^/  /'
fi

# 5. 显示目录结构
echo ""
echo "=== 目录结构 ==="
echo "/app:"
ls -la /app | grep -v "backup" | sed 's/^/  /'

echo "/app/static:"
if [ -d "/app/static" ]; then
    ls -la /app/static | sed 's/^/  /'
fi

echo "/app/config:"
if [ -d "/app/config" ]; then
    ls -la /app/config | sed 's/^/  /'
fi

echo ""
echo "=== 启动应用 ==="
echo "执行: /app/dmail-server-rest-api $CONFIG_PATH"
echo "========================================"

# 启动应用
exec /app/dmail-server-rest-api $CONFIG_PATH