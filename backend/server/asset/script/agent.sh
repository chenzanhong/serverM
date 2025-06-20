#!/bin/bash

set -e

GITHUB_REPO="{{ .GithubRepoUrl }}"
HOSTNAME="{{ .HostName }}"
TOKEN={{ .Token }}
AGENT_DIR="/opt/monitor"
SUDO=""

# 安装依赖
detect_os() {
  if [ -f /etc/os-release ]; then
    . /etc/os-release
    echo "$ID"
  else
    echo "unknown"
  fi
}

OS=$(detect_os)

# 判断是否使用 sudo
if command -v sudo &> /dev/null; then
  SUDO="sudo"
else
  SUDO=""
fi

# 判断是否支持 systemd
if ! command -v systemctl &> /dev/null; then
  echo "[!] 当前系统不支持 systemd，无法继续安装服务"
  exit 1
fi

# 检查 monitor_agent.service 是否存在或正在运行
SERVICE_EXISTS=false

if ${SUDO} systemctl is-active --quiet monitor_agent.service || ${SUDO} systemctl is-enabled --quiet monitor_agent.service || [ -f /etc/systemd/system/monitor_agent.service ]; then
    SERVICE_EXISTS=true
fi

if $SERVICE_EXISTS; then
    read -p "[!] 检测到已存在的 monitor_agent 服务，是否清理并重新安装？(y/N): " CONFIRM
    case "$CONFIRM" in
        y|Y|yes|Yes|YES)
            echo "[*] 用户选择继续清理旧服务..."
            ;;
        *)
            echo "[*] 用户取消操作，退出安装脚本。"
            exit 0
            ;;
    esac

    # 开始清理旧服务
    if ${SUDO} systemctl is-active --quiet monitor_agent.service; then
        echo "[*] 发现现有 monitor_agent 服务，正在停止..."
        ${SUDO} systemctl stop monitor_agent.service || { echo "无法停止现有服务"; exit 1; }
    fi

    if ${SUDO} systemctl is-enabled --quiet monitor_agent.service; then
        echo "[*] 禁用现有 monitor_agent 服务..."
        ${SUDO} systemctl disable monitor_agent.service || { echo "无法禁用现有服务"; exit 1; }
    fi

    if [ -f /etc/systemd/system/monitor_agent.service ]; then
        echo "[*] 正在删除旧的 monitor_agent.service 文件..."
        ${SUDO} rm /etc/systemd/system/monitor_agent.service || { echo "无法删除旧的服务文件"; exit 1; }
    fi

    ${SUDO} systemctl daemon-reload
else
    echo "[+] 准备开始安装代理程序。"
fi

case "$OS" in
  ubuntu|debian)
    # 移除旧的 sbt 源（避免 apt 报错）
    if [ -f /etc/apt/sources.list.d/sbt.list ]; then
        ${SUDO} mv /etc/apt/sources.list.d/sbt.list /etc/apt/sources.list.d/sbt.list.bak
    fi
    ${SUDO} apt update && ${SUDO} apt install -y git
    ;;
  centos|rhel)
    ${SUDO} yum install -y git
    ;;
  fedora)
    ${SUDO} dnf install -y git
    ;;
  alpine)
    su root -c "apk add --no-cache git"
    ;;
  *)
    echo "不支持的操作系统: $OS"
    exit 1
    ;;
esac

# 确保 AGENT_DIR 是干净的（如果存在）
if [ -d "$AGENT_DIR" ]; then
    echo "[!] 检测到代理目录已存在：$AGENT_DIR"
    read -p "是否删除该目录以继续安装？(y/N): " CONFIRM_DELETE

    case "$CONFIRM_DELETE" in
        y|Y|yes|Yes|YES)
            echo "[*] 用户选择删除目录: $AGENT_DIR"
            ${SUDO} rm -rf "$AGENT_DIR"
            if [ $? -eq 0 ]; then
                echo "[+] 成功删除目录: $AGENT_DIR"
            else
                echo "[!] 删除目录失败，请检查权限或路径"
                exit 1
            fi
            ;;
        *)
            echo "[*] 用户取消操作，退出安装脚本。"
            exit 0
            ;;
    esac
fi

${SUDO} mkdir -p "$AGENT_DIR"
cd "$AGENT_DIR" || { echo "无法创建或进入目录 $AGENT_DIR"; exit 1; }

echo "代理程序agent所在目录："
pwd

git clone "$GITHUB_REPO" .
cd agent || { echo "找不到目录 agent，请检查仓库结构"; exit 1; }

# 编译 main
# go build -o main .

# 授予执行权限并运行主程序
chmod +x main
# ./main -hostname=${HOSTNAME} -token=${TOKEN} &

cat > /tmp/monitor_agent.service <<EOF
[Unit]
Description=Main Program Startup Service
After=network.target

[Service]
Type=simple
ExecStart=$AGENT_DIR/agent/main -hostname=${HOSTNAME} -token=${TOKEN}
Restart=always

[Install]
WantedBy=multi-user.target
EOF

${SUDO}  mv /tmp/monitor_agent.service /etc/systemd/system/monitor_agent.service
${SUDO}  systemctl daemon-reload
${SUDO}  systemctl enable monitor_agent.service
${SUDO}  systemctl start monitor_agent.service
${SUDO}  systemctl status monitor_agent.service

echo "[+] Agent 安装完成！已启动 monitor_agent 服务"