package main

import (
	"fmt"
	"log"
	"time"

	// "serverM/server/handle/server/transfer/global"
	trans "serverM/server/handle/transfer/trans-init" // 请替换为您的实际项目路径
)

func main() {
	// 初始化SSH连接池
	pool := trans.NewSSHConnectionPool(10, 5*time.Minute)
	stopChan := make(chan struct{})
	defer close(stopChan)
	go pool.Cleanup(stopChan) // 启动清理协程

	// 示例：为两个服务器创建SSH连接并添加到池中
	err := trans.CreateConnectionToPool(pool, "192.168.202.128", "root", "czh_centos")
	if err != nil {
		log.Printf("创建与服务器 192.168.202.128 的连接失败: %v", err)
		return
	}
	err = trans.CreateConnectionToPool(pool, "113.44.170.52", "root", "czh2004_centos")
	if err != nil {
		log.Printf("创建与服务器 113.44.170.52 的连接失败: %v", err)
		return
	}

	// 创建FileTransferService实例
	service := trans.NewFileTransferService(pool)

	// 执行文件传输任务
	taskID, err := service.CreateTransferBetween2STask(
		"192.168.202.128",      // 源服务器IP
		"/home/czh/docker.txt", // 源文件路径
		"113.44.170.52",        // 目标服务器IP
		"/root/docker.txt",     // 目标文件路径
	)
	if err != nil {
		log.Printf("文件传输失败: %v", err)
	}

	fmt.Printf("文件传输任务已完成，任务ID: %s\n", taskID)
}
