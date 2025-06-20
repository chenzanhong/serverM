package port

import (
	"errors"
	"net/http"
	m_init "serverM/server/model/init"
	u "serverM/server/model/user"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 获取可用端口的函数
func GetUnusedPort() (int, error) {
	var port u.SSHPort
	// 优先获取未被使用过的端口
	if err := m_init.DB.Where("is_used = false and hostname = ''").First(&port).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			// 尝试获取曾经被使用过，但是现在未被使用的port
			if newErr := m_init.DB.Where("is_used = false").First(&port).Error; newErr != nil {
				if newErr != gorm.ErrRecordNotFound {
					return -1, errors.New("没有可用端口了")
				}
			}
		}
		return -1, err
	}

	return port.Port, nil
}

// 添加服务器成功后，更新ssh_port的is_used为true
func UpdatePortInDB(port int) error {
	var sshPort u.SSHPort
	if err := m_init.DB.First(&sshPort, port).Error; err != nil {
		return err
	}
	sshPort.IsUsed = true
	sshPort.UpdatedAt = time.Now()
	return m_init.DB.Save(&sshPort).Error
}

// 释放ssh_port的端口
func releasePortInDB(port int) error {
	var sshPort u.SSHPort
	if err := m_init.DB.First(&sshPort, port).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("端口号不存在")
		} else {
			return errors.New("查询端口号失败：" + err.Error())
		}
	}

	sshPort.IsUsed = false
	sshPort.UpdatedAt = time.Now()
	err := m_init.DB.Save(&sshPort).Error
	if err != nil {
		return errors.New("释放端口号失败：" + err.Error())
	} else {
		return nil
	}
}

// 获取可用端口
func GetAvailablePort(c *gin.Context) {
	port, err := GetUnusedPort()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "端口号获取失败：" + err.Error(), "port": port})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "端口号获取成功", "port": port})
}

// 释放ssh_port的端口
func ReleasePort(c *gin.Context) {
	portStr := c.Query("port")
	if portStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "端口号不能为空"})
		return
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "端口号无效"})
		return
	}

	if err := releasePortInDB(port); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "释放端口" + portStr + "失败：" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "端口" + portStr + "释放成功"})
}
