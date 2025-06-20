package login

import (
	"errors"
	"net/http"
	"regexp"
	"serverM/server/middlewire"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gorm.io/gorm"

	m_init "serverM/server/model/init"
	u "serverM/server/model/user"

	e "serverM/server/handle/email"
)

// RegisterRequest 定义注册请求的数据结构
type RegisterRequest struct {
	Company  string `json:"company"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// LoginRequest 定义登录请求的数据结构
type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// 正则表达式验证邮箱格式
func IsValidEmail(email string) bool {
	// 正则表达式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	return emailRegex.MatchString(email)
}

// Register 用户注册接口
//
// @Summary 用户注册
// @Description 通过提供用户名、邮箱和密码来注册新用户。
// @Tags User
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 200 {object} map[string]string "注册成功"
// @Failure 400 {object} map[string]string "请求数据格式错误"
// @Failure 422 {object} map[string]string "用户名或邮箱验证失败"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /register [post]
func Register(c *gin.Context) {
	// 定义用于接收 JSON 数据的结构体
	var input RegisterRequest

	// 解析 JSON 数据
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}

	// 检查token是否正确
	e.EmailTokenMutex.Lock()
	tokenInfo, exists := e.EmailToken[input.Email]
	e.EmailTokenMutex.Unlock()
	if !exists {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "未请求验证码或验证码已过期，请重新获取。"})
		return
	}

	if time.Now().After(tokenInfo.ExpiresAt) {
		// 清除过期记录
		e.EmailTokenMutex.Lock()
		delete(e.EmailToken, input.Email)
		e.EmailTokenMutex.Unlock()

		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "验证码已过期，请重新获取。"})
		return
	}

	if tokenInfo.Token != input.Token {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "验证码错误，请重新输入。"})
		return
	}

	// 验证通过，清除该邮箱的验证码

	// 去除前后空格
	input.Email = strings.TrimSpace(input.Email)
	e.EmailTokenMutex.Lock()
	delete(e.EmailToken, input.Email)
	e.EmailTokenMutex.Unlock()

	// 数据验证
	if len(input.Name) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "用户名不能为空"})
		return
	} else if len(input.Email) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "邮箱不能为空"})
		return
	} else if !IsValidEmail(input.Email) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "邮箱格式不正确"})
		return
	} else if len(input.Password) < 6 || len(input.Password) > 16 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "密码长度应该不小于6，不大于16"})
		return
	}

	// 检查用户名是否存在
	var user u.User
	err := m_init.DB.Where("name = ?", input.Name).First(&user).Error
	if err == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "用户名已存在"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果 err 不为 nil 且不是因为记录未找到导致的，则是其他数据库错误
		c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库查询用户名失败"})
		return
	}
	// 检查邮箱是否存在
	err = m_init.DB.Where("email = ?", input.Email).First(&user).Error
	if err == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "邮箱已存在"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果 err 不为 nil 且不是因为记录未找到导致的，则是其他数据库错误
		c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库查询邮箱失败"})
		return
	}

	// 创建用户
	newUser := u.User{
		Name:       input.Name,
		Realname:   input.Name,
		Email:      input.Email,
		Password:   input.Password,
		RoleId:     0,
		IsVerified: true,
	}
	err = m_init.DB.Create(&newUser).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "用户创建失败"})
		return
	}

	// 返回结果，包括加密后的密码
	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
	})
}

// Login 用户登录接口
//
// @Summary 用户登录
// @Description 用于用户登录，需要提供用户名和密码，并返回JWT token。
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} map[string]string "登录成功"
// @Failure 400 {object} map[string]string "请求数据格式错误"
// @Failure 401 {object} map[string]string "用户不存在或密码错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /login [post]
func Login(c *gin.Context) {
	// 定义用于接收 JSON 数据的结构体
	var input LoginRequest

	// 解析 JSON 数据
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}

	// 查找用户
	var user u.User
	err := m_init.DB.Where("name = ?", input.Name).First(&user).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户不存在"})
		return
	}

	// 验证密码
	if user.Password != input.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "密码错误"})
		return
	}

	// 生成 JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &middlewire.Claims{
		Username: input.Name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(middlewire.JwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "生成 token 错误"})
		return
	}

	// 查询用户角色
	var role u.Role
	if err := m_init.DB.Table("roles").Where("id = ?", user.RoleId).First(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询用户角色失败"})
		return
	}

	// 登录成功
	c.JSON(http.StatusOK, gin.H{
		"message":   "登录成功",
		"role":      role.Name,
		"token":     tokenString,
		"username":  user.Name,
		"companyId": user.CompanyId,
	})
}
