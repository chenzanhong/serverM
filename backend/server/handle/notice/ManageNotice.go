package notice

import (
	"fmt"
	"serverM/server/logs"
	model "serverM/server/model"
	m_init "serverM/server/model/init"
	m_user "serverM/server/model/user"

	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func ManageNotice(c *gin.Context) {
	Username, exists := c.Get("username")
	if !exists {
		log.Printf("用户还未登录")
		c.JSON(401, gin.H{
			"message": "用户未登录",
		})
		return
	}
	username := Username.(string)

	var requestBody m_user.Notice
	err := c.BindJSON(&requestBody)
	if err != nil {
		log.Printf("请求数据格式错误")
		logs.Sugar.Errorw("通知处理", "username", username, "detail", "解析请求失败，请检查请求格式是否正确")
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}

	// 根据通知id查找对应的通知
	var notice m_user.Notice
	err = m_init.DB.Where("id = ?", requestBody.ID).First(&notice).Error
	if err != nil {
		log.Println("通知不存在")
		c.JSON(http.StatusBadRequest, gin.H{"message": "通知不存在"})
		return
	}

	var detail = fmt.Sprintf("发送者:%s,接收人:%s,内容:%s", notice.Send, notice.Receive, notice.Content)

	if notice.State == "processed" || notice.State == "expired" {
		log.Println("通知已处理或已过期")
		logs.Sugar.Errorw("通知处理", "username", username, "detail", "通知已处理或已过期。"+detail)
		c.JSON(http.StatusBadRequest, gin.H{"message": "通知已处理或已过期"})
		return
	}

	if notice.Receive != username {
		log.Println("处理的用户不是接收人")
		logs.Sugar.Errorw("通知处理", "username", username, "detail", "处理的用户不是接收人。"+detail)
		c.JSON(http.StatusBadRequest, gin.H{"message": "处理的用户不是接收人"})
		return
	}

	// 检查通知是否已经过期(通知的有效期限是14天)
	createAtTime, err := time.Parse(time.RFC3339, notice.CreateAt)
	if err != nil {
		log.Println("时间格式错误")
		logs.Sugar.Errorw("通知处理", "username", username, "detail", "时间格式错误。"+detail)
		c.JSON(http.StatusBadRequest, gin.H{"message": "时间格式错误"})
		return
	}

	if time.Since(createAtTime) > 14*24*time.Hour {
		update := `UPDATE notices SET state = 'expired' WHERE id = $1`
		_, err := model.DB.Exec(update, requestBody.ID)
		if err != nil {
			log.Println("更新通知状态失败:", err)
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "更新通知状态失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "更新通知状态失败"})
			return
		}
		logs.Sugar.Errorw("通知处理", "username", username, "detail", "通知已过期。"+detail)
		c.JSON(http.StatusBadRequest, gin.H{"message": "通知已过期"})
		return
	}

	// 查询权限
	var role_id int
	query := "SELECT role_id FROM users WHERE name = $1"
	err = model.DB.QueryRow(query, username).Scan(&role_id)
	if err != nil {
		log.Println("查询用户权限失败:", err)
		logs.Sugar.Errorw("通知处理", "username", username, "detail", "查询用户权限失败。"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询用户权限失败"})
		return
	}

	log.Printf("开始处理通知内容: %s", notice.Content)

	// 处理更换管理员的通知
	if strings.Contains(notice.Content, "更换") && strings.Contains(notice.Content, "公司管理,新管理员姓名:") {
		parts := strings.Split(notice.Content, "更换")
		if len(parts) < 2 {
			log.Println("通知内容格式错误：缺少 '更换' 分隔符")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "通知内容格式错误。"+detail)
			c.JSON(http.StatusBadRequest, gin.H{"message": "通知内容格式错误"})
			return
		}
		old_admin := strings.TrimSpace(parts[0])
		log.Println(old_admin)

		parts = strings.Split(parts[1], ",新管理员用户名:")
		if len(parts) < 2 {
			log.Println("通知内容格式错误：缺少 ',新管理员用户名:' 分隔符")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "通知内容格式错误。"+detail)
			c.JSON(http.StatusBadRequest, gin.H{"message": "通知内容格式错误"})
			return
		}
		parts = strings.Split(parts[1], ",新管理员邮箱:")
		if len(parts) < 2 {
			log.Println("通知内容格式错误：缺少 ',新管理员邮箱:' 分隔符")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "通知内容格式错误。"+detail)
			c.JSON(http.StatusBadRequest, gin.H{"message": "通知内容格式错误"})
		}
		new_admin := strings.TrimSpace(parts[0])
		log.Println(new_admin)

		// 获取原管理员公司编号
		var old_admin_company_id int
		query = "select company_id from users where name = $1"
		if err := m_init.DB.Raw(query, old_admin).Scan(&old_admin_company_id).Error; err != nil {
			log.Println("获取原管理员公司编号失败")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "获取原管理员公司编号失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "获取原管理员公司编号失败"})
			return
		}

		// 获取公司新管理员编号
		var new_admin_id int
		query = "select id from users where name = $1"
		if err := m_init.DB.Raw(query, new_admin).Scan(&new_admin_id).Error; err != nil {
			log.Println("获取新管理员编号失败")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "获取新管理员编号失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "获取新管理员编号失败"})
			return
		}

		var update string
		// 更换老公司管理员权限变为0——普通用户
		if old_admin != "root" {
			update = "update users set role_id = 0 where name = $1"
			if err := m_init.DB.Exec(update, old_admin).Error; err != nil {
				log.Println("更新旧管理员权限失败")
				logs.Sugar.Errorw("通知处理", "username", username, "detail", "更新旧管理员权限失败。"+detail)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "更新旧管理员权限失败"})
				return
			}
		}

		// 将老管理员所有服务器的company_id字段置为0，不再属于公司，恢复为属于个人
		update = "update host_info set company_id = 0 where user_name = $1"
		if err := m_init.DB.Exec(update, old_admin).Error; err != nil {
			log.Println("更新旧管理员所有服务器的归属失败")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "更新旧管理员所有服务器的归属失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "更新旧管理员所属公司失败"})
			return
		}

		// 更换新公司管理员权限变为1——公司管理员
		if new_admin != "root" {
			update = "update users set role_id = 1 where name = $1"
			if err := m_init.DB.Exec(update, new_admin).Error; err != nil {
				log.Println("更新新管理员权限失败")
				logs.Sugar.Errorw("通知处理", "username", username, "detail", "更新新管理员权限失败。"+detail)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "更新新管理员权限失败"})
				return
			}
		}

		// 将新管理员的所有服务器的company_id字段改为当前公司
		update = "update host_info set company_id = $1 where user_name = $2"
		err := m_init.DB.Exec(update, old_admin_company_id, new_admin).Error
		if err != nil {
			log.Println("更新新管理员所有服务器的归属失败")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "更新新管理员所有服务器的归属失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "更新新管理员所有服务器的归属失败"})
			return
		}

		// 更新新管理员的company_id字段
		update = "update users set company_id = $1 where name = $2"
		err = m_init.DB.Exec(update, old_admin_company_id, new_admin).Error
		if err != nil {
			log.Println("更新新管理员所属公司失败")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "更新新管理员所属公司失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "更新新管理员所属公司失败"})
			return
		}

		// 更换公司的管理员
		update = "update companies set admin_id = $1 where id = $2"
		if err := m_init.DB.Exec(update, new_admin_id, old_admin_company_id).Error; err != nil {
			log.Println("更新公司管理员失败")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "更新公司管理员失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "更新公司管理员id失败"})
			return
		}

		// 更新通知的状态为已处理
		update = "update notices set state = $1 where id = $2"
		if err := m_init.DB.Exec(update, "processed", notice.ID).Error; err != nil {
			log.Println("更新通知状态失败")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "更新通知状态失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "更新通知状态失败"})
			return
		}

		//发通知提醒root已经更换管理员
		const layout = "2006-01-02 15:04:05.000000"
		createAt := time.Now().Format(layout)
		notices := m_user.Notice{
			Send:     old_admin,
			Receive:  "root",
			Content:  notice.Content,
			State:    "processed",
			CreateAt: createAt,
		}
		if err := m_init.DB.Create(&notices).Error; err != nil {
			log.Println("数据库插入更换管理员通知失败")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "数据库插入更换管理员通知失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库插入更换管理员通知失败"})
			return
		}

		logs.Sugar.Infow("通知处理", "username", username, "detail", "通知处理成功。"+detail)
		c.JSON(http.StatusOK, gin.H{"message": "通知处理成功"})
		return
	}

	// 处理邀请加入公司的通知
	if strings.Contains(notice.Content, "邀请") && strings.Contains(notice.Content, "加入") {
		parts := strings.Split(notice.Content, "加入")
		if len(parts) < 2 || parts[1] == "" {
			log.Println("通知内容格式错误：缺少 '加入' 后的内容")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "通知内容格式错误。"+detail)
			c.JSON(http.StatusBadRequest, gin.H{"message": "通知内容格式错误"})
			return
		}
		companyName := strings.TrimSpace(parts[1])

		// 查找该公司的id
		var companyId int
		query := "SELECT id FROM companies WHERE name = $1"
		err := m_init.DB.Raw(query, companyName).Scan(&companyId).Error
		if err != nil {
			log.Println("数据库查询公司失败")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "数据库查询公司失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库查询公司失败"})
			return
		}

		var user m_user.User
		err = m_init.DB.Where("name = ?", username).First(&user).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "处理邀请成员通知，查询用户失败"})
			return
		}

		// 如果新成员之前不属于当前公司，更新当前公司成员数量，+1
		update := "UPDATE companies SET memberNum = memberNum + 1 WHERE id = $1"
		err = m_init.DB.Exec(update, companyId).Error
		if err != nil {
			log.Println("数据库更新公司成员数量失败")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "数据库更新公司成员数量失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库更新公司成员数量失败"})
			return
		}

		// 如果新成员之前属于其他某个公司A，则A的人数-1
		if user.CompanyId != 0 && user.CompanyId != companyId {
			err = m_init.DB.Exec("UPDATE companies SET memberNum = memberNum - 1 WHERE id = $1", user.CompanyId).Error
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库更新公司成员数量失败"})
				return
			}
		}

		// 如果新成员之前是某一个公司的管理员，则把原来公司的管理员置为0，新成员的所有服务器的company_id置为0
		if user.RoleId != 0 {
			update = "UPDATE companies SET admin_id = 0 WHERE id = $1"
			err = m_init.DB.Exec(update, user.CompanyId).Error
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库更新公司管理员失败"})
				return
			}
			update = "UPDATE host_info SET company_id = 0 WHERE user_name = $1"
			err = m_init.DB.Exec(update, username).Error
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库更新用户所属公司失败"})
				return
			}
		}

		// 更改新成员公司所属
		update = "UPDATE users SET company_id = $1 WHERE name = $2"
		err = m_init.DB.Exec(update, companyId, username).Error
		if err != nil {
			log.Println("数据库更新用户所属公司失败")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "数据库更新用户所属公司失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库更新用户所属公司失败"})
			return
		}

		// 更新通知的状态
		update = "UPDATE notices SET state = $1 WHERE id = $2"
		err = m_init.DB.Exec(update, "processed", notice.ID).Error
		if err != nil {
			log.Println("数据库更新通知状态失败")
			logs.Sugar.Errorw("通知处理", "username", username, "detail", "数据库更新通知状态失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库更新通知状态失败"})
			return
		}

		logs.Sugar.Infow("通知处理", "username", username, "detail", "通知处理成功。"+detail)
		c.JSON(http.StatusOK, gin.H{"message": "通知处理成功"})
	}
}
