package main

import (
	cf "agent/config"
	"agent/data"
	"path/filepath"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-co-op/gocron"
)

func main() {
	// 定义一个字符串变量，用来接收传入的 hostname 与token参数
	hostName := flag.String("hostname", "", "The hostname for the agent")
	token := flag.String("token", "", "A string of 16 characters")

	// 解析命令行参数
	flag.Parse()
	fmt.Printf("hostname:%s,token:%s\n", *hostName, *token)

	// 获取当前可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("无法获取可执行文件路径: %v", err)
	}
	rootDir := filepath.Dir(exePath) // 获取可执行文件所在目录

	// 构建 config.yaml 的绝对路径
	configPath := filepath.Join(rootDir, "config", "config.yaml")
	if err != nil {
		log.Fatalf("failed to open config file:%v", err)
		return
	}

	config, err := cf.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to open config file:%v", err)
		return
	}
	// 定义服务器端点的URL
	serverURL := fmt.Sprintf("http://%s:%s/agent/addSystem_info", config.Url, config.Port) // 你的服务器URL

	//创建调度器
	s := gocron.NewScheduler(time.UTC)
	// 每分钟执行一次任务
	s.Every(config.Second).Second().Do(func() {
		// 收集监控数据
		datas, err := data.CollectMonitorData(*hostName, *token)
		if err != nil {
			fmt.Println("收集数据错误")
		}
		fmt.Println(datas.HostInfo.Hostname)
		// 发送收集到的数据到服务器
		err = data.SendMonitorData(serverURL, datas)
		if err != nil {
			fmt.Printf("发送数据错误%v", err)
		}
	})
	s.StartAsync()
	select {} // 防止主 goroutine 结束
}
