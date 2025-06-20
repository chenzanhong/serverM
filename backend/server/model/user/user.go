package model

import "time"

// import "time"

type User struct {
	ID         int    `json:"id" gorm:"primarykey;autoIncrement"`
	Name       string `json:"name" gorm:"column:name; not null"`
	Realname   string `json:"realname" gorm:"column:realname; not null"`
	Email      string `json:"email" gorm:"unique;not null"`
	Password   string `json:"password" gorm:"not null"`
	RoleId     int    `json:"role_id" gorm:"column:role_id;default:0"` // 2:ROOT: , 1: ADMIN, 0: USER
	CompanyId  int    `json:"company_id" gorm:"column:company_id;default:0"`
	IsVerified bool   `json:"is_verified" gorm:"column:is_verified"`
	Token      string `json:"token" gorm:"column:token"`
}

type Company struct {
	ID               int    `json:"id" gorm:"primarykey;autoIncrement"`
	Name             string `json:"name" gorm:"column:name; not null"`
	SocialCreditCode string `json:"social_credit_code" gorm:"column:social_credit_code; not null"`
	Description      string `json:"description" gorm:"column:description; not null"`
	AdminID          int    `json:"admin_id" gorm:"column:admin_id; not null"`
	MemberNum        int    `json:"membernum" gorm:"column:membernum; default:0"`
	SystemNum        int    `json:"systemnum" gorm:"column:systemnum; default:0"`
}

type Role struct {
	ID          int    `json:"id" gorm:"column:id"`
	Name        string `json:"name" gorm:"column:role_name"`
	Description string `json:"description" gorm:"description"`
}

type SSHKey struct {
	ID       int    `json:"id" gorm:"primarykey;autoIncrement"`
	Hostname string `json:"host_name" gorm:"column:host_name"`
	SSHKey   string `json:"sshkey" gorm:"column:sshkey"`
}

type Notice struct {
	ID       int    `json:"id" gorm:"primarykey;autoIncrement"`
	Send     string `json:"send" gorm:"column:send"`
	Receive  string `json:"receive" gorm:"column:receive"`
	Content  string `json:"content" gorm:"column:content"`
	State    string `json:"state" gorm:"column:state"`
	CreateAt string `json:"created_at" gorm:"column:created_at"`
}

// host_info表
type HostInfo struct {
	ID           int     `json:"id" gorm:"column:id"`
	UserName     string  `json:"user_name" gorm:"column:user_name"`
	Hostname     string  `json:"host_name" gorm:"column:host_name"`
	CompanyID    int     `json:"company_id,omitempty" gorm:"column:company_id"`
	IP           string  `json:"ip" gorm:"column:ip"`
	OS           string  `json:"os" gorm:"column:os"`
	Platform     string  `json:"platform" gorm:"column:platform"`
	KernelArch   string  `json:"kernel_arch" gorm:"column:kernel_arch"`
	CPUThreshold float64 `json:"cpu_threshold" gorm:"column:cpu_threshold"`
	MemThreshold float64 `json:"mem_threshold" gorm:"column:mem_threshold"`
	CreatedAt    string  `json:"host_info_created_at" gorm:"column:created_at"`
}

// TableName 指定表名
func (HostInfo) TableName() string {
	return "host_info" // 数据库表名对应
}

type HostAndToken struct {
	ID            int    `json:"id"`
	HostName      string `json:"host_name"`
	Token         string `json:"token"`
	LastHeartBeat string `json:"last_heartbeat"`
	Status        string `json:"status" gorm:"default 'offline'"`
}

func (HostAndToken) TableName() string {
	return "hostandtoken" // 数据库表名对应
}

type SSHPort struct {
	ID        int       `json:"id"`
	Port      int       `json:"port" gorm:"column:port"`
	IsUsed    bool      `json:"is_used" gorm:"default:false"`
	Hostname  string    `json:"hostname" gorm:"column:hostname"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// Warning 定义告警记录结构体
type Warning struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	HostName     string    `json:"host_name"`
	Username     string    `json:"username"`
	WarningType  string    `json:"warning_type"`
	WarningTitle string    `json:"warning_title"`
	WarningTime  time.Time `json:"warning_time"`
}
