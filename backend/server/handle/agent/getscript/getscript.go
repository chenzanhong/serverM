package getscript

import (
	"bytes"
	"fmt"
	"net/http"
	cf "serverM/server/config"
	pt "serverM/server/handle/agent/port"
	m_init "serverM/server/model/init"
	u "serverM/server/model/user"
	"text/template"

	"github.com/gin-gonic/gin"
)

const agentTemplate = `#!/bin/bash

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
`

const sshTunnelTemplate = `#!/bin/bash

set -e

PUBLIC_SERVER_IP="{{ .PublicServerIP }}"
SSH_TUNNEL_PORT={{ .Port }}
SSH_TUNNEL_USER="{{ .SshTunnelUsername }}"
SSH_TUNNEL_PASS="{{ .SshTunnelPassword }}"

# 安装 autossh
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

# 定义服务名称
SERVICE_NAME="reversetunnel@${SSH_TUNNEL_PORT}"

# 检查 reversetunnel@port.service 是否存在或正在运行
SERVICE_EXISTS=false

if ${SUDO} systemctl is-active --quiet $SERVICE_NAME || ${SUDO} systemctl is-enabled --quiet $SERVICE_NAME || [ -f /etc/systemd/system/${SERVICE_NAME}.service ]; then
    SERVICE_EXISTS=true
fi

if $SERVICE_EXISTS; then
    read -p "[!] 检测到已存在的 $SERVICE_NAME 服务，是否清理并重新安装？(y/N): " CONFIRM
    case "$CONFIRM" in
        y|Y|yes|Yes|YES)
            echo "[*] 用户选择继续清理旧服务..."
            ;;
        *)
            echo "[*] 用户取消操作，保留现有服务并退出安装脚本。"
            exit 0
            ;;
    esac

    # 开始清理旧服务
    if ${SUDO} systemctl is-active --quiet $SERVICE_NAME; then
        echo "[*] 发现现有 $SERVICE_NAME 服务，正在停止..."
        ${SUDO} systemctl stop $SERVICE_NAME || { echo "无法停止现有服务"; exit 1; }
    fi

    if ${SUDO} systemctl is-enabled --quiet $SERVICE_NAME; then
        echo "[*] 禁用现有 $SERVICE_NAME 服务..."
        ${SUDO} systemctl disable $SERVICE_NAME || { echo "无法禁用现有服务"; exit 1; }
    fi

    if [ -f /etc/systemd/system/${SERVICE_NAME}.service ]; then
        echo "[*] 正在删除旧的 ${SERVICE_NAME}.service 文件..."
        ${SUDO} rm /etc/systemd/system/${SERVICE_NAME}.service || { echo "无法删除旧的服务文件"; exit 1; }
    fi

    ${SUDO} systemctl daemon-reload
else
    echo "[+] 准备开始配置反向隧道服务。"
fi

# 安装 autossh
if command -v autossh &> /dev/null; then
  echo "[*] autossh 已安装"
else
  case "$OS" in
    ubuntu|debian)
      # 移除旧的 sbt 源（避免 apt 报错）
      if [ -f /etc/apt/sources.list.d/sbt.list ]; then
          ${SUDO} mv /etc/apt/sources.list.d/sbt.list /etc/apt/sources.list.d/sbt.list.bak
      fi
      ${SUDO} apt update && ${SUDO} apt install -y autossh
      ;;
    centos|rhel)
      ${SUDO} yum install -y autossh
      ;;
    fedora)
      ${SUDO} dnf install -y autossh
      ;;
    alpine)
      su root -c "apk add --no-cache autossh"
      ;;
    *)
      echo "不支持的操作系统: $OS"
      exit 1
      ;;
  esac
fi

# 安装 sshpass（如果还没有安装）
if ! command -v sshpass &> /dev/null; then
  case "$OS" in
    ubuntu|debian)
      ${SUDO} apt install -y sshpass
      ;;
    centos|rhel)
      ${SUDO} yum install -y sshpass
      ;;
    fedora)
      ${SUDO} dnf install -y sshpass
      ;;
    alpine)
      su root -c "apk add --no-cache sshpass"
      ;;
    *)
      echo "不支持的操作系统: $OS, 或者无法安装 sshpass"
      exit 1
      ;;
  esac
fi

# 创建 systemd 服务文件
cat > /tmp/${SERVICE_NAME}.service <<EOF
[Unit]
Description=Reverse SSH Tunnel on Port %i
After=network.target

[Service]
User=$(whoami)

ExecStart=/usr/bin/sshpass -p '${SSH_TUNNEL_PASS}' /usr/bin/autossh -M 0 -N \
  -o "StrictHostKeyChecking=no" \
  -o "UserKnownHostsFile=/dev/null" \
  -R %i:localhost:22 ${SSH_TUNNEL_USER}@${PUBLIC_SERVER_IP}
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 移动并启用服务
${SUDO} mv /tmp/${SERVICE_NAME}.service /etc/systemd/system/
${SUDO} systemctl daemon-reload
${SUDO} systemctl enable $SERVICE_NAME
${SUDO} systemctl start $SERVICE_NAME
${SUDO} systemctl status $SERVICE_NAME

echo "[+] 反向 SSH 隧道配置完成！已启动隧道（端口: $SSH_TUNNEL_PORT）"
`

// 获取安装代理程序的脚本
func GetAgentScript(c *gin.Context) {
	hostname := c.Query("hostname")
	tmpl, err := template.New("agent").Parse(agentTemplate)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// 查询token
	var hostandtoken u.HostAndToken
	err = m_init.DB.Where("host_name = ?", hostname).First(&hostandtoken).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, "查询hostandtoken表失败："+err.Error())
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=install_agent.sh")

	err = tmpl.Execute(c.Writer, struct {
		GithubRepoUrl string
		HostName      string
		Token         string
	}{
		GithubRepoUrl: cf.GithubRepoUrl,
		HostName:      hostname,
		Token:         hostandtoken.Token,
	})

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}

func GenerateAgentScriptBytes(hostname, token string) ([]byte, error) {
	tmpl, err := template.New("agent").Parse(agentTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		GithubRepoUrl string
		HostName      string
		Token         string
	}{
		GithubRepoUrl: cf.GithubRepoUrl,
		HostName:      hostname,
		Token:         token,
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// 获取配置反向SSH的脚本
func GetSSHScript(c *gin.Context) {
	hostname := c.Query("hostname")
	if hostname == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "hostname参数不能为空"})
		return
	}
	var sshport u.SSHPort
	var port int
	var err error

	// 先尝试查找是否有已分配给该 hostname 的端口记录
	err = m_init.DB.Where("hostname = ?", hostname).First(&sshport).Error
	if err == nil {
		// 已存在记录，使用现有端口
		port = sshport.Port
	} else {
		// 不存在，获取一个新端口并更新记录
		port, err = pt.GetUnusedPort()
		if port == -1 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"message": fmt.Sprintf("获取 SSH 隧道端口失败: %v", err.Error)})
			return
		}

		// 修改ssh_port表中port对应记录的hostname
		var sshport u.SSHPort
		err = m_init.DB.Where("port = ?", port).First(&sshport).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("查询 ssh_ports 表失败：%v", err)})
			return
		}
		sshport.Hostname = hostname
		sshport.IsUsed = true
		err = m_init.DB.Save(&sshport).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("更新 ssh_ports 表失败：%v", err)})
			return
		}
	}

	tmpl, err := template.New("ssh").Parse(sshTunnelTemplate)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=setup_ssh_tunnel.sh")

	data := struct {
		PublicServerIP    string
		SshTunnelUsername string
		SshTunnelPassword string
		Port              int
	}{
		PublicServerIP:    cf.PublicServerIP,
		SshTunnelUsername: cf.SshTunnelUsername,
		SshTunnelPassword: cf.SshTunnelPassword,
		Port:              port,
	}

	// 执行模板并将结果写入响应
	if err := tmpl.Execute(c.Writer, data); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}

// 联合脚本
const combinedScriptTemplate = `#!/bin/bash

set -e

# 第一部分：安装代理程序

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

    # 删除 AGENT_DIR（如果存在）
    if [ -d "$AGENT_DIR" ]; then
        echo "[*] 正在删除代理目录: $AGENT_DIR"
        ${SUDO} rm -rf "$AGENT_DIR"
        if [ $? -eq 0 ]; then
            echo "[+] 成功删除代理目录: $AGENT_DIR"
        else
            echo "[!] 删除代理目录失败，请检查权限或路径"
        fi
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

${SUDO} git clone "$GITHUB_REPO" .
cd agent || { echo "找不到目录 agent，请检查仓库结构"; exit 1; }

# 编译 main
# go build -o main .

# 授予执行权限并运行主程序
${SUDO} chmod +x main
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

echo "[+] Agent 安装完成！已启动 monitor_agent.service 服务"



# 第二部分：配置反向SSH隧道

PUBLIC_SERVER_IP="{{ .PublicServerIP }}"
SSH_TUNNEL_PORT={{ .Port }}
SSH_TUNNEL_USER="{{ .SshTunnelUsername }}"
SSH_TUNNEL_PASS="{{ .SshTunnelPassword }}"

# 安装 autossh
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

# 定义服务名称
SERVICE_NAME="reversetunnel@${SSH_TUNNEL_PORT}"

# 检查 reversetunnel@port.service 是否存在或正在运行
SERVICE_EXISTS=false

if ${SUDO} systemctl is-active --quiet $SERVICE_NAME || ${SUDO} systemctl is-enabled --quiet $SERVICE_NAME || [ -f /etc/systemd/system/${SERVICE_NAME}.service ]; then
    SERVICE_EXISTS=true
fi

if $SERVICE_EXISTS; then
    read -p "[!] 检测到已存在的 $SERVICE_NAME 服务，是否清理并重新安装？(y/N): " CONFIRM
    case "$CONFIRM" in
        y|Y|yes|Yes|YES)
            echo "[*] 用户选择继续清理旧服务..."
            ;;
        *)
            echo "[*] 用户取消操作，保留现有服务并退出安装脚本。"
            exit 0
            ;;
    esac

    # 开始清理旧服务
    if ${SUDO} systemctl is-active --quiet $SERVICE_NAME; then
        echo "[*] 发现现有 $SERVICE_NAME 服务，正在停止..."
        ${SUDO} systemctl stop $SERVICE_NAME || { echo "无法停止现有服务"; exit 1; }
    fi

    if ${SUDO} systemctl is-enabled --quiet $SERVICE_NAME; then
        echo "[*] 禁用现有 $SERVICE_NAME 服务..."
        ${SUDO} systemctl disable $SERVICE_NAME || { echo "无法禁用现有服务"; exit 1; }
    fi

    if [ -f /etc/systemd/system/${SERVICE_NAME}.service ]; then
        echo "[*] 正在删除旧的 ${SERVICE_NAME}.service 文件..."
        ${SUDO} rm /etc/systemd/system/${SERVICE_NAME}.service || { echo "无法删除旧的服务文件"; exit 1; }
    fi

    ${SUDO} systemctl daemon-reload
else
    echo "[+] 准备开始配置反向隧道服务。"
fi

# 安装 autossh
if command -v autossh &> /dev/null; then
  echo "[*] autossh 已安装"
else
  case "$OS" in
    ubuntu|debian)
      # 移除旧的 sbt 源（避免 apt 报错）
      if [ -f /etc/apt/sources.list.d/sbt.list ]; then
          ${SUDO} mv /etc/apt/sources.list.d/sbt.list /etc/apt/sources.list.d/sbt.list.bak
      fi
      ${SUDO} apt update && ${SUDO} apt install -y autossh
      ;;
    centos|rhel)
      ${SUDO} yum install -y autossh
      ;;
    fedora)
      ${SUDO} dnf install -y autossh
      ;;
    alpine)
      su root -c "apk add --no-cache autossh"
      ;;
    *)
      echo "不支持的操作系统: $OS"
      exit 1
      ;;
  esac
fi

# 安装 sshpass（如果还没有安装）
if ! command -v sshpass &> /dev/null; then
  case "$OS" in
    ubuntu|debian)
      ${SUDO} apt install -y sshpass
      ;;
    centos|rhel)
      ${SUDO} yum install -y sshpass
      ;;
    fedora)
      ${SUDO} dnf install -y sshpass
      ;;
    alpine)
      su root -c "apk add --no-cache sshpass"
      ;;
    *)
      echo "不支持的操作系统: $OS, 或者无法安装 sshpass"
      exit 1
      ;;
  esac
fi

# 创建 systemd 服务文件
cat > /tmp/${SERVICE_NAME}.service <<EOF
[Unit]
Description=Reverse SSH Tunnel on Port %i
After=network.target

[Service]
User=$(whoami)

ExecStart=/usr/bin/sshpass -p '${SSH_TUNNEL_PASS}' /usr/bin/autossh -M 0 -N \
  -o "StrictHostKeyChecking=no" \
  -o "UserKnownHostsFile=/dev/null" \
  -R %i:localhost:22 ${SSH_TUNNEL_USER}@${PUBLIC_SERVER_IP}

Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 移动并启用服务
${SUDO} mv /tmp/${SERVICE_NAME}.service /etc/systemd/system/
${SUDO} systemctl daemon-reload
${SUDO} systemctl enable $SERVICE_NAME
${SUDO} systemctl start $SERVICE_NAME
${SUDO} systemctl status $SERVICE_NAME

echo "[+] 反向 SSH 隧道配置完成！已启动隧道（端口: $SSH_TUNNEL_PORT）"
`

// 获取合并后的脚本——包含安装代理程序和配置反向SSH隧道
func GetCombinedScript(c *gin.Context) {
	hostname := c.Query("hostname")
	// 检查hostname参数是否为空
	if hostname == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "hostname参数不能为空"})
		return
	}

	var sshport u.SSHPort
	var port int
	var err error

	// 先尝试查找是否有已分配给该 hostname 的端口记录
	err = m_init.DB.Where("hostname = ?", hostname).First(&sshport).Error
	if err == nil {
		// 已存在记录，使用现有端口
		port = sshport.Port
	} else {
		// 不存在，获取一个新端口并更新记录
		port, err = pt.GetUnusedPort()
		if port == -1 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"message": fmt.Sprintf("获取 SSH 隧道端口失败: %v", err.Error)})
			return
		}

		// 修改ssh_port表中port对应记录的hostname
		var sshport u.SSHPort
		err = m_init.DB.Where("port = ?", port).First(&sshport).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("查询 ssh_ports 表失败：%v", err)})
			return
		}
		sshport.Hostname = hostname
		sshport.IsUsed = true
		err = m_init.DB.Save(&sshport).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("更新 ssh_ports 表失败：%v", err)})
			return
		}
	}

	// 查询token
	var hostandtoken u.HostAndToken
	err = m_init.DB.Where("host_name = ?", hostname).First(&hostandtoken).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, "查询hostandtoken表失败："+err.Error())
		return
	}

	tmpl, err := template.New("combined").Parse(combinedScriptTemplate)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=install_agent_and_ssh_tunnel.sh")

	data := struct {
		GithubRepoUrl     string
		PublicServerIP    string
		HostName          string
		Token             string
		SshTunnelUsername string
		SshTunnelPassword string
		Port              int
	}{
		GithubRepoUrl:     cf.GithubRepoUrl,
		PublicServerIP:    cf.PublicServerIP,
		HostName:          hostname,
		Token:             hostandtoken.Token,
		SshTunnelUsername: cf.SshTunnelUsername,
		SshTunnelPassword: cf.SshTunnelPassword,
		Port:              port,
	}

	if err := tmpl.Execute(c.Writer, data); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}

// GenerateCombinedScriptBytes 生成包含 agent 安装和 SSH 隧道配置的完整脚本字节流
func GenerateCombinedScriptBytes(hostname, token string) ([]byte, error) {
	// 检查必要配置是否存在
	if cf.PublicServerIP == "" || cf.SshTunnelUsername == "" {
		return nil, fmt.Errorf("公共服务器 IP 或隧道用户名未配置")
	}

	var sshport u.SSHPort
	var port int
	var err error
	// 先尝试查找是否有已分配给该 hostname 的端口记录
	err = m_init.DB.Where("hostname = ?", hostname).First(&sshport).Error

	if err == nil {
		// 已存在记录，使用现有端口
		port = sshport.Port
	} else {
		// 不存在，获取一个新端口并更新记录
		port, err = pt.GetUnusedPort()
		if port == -1 {
			return nil, fmt.Errorf("获取 SSH 隧道端口失败: %v", err)
		}

		// 修改ssh_port表中port对应记录的hostname
		var sshport u.SSHPort
		err = m_init.DB.Where("port = ?", port).First(&sshport).Error
		if err != nil {
			return nil, fmt.Errorf("查询ssh_port表失败：%v", err)
		}
		sshport.Hostname = hostname
		sshport.IsUsed = true
		err = m_init.DB.Save(&sshport).Error
		if err != nil {
			return nil, fmt.Errorf("更新ssh_port表失败：%v", err)
		}
	}

	// 解析模板
	tmpl, err := template.New("combined").Parse(combinedScriptTemplate)
	if err != nil {
		return nil, fmt.Errorf("解析 combined 模板失败: %v", err)
	}

	// 构造数据上下文
	data := struct {
		GithubRepoUrl     string
		PublicServerIP    string
		HostName          string
		Token             string
		SshTunnelUsername string
		SshTunnelPassword string
		Port              int
	}{
		GithubRepoUrl:     cf.GithubRepoUrl,
		PublicServerIP:    cf.PublicServerIP,
		HostName:          hostname,
		Token:             token,
		SshTunnelUsername: cf.SshTunnelUsername,
		SshTunnelPassword: cf.SshTunnelPassword,
		Port:              port,
	}

	// 渲染模板到缓冲区
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("执行 combined 模板失败: %v", err)
	}

	return buf.Bytes(), nil
}

// ---------------------------------------------------------- 服务卸载
const deleteMonitorAgentScriptTemplate = `#!/bin/bash

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
`

// DeleteMonitorAgentScript 返回一个可下载的卸载脚本，仅用于删除 monitor_agent 服务
func GetAgentUninstallScript(c *gin.Context) {
	// hostname := c.Query("hostname")
	// if hostname == "" {
	// 	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Query参数hostname不能为空"})
	// 	return
	// }

	// // 判断hostname是否存在，如果不存在则不返回
	// var hostinfo u.HostInfo
	// err := m_init.DB.Where("hostname = ?", hostname).First(&hostinfo).Error
	// if err != nil {
	//   c.JSON(http.StatusInternalServerError, gin.H{"message": "查询 host_info 表失败: " + err.Error()})
	//   return
	// }

	// 设置响应头为文件下载
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=uninstall_monitor_agent.sh")

	_, err := c.Writer.Write([]byte(deleteMonitorAgentScriptTemplate))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}

const uninstallCombinedScriptTemplate = `#!/bin/bash

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
`

// GetUninstallScript 返回一个可下载的卸载脚本，用于删除 monitor_agent 和 reversetunnel 服务
func GetCombinedUninstallScript(c *gin.Context) {
	// hostname := c.Query("hostname")
	// if hostname == "" {
	// 	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "hostname参数不能为空"})
	// 	return
	// }

	// // 查询 ssh_ports 表获取 Port 和 SshTunnelUsername
	// var sshport u.SSHPort
	// err := m_init.DB.Where("hostname = ?", hostname).First(&sshport).Error
	// if err != nil {
	// 	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "请确认hostname参数值是否正确: " + err.Error()})
	// 	return
	// }

	// 使用模板生成卸载脚本
	tmpl, err := template.New("uninstall").Parse(uninstallCombinedScriptTemplate)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// 设置响应头为文件下载
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=uninstall_monitor_and_tunnel.sh")

	// // 数据填充
	// data := struct {
	// 	// HostName          string
	// 	// Token             string
	// 	// SshTunnelUsername string
	// 	Port              int
	// 	// GithubRepoUrl     string
	// 	// PublicServerIP    string
	// }{
	// 	Port:              sshport.Port,
	// }

	// 执行模板渲染并写入响应
	if err := tmpl.Execute(c.Writer, struct{}{}); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}
