package config

import (
	"log"
	"os"
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
	StartPort         string `yaml:"start_port"`
	EndPort           string `yaml:"end_port"`
	SshTunnelUsername string `yaml:"ssh_tunnel_username"`
	SshTunnelPassword string `yaml:"ssh_tunnel_password"`
	PublicServerIP    string `yaml:"public_server_ip"`
	GithubRepoUrl     string `yaml:"github_repo_url"`
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
	StartPort         int    // 配置反向ssh服务可用的最小端口号
	EndPort           int    // 配置反向ssh服务可用的最大端口号
	PublicServerIP    string // 部署该项目的公网服务器
	GithubRepoUrl     string
	SshTunnelUsername string
	SshTunnelPassword string
)

// getDBConfigPath 获取数据库配置文件的路径
func getDBConfigPath() string {
	_, filename, _, ok := runtime.Caller(2) // 获取调用者的文件名
	if !ok {
		log.Printf("警告: 无法获取运行时调用者信息，使用默认配置路径")
		return "/app/config/configs/config.yaml"
	}

	// 获取当前文件所在的目录
	currentDir := filepath.Dir(filename)

	// 构建到项目根目录的相对路径
	dbConfigPath := filepath.Join(currentDir, "..", "config", "configs", "config.yaml")

	// 将路径转换为绝对路径并简化路径
	absPath, err := filepath.Abs(dbConfigPath)
	if err != nil {
		log.Printf("警告: 无法获取绝对路径: %v，使用默认配置路径", err)
		return "/app/config/configs/config.yaml"
	}

	simplifiedPath := filepath.Clean(absPath)
	return simplifiedPath
}

// GetDBConfigPath 返回数据库配置文件的路径
func GetDBConfigPath() string {
	return getDBConfigPath()
}

// LoadConfig 加载配置文件并返回 Config
func LoadConfig() (*Config, error) {
	// 首先尝试从环境变量获取配置路径
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = GetDBConfigPath()
	}
	
	log.Printf("正在加载配置文件: %s", configPath)
	
	// 使用 os.ReadFile 替代已弃用的 ioutil.ReadFile
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("警告: 读取配置文件失败: %v，检查配置文件是否存在", err)
		// 创建默认配置
		config := createDefaultConfig()
		log.Println("使用默认配置继续")
		return config, nil
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Printf("错误: 解析配置文件失败: %v", err)
		return nil, err
	}
	
	// 从环境变量中覆盖配置
	// 数据库配置覆盖
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.DB.Host = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		config.DB.Port = dbPort
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.DB.Name = dbName
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		config.DB.User = dbUser
	}
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		config.DB.Password = dbPassword
	}
	
	// Redis配置覆盖
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		config.Redis.Host = redisHost
	}
	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		config.Redis.Port = redisPort
	}
	if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
		config.Redis.DB = redisDB
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config.Redis.Password = redisPassword
	}
	
	// Email配置覆盖
	if emailName := os.Getenv("EMAIL_NAME"); emailName != "" {
		config.Email.Name = emailName
	}
	if emailPassword := os.Getenv("EMAIL_PASSWORD"); emailPassword != "" {
		config.Email.Password = emailPassword
	}
	
	// SMTP配置覆盖
	if smtpHost := os.Getenv("SMTP_SERVER_HOST"); smtpHost != "" {
		config.SMTPServer.Host = smtpHost
	}
	if smtpPort := os.Getenv("SMTP_SERVER_PORT"); smtpPort != "" {
		config.SMTPServer.Port = smtpPort
	}
	
	// 脚本配置覆盖
	if startPort := os.Getenv("SCRIPT_START_PORT"); startPort != "" {
		config.Script.StartPort = startPort
	}
	if endPort := os.Getenv("SCRIPT_END_PORT"); endPort != "" {
		config.Script.EndPort = endPort
	}
	if publicServerIP := os.Getenv("PUBLIC_SERVER_IP"); publicServerIP != "" {
		config.Script.PublicServerIP = publicServerIP
	}
	if githubRepoUrl := os.Getenv("GITHUB_REPO_URL"); githubRepoUrl != "" {
		config.Script.GithubRepoUrl = githubRepoUrl
	}
	if sshTunnelUsername := os.Getenv("SSH_TUNNEL_USERNAME"); sshTunnelUsername != "" {
		config.Script.SshTunnelUsername = sshTunnelUsername
	}
	if sshTunnelPassword := os.Getenv("SSH_TUNNEL_PASSWORD"); sshTunnelPassword != "" {
		config.Script.SshTunnelPassword = sshTunnelPassword
	}
	
	// 初始化全局变量
	StartPort, err = strconv.Atoi(config.Script.StartPort)
	if err != nil {
		log.Printf("警告: 解析 StartPort 失败: %v，使用默认值 30000", err)
		StartPort = 30000
	}
	
	EndPort, err = strconv.Atoi(config.Script.EndPort)
	if err != nil {
		log.Printf("警告: 解析 EndPort 失败: %v，使用默认值 30100", err)
		EndPort = 30100
	}
	
	// 验证端口范围有效性
	if StartPort >= EndPort {
		log.Printf("警告: 端口范围无效(StartPort >= EndPort)，使用默认值")
		StartPort = 30000
		EndPort = 30100
	}
	
	// 设置其他全局变量，添加默认值
	PublicServerIP = config.Script.PublicServerIP
	if PublicServerIP == "" {
		log.Printf("警告: 未设置 PublicServerIP，使用默认值 localhost")
		PublicServerIP = "localhost"
	}
	
	GithubRepoUrl = config.Script.GithubRepoUrl
	if GithubRepoUrl == "" {
		log.Printf("警告: 未设置 GithubRepoUrl")
	}
	
	SshTunnelUsername = config.Script.SshTunnelUsername
	if SshTunnelUsername == "" {
		log.Printf("警告: 未设置 SshTunnelUsername")
	}
	
	SshTunnelPassword = config.Script.SshTunnelPassword
	if SshTunnelPassword == "" {
		log.Printf("警告: 未设置 SshTunnelPassword")
	}
	
	log.Println("配置加载成功")
	return &config, nil
}

// createDefaultConfig 创建默认配置
func createDefaultConfig() *Config {
	return &Config{
		DB: DBConfig{
			Host:     "localhost",
			Port:     "5432",
			Name:     "serverm",
			User:     "postgres",
			Password: "postgres",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       "0",
		},
		Script: Script{
			StartPort:         "30000",
			EndPort:           "30100",
			PublicServerIP:    "localhost",
			GithubRepoUrl:     "",
			SshTunnelUsername: "",
			SshTunnelPassword: "",
		},
		Email: EMAILConfig{
			Name:     "",
			Password: "",
		},
		SMTPServer: SMTPServerConfig{
			Host: "",
			Port: "25",
		},
	}
}
