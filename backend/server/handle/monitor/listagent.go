package monitor

import (
	"log"
	"net/http"
	"serverM/server/logs"
	m_init "serverM/server/model/init"
	u "serverM/server/model/user"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// ListAgent 用于查询所有主机信息
func HostInfoList(c *gin.Context) {
	// 从上下文中获取用户名
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

	// 解析时间查询参数
	from := c.Query("from")
	to := c.Query("to")

	if from == "" {
		from = "1970-01-01T00:00:00Z"
	}
	if to == "" {
		to = "9999-12-31T23:59:59Z"
	}

	fromTime, err := time.Parse(time.RFC3339, from)
	if err != nil {
		log.Println(logs.GetLogPrefix(2) + "无效的 from 时间格式")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 from 时间格式"})
		return
	}
	toTime, err := time.Parse(time.RFC3339, to)
	if err != nil {
		log.Println(logs.GetLogPrefix(2) + "无效的 to 时间格式")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 to 时间格式"})
		return
	}

	// 查询用户所在公司
	var user u.User
	err = m_init.DB.Where("name = ?", username).First(&user).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询用户的公司失败"})
		return
	}

	// 使用 GORM 查询
	var hosts []u.HostInfo
	if username == "root" { // 系统管理员可见所有服务器
		err = m_init.DB.Table("host_info").
			Where("created_at BETWEEN ? AND ?", fromTime, toTime).
			Order("created_at DESC"). // 可选排序
			Find(&hosts).Error

	} else { // 其他用户可见自己的服务器和公司服务器
		if user.CompanyId != 0 {
			err = m_init.DB.Table("host_info").
				Where("( user_name = ? OR company_id = ?) AND created_at BETWEEN ? AND ?", username, user.CompanyId, fromTime, toTime).
				Order("created_at DESC"). // 可选排序
				Find(&hosts).Error
		} else {
			err = m_init.DB.Table("host_info").
				Where("user_name = ? AND created_at BETWEEN ? AND ?", username, fromTime, toTime).
				Order("created_at DESC"). // 可选排序
				Find(&hosts).Error
		}
	}

	if err != nil {
		log.Printf("Failed to query host_info; details: %s", err.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to query host_info",
			"details": err.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hosts": hosts})
}
