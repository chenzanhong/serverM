#!/bin/bash

set -e

# 从参数中继承
AGENT_DIR="/opt/monitor"
SERVICE_NAME="monitor_agent"

# 检测操作系统（假设 detect_os 函数已定义）
OS=$(detect_os)

# 判断是否使用 sudo
if command -v sudo &> /dev/null; then
  SUDO="sudo"
else
  SUDO=""
fi

# 日志记录
${SUDO} exec > >(tee -a /tmp/uninstall_monitor_agent_$(date +%Y%m%d).log) 2>&1
echo "[*] 开始卸载 $SERVICE_NAME..."

# 用户确认
read -p "[?] 确定要删除 $SERVICE_NAME 服务和相关目录吗？(y/N): " confirm
case "$confirm" in
    y|Y|yes|Yes|YES)
        echo "[*] 用户选择继续..."
        ;;
    *)
        echo "[*] 用户取消操作，退出。"
        exit 0
        ;;
esac

# 判断是否支持 systemd
if ! command -v systemctl &> /dev/null; then
  echo "[!] 当前系统不支持 systemd，无法继续清理服务"
  exit 1
fi

# 停止服务
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "[-] 正在停止 $SERVICE_NAME..."
    ${SUDO} systemctl stop "$SERVICE_NAME"
fi

# 禁用开机启动
if systemctl is-enabled --quiet "$SERVICE_NAME"; then
    echo "[-] 正在禁用 $SERVICE_NAME..."
    ${SUDO} systemctl disable "$SERVICE_NAME"
fi

# 删除服务文件
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
if [ -f "$SERVICE_FILE" ]; then
    echo "[-] 正在删除 $SERVICE_FILE..."
    ${SUDO} rm -f "$SERVICE_FILE"
fi

# 重载 systemd
${SUDO} systemctl daemon-reload

# 删除 agent 目录
if [ -d "$AGENT_DIR" ]; then
    if [ -L "$AGENT_DIR" ]; then
        echo "[!] 警告: $AGENT_DIR 是软链接，跳过删除"
    else
        echo "[-] 正在删除目录 $AGENT_DIR..."
        ${SUDO} rm -rf "$AGENT_DIR"
    fi
else
    echo "[!] 警告: 目录 $AGENT_DIR 不存在"
fi

echo "[+] $SERVICE_NAME 已成功卸载！"