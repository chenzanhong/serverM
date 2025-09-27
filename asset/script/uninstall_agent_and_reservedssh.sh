#!/bin/bash

set -e

# 从参数中继承
AGENT_DIR="/opt/monitor"
SERVICE_NAME="monitor_agent"
SSH_TUNNEL_SERVICE_PREFIX="reversetunnel"

# 日志记录
exec > >(tee -a /tmp/uninstall_combined_monitor_$(date +%Y%m%d).log) 2>&1
echo "[*] 开始卸载 $SERVICE_NAME 和 reversetunnel@${PORT} 服务..."

# 判断是否使用 sudo
if command -v sudo &> /dev/null; then
  SUDO="sudo "
else
  SUDO=""
fi

# 用户确认
read -p "[?] 确定要删除 $SERVICE_NAME 服务及 reversetunnel@* 服务吗？(y/N): " confirm
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

# ================================
# 1. 卸载 monitor_agent 主服务
# ================================

if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "[-] 正在停止 $SERVICE_NAME..."
    ${SUDO}systemctl stop "$SERVICE_NAME"
fi

if systemctl is-enabled --quiet "$SERVICE_NAME"; then
    echo "[-] 正在禁用 $SERVICE_NAME..."
    ${SUDO}systemctl disable "$SERVICE_NAME"
fi

SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
if [ -f "$SERVICE_FILE" ]; then
    echo "[-] 正在删除 $SERVICE_FILE..."
    ${SUDO}rm -f "$SERVICE_FILE"
fi

# 删除 agent 目录（排除软链接）
if [ -d "$AGENT_DIR" ]; then
    if [ -L "$AGENT_DIR" ]; then
        echo "[!] 警告: $AGENT_DIR 是软链接，跳过删除"
    else
        echo "[-] 正在删除目录 $AGENT_DIR..."
        ${SUDO}rm -rf "$AGENT_DIR"
    fi
else
    echo "[!] 警告: 目录 $AGENT_DIR 不存在"
fi

# ================================
# 2. 卸载所有 reversetunnel@<port> 服务
# ================================

# 获取所有 reversetunnel 服务列表
TUNNEL_SERVICES=$(systemctl list-units --type=service --all | grep "${SSH_TUNNEL_SERVICE_PREFIX}" | awk '{print $1}')

if [ -z "$TUNNEL_SERVICES" ]; then
    echo "[*] 没有找到任何 reversetunnel 服务，跳过卸载。"
else
    for SERVICE in $TUNNEL_SERVICES; do
        echo "[-] 正在处理服务: $SERVICE"

        if systemctl is-active --quiet "$SERVICE"; then
            echo "    正在停止 $SERVICE..."
            ${SUDO}systemctl stop "$SERVICE"
        fi

        if systemctl is-enabled --quiet "$SERVICE"; then
            echo "    正在禁用 $SERVICE..."
            ${SUDO}systemctl disable "$SERVICE"
        fi

        SERVICE_FILE="/etc/systemd/system/$SERVICE"
        if [ -f "$SERVICE_FILE" ]; then
            echo "    正在删除 $SERVICE_FILE..."
            ${SUDO}rm -f "$SERVICE_FILE"
        fi
    done
fi

# 重载 systemd
${SUDO}systemctl daemon-reload

echo "[+] 所有指定服务已成功卸载！"