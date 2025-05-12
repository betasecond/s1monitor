#!/bin/bash
# S1Monitor 部署脚本 - Linux版本

echo "===== S1Monitor 部署脚本 ====="
echo "此脚本将帮助您设置 S1Monitor 作为系统服务"
echo

# 检查是否以root运行
if [ "$EUID" -ne 0 ]; then 
  echo "警告: 此脚本需要以 root 权限运行才能安装系统服务"
  echo "请使用 sudo 重新运行此脚本"
  exit 1
fi

# 定义变量
INSTALL_DIR="/opt/s1monitor"
SERVICE_FILE="/etc/systemd/system/s1monitor.service"
BINARY_NAME="s1monitor_linux_amd64"
CONFIG_FILE="config.yaml"

# 创建安装目录
echo "创建安装目录..."
mkdir -p $INSTALL_DIR

# 复制文件
echo "复制必要文件..."
cp $BINARY_NAME $INSTALL_DIR/s1monitor
chmod +x $INSTALL_DIR/s1monitor

# 检查配置文件
if [ -f "$CONFIG_FILE" ]; then
  cp $CONFIG_FILE $INSTALL_DIR/
else
  echo "警告: 未找到配置文件 $CONFIG_FILE"
  echo "将在首次运行时创建默认配置文件"
fi

# 创建服务文件
echo "创建systemd服务..."
cat > $SERVICE_FILE << EOF
[Unit]
Description=S1 Forum Monitor
After=network.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/s1monitor -d
WorkingDirectory=$INSTALL_DIR
Restart=always
RestartSec=5
User=root
# 如果不希望以root运行，可以改为其他用户
# User=nobody

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
echo "启用并启动服务..."
systemctl daemon-reload
systemctl enable s1monitor
systemctl start s1monitor

echo
echo "部署完成!"
echo "S1Monitor 已安装到 $INSTALL_DIR 并作为系统服务启动"
echo "可以使用以下命令查看状态:"
echo "  systemctl status s1monitor"
echo "可以使用以下命令查看日志:"
echo "  journalctl -u s1monitor -f"
echo "或检查日志文件:"
echo "  $INSTALL_DIR/s1monitor.log"
echo
