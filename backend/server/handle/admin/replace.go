package admin

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"serverM/server/logs"
	m_init "serverM/server/model/init"
	u "serverM/server/model/user"
)

type RepalceRequest struct {
	Realname string `json:"realname"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// 更换公司管理员
func ReplaceAdmin(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		log.Printf("未找到用户信息")
		c.JSON(401, gin.H{
			"message": "未找到用户信息",
		})
		return
	}
	Username := username.(string)

	var input RepalceRequest
	if err := c.BindJSON(&input); err != nil {
		log.Printf("请求数据格式错误")
		logs.Sugar.Errorw("更换公司管理员", "username", username, "detail", "解析请求失败，请检查请求格式是否正确")
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}

	var detail = fmt.Sprintf("新管理员实名:%s,新管理员用户名:%s,新管理员邮箱:%s", input.Realname, input.Username, input.Email)

	//查看当前管理员信息
	var oldAdmin u.User
	if err := m_init.DB.Where("name =?", Username).First(&oldAdmin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("管理员不存在")
			logs.Sugar.Errorw("更换公司管理员", "username", username, "detail", "管理员不存在"+detail)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "管理员不存在"})
			return
		}
		log.Println("数据库查询管理员失败")
		logs.Sugar.Errorw("更换公司管理员", "username", username, "detail", "数据库查询管理员失败"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库查询管理员失败"})
		return
	}
	if oldAdmin.RoleId == 0 {
		log.Println("没有更换管理员权限")
		logs.Sugar.Errorw("更换公司管理员", "username", username, "detail", "没有更换管理员权限"+detail)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "没有更换管理员权限"})
		return
	}

	//查找新管理员信息
	var user u.User
	if err := m_init.DB.Where("name =?", input.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("新管理员不存在")
			logs.Sugar.Errorw("更换公司管理员", "username", username, "detail", "新管理员不存在"+detail)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "新管理员不存在"})
			return
		}
		log.Println("数据库查询管理员失败")
		logs.Sugar.Errorw("更换公司管理员", "username", username, "detail", "数据库查询管理员失败"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库查询管理员失败"})
		return
	}
	if user.Realname == "" {
		log.Println("新管理员未实名,请在个人信息中实名")
		logs.Sugar.Errorw("更换公司管理员", "username", username, "detail", "新管理员未实名,请在个人信息中实名"+detail)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "新管理员未实名,请在个人信息中实名"})
		return
	}
	if user.Email != input.Email {
		log.Println("新管理员邮箱不匹配")
		logs.Sugar.Errorw("更换公司管理员", "username", username, "detail", "新管理员邮箱不匹配"+detail)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "新管理员邮箱不匹配"})
		return
	}
	// if user.CompanyId != oldAdmin.CompanyId {
	// 	log.Println("新管理员和你不在一个公司")
	// 	logs.Sugar.Errorw("更换公司管理员", "username", username, "detail", "新管理员和你不在一个公司" + detail)
	// 	c.JSON(http.StatusUnauthorized, gin.H{"message": "新管理员和你不在一个公司"})
	// 	return
	// }

	//获取公司名
	var company u.Company
	if err := m_init.DB.Where("id =?", oldAdmin.CompanyId).First(&company).Error; err != nil {
		log.Println("数据库查询公司失败")
		logs.Sugar.Errorw("更换公司管理员", "username", username, "detail", "数据库查询公司失败"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库查询公司失败"})
		return
	}

	//发出更换管理员申请
	content := Username + "更换" + company.Name + "公司管理,新管理员姓名:" + input.Realname + ",新管理员用户名:" +
		input.Username + ",新管理员邮箱:" + input.Email

	const layout = "2006-01-02 15:04:05.000000"
	createAt := time.Now().Format(layout)

	notice := u.Notice{
		Content:  content,
		Send:     Username,
		Receive:  user.Name,
		State:    "unprocessed",
		CreateAt: createAt,
	}
	if err := m_init.DB.Create(&notice).Error; err != nil {
		log.Println("数据库插入申请失败")
		logs.Sugar.Errorw("更换公司管理员", "username", username, "detail", "数据库插入更换管理员申请失败"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库插入更换管理员申请失败"})
		return
	}

	logs.Sugar.Infow("更换公司管理员", "username", username, "detail", "成功发出更换管理员申请。"+detail)
	c.JSON(http.StatusOK, gin.H{
		"message": "发出更换管理员申请",
	})
}
