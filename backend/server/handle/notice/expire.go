package notice

import (
	"log"
	m_init "serverM/server/model/init"
	"time"
)

func ExpireOldNotices() {
	// 获取当前时间并计算明天凌晨 2 点的时间
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if now.After(nextRun) {
		nextRun = nextRun.Add(24 * time.Hour)
	}

	// 计算第一次运行前等待的时间
	firstDelay := time.Until(nextRun)
	log.Printf("First run scheduled at: %v (in %v)", nextRun, firstDelay)

	time.Sleep(firstDelay)

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Running daily task to expire old notices...")

			// 使用 GORM 执行原生 SQL
			err := m_init.DB.Exec(`UPDATE notices SET state = 'expired' WHERE state = 'unprocessed' AND created_at < NOW() - INTERVAL '30 days'`).Error
			if err != nil {
				log.Printf("Error updating expired notices: %v", err)
			} else {
				log.Println("Expired notices updated successfully.")
			}

		}
	}
}
