package company

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"serverM/server/logs"
	m_init "serverM/server/model/init"
	u "serverM/server/model/user"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRequest 定义了公司注册请求结构体
type RegisterRequest struct {
	Company            string `json:"company"`
	Social_Credit_Code string `json:"social_credit_code"`
	Legal_Name         string `json:"legal_name"`
	Admin_Name         string `json:"admin_name"`
	Admin_Email        string `json:"admin_email"`
}

// 申请注册公司
func Register(c *gin.Context) {

	Username, exists := c.Get("username")
	if !exists {
		log.Printf("未找到用户信息")
		c.JSON(401, gin.H{
			"message": "未找到用户信息",
		})
		return
	}
	username := Username.(string)

	//定义用于接收JSON数据的请求体
	var input RegisterRequest

	// 解析JSON数据
	if err := c.BindJSON(&input); err != nil {
		log.Printf("请求数据格式错误")
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "解析请求失败，请检查请求格式是否正确")
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}

	var detail = fmt.Sprintf("公司名称:%s,公司法人:%s,公司管理员:%s,社会信用代码:%s,管理员邮箱:%s", input.Company, input.Legal_Name, input.Admin_Name, input.Social_Credit_Code, input.Admin_Email)

	//检查公司名是否已经存在
	var company u.Company
	if err := m_init.DB.Where("name = ?", input.Company).First(&company).Error; err == nil {
		log.Printf("公司已存在")
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "公司已存在。"+detail)
		c.JSON(http.StatusBadRequest, gin.H{"message": "公司名已存在"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("数据库查询失败")
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "查询公司名失败。"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询公司名失败"})
		return
	}

	//简单检查社会信用代码格式是否正确
	//if len(input.Social_Credit_Code) != 18 {
	//	c.JSON(http.StatusBadRequest, gin.H{"message": "社会信用代码应该为18位"})
	//	return
	//}
	//检测公司统一社会信用代码是否已经存在
	if err := m_init.DB.Where("social_credit_code = ?", input.Social_Credit_Code).First(&company).Error; err == nil {
		log.Printf("公司已存在")
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "公司统一社会信用代码已经存在。"+detail)
		c.JSON(http.StatusBadRequest, gin.H{"message": "公司统一社会信用代码已存在"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("查询公司统一社会信用代码失败")
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "查询公司统一社会信用代码失败。"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询公司统一社会信用代码失败"})
		return
	}

	//查看公司管理员的用户信息
	var admin u.User
	if err := m_init.DB.Where("name = ?", username).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("管理员用户信息不存在")
			logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "管理员用户信息不存在。"+detail)
			c.JSON(http.StatusBadRequest, gin.H{"message": "管理员用户信息不存在"})
			return
		}
		log.Printf("数据库查询用户失败")
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "查询管理员用户信息失败。"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询管理员用户信息失败"})
		return
	}
	if admin.Realname == "" {
		log.Printf("管理员未实名")
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "管理员未实名。"+detail)
		c.JSON(http.StatusBadRequest, gin.H{"message": "管理员需要实名,请在个人信息中实名"})
		return
	}
	if admin.Realname != input.Admin_Name {
		log.Printf("管理员实名信息不匹配")
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "管理员实名信息不匹配。"+detail)
		c.JSON(http.StatusBadRequest, gin.H{"message": "管理员实名信息不匹配"})
		return
	}
	if admin.Email != input.Admin_Email {
		log.Printf("管理员邮箱不匹配")
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "管理员邮箱不匹配。"+detail)
		c.JSON(http.StatusBadRequest, gin.H{"message": "管理员邮箱不匹配"})
		return
	}
	if admin.RoleId == 1 {
		log.Printf("已注册有公司")
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "已注册有公司。"+detail)
		c.JSON(http.StatusBadRequest, gin.H{"message": "已注册有公司"})
		return
	}

	//创建公司
	newCompany := u.Company{
		AdminID:          admin.ID,
		Name:             input.Company,
		SocialCreditCode: input.Social_Credit_Code,
		MemberNum:        1,
		Description:      "暂无",
	}

	if err := m_init.DB.Create(&newCompany).Error; err != nil {
		log.Println(logs.GetLogPrefix(2) + "创建公司失败")
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "创建公司失败。"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "创建公司失败"})
		return
	}

	// 同步更新公司管理员的company_id和role_id
	var roleId = 1
	if username == "root" {
		roleId = 2
	}
	err := m_init.DB.Model(&admin).Select("company_id", "role_id").Updates(map[string]interface{}{
		"company_id": newCompany.ID,
		"role_id":    roleId,
	}).Error
	if err != nil {
		log.Println(logs.GetLogPrefix(2) + "更新公司管理员信息失败")
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "更新公司管理员信息失败。"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新公司管理员信息失败"})
		return
	}

	// 将管理员的服务器的company_id改为公司id
	if err := m_init.DB.Exec("UPDATE host_info SET company_id = $1 WHERE user_name = $2", newCompany.ID, username).Error; err != nil {
		// 处理错误
		log.Println("更新管理员的服务器信息失败:", err)
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "更新管理员的服务器信息失败。")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新管理员的服务器信息失败"})
		return
	}

	//团队申请
	content := username + "申请注册公司:" + input.Company + "，法人:" + input.Legal_Name +
		",管理员:" + input.Admin_Name + ",社会信用代码:" + input.Social_Credit_Code +
		",管理员邮箱:" + input.Admin_Email

	const layout = "2006-01-02 15:04:05.000000"
	createAt := time.Now().Format(layout)

	notice := u.Notice{
		Content:  content,
		Send:     username,
		Receive:  "root",
		State:    "processed",
		CreateAt: createAt,
	}
	if err := m_init.DB.Create(&notice).Error; err != nil {
		log.Printf("数据库插入申请失败")
		logs.Sugar.Errorw("申请注册公司", "username", username, "detail", "数据库插入注册公司申请失败。"+detail)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库插入申请失败"})
		return
	}

	logs.Sugar.Infow("申请注册公司", "username", username, "detail", "成功注册公司。"+detail)
	c.JSON(http.StatusOK, gin.H{
		"message": "成功注册公司",
	})
}
