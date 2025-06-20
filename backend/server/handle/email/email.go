package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

// 邮箱验证码存储结构
type TokenInfo struct {
	Token     string
	ExpiresAt time.Time
}

var (
	// 全局变量：邮箱 -> 验证码 + 过期时间
	EmailToken      = make(map[string]TokenInfo)
	EmailTokenMutex sync.Mutex // 并发安全锁
)

type EmailRequest struct {
	Email       string `json:"email" form:"email" binding:"required"`
	Message     string `json:"message" form:"message"`
	HTMLMessage string `json:"html_message" form:"html_message"`
	Subject     string `json:"subject" form:"subject"`
}

func IsValidEmail(email string) bool {
    // 正则表达式
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
    
    return emailRegex.MatchString(email)
}

// SendEmail 是一个通用的邮件发送函数
func SendEmail(toEmail, subject, message string) error {
	myEmail := os.Getenv("EMAIL_NAME")
	myPassword := os.Getenv("EMAIL_PASSWORD")
	smtpServerHost := os.Getenv("SMTP_SERVER_HOST")
	smtpServerPortStr := os.Getenv("SMTP_SERVER_PORT")

	if myEmail == "" || myPassword == "" || smtpServerHost == "" || smtpServerPortStr == "" {
		log.Fatalf("环境变量未正确设置")
		return errors.New("环境变量未正确设置")
	}

	smtpServerPort, err := strconv.Atoi(smtpServerPortStr)
	if err != nil {
		log.Fatalf("将端口号转换为整数时出错: %v", err)
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", myEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", message)

	d := gomail.NewDialer(smtpServerHost, smtpServerPort, myEmail, myPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // 跳过证书验证，生产环境中应谨慎使用

	if err := d.DialAndSend(m); err != nil {
		log.Printf("发送邮件失败: %v", err)
		if strings.Contains(err.Error(), "535") { // 检查错误消息中是否包含 SMTP 身份验证失败的代码
			log.Printf("可能是 SMTP 身份验证错误")
			return errors.New("发送邮件失败可能是 SMTP 身份验证错误")
		} else if strings.Contains(err.Error(), "connection refused") {
			log.Printf("SMTP 服务器连接被拒绝")
			return errors.New("发送邮件失败：SMTP 服务器连接被拒绝")
		}
		return err
	}
	log.Println("邮件发送成功")
	return nil
}

// 发送邮件
func SendEmailHandler(c *gin.Context) {
	var request EmailRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"success": false,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	subject := request.Subject
	if subject == "" {
		subject = "你有一条新消息"
	}

	message := request.HTMLMessage
	if message == "" {
		message = request.Message
	}

	err := SendEmail(request.Email, subject, message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"success": false,
			"message": "发送邮件失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"success": true,
		"message": "邮件发送成功",
	})
}

// 发送重置密码邮件，包含验证码
func SendResetPasswordEmail(email, token string) error {
	myEmail := os.Getenv("EMAIL_NAME")
	myPassword := os.Getenv("EMAIL_PASSWORD")
	baseUrl := os.Getenv("BASE_URL")
	smtpServerHost := os.Getenv("SMTP_SERVER_HOST")
	smtpServerPortStr := os.Getenv("SMTP_SERVER_PORT")

	if myEmail == "" || myPassword == "" || smtpServerHost == "" || smtpServerPortStr == "" {
		log.Fatalf("环境变量未正确设置")
		return errors.New("环境变量未正确设置")
	}

	smtpServerPort, err := strconv.Atoi(smtpServerPortStr)
	if err != nil {
		log.Fatalf("将端口号转换为整数时出错: %v", err)
		return err
	}

	log.Printf("Email: %s, Password: %s, SMTP Server: %s, Port: %d, BaseUrl: %s", myEmail, myPassword, smtpServerHost, smtpServerPort, baseUrl)

	m := gomail.NewMessage()
	m.SetHeader("From", myEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Password Reset Request")
	m.SetBody("text/html", fmt.Sprintf(`
		<h1>密码找回</h1>
		<p>这是你的验证码：%s</p>
	`, token))

	d := gomail.NewDialer(smtpServerHost, smtpServerPort, myEmail, myPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // 跳过证书验证，生产环境中应谨慎使用
	if err := d.DialAndSend(m); err != nil {
		log.Printf("发送邮件失败: %v", err)
		if strings.Contains(err.Error(), "535") { // 例如，检查错误消息中是否包含 SMTP 身份验证失败的代码
			log.Printf("可能是 SMTP 身份验证错误")
			return errors.New("发送邮件失败可能是 SMTP 身份验证错误")
		} else if strings.Contains(err.Error(), "connection refused") {
			log.Printf("SMTP 服务器连接被拒绝")
			return errors.New("发送邮件失败：SMTP 服务器连接被拒绝")
		}
	} else {
		log.Println("邮件发送成功")
	}
	return nil
}

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

// 发送验证码
func SendVerificationCode(c *gin.Context) {
	var request struct {
		Email string `json:"email"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

    // 去除前后空格
	email := strings.TrimSpace(request.Email)
	if !IsValidEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "邮箱格式不正确"})
		return
	}

	// 生成6位随机token
	token := GenerateRandomToken(6)

	// 设置过期时间（例如5分钟）
	expireTime := time.Now().Add(5 * time.Minute)

	// 写入全局变量（带锁）
	EmailTokenMutex.Lock()
	EmailToken[email] = TokenInfo{
		Token:     token,
		ExpiresAt: expireTime,
	}
	EmailTokenMutex.Unlock()

	// 构造邮件内容
	subject := "您的注册验证码"
	message := fmt.Sprintf("<h3>您的验证码是：<strong>%s</strong></h3><p>请在5分钟内完成注册。</p>", token)

	// 发送邮件
	err := SendEmail(email, subject, message)
	if err != nil {
		log.Printf("发送邮件失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "验证码发送失败，请重试。"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
}
