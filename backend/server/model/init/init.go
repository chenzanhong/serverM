package init

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	cf "serverM/server/config"
	u "serverM/server/model/user"
	"strconv"
	"time"

	//"os/signal"

	//"regexp"
	"strings"

	"database/sql"

	"github.com/go-redis/redis/v8"
	_ "github.com/taosdata/driver-go/v3/taosSql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 属性均用驼峰命名转换后的含_的，表名就不含_。
const createTableSQL = `
-- roles 表
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
);

-- users 表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR UNIQUE NOT NULL,
	realname VARCHAR ,
    email VARCHAR UNIQUE NOT NULL,
    password VARCHAR NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE,
    role_id INT REFERENCES roles(id) DEFAULT 0,
    company_id INT DEFAULT 0,
	token TEXT, -- 存储更新密码时使用的唯一凭证，用于验证用户身份与验证码是否匹配
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- company 表
CREATE TABLE IF NOT EXISTS companies (
    id SERIAL PRIMARY KEY,
    name VARCHAR UNIQUE NOT NULL,
	social_credit_code VARCHAR(18) UNIQUE NOT NULL,
	memberNum int DEFAULT 0,
	systemNum int DEFAULT 0,
    admin_id INT REFERENCES users(id) DEFAULT 0,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
	
-- host表
CREATE TABLE IF NOT EXISTS host_info (
	id SERIAL PRIMARY KEY,
    user_name VARCHAR, -- REFERENCES users(name),
	host_name VARCHAR(255)  UNIQUE,
	ip VARCHAR(255)  UNIQUE,
	port INT DEFAULT 22,
	company_id INT DEFAULT 0, -- REFERENCES company(id),
	os TEXT NOT NULL,
	platform TEXT NOT NULL,
	kernel_arch TEXT NOT NULL,
	cpu_threshold FLOAT DEFAULT 0.9,
	mem_threshold FLOAT DEFAULT 0.9,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- TIMESTAMP WITH TIME ZONE 加上时区
);

-- token表
CREATE TABLE IF NOT EXISTS hostandtoken (
	id SERIAL PRIMARY KEY,
	host_name VARCHAR(255) , -- REFERENCES host_info(host_name),
	token TEXT NOT NULL,
	last_heartbeat TIMESTAMP DEFAULT NOW(),
	status VARCHAR(10) DEFAULT 'offline'
);

-- sshkey表
CREATE TABLE IF NOT EXISTS ssh_keys (
    id SERIAL PRIMARY KEY,
    host_name VARCHAR(255) , 
    sshkey TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ssh_ports表，用于生成配置反向ssh的脚本
CREATE TABLE IF NOT EXISTS ssh_ports (
	id SERIAL PRIMARY KEY,
    port INT,
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    hostname VARCHAR(255), 
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- notice 表
CREATE TABLE IF NOT EXISTS notices (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
	state VARCHAR DEFAULT 'unprocessed' ,
	send VARCHAR, -- REFERENCES users(name),
	receive VARCHAR, -- REFERENCES users(name),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- warning 表 预警记录
CREATE TABLE IF NOT EXISTS warnings (
    id SERIAL PRIMARY KEY,
	host_name VARCHAR(255) ,
	username VARCHAR(255) ,
    warning_type VARCHAR ,
	warning_title VARCHAR,
	warning_time TIMESTAMP DEFAULT NOW()
);

-- warning_states 预警状态状态表
CREATE TABLE IF NOT EXISTS warning_states (
 	id SERIAL PRIMARY KEY,
	host_name VARCHAR(255) ,
    warning_type VARCHAR(255) ,
	is_warning BOOLEAN DEFAULT FALSE, -- 准确点说应该是“是否预警过”，用于区分是否第一次预警，其实也可以不要这个字段，应该是否需要预警是通过冷静期的，只需要last_time
	last_time TIMESTAMP DEFAULT NOW()
); 

-- 在hostandtoken表的host_name字段上创建索引，加速主机名查找
CREATE INDEX IF NOT EXISTS idx_hostandtoken_host_name ON hostandtoken(host_name);

-- 如果经常按last_heartbeat查询或排序，可以在此字段上创建索引
CREATE INDEX IF NOT EXISTS idx_hostandtoken_last_heartbeat ON hostandtoken(last_heartbeat);

`

// 创建system的超级表
const systemSuperTable = `
CREATE STABLE if not exists system_info (
	created_at TIMESTAMP,
    host_name VARCHAR(255),
	host_info VARCHAR(4096),
    cpu_info VARCHAR(4096),
    memory_info VARCHAR(4096),
    process_info VARCHAR(4096),
    network_info VARCHAR(4096)
) TAGS (
    tags_host_name VARCHAR(255)
);
`

var DB *gorm.DB
var TDengineDB *sql.DB

var CTX = context.Background()
var RDB *redis.Client

func InitRedis() error {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDBstr := os.Getenv("REDIS_DB")
	redisDB, err := strconv.Atoi(redisDBstr)
	if err != nil {
		log.Fatalf("Failed to parse Redis DB number: %v", err)
		return err
	}

	if redisAddr == "" {
		log.Fatal("Redis configuration is missing")
		return fmt.Errorf("Redis configuration is missing")
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     redisAddr,     // Redis地址
		Password: redisPassword, // 无密码
		DB:       redisDB,       // 使用默认DB
	})

	// 测试连接
	_, err = RDB.Ping(CTX).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
		return err
	}
	return nil
}

// ConnectDatabase 连接到数据库
func ConnectDatabase() error {
	var err error

	// 获取数据库连接信息
	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	// 使用gorm打开数据库连接
	DB, err = gorm.Open(postgres.Open(dsn))
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(10) // 默认值2容易导致连接不足
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err != nil {
		return err // 返回连接错误
	}
	return nil
}

// 连接TDengine数据库
func ConnectTDengine() error {
	var err error

	// 获取数据库连接信息
	user := os.Getenv("TDENGINE_USER")
	password := os.Getenv("TDENGINE_PASSWORD")
	dbname := os.Getenv("TDENGINE_NAME")
	host := os.Getenv("TDENGINE_HOST")
	port := os.Getenv("TDENGINE_PORT")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbname)

	//打开数据库
	TDengineDB, err = sql.Open("taosSql", dsn)
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}

	// 测试连接
	if err := TDengineDB.Ping(); err != nil {
		log.Fatalf("Failed to ping TDengine:%v", err)
		return err
	}

	return nil
}

// InitDB 初始化数据库，创建所需的表
func InitDB() error {
	//初始化postgresql数据库
	if DB == nil {
		return fmt.Errorf("database connection is not initialized") // 检查数据库连接是否已初始化
	}

	tx := DB.Begin() // 开始事务
	if tx.Error != nil {
		return tx.Error // 返回事务错误
	}

	if err := tx.Exec(createTableSQL).Error; err != nil {
		tx.Rollback() // 回滚事务
		return err    // 返回创建表时的错误
	}

	if err := tx.Commit().Error; err != nil {
		return err // 返回提交事务时的错误
	}

	//初始化TDengine数据库
	// if TDengineDB == nil {
	// 	return fmt.Errorf("TDengine database connection is not initialized") // 检查数据库连接是否已初始化
	// }

	// 创建超级表
	// if _, err := TDengineDB.Exec(systemSuperTable); err != nil {
	// 	return err
	// }

	return nil
}

func isValidJSON(data string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(data), &js) == nil
}

// InitDBData 初始化数据库的基本数据
func InitDBData() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized") // 检查数据库连接是否已初始化
	}

	tx := DB.Begin() // 开始事务
	if tx.Error != nil {
		return tx.Error // 返回事务错误
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 如果发生panic，回滚事务
		}
	}()

	var user u.User
	result := tx.Where("name=?", "root").First(&user) // 查找用户名为root的用户

	if result.Error == nil {
		log.Printf("Root already exists") // 用户已存在
		tx.Commit()                       // 提交事务
		return nil
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("Failed to find user: %v", result.Error)
		tx.Rollback()       // 回滚事务
		return result.Error // 返回查找用户错误
	}

	// 插入角色数据
	if err := insertRoles(tx); err != nil {
		tx.Rollback() // 回滚事务
		return err    // 返回插入角色时的错误
	}
	fmt.Println("1---------------")

	// 插入用户数据
	if err := insertUsers(tx); err != nil {
		tx.Rollback() // 回滚事务
		return err    // 返回插入用户时的错误
	}
	fmt.Println("2---------------")

	// 插入公司数据
	if err := insertCompanies(tx); err != nil {
		tx.Rollback() // 回滚事务
		return err    // 返回插入用户时的错误
	}
	fmt.Println("3---------------")

	// 插入 host_info 数据
	if err := insertHostInfo(tx); err != nil {
		tx.Rollback() // 回滚事务
		return err    // 返回插入主机信息时的错误
	}
	fmt.Println("4---------------")

	// 插入 system_info 数据
	//if err := insertSystemInfo(tx); err != nil {
	//	tx.Rollback() // 回滚事务
	//	return err    // 返回插入系统信息时的错误
	//}
	//fmt.Println("5---------------")

	// 插入 hostandtoken 数据
	if err := insertHostAndToken(tx); err != nil {
		tx.Rollback() // 回滚事务
		return err    // 返回插入 token 信息时的错误
	}
	fmt.Println("5---------------")

	//插入 sshkeys数据
	if err := insertSSHKeys(tx); err != nil {
		tx.Rollback() // 回滚事务
		return err    // 返回插入 sshkeys 信息时的错误
	}
	fmt.Println("6---------------")

	//插入 notices数据
	if err := insertNotices(tx); err != nil {
		tx.Rollback() // 回滚事务
		return err    // 返回插入 notices 信息时的错误
	}
	fmt.Println("7---------------")

	// 插入端口池数据
	if err := initPortPool(tx); err != nil {
		tx.Rollback()
		return err
	}
	fmt.Println("8---------------")

	//插入warning数据
	if err := insertWarning(tx); err != nil {
		tx.Rollback()
		return err
	}
	fmt.Println("9---------------")

	if err := tx.Commit().Error; err != nil {
		return err // 返回提交事务时的错误
	}

	return nil
}

// 初始化TDengine数据
func InitTDengine() error {

	if TDengineDB == nil {
		return fmt.Errorf("TDengine database connection is not initialized")
	}
	/* 	// 设置信号处理
	   	signals := make(chan os.Signal, 1)
	   	signal.Notify(signals, os.Interrupt, os.Kill)
	   	go func() {
	   		<-signals
	   		fmt.Println("Received signal, closing database connection...")
	   		TDengineDB.Close()
	   		os.Exit(1)
	   	}() */

	//插入system_info子表数据
	if err := insertSystemInfo(TDengineDB); err != nil {
		return err
	}
	fmt.Print("initTDengine---------------")
	return nil
}

// insertRoles 函数从 roles.txt 文件中读取角色数据
func insertRoles(tx *gorm.DB) error {
	file, err := os.Open("asset/example/roles.txt")
	if err != nil {
		return fmt.Errorf("failed to open roles file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 检查是否以 "//" 开头
		if strings.HasPrefix(line, "//") {
			fmt.Println("Encountered a comment line, exiting the loop.")
			break // 退出循环
		}
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			return fmt.Errorf("invalid line format: %s", line)
		}

		id := parts[0]
		roleName := parts[1]
		description := parts[2]

		if err := tx.Exec("INSERT INTO roles (id, role_name, description) VALUES (?, ?, ?)", id, roleName, description).Error; err != nil {
			return fmt.Errorf("failed to insert role %s: %w", roleName, err)
		}
	}
	return scanner.Err() // 返回扫描器的错误（如果有）
}

// insertUsers 函数从 users.txt 文件中读取用户数据
func insertUsers(tx *gorm.DB) error {
	file, err := os.Open("asset/example/users.txt")
	if err != nil {
		return fmt.Errorf("failed to open users file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 检查是否以 "//" 开头
		if strings.HasPrefix(line, "//") {
			fmt.Println("Encountered a comment line, exiting the loop.")
			break // 退出循环
		}
		parts := strings.Split(line, ",")
		if len(parts) < 6 {
			return fmt.Errorf("invalid line format: %s", line)
		}
		name := parts[0]
		realname := parts[1]
		email := parts[2]
		password := parts[3]
		roleID := parts[4]
		companyID := parts[5]

		if err := tx.Exec("INSERT INTO users (name, realname,email, password, role_id, company_id) VALUES (?, ?, ?, ?, ?, ?)", name, realname, email, password, roleID, companyID).Error; err != nil {
			return fmt.Errorf("failed to insert user %s: %w", name, err)
		}
	}
	return scanner.Err()
}

// insertHostInfo 函数从 host_info.txt 文件中读取主机信息
func insertHostInfo(tx *gorm.DB) error {
	file, err := os.Open("asset/example/host_info.txt")
	if err != nil {
		return fmt.Errorf("failed to open host_info file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 检查是否以 "//" 开头
		if strings.HasPrefix(line, "//") {
			fmt.Println("Encountered a comment line, exiting the loop.")
			break // 退出循环
		}
		parts := strings.Split(line, ",")
		if len(parts) < 6 {
			return fmt.Errorf("invalid line format: %s", line)
		}

		userName := parts[0]
		hostname := parts[1]
		ip := parts[2]
		companyId := parts[3]
		os := parts[4]
		platform := parts[5]
		kernelArch := parts[6]

		if err := tx.Exec("INSERT INTO host_info (user_name, host_name, ip, company_id, os, platform, kernel_arch) VALUES (?, ?, ?, ?, ?, ?, ?)", userName, hostname, ip, companyId, os, platform, kernelArch).Error; err != nil {
			return fmt.Errorf("failed to insert host_info for %s: %w", hostname, err)
		}
	}
	return scanner.Err()
}

// insertUsers 函数从 users.txt 文件中读取用户数据
func insertCompanies(tx *gorm.DB) error {
	file, err := os.Open("asset/example/companies.txt")
	if err != nil {
		return fmt.Errorf("failed to open companies file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 检查是否以 "//" 开头
		if strings.HasPrefix(line, "//") {
			fmt.Println("Encountered a comment line, exiting the loop.")
			break // 退出循环
		}
		parts := strings.Split(line, ",")
		if len(parts) < 6 {
			return fmt.Errorf("invalid line format: %s", line)
		}

		name := parts[0]
		companyCode := parts[1]
		memberNum := parts[2]
		systemNum := parts[3]
		admin_id := parts[4]
		description := parts[5]
		if err := tx.Exec("INSERT INTO companies (name, social_credit_code ,memberNum, systemNum, admin_id, description) VALUES (?, ?, ?, ?, ?, ?)", name, companyCode, memberNum, systemNum, admin_id, description).Error; err != nil {
			return fmt.Errorf("failed to insert user %s: %w", name, err)
		}
	}
	return scanner.Err()
}

// insertSystemInfo 函数从 system_info.txt 文件中读取系统信息
func insertSystemInfo(t *sql.DB) error {
	file, err := os.Open("asset/example/system_info.txt")
	if err != nil {
		return fmt.Errorf("failed to open system_info file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 检查是否以 "//" 开头
		if strings.HasPrefix(line, "//") {
			fmt.Println("Encountered a comment line, exiting the loop.")
			break // 退出循环
		}
		// fmt.Println("Read line:", line) // 输出读取的行（调试用）

		parts := strings.Split(line, ",,")
		// fmt.Println()
		// fmt.Println("len:",len(parts))
		// fmt.Println()
		if len(parts) < 6 {
			return fmt.Errorf("invalid line format: %s", line)
		}
		hostName := parts[0]
		hostInfo := parts[1]
		cpuInfo := parts[2]
		memoryInfo := parts[3]
		processInfo := parts[4]
		networkInfo := parts[5]
		// fmt.Println("hostName:", hostName)
		// fmt.Println("hostInfo:", hostInfo)
		// fmt.Println("cpuInfo:", cpuInfo)
		// fmt.Println("memoryInfo:", memoryInfo)
		// fmt.Println("processInfo:", processInfo)
		// fmt.Println("networkInfo:", networkInfo)

		// 验证每个 JSON 字符串的有效性
		if !isValidJSON(hostInfo) || !isValidJSON(cpuInfo) || !isValidJSON(memoryInfo) || !isValidJSON(processInfo) || !isValidJSON(networkInfo) {
			return fmt.Errorf("invalid JSON data for host %s", hostName)
		}

		// 插入数据库（注意：这里假设数据库表 system_info 的对应字段已经设置为接受 jsonb 类型）
		//if err := tx.Exec(
		//	"INSERT INTO system_info (host_name, host_info_id, cpu_info, memory_info, process_info, network_info) VALUES (?, ?, ?::jsonb, ?::jsonb, ?::jsonb, ?::jsonb)",
		//	hostName, hostInfoID, cpuInfo, memoryInfo, processInfo, networkInfo,
		//).Error; err != nil {
		//	return fmt.Errorf("failed to insert system info for host %s: %w", hostName, err)
		//}

		// 拼接子表名
		tableName := fmt.Sprintf("%s_system_info", hostName)

		createTableQuery := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s USING system_info TAGS('%s')
		`, tableName, hostName)

		_, err := t.Exec(createTableQuery)
		if err != nil {
			log.Printf("Error creating table %s: %v\n", tableName, err)
			return err
		}

		currentTime := time.Now().Format("2006-01-02 15:04:05") // 格式化时间为TDengine接受的格式

		hostInfoJSON, err := json.Marshal(hostInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal hostInfo: %w", err)
		}
		cpuInfoJSON, err := json.Marshal(cpuInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal cpuInfo: %w", err)
		}
		memoryInfoJSON, err := json.Marshal(memoryInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal memoryInfo: %w", err)
		}
		processInfoJSON, err := json.Marshal(processInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal processInfo: %w", err)
		}
		networkInfoJSON, err := json.Marshal(networkInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal networkInfo: %w", err)
		}

		insertDataQuery := fmt.Sprintf(`
			INSERT INTO %s (created_at, host_name, host_info, cpu_info, memory_info, process_info, network_info) 
			VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s')
		`, tableName, currentTime, hostName, string(hostInfoJSON), string(cpuInfoJSON), string(memoryInfoJSON), string(processInfoJSON), string(networkInfoJSON))

		_, err = t.Exec(insertDataQuery)
		if err != nil {
			log.Printf("Error inserting data into table %s: %v\n", tableName, err)
			return err
		}
	}
	return scanner.Err() // 返回读取文件的错误（如果有）
}

// insertHostAndToken 函数从 hostandtoken.txt 文件中读取 token 数据
func insertHostAndToken(tx *gorm.DB) error {
	file, err := os.Open("asset/example/hostandtoken.txt")
	if err != nil {
		return fmt.Errorf("failed to open hostandtoken file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 检查是否以 "//" 开头
		if strings.HasPrefix(line, "//") {
			fmt.Println("Encountered a comment line, exiting the loop.")
			break // 退出循环
		}
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			return fmt.Errorf("invalid line format: %s", line)
		}

		hostName := parts[0]
		token := parts[1]
		status := parts[2]

		if err := tx.Exec("INSERT INTO hostandtoken (host_name, token, status) VALUES (?, ?, ?)", hostName, token, status).Error; err != nil {
			return fmt.Errorf("failed to insert token for host %s: %w", hostName, err) // 返回详细错误
		}
	}
	return scanner.Err()
}

// insertSSHKeys 函数从 sshkeys.txt 文件中读取 SSH 密钥数据
func insertSSHKeys(tx *gorm.DB) error {
	file, err := os.Open("asset/example/sshkeys.txt")
	if err != nil {
		return fmt.Errorf("failed to open sshkeys file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 检查是否以 "//" 开头
		if strings.HasPrefix(line, "//") {
			fmt.Println("Encountered a comment line, exiting the loop.")
			break // 退出循环
		}
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			return fmt.Errorf("invalid line format: %s", line)
		}

		hostname := parts[0]
		sshkey := parts[1]

		if err := tx.Exec("INSERT INTO ssh_keys (host_name,sshkey) VALUES (?, ?)", hostname, sshkey).Error; err != nil {
			return fmt.Errorf("failed to insert ssh_keys for %s: %w", hostname, err)
		}
	}
	return scanner.Err()
}

// insertNotices 函数从 notices.txt 文件中读取通知数据
func insertNotices(tx *gorm.DB) error {
	file, err := os.Open("asset/example/notices.txt")
	if err != nil {
		return fmt.Errorf("failed to open notices file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 检查是否以 "//" 开头
		if strings.HasPrefix(line, "//") {
			fmt.Println("Encountered a comment line, exiting the loop.")
			break // 退出循环
		}
		parts := strings.Split(line, ",,")
		if len(parts) < 4 {
			return fmt.Errorf("invalid line format: %s", line)
		}

		content := parts[0]
		send := parts[1]
		receive := parts[2]
		state := parts[3]

		if err := tx.Exec("INSERT INTO notices (content,send,receive,state) VALUES (?, ?, ?, ?)", content, send, receive, state).Error; err != nil {
			return fmt.Errorf("failed to insert notices for %s: %w", content, err)
		}
	}
	return scanner.Err()
}

// insertWarning  函数从 warning.txt 文件中读取警告数据
func insertWarning(tx *gorm.DB) error {
	file, err := os.Open("asset/example/warning.txt")
	if err != nil {
		return fmt.Errorf("failed to open warning file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 检查是否以 "//" 开头
		if strings.HasPrefix(line, "//") {
			fmt.Println("Encountered a comment line, exiting the loop.")
			break // 退出循环
		}
		parts := strings.Split(line, ",")
		if len(parts) < 4 {
			return fmt.Errorf("invalid line format: %s", line)
		}

		hostname := parts[0]
		username := parts[1]
		warning_type := parts[2]
		warning_title := parts[3]
		warning_time := time.Now().Format("2006-01-02 15:04:05")
		fmt.Println(hostname)
		fmt.Println(username)
		fmt.Println(warning_type)
		fmt.Println(warning_title)
		fmt.Println(warning_time)

		if err := tx.Exec("INSERT INTO warnings (host_name, username, warning_type, warning_title, warning_time) VALUES (?, ?, ?, ?, ?)", hostname, username, warning_type, warning_title, warning_time).Error; err != nil {
			return fmt.Errorf("failed to insert warning for %s: %w", hostname, err)
		}
	}
	return scanner.Err()
}

// insertPortPool 根据配置文件config.yaml的参数初始化ssh_ports表
func initPortPool(tx *gorm.DB) error {
	startPort := cf.StartPort
	endPort := cf.EndPort
	time := time.Now()
	for port := startPort; port <= endPort; port++ {
		err := tx.Create(&u.SSHPort{Port: port, UpdatedAt: time}).Error
		if err != nil {
			return fmt.Errorf("failed to insert port %d: %w", port, err)
		}
	}
	return nil
}
