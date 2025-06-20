package update

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	e "serverM/server/handle/email"
	"serverM/server/logs"
	m_init "serverM/server/model/init"
	u "serverM/server/model/user"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// 修改真实姓名、密码、邮箱
func UpdateUserInfo(c *gin.Context) {
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

	// 解析请求体	前端可只传递要修改的字段
	var request struct {
		// NewName     string `json:"new_name"`
		NewPassword string `json:"new_password"`
		Email       string `json:"new_email"`
		RealName    string `json:"realname"`
	}
	if err := c.BindJSON(&request); err != nil {
		log.Printf("解析请求数据失败: %v", err)
		logs.Sugar.Errorw("修改个人信息", "username", username, "detail", "解析请求失败，请检查请求格式是否正确")
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误", "error": err.Error()})
		return
	}

	var detail = fmt.Sprintf("密码:%s, 邮箱:%s, 真实姓名:%s", request.NewPassword, request.Email, request.RealName)

	var user u.User
	if err := m_init.DB.Where("name =?", username).First(&user).Error; err != nil {
		log.Printf("未找到用户名")
		logs.Sugar.Errorw("修改个人信息", "username", username, "detail", "未找到用户名。"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取用户信息失败", "error": err.Error()})
		return
	}

	// 检查新密码是否为空
	if request.NewPassword != "" && request.NewPassword != user.Password {
		// 执行密码更新操作
		if err := m_init.DB.Model(&u.User{}).Where("name =?", username).Updates(map[string]interface{}{"password": request.NewPassword}).Error; err != nil {
			log.Printf("更新密码失败")
			logs.Sugar.Errorw("修改个人信息", "username", username, "detail", "更新密码失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "更新密码失败", "error": err.Error()})
			return
		}
	}

	// 检查新邮箱是否为空
	if request.Email != "" && request.Email != user.Email {
		// 执行邮箱更新操作
		if err := m_init.DB.Model(&u.User{}).Where("name =?", username).Updates(map[string]interface{}{"email": request.Email}).Error; err != nil {
			log.Printf("更新邮箱失败")
			logs.Sugar.Errorw("修改个人信息", "username", username, "detail", "更新邮箱失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "更新邮箱失败", "error": err.Error()})
			return
		}
	}

	// 检查是否传入真实姓名
	if request.RealName != "" && request.RealName != user.Realname {
		// 执行真实姓名更新操作
		if err := m_init.DB.Model(&u.User{}).Where("name =?", username).Updates(map[string]interface{}{"realname": request.RealName}).Error; err != nil {
			log.Printf("更新真实姓名失败")
			logs.Sugar.Errorw("修改个人信息", "username", username, "detail", "更新真实姓名失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "更新真实姓名失败", "error": err.Error()})
			return
		}
	}

	logs.Sugar.Infow("修改个人信息", "username", username, "detail", "修改个人信息成功。"+detail)
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// 密码找回：-----------------------------------
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRandomToken(length int) string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	token := make([]byte, length)
	for i := range token {
		token[i] = charset[r.Intn(len(charset))]
	}

	return string(token)
}

// 处理重置密码请求
func RequestResetPassword(c *gin.Context) {
	// 实现请求重置密码的逻辑
	var request struct {
		Email string `json:"email"`
	}

	if err := c.BindJSON(&request); err != nil {
		log.Printf("解析请求数据失败: %v", err)
		logs.Sugar.Errorw("重置密码请求", "detail", "解析请求失败，请检查请求格式是否正确")
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}

	var detail = fmt.Sprintf("邮箱:%s", request.Email)

	// 查找用户
	var user u.User
	err := m_init.DB.Where("email = ?", request.Email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("用户未找到")
			logs.Sugar.Errorw("重置密码请求", "detail", "用户未找到。"+detail)
			c.JSON(http.StatusNotFound, gin.H{"message": "用户未找到"})
		} else {
			log.Printf("数据库查询失败")
			logs.Sugar.Errorw("重置密码请求", "detail", "数据库查询失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库查询失败"})
		}
		return
	}

	// 生成唯一的重置密码 token
	// token := fmt.Sprintf("%d", time.Now().UnixNano())
	token := GenerateRandomToken(6) // 生成6位长度的token
	fmt.Println("密码找回时生成的token为：", token)
	// 在数据库中保存 token
	err = m_init.DB.Model(&user).Update("token", token).Error
	if err != nil {
		log.Printf("保存 token 失败")
		logs.Sugar.Errorw("重置密码请求", "detail", "保存 token 失败。"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "保存 token 失败"})
		return
	}

	// 发送重置密码邮件
	err = e.SendResetPasswordEmail(request.Email, token)
	if err != nil {
		log.Printf("发送重置密码邮件失败")
		logs.Sugar.Errorw("重置密码请求", "detail", "发送重置密码邮件失败。"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "发送重置密码邮件失败"})
		return
	}

	logs.Sugar.Infow("重置密码请求", "detail", "重置密码请求成功。"+detail)
	c.JSON(http.StatusOK, gin.H{
		"message": "重置密码请求成功",
	})
}

// 重置密码
func ResetPassword(c *gin.Context) {
	// 实现重置密码的逻辑
	var request struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BindJSON(&request); err != nil {
		logs.Sugar.Errorw("重置密码", "detail", "解析请求数据失败，请检查请求格式是否正确")
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}
	fmt.Println("The new password is : ", request.NewPassword, ", and the token is : ", request.Token)

	var detail = fmt.Sprintf("token:%s, 新密码:%s", request.Token, request.NewPassword)

	var user u.User
	err := m_init.DB.Where("token = ?", request.Token).First(&user).Error
	// username, _ := c.Get("username")
	// err := m_init.DB.Where("token = ? and name = ?", request.Token, username.(string)).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("无效的重置密码 token")
			logs.Sugar.Errorw("重置密码", "detail", "无效的重置密码 token。"+detail)
			c.JSON(http.StatusNotFound, gin.H{"message": "无效的重置密码 token"})
		} else {
			log.Printf("数据库查询失败")
			logs.Sugar.Errorw("重置密码", "detail", "数据库查询失败。"+detail)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库查询失败"})
		}
		return
	}

	err = m_init.DB.Model(&user).Update("password", request.NewPassword).Error
	if err != nil {
		log.Printf("密码重置失败")
		logs.Sugar.Errorw("重置密码", "detail", "密码重置失败。"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "密码重置失败"})
		return
	}

	err = m_init.DB.Model(&user).Update("token", nil).Error
	if err != nil {
		log.Printf("密码重置成功，但是 token 重置失败")
		logs.Sugar.Errorw("重置密码", "detail", "密码重置成功，但是 token 重置失败。"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "密码重置成功，但是 token 重置失败"})
		return
	}

	logs.Sugar.Infow("重置密码", "detail", "重置密码成功。"+detail)
	c.JSON(http.StatusOK, gin.H{
		"message": "重置密码成功",
	})
}

// 方式一 发送token
// func sendResetPasswordEmail(email, token string) error {
// 	myEmail := os.Getenv("EMAIL_NAME")
// 	myPassword := os.Getenv("EMAIL_PASSWORD")
// 	baseUrl := os.Getenv("BASE_URL")
// 	smtpServerHost := os.Getenv("SMTP_SERVER_HOST")
// 	smtpServerPortStr := os.Getenv("SMTP_SERVER_PORT")

// 	if myEmail == "" || myPassword == "" || smtpServerHost == "" || smtpServerPortStr == "" {
// 		log.Fatalf("环境变量未正确设置")
// 		return errors.New("环境变量未正确设置")
// 	}

// 	smtpServerPort, err := strconv.Atoi(smtpServerPortStr)
// 	if err != nil {
// 		log.Fatalf("将端口号转换为整数时出错: %v", err)
// 		return err
// 	}

// 	log.Printf("Email: %s, Password: %s, SMTP Server: %s, Port: %d, BaseUrl: %s", myEmail, myPassword, smtpServerHost, smtpServerPort, baseUrl)

// 	m := gomail.NewMessage()
// 	m.SetHeader("From", myEmail)
// 	m.SetHeader("To", email)
// 	m.SetHeader("Subject", "Password Reset Request")
// 	m.SetBody("text/html", fmt.Sprintf(`
// 		<h1>密码找回</h1>
// 		<p>这是你的验证码：%s</p>
// 	`, token))

// 	d := gomail.NewDialer(smtpServerHost, smtpServerPort, myEmail, myPassword)
// 	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // 跳过证书验证，生产环境中应谨慎使用
// 	if err := d.DialAndSend(m); err != nil {
// 		log.Printf("发送邮件失败: %v", err)
// 		if strings.Contains(err.Error(), "535") { // 例如，检查错误消息中是否包含 SMTP 身份验证失败的代码
// 			log.Printf("可能是 SMTP 身份验证错误")
// 			return errors.New("发送邮件失败可能是 SMTP 身份验证错误")
// 		} else if strings.Contains(err.Error(), "connection refused") {
// 			log.Printf("SMTP 服务器连接被拒绝")
// 			return errors.New("发送邮件失败：SMTP 服务器连接被拒绝")
// 		}
// 	} else {
// 		log.Println("邮件发送成功")
// 	}
// 	return nil
// }

// 发送链接
// func sendResetPasswordEmail(email, token string) {
// 	myEmail := os.Getenv("EMAIL_NAME")
// 	myPassword := os.Getenv("EMAIL_PASSWORD")
// 	baseUrl := os.Getenv("BASE_URL")
// 	smtpServerHost := os.Getenv("SMTP_SERVER_HOST")
// 	smtpServerPortStr := os.Getenv("SMTP_SERVER_PORT")

// 	if myEmail == "" || myPassword == "" || baseUrl == "" || smtpServerHost == "" || smtpServerPortStr == "" {
// 		log.Fatalf("环境变量未正确设置")
// 	}

// 	smtpServerPort, err := strconv.Atoi(smtpServerPortStr)
// 	if err != nil {
// 		log.Fatalf("将端口号转换为整数时出错: %v", err)
// 	}

// 	log.Printf("Email: %s, Password: %s, SMTP Server: %s, Port: %d, BaseUrl: %s", myEmail, myPassword, smtpServerHost, smtpServerPort, baseUrl)

// 	m := gomail.NewMessage()
// 	m.SetHeader("From", myEmail)
// 	m.SetHeader("To", email)
// 	m.SetHeader("Subject", "Password Reset Request")
// m.SetBody("text/html", fmt.Sprintf(`
// 	<h1>Password Reset</h1>
// 	<p>Click the link to reset your password: <a href="%s/static/reset_password.html?token=%s" >Reset Password</a></p>
// `, baseUrl, token))

// 	d := gomail.NewDialer(smtpServerHost, smtpServerPort, myEmail, myPassword)
// 	if err := d.DialAndSend(m); err != nil {
// 		log.Printf("发送邮件失败: %v", err)
// 		if strings.Contains(err.Error(), "535") { // 例如，检查错误消息中是否包含 SMTP 身份验证失败的代码
// 			log.Printf("可能是 SMTP 身份验证错误")
// 		} else if strings.Contains(err.Error(), "connection refused") {
// 			log.Printf("SMTP 服务器连接被拒绝")
// 		}
// 	} else {
// 		log.Println("邮件发送成功")
// 	}
// }
