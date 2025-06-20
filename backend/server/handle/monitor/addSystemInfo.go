package monitor

import (
	"serverM/server/handle/email"
	"serverM/server/model"
	m_init "serverM/server/model/init"
	"serverM/server/redis"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
)

type AlertTask struct {
	RequestData RequestData
}

const (
	MaxWorkers    = 10   // 根据机器 CPU 和 IO 能力调整
	TaskQueueSize = 1000 // 队列长度，防止突发流量压垮系统
)

var (
	TaskQueue   chan AlertTask
	initialized bool
	mutexMap    sync.Map // 确保写操作互斥，避免同时多个handleAlert而导致发送了多个邮件
)

func StartWorkerPool(maxWorkers int, queueSize int) error {
	if initialized {
		return fmt.Errorf("worker pool already started")
	}

	TaskQueue = make(chan AlertTask, queueSize)
	for i := 0; i < maxWorkers; i++ {
		go worker()
	}
	initialized = true
	return nil
}

func worker() {
	for task := range TaskQueue {
		handleAlert(task.RequestData)
	}
}

// RequestData 用于接收系统监控数据的请求体
// @Description RequestData 包含所有需要收集的系统信息
type RequestData struct {
	CPUInfo  []model.CPUInfo     `json:"cpu_info"`  // CPU 信息
	HostInfo model.HostInfo      `json:"host_info"` // 主机信息
	MemInfo  model.MemoryInfo    `json:"mem_info"`  // 内存信息
	ProInfo  []model.ProcessInfo `json:"pro_info"`  // 进程信息
	NetInfo  []model.NetworkInfo `json:"net_info"`  // 网络信息
}

func sendEmailNotification(username, hostname, alertMessages string, cpu_percent, mem_percent float64) {
	var userEmail string
	err := m_init.DB.Raw("SELECT email FROM users WHERE name = ?", username).Scan(&userEmail).Error
	if err != nil {
		log.Printf("预警，查询用户邮箱失败: %s", err)
	} else if userEmail != "" {
		// 发送邮件通知
		subject := fmt.Sprintf("系统告警通知 - %s", hostname)
		message := fmt.Sprintf(`
				<h2>系统告警通知</h2>
				<p>主机名: %s</p>
				<p>告警类型: %s</p>
				<p>CPU使用率: %.2f%%</p>
				<p>内存使用率: %.2f%%</p>
				<p>时间: %s</p>
			`, hostname, alertMessages, cpu_percent, mem_percent, time.Now().Format("2006-01-02 15:04:05"))

		err = email.SendEmail(userEmail, subject, message)
		if err != nil {
			log.Printf("预警，发送邮件通知失败: %s", err)
		}
	}
}

// shouldAlert 判断指定 hostname 是否满足告警的冷却时间要求
func ShouldAlert(hostname, warningType string, warning_time time.Time) bool {
	var latestTime time.Time
	err := m_init.DB.Raw("SELECT last_time FROM warning_states WHERE host_name = ? AND is_warning = TRUE LIMIT 1", hostname).Scan(&latestTime).Error

	if err != nil {
		log.Println("shouldAlert 查询last_time失败，不预警")
		return false // Scan失败或查询失败
	}

	if latestTime.IsZero() { // 查询无记录
		fmt.Println("last_time 是 NULL，视为无记录，要预警")
		return true 
	}

	coolDownPeriod := 10 * time.Minute
	timeInterval := warning_time.Sub(latestTime)
	exceedsCoolDown := timeInterval >= coolDownPeriod

	fmt.Printf("是否超过冷静期（%v）: %v\n", coolDownPeriod, exceedsCoolDown)

	if !exceedsCoolDown {
		// 未超过冷静期
		// 判断预警类型是否改变
		var oldWarningType string
		err = m_init.DB.Raw("SELECT warning_type FROM warning_states WHERE host_name = ? LIMIT 1", hostname).Scan(&oldWarningType).Error
		if err != nil {
			log.Printf("预警，查询旧预警类型失败: %s", err)
			return false
		}
		// 如果预警类型改变且时间间隔大于半个冷静期，还是要记录新预警记录，并更新预警时间
		if oldWarningType != warningType && (timeInterval > coolDownPeriod / 2){
			return true
		}
		return false
	}

	// 超过冷静期 
	return true
}

func handleAlert(requestData RequestData) {
	ctx := context.Background()
	hostname := requestData.HostInfo.Hostname
	fmt.Println("hostname:", hostname)
	// 从 Redis 读取阈值
	memKey := fmt.Sprintf("mem_threshold:%s", hostname)
	cpuKey := fmt.Sprintf("cpu_threshold:%s", hostname)
	memThreshold, err := redis.Rdb.Get(ctx, memKey).Float64()
	if err != nil {
		log.Printf("预警，获取内存阈值失败: %s", err)
		memThreshold = 0.9 // 设置为默认值
	}
	cpuThreshold, err := redis.Rdb.Get(ctx, cpuKey).Float64()
	if err != nil {
		log.Printf("预警，获取 CPU 阈值失败: %s", err)
		cpuThreshold = 0.9
	}

	warningType := ""
	alertMessages := ""

	// 判断类型
	cpuAlert := false
	memAlert := false
	if requestData.MemInfo.UserPercent > memThreshold*100 {
		memAlert = true
	}
	// 计算所有 CPU 核心的平均使用率
	var totalCPUPercent float64
	var coreCount int
	for _, data := range requestData.CPUInfo {
		totalCPUPercent += data.Percent
		coreCount++
	}

	var avgCPUPercent float64
	if coreCount > 0 {
		avgCPUPercent = totalCPUPercent / float64(coreCount)
	}
	// fmt.Println(requestData.MemInfo.UserPercent, " ", memThreshold, " ", avgCPUPercent, " ", cpuThreshold)

	// 判断平均是否超阈值
	cpuAlert = avgCPUPercent > cpuThreshold*100

	if cpuAlert && memAlert {
		warningType = "CPU与内存"
		alertMessages = "CPU与内存告警"
	} else if cpuAlert {
		warningType = "CPU"
		alertMessages = "CPU告警"
	} else if memAlert {
		warningType = "内存"
		alertMessages = "内存告警"
	}
	// fmt.Println("handleAlert:", warningType)
	// 如果有告警信息，存储到数据库并发送邮件通知
	if warningType != "" {
		// 尝试获取锁
		var mu *sync.Mutex
		if v, ok := mutexMap.Load(hostname); ok {
			mu = v.(*sync.Mutex)
		} else {
			newMu := &sync.Mutex{}
			v, loaded := mutexMap.LoadOrStore(hostname, newMu)
			if loaded {
				// 如果有其他goroutine 已经插入了这个锁，则使用已存在的锁
				mu = v.(*sync.Mutex)
			} else {
				mu = newMu
			}
		}

		mu.Lock() // 获取锁，确保对预警状态表的互斥访问
		if !ShouldAlert(hostname, warningType, requestData.HostInfo.CreatedAt.UTC()) {
			// 不需要预警
			return
		}
		// 需要预警
		// 更新warning_states中的记录
		err = m_init.DB.Exec(`
			UPDATE warning_states
			SET last_time = $1, warning_type = $2, is_warning = TRUE
			WHERE host_name = $3`,
			requestData.HostInfo.CreatedAt.UTC(), warningType, hostname).Error
		if err != nil {
			log.Printf("预警，更新预警状态失败: %s", err)
			// return false
			// 不return false，因为要确保成功预警，避免遗漏预警，但如果持续的更新失败，会导致很多没必要的重复预警
		}
		mu.Unlock()
		mutexMap.Delete(hostname) // 解锁后删除锁

		// 查询用户名，这里直接查询单个字段而非整个结构体
		var username string
		err := m_init.DB.Table("host_info").Select("user_name").Where("host_name = ?", hostname).Scan(&username).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("预警，未找到主机 %s 的信息", hostname)
				return // 或者采取其他措施，比如记录错误日志后返回
			} else {
				log.Printf("预警，查询主机信息失败: %s", err)
				return // 同样可以考虑记录错误并决定是否继续执行
			}
		}

		// 存储告警信息到数据库
		alertContent := fmt.Sprintf("主机 %s 发生 %s，CPU使用率: %.2f%%，内存使用率: %.2f%%",
			hostname, alertMessages, requestData.CPUInfo[0].Percent, requestData.MemInfo.UserPercent)

		err = m_init.DB.Exec(`
			INSERT INTO warnings (host_name, username, warning_type, warning_title, warning_time)
			VALUES (?, ?, ?, ?, ?)`,
			hostname, username, warningType, alertContent, requestData.HostInfo.CreatedAt.UTC()).Error

		if err != nil {
			log.Printf("存储告警信息失败: %s", err)
		}
		// 查询用户邮箱
		go sendEmailNotification(username, hostname, alertMessages, requestData.CPUInfo[0].Percent, requestData.MemInfo.UserPercent)
	}
}

func ReceiveAndStoreSystemMetrics(c *gin.Context) {
	// 解析请求数据
	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		s := fmt.Sprintf("Invalid JSON data: %s", err)
		log.Printf("Invalid JSON data: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": s})
		return
	}

	// 将数据插入 Redis
	ctx := context.Background()
	timestamp := time.Now().Unix() // 获取当前时间戳
	key := fmt.Sprintf("system_info:%s:%d", requestData.HostInfo.Hostname, timestamp)
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		log.Printf(" Failed to marshal data to JSON: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal data to JSON"})
		return
	}

	// 将 JSON 字符串存储到 Redis
	err = redis.Rdb.Set(ctx, key, jsonData, 30*time.Minute).Err()
	if err != nil {
		log.Printf("Failed to insert data into Redis: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data into Redis"})
		return
	}

	// 入队而不是直接开 goroutine
	select {
	case TaskQueue <- AlertTask{RequestData: requestData}:
		// 成功入队
		// fmt.Println("成功入队")
	default:
		// 任务队列已满，启动单独的协程进行处理
		fmt.Println("管道已满")
		go handleAlert(requestData)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "System information inserted successfully"})
}
