package monitor

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"serverM/server/logs"
	m_init "serverM/server/model/init"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DeleteSystemInfoRequest struct {
	IP       string `json:"ip"`
	HostName string `json:"host_name"`
}

func DeleteSystemInfo(c *gin.Context) {
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

	// 解析 json body 到结构体 DeleteSystemInfoRequest
	var deleteSystemInfoRequest DeleteSystemInfoRequest
	err := c.BindJSON(&deleteSystemInfoRequest)
	if err != nil {
		log.Println("解析 json body 失败")
		logs.Sugar.Errorw("删除服务器", "username", username, "detail", "解析请求失败，请检查请求格式是否正确")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// 开始事务
	tx := m_init.DB.Begin()
	if tx.Error != nil {
		log.Println("开始事务失败")
		logs.Sugar.Errorw("删除服务器", "username", username, "detail", "开始事务失败")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "开始事务失败"})
		return
	}

	var detail = fmt.Sprintf("删除服务器,ip: %s, 主机名: %s", deleteSystemInfoRequest.IP, deleteSystemInfoRequest.HostName)

	var existingID int
	// 检查在 host_info 表是否存在对应的数据
	query := `
        SELECT id
        FROM host_info
        WHERE host_name = $1 AND ip = $2 AND user_name = $3
    `
	err = tx.Raw(query, deleteSystemInfoRequest.HostName, deleteSystemInfoRequest.IP, username).Scan(&existingID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("数据库没有相应的 host_info 数据")
			logs.Sugar.Errorw("删除服务器", "username", username, "detail", "该服务器不是你的，你完全删除")
			tx.Rollback() // 回滚事务
			c.JSON(http.StatusInternalServerError, gin.H{"message": "该服务器不是你的，你无权删除"})
			return
		}
		log.Printf("查询 host_info 失败")
		tx.Rollback() // 回滚事务
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询 host_info 失败"})
		return
	}

	if existingID > 0 {
		deleteQuery := `
            DELETE FROM host_info
            WHERE host_name = $1 AND ip = $2 AND user_name = $3
        `
		if err = tx.Exec(deleteQuery, deleteSystemInfoRequest.HostName, deleteSystemInfoRequest.IP, username).Error; err != nil {
			log.Printf("删除 host_info 数据失败")
			logs.Sugar.Errorw("删除服务器", "username", username, "detail", "数据库删除 host_info 数据失败。"+detail)
			tx.Rollback() // 回滚事务
			c.JSON(http.StatusInternalServerError, gin.H{"message": "删除 host_info 数据失败"})
			return
		}
	} else {
		log.Println("数据库没有相应的 host_info 数据")
		logs.Sugar.Errorw("删除服务器", "username", username, "detail", "不是你的服务器，不可删除。"+detail)
		tx.Rollback() // 回滚事务
		c.JSON(http.StatusBadRequest, gin.H{"message": "数据库没有相应的 host_info 数据"})
		return
	}

	// 查询该主机在 hostandtoken 表是否存在
	query = `
        SELECT id
        FROM hostandtoken
        WHERE host_name = $1
    `
	err = tx.Raw(query, deleteSystemInfoRequest.HostName).Scan(&existingID).Error
	if err != nil {
		log.Println("数据库查询失败")
		logs.Sugar.Errorw("删除服务器", "username", username, "detail", "数据库查询失败。"+detail)
		tx.Rollback() // 回滚事务
		c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库查询失败"})
		return
	}

	if existingID > 0 {
		deleteQuery := `
            DELETE FROM hostandtoken
            WHERE host_name = $1 
        `
		if err = tx.Exec(deleteQuery, deleteSystemInfoRequest.HostName).Error; err != nil {
			log.Println("删除 hostandtoken 数据失败")
			logs.Sugar.Errorw("删除服务器", "username", username, "detail", "数据库删除 hostandtoken 数据失败。"+detail)
			tx.Rollback() // 回滚事务
			c.JSON(http.StatusInternalServerError, gin.H{"message": "删除 hostandtoken 的数据失败"})
			return
		}
	} else {
		log.Println("数据库没有相应的 hostandtoken 数据")
		logs.Sugar.Errorw("删除服务器", "username", username, "detail", "数据库没有相应的 hostandtoken 数据。"+detail)
		// tx.Rollback() // 回滚事务
		// c.JSON(http.StatusBadRequest, gin.H{"message": "数据库没有相应的 hostandtoken 数据"})
		// return
	}

	// 更新 ssh_ports 表：设置 is_used = false 并但是不清空 hostname，获取删除代理服务脚本时要用
	updateSSHPortSQL := `
        UPDATE ssh_ports
        SET is_used = false
        WHERE hostname = $1
    `
	result := tx.Exec(updateSSHPortSQL, deleteSystemInfoRequest.HostName)
	if result.Error != nil {
		log.Printf("更新 ssh_ports 表失败: %v", result.Error)
		logs.Sugar.Errorw("删除服务器", "username", username, "detail", "更新 ssh_ports 表失败。"+detail)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新 ssh_ports 表失败"})
		return
	}

	// 可选：检查是否影响了记录
	if result.RowsAffected == 0 {
		log.Printf("未找到与 hostname=%s 关联的 ssh_ports 记录，跳过更新", deleteSystemInfoRequest.HostName)
		logs.Sugar.Errorw("删除服务器", "username", username, "detail", "未找到与 hostname="+deleteSystemInfoRequest.HostName+"关联的 ssh_ports 记录，跳过更新。"+detail)
		// 这里可以选择继续提交事务，或者返回警告
	}

	// 删除warning_states的记录
	deleteWarningStatesSQL := `
        DELETE FROM warning_states
        WHERE host_name = $1
    `
	if err = tx.Exec(deleteWarningStatesSQL, deleteSystemInfoRequest.HostName).Error; err != nil {
		log.Printf("删除 warning_states 数据失败")
		logs.Sugar.Errorw("删除服务器", "username", username, "detail", "数据库删除 warning_states 数据失败。"+detail)
		tx.Rollback() // 回滚事务
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除 warning_states 的数据失败"})
		return
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		log.Println("提交事务失败")
		logs.Sugar.Errorw("删除服务器", "username", username, "detail", "提交事务失败。"+detail)
		tx.Rollback() // 回滚事务
		c.JSON(http.StatusInternalServerError, gin.H{"message": "提交事务失败"})
		return
	}

	// 成功响应
	logs.Sugar.Infow("删除服务器", "username", username, "detail", "删除服务器成功。"+detail)
	c.JSON(http.StatusOK, gin.H{
		"message": "采集器删除成功",
	})
}
