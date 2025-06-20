package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"serverM/server/handle/monitor"
	"serverM/server/handle/notice"
	"serverM/server/handle/transfer/global"
	"serverM/server/model"
	m_init "serverM/server/model/init"
	"serverM/server/redis"
	"serverM/server/router"
	"syscall"
	"time"
)

func main() {
	r := router.SetupRouter()

	// 清理redis中的过期监控数据
	go monitor.CheckServerStatus()

	// 通知notice的过期判断和处理（默认30天未处理就过期）：每天2点执行一次判断和修改，确保过期通知的状态为“已过期”
	go notice.ExpireOldNotices()

	// 预警的消费，异步处理预警
	err := monitor.StartWorkerPool(100, 10000) // 100个消费者，管道长度10000
	if err != nil {
		log.Fatalf("Failed to start worker pool: %v", err)
	}

	// 创建 HTTP 服务实例
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 创建 context 监听系统信号
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 启动 HTTP 服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe error: %v", err)
		}
	}()

	log.Println("Server started on :8080")

	// 等待中断信号
	<-ctx.Done()

	log.Println("Shutting down server...")

	// 创建一个超时 context 控制优雅关闭时间
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. 停止 HTTP 服务
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println("HTTP server Shutdown error: %v", err)
	} else {
		log.Println("HTTP server gracefully stoped")
	}

	// 2. 关闭 Redis
	if err := redis.Close(); err != nil {
		log.Println("Redis close error: %v", err)
	} else {
		log.Println("Redis connection closed")
	}

	// 3. 关闭 pg 数据库连接 *sql.DB
	if model.DB != nil {
		if err := model.DB.Close(); err != nil {
			log.Println("PostgreSQL SQL DB Close error: %v", err)
		} else {
			log.Println("PostgreSQL SQL DB closed")
		}
	}

	// 4. 关闭 pg 数据库连接 *gorm.DB
	sqlDB, gormErr := m_init.DB.DB()
	if gormErr == nil {
		if err := sqlDB.Close(); err != nil {
			log.Println("PostgreSQL GORM DB Close error: %v", err)
		} else {
			log.Println("PostgreSQL GORM DB closed")
		}
	} else {
		log.Println("Failed to get underlying SQL DB from GORM")
	}

	// 5. 关闭 SSH 连接池
	if global.Pool != nil {
		global.Pool.Close()
		log.Println("SSH connection pool closed")
	}

	log.Println("Server exited")
}
