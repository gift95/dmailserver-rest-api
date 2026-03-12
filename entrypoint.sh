#!/bin/sh
# 检查config/config.yaml是否存在
if [ ! -f "./config/config.yaml" ]; then
    echo "config/config.yaml not found, copying from source..."
    
    # 确保config目录存在
    mkdir -p ./config
    
    # 从config-source复制配置文件
    if [ -f "./config-source/config.yaml" ]; then
        cp ./config-source/config.yaml ./config/config.yaml
        echo "config.yaml copied successfully."
    else
        echo "Warning: config-source/config.yaml not found."
    fi
fi

/app/dmail-server-rest-api $CONFIG_PATH