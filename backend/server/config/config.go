package config

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"strconv"

	"gopkg.in/yaml.v2"
)

// DBConfig 用于保存数据库配置
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// TDengineConfig 用于保存TDengine数据库配置
type TDengineConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}

type EMAILConfig struct {
	Name     string `yaml:"email_name"`
	Password string `yaml:"email_password"`
}
type SMTPServerConfig struct {
	Host string `yaml:"SMTPServer_host"`
	Port string `yaml:"SMTPServer_port"`
}

type Script struct {
	StartPort string `yaml:"start_port"`
	EndPort   string `yaml:"end_port"`
	SshTunnelUsername string `yaml:"ssh_tunnel_username"`
	SshTunnelPassword string `yaml:"ssh_tunnel_password"`
	PublicServerIP string `yaml:"public_server_ip"`
	GithubRepoUrl  string `yaml:"github_repo_url"`
}

// Config 用于保存所有配置项
type Config struct {
	DB         DBConfig         `yaml:"db"`
	TDengine   TDengineConfig   `yaml:"tdengine"`
	Redis      RedisConfig      `yaml:"redis"`
	Email      EMAILConfig      `yaml:"email"`
	SMTPServer SMTPServerConfig `yaml:"smtp_server"`
	Script     Script           `yaml:"script"`
}

var (
	StartPort      int // 配置反向ssh服务可用的最小端口号
	EndPort        int // 配置反向ssh服务可用的最大端口号
	PublicServerIP string // 部署该项目的公网服务器
	GithubRepoUrl  string
	SshTunnelUsername string
	SshTunnelPassword string
)

// getDBConfigPath 获取数据库配置文件的路径
func getDBConfigPath() string {
	_, filename, _, ok := runtime.Caller(2) // 获取调用者的文件名
	if !ok {
		log.Fatal("无法获取运行时调用者信息")
	}

	// 获取当前文件所在的目录
	currentDir := filepath.Dir(filename)

	// 构建到项目根目录的相对路径
	dbConfigPath := filepath.Join(currentDir, "..", "config", "configs", "config.yaml")

	// 将路径转换为绝对路径并简化路径
	absPath, err := filepath.Abs(dbConfigPath)
	if err != nil {
		log.Printf("无法获取绝对路径: %v", err)
	}

	simplifiedPath := filepath.Clean(absPath)

	return simplifiedPath
}

// GetDBConfigPath 返回数据库配置文件的路径
func GetDBConfigPath() string {
	return getDBConfigPath()
}

// LoadConfig 加载配置文件并返回 DBConfig
func LoadConfig() (*Config, error) {
	configPath := GetDBConfigPath()
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Printf("读取配置文件失败: %v", err)
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Printf("解析配置文件失败: %v", err)
		return nil, err
	}
	StartPort, err = strconv.Atoi(config.Script.StartPort)
	if err != nil {
		log.Printf("Atoi(config.SSHPort.StartPort)失败: %v", err)
		return nil, err
	}
	EndPort, err = strconv.Atoi(config.Script.EndPort)
	if err != nil {
		log.Printf("Atoi(config.SSHPort.EndPort)失败: %v", err)
		return nil, err
	}
	PublicServerIP = config.Script.PublicServerIP
	GithubRepoUrl = config.Script.GithubRepoUrl
	SshTunnelUsername = config.Script.SshTunnelUsername
	SshTunnelPassword = config.Script.SshTunnelPassword
	return &config, nil
}
