package threshold

import (
	"context"
	"fmt"
	"net/http"
	"serverM/server/model"
	"serverM/server/redis"

	"github.com/gin-gonic/gin"
)

type ThresholdReuqest struct {
	Hostname     string  `json:"hostname"`
	CPUThreshold float64 `json:"cpu_threshold"`
	MemThreshold float64 `json:"mem_threshold"`
}

func UpdateThreshold(c *gin.Context) {
	var request ThresholdReuqest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数解析错误" + err.Error()})
		return
	}

	if request.CPUThreshold < 0 || request.CPUThreshold > 100 || request.MemThreshold < 0 || request.MemThreshold > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "CPU 和内存阈值必须在 0 到 100 之间"})
		return
	}

	cpuThreshold := request.CPUThreshold / 100.0
	memThreshold := request.MemThreshold / 100.0

	// 更新 Redis 中的阈值
	memKey := "mem_threshold:" + request.Hostname
	cpuKey := "cpu_threshold:" + request.Hostname
	fmt.Println(cpuThreshold, " ", memThreshold)
	ctx := context.Background()
	if err := redis.Rdb.Set(ctx, cpuKey, cpuThreshold, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "更新CPU阈值失败",
			"error":   "Failed to update CPU threshold in Redis",
		})
		return
	}
	if err := redis.Rdb.Set(ctx, memKey, memThreshold, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "更新内存阈值失败",
			"error":   "Failed to update memory threshold in Redis",
		})
		return
	}

	// 更新pg数据库中的阈值
	tx, err := model.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			fmt.Printf("UpdateThresholds: recovered from panic: %v, transaction rolled back", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}()

	query := `
		UPDATE host_info 
		SET cpu_threshold = $1, mem_threshold = $2 
		WHERE host_name = $3`
	result, err := tx.Exec(query, cpuThreshold, memThreshold, request.Hostname)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update thresholds in database",
		})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "没有找到对应的hostname，请确认hostname是否正确",
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// 成功响应
	c.JSON(http.StatusOK, gin.H{
		"message":       "Thresholds updated successfully",
		"hostname":      request.Hostname,
		"cpu_threshold": cpuThreshold,
		"mem_threshold": memThreshold,
	})
}
