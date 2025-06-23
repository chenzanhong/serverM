package install

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	gs "serverM/server/handle/agent/getscript"
	"serverM/server/logs"
	"serverM/server/model"
	"time"

	"serverM/server/redis"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
)

type SshInfo struct {
	Host         string  `json:"host"`
	User         string  `json:"user"`
	Password     string  `json:"password"`
	Port         int     `json:"port"`
	Host_Name    string  `json:"host_name"`
	OS           string  `json:"os"`
	Platform     string  `json:"platform"`
	KernelArch   string  `json:"kernel_arch"`
	CPUThreshold float64 `json:"cpu_threshold"`
	MemThreshold float64 `json:"mem_threshold"`
	Token        string  `json:"token"`
}

// InstallAgent 添加服务器
func InstallAgent(c *gin.Context) {
	Username, exists := c.Get("username")
	if !exists {
		log.Printf("未找到用户名")
		c.JSON(401, gin.H{
			"code":    401,
			"success": false,
			"message": "未找到用户信息",
		})
		return
	}
	username := Username.(string)

	// 启动事务
	tx, err := model.DB.Begin()
	if err != nil {
		log.Printf("InstallAgent: failed to begin transaction: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}

	// 使用 defer 语句来处理事务的提交或回滚
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("InstallAgent: recovered from panic: %v, transaction rolled back", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}()

	// 解析json body 到结构体 SshInfo
	var agentInfo SshInfo
	if err := c.BindJSON(&agentInfo); err != nil {
		log.Printf("解析请求失败", err)
		logs.Sugar.Errorw("添加服务器", "username", username, "detail", "解析请求失败，请检查请求格式是否正确")
		tx.Rollback()
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var detail = fmt.Sprintf("添加服务器,ip:%s,user:%s,password:%s,port:%d,host_name:%s,os:%s,platform:%s,kernel_arch:%s,cpu_threshold:%f,mem_threshold:%f",
		agentInfo.Host, agentInfo.User, agentInfo.Password, agentInfo.Port, agentInfo.Host_Name, agentInfo.OS, agentInfo.Platform, agentInfo.KernelArch, agentInfo.CPUThreshold, agentInfo.MemThreshold)

	// 检查数据库中是否存在相同的 host_name
	var exist bool
	query := `SELECT EXISTS (SELECT 1 FROM host_info WHERE host_name = $1)`
	err = tx.QueryRow(query, agentInfo.Host_Name).Scan(&exist)
	if err != nil {
		log.Printf("数据库查询hostname失败")
		logs.Sugar.Errorw("添加服务器", "username", username, "detail", "数据库查询hostname失败。"+detail)
		tx.Rollback()
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to check host_name in database"})
		return
	}

	// 如果 host_name 已存在，返回错误并停止安装
	if exist {
		logs.Sugar.Errorw("添加服务器", "username", username, "detail", "host_name 已经存在。"+detail)
		tx.Commit() //  提交事务
		c.IndentedJSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("host_name '%s' already exists", agentInfo.Host_Name)})
		return
	}

	// 生成16位随机token
	token, err := generateToken(16)
	if err != nil {
		logs.Sugar.Errorw("添加服务器", "username", username, "detail", "生成token失败。"+detail)
		tx.Rollback()
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	agentInfo.Token = token

	//查找company_id
	var company_id int
	query = `SELECT company_id FROM users  WHERE name = $1`
	err = tx.QueryRow(query, username).Scan(&company_id)
	if err != nil {
		log.Println(logs.GetLogPrefix(2) + "获取company_id失败")
		logs.Sugar.Errorw("添加服务器", "username", username, "detail", "获取company_id失败。"+detail)
		tx.Rollback()
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get company_id"})
		return
	}

	// 插入 host_info 表
	var hostInfo model.HostInfo
	hostInfo.Hostname = agentInfo.Host_Name
	hostInfo.IP = agentInfo.Host
	hostInfo.OS = agentInfo.OS
	hostInfo.Platform = agentInfo.Platform
	hostInfo.KernelArch = agentInfo.KernelArch
	hostInfo.Token = agentInfo.Token

	cpuThreshold := agentInfo.CPUThreshold / 100.0
	memThreshold := agentInfo.MemThreshold / 100.0

	hostInfo.CPUThreshold = cpuThreshold
	hostInfo.MemThreshold = memThreshold
	hostInfo.CreatedAt = time.Now()
	hostInfo.CompanyID = company_id
	err = model.InsertHostInfoTx(tx, hostInfo, username)
	if err != nil {
		log.Println(logs.GetLogPrefix(2) + "插入host_info表失败")
		logs.Sugar.Errorw("添加服务器", "username", username, "detail", "插入host_info表失败。"+detail)
		tx.Rollback()
		s := fmt.Sprintf("Failed to insert host info: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": s})
		return
	}

	// 将阈值存入 Redis
	memKey := fmt.Sprintf("mem_threshold:%s", agentInfo.Host_Name)
	cpuKey := fmt.Sprintf("cpu_threshold:%s", agentInfo.Host_Name)
	err = redis.Rdb.Set(context.Background(), memKey, memThreshold, 0).Err()
	if err != nil {
		log.Println(logs.GetLogPrefix(2) + "存储内存阈值失败")
		logs.Sugar.Errorw("添加服务器", "username", username, "detail", "存储内存阈值失败。"+detail)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store memory threshold in Redis"})
		return
	}
	err = redis.Rdb.Set(context.Background(), cpuKey, cpuThreshold, 0).Err()
	if err != nil {
		log.Println(logs.GetLogPrefix(2) + "存储CPU阈值失败")
		logs.Sugar.Errorw("添加服务器", "username", username, "detail", "存储CPU阈值失败。"+detail)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store CPU threshold in Redis"})
		return
	}

	// 存储host_name和token到数据库
	err = model.InsertHostandTokenTx(tx, agentInfo.Host_Name, agentInfo.Token)
	if err != nil {
		log.Println(logs.GetLogPrefix(2) + "存储host_name和token失败")
		logs.Sugar.Errorw("添加服务器", "username", username, "detail", "存储host_name和token失败。"+detail)
		tx.Rollback()
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert host info into database"})
		return
	}

	// 添加阈值状态记录
	if err = model.InsertWarningStates(tx, agentInfo.Host_Name); err != nil {
		log.Println("添加服务器，创建阈值状态记录失败")
		logs.Sugar.Errorw("添加服务器", "username", username, "detail", "创建阈值状态记录失败。"+detail)
		tx.Rollback()
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert warning state into database"})
		return
	}

	// 改为在前端向用户展示下载、安装、执行代理的步骤，后端只负责数据库操作
	// // 添加服务器
	// err = DoInstallAgent(agentInfo)
	// if err != nil {
	// 	c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	scriptBytes, err := gs.GenerateCombinedScriptBytes(agentInfo.Host_Name, agentInfo.Token)
	if err != nil {
		log.Printf("InstallAgent: 生成脚本错误: %v", err)
		logs.Sugar.Errorw("添加服务器", "username", username, "detail", "生成脚本错误。"+detail)
		tx.Rollback()
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/octet-stream")
	filename := "install_agent.sh"
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	// 返回脚本文件
	if _, err := c.Writer.Write(scriptBytes); err != nil { // 注意检查 Write 的错误
		log.Printf("InstallAgent: 写入响应体错误: %v", err)
		logs.Sugar.Errorw("添加服务器", "username", username, "detail", "写入响应体错误。"+detail)
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to write script to response",
		})
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Printf("InstallAgent: failed to commit transaction: %v", err)
		logs.Sugar.Errorw("添加服务器", "username", username, "detail", "提交事务失败。"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit database changes"})
		return
	}

	logs.Sugar.Infow("添加服务器", "username", username, "detail", "添加服务器成功。"+detail)
	// 安装成功，返回成功信息
	// c.IndentedJSON(http.StatusOK, gin.H{"message": "Agent installed successfully", "host_name": agentInfo.Host_Name, "token": agentInfo.Token})
}

// 随机生成指定长度的随机token
func generateToken(length int) (string, error) {
	bytes := make([]byte, length/2) // 16字节 = 16位字符
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// DoInstallAgent 执行 agent 安装
func DoInstallAgent(ss SshInfo) error {
	// SSH 配置
	config := &ssh.ClientConfig{
		User: ss.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(ss.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// 建立 SSH 连接
	s := fmt.Sprintf("%s:%v", ss.Host, ss.Port)
	client, err := ssh.Dial("tcp", s, config)
	if err != nil {
		fmt.Printf("Failed to dial: %s", err)
		return err
	}
	defer client.Close()

	// 创建新会话
	session, err := client.NewSession()
	if err != nil {
		fmt.Printf("Failed to create session: %s", err)
		return err
	}
	defer session.Close()
	packageCmd := ""
	switch ss.Platform {
	case "ubuntu", "debian":
		packageCmd = "apt update && apt install -y git"
	case "centos", "rhel", "fedora":
		packageCmd = "yum install -y git"
	default:
		return fmt.Errorf("unsupported platform: %s", ss.Platform)
	}
	cmd := fmt.Sprintf(
		`#!/bin/bash
# 克隆代码仓库
%s

git clone https://gitee.com/wu-jinhao111/agent.git
cd agent/agent || exit

# 授予执行权限并运行主程序
chmod +x main
./main -hostname="%s" -token="%s" &

# 创建 systemd 服务文件
cat <<EOF | sudo tee /etc/systemd/system/main_startup.service
[Unit]
Description=Main Program Startup Service
After=network.target

[Service]
Type=simple
ExecStart=$HOME/agent/agent/main -hostname=%s -token=%s
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
sudo systemctl enable main_startup.service
sudo systemctl start main_startup.service

		`,
		packageCmd, ss.Host_Name, ss.Token, ss.Host_Name, ss.Token)
	// 使用 channel 来传递执行结果
	resultChan := make(chan error)

	// 启动 goroutine 执行命令
	go func() {
		err := session.Start(cmd)
		if err != nil {
			resultChan <- fmt.Errorf("failed to run command: %s", err)
			return
		}

		// 等待命令完成
		err = session.Wait()
		if err != nil {
			resultChan <- fmt.Errorf("command execution failed: %s", err)
			return
		}

		resultChan <- nil
	}()

	// 设置超时时间为 30 秒
	select {
	case err := <-resultChan:
		return err
	case <-time.After(30 * time.Second):
		return nil
	}
}
