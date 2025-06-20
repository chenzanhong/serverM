package monitor

import (
	m_init "serverM/server/model/init"
	u "serverM/server/model/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetWarningRecordsByHostname 获取指定 hostname 的所有告警记录
func GetWarningRecordsByHostname(c *gin.Context) {
	Username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username not found in context",
		})
		return
	}
	username := Username.(string)

	var records []u.Warning
	result := m_init.DB.Where("username = ?", username).Order("warning_time DESC").Find(&records)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询告警记录失败：" + result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": records,
	})
}
