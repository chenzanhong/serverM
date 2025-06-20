package logs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	// "github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Sugar *zap.SugaredLogger

func InitZapSugarDefault() {
	_, err := os.OpenFile("./logs/user_operation_log/user_operation_log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("创建/打开日志文件失败：%v", err)
		return
	}
	// 配置日志文件
	fileLooger := &lumberjack.Logger{
		Filename:   "./logs/user_operation_log/user_operation_log.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     60,
		Compress:   false,
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(fileLooger),
		zapcore.InfoLevel,
	)

	// logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	logger := zap.New(core)

	Sugar = logger.Sugar()
}

func SetupZapSugar(path string, level zapcore.Level) {
	fileLooger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     60,
		Compress:   false,
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(fileLooger),
		level,
	)

	// logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	logger := zap.New(core)

	Sugar = logger.Sugar()
}

type Log struct {
	Level     string `json:"level"`
	Timestamp string `json:"ts"`
	Msg       string `json:"msg"`
	Username  string `json:"username"`
	Detail    string `json:"detail"`
}

type LogRequest struct {
	Username  string `json:"username" form:"username"`
	FromTime  string `json:"fromTime" form:"fromTime"`
	ToTime    string `json:"toTime" form:"toTime"`
	Operation string `json:"operation" form:"operation"`
}

const timeFormat = "2006-01-02T15:04:05.999Z0700" // iso8601格式

// 过滤日志的核心逻辑
func FilterLogs(scanner *bufio.Scanner, logRequest LogRequest, username string) []Log {
	var logs []Log
	var _log Log
	var fromTime, toTime, logTime time.Time
	var err error

	if logRequest.FromTime != "" {
		fromTime, err = time.Parse(timeFormat, logRequest.FromTime)
		if err != nil {
			log.Printf("解析开始时间失败：%v", err)
			return nil
		}
	}

	if logRequest.ToTime != "" {
		toTime, err = time.Parse(timeFormat, logRequest.ToTime)
		if err != nil {
			log.Printf("解析结束时间失败：%v", err)
			return nil
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &_log); err != nil {
			log.Printf("解析日志失败：%v", err)
			continue
		}

		// 检查用户名
		if username != "" && _log.Username != username {
			continue
		}

		// 筛选操作类型
		if logRequest.Operation != "" && _log.Msg != logRequest.Operation {
			continue
		}

		logTime, err = time.Parse(timeFormat, _log.Timestamp)
		if err != nil {
			log.Printf("解析日志时间失败：%v", err)
			continue
		}

		// 时间范围筛选
		if logRequest.FromTime != "" && logTime.Before(fromTime) {
			continue
		}
		if logRequest.ToTime != "" && logTime.After(toTime) {
			continue
		}

		logs = append(logs, _log)
	}

	// 逆序，确保时间最新的日志放在最前面
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	return logs
}

// 获取用户操作日志，支持按时间段、操作类型、按用户名筛选
func GetUserOperationLogs(c *gin.Context) {
	usernameInterface, exists := c.Get("username")
	if !exists {
		log.Print("获取用户信息失败")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "获取用户信息失败"})
		return
	}
	username := usernameInterface.(string)

	var logRequest LogRequest
	if err := c.ShouldBind(&logRequest); err != nil {
		log.Printf("解析请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("解析请求失败: %v", err)})
		return
	}

	file, err := os.OpenFile("./logs/user_operation_log/user_operation_log.log", os.O_RDONLY, 0644)
	if err != nil {
		log.Printf("打开日志文件失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "打开日志文件失败"})
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var logs []Log
	if username == "root" { // 管理员
		if logRequest.Username == "" {
			logs = FilterLogs(scanner, logRequest, "")
		} else {
			logs = FilterLogs(scanner, logRequest, logRequest.Username)
		}
	} else { // 非管理员
		logs = FilterLogs(scanner, logRequest, username)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("读取日志文件失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "读取日志文件失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// --------------------------------------------------------------------------------
// GetRelativePath 获取调用者的相对路径和行号
func GetRelativePath() (file string, line int) {
	_, filepath, line, _ := runtime.Caller(0)
	i := strings.Index(filepath, "/server/")
	if i != -1 {
		filepath = "/" + filepath[i+len("/server/"):] // 加1是为了跳过"/"
	} else {
		filepath = "" // 或者其他默认值/错误处理
	}
	return filepath, line
}

func GetLogPrefix(skip int) (logPrefix string) {
	_, filepath, line, _ := runtime.Caller(skip)
	i := strings.Index(filepath, "/server/")
	if i != -1 {
		filepath = "/" + filepath[i+len("/server/"):]
	} else {
		filepath = "" // 或者其他默认值/错误处理
	}
	return fmt.Sprintf("%s %d: ", filepath, line)
}
