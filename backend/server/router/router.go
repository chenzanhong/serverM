package router

import (
	"log"
	"os"
	"time"

	"serverM/server/config"
	"serverM/server/handle/admin"
	"serverM/server/handle/agent/getscript"
	"serverM/server/handle/agent/install"
	pt "serverM/server/handle/agent/port"
	"serverM/server/handle/agent/threshold"
	"serverM/server/handle/company"
	e "serverM/server/handle/email"
	"serverM/server/handle/monitor" // 引入 monitor 包
	"serverM/server/handle/notice"
	"serverM/server/handle/transfer"
	g "serverM/server/handle/transfer/global"
	trans "serverM/server/handle/transfer/trans-init"
	"serverM/server/handle/user/info"
	"serverM/server/handle/user/login"
	"serverM/server/handle/user/update"
	"serverM/server/logs"
	"serverM/server/middlewire"
	"serverM/server/middlewire/cors"
	db "serverM/server/model/init"
	"serverM/server/redis"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	logs.InitZapSugarDefault()

	// 读取DBConfig.yaml文件
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置环境变量
	setEnvVariables(config)

	router := gin.Default()
	router.Use(cors.CORSMiddleware())

	// 初始化 Redis 连接
	if err := redis.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// 连接pg数据库
	if err := db.ConnectDatabase(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 连接TDengine数据库
	// if err := db.ConnectTDengine(); err != nil {
	// 	log.Fatalf("Failed to connect to TDengine database: %v", err)
	// }

	// 初始化数据库
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化数据库数据
	if err := db.InitDBData(); err != nil {
		log.Fatalf("Failed to initialize data: %v", err)
	}

	//初始化TDengine
	// if err := db.InitTDengine(); err != nil {
	// 	log.Fatalf("Failed to initialize TDengine: %v", err)
	// }

	// 初始化SSH连接池及文件传输服务
	g.Pool = trans.NewSSHConnectionPool(10, 10*time.Minute)
	stopChan := make(chan struct{})
	go g.Pool.Cleanup(stopChan)          // 启动清理协程
	trans.NewFileTransferService(g.Pool) // 初始化文件传输服务
	g.FTS = trans.NewFileTransferService(g.Pool)

	router.Static("/static", "./static")

	// 注册 Swagger 路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/docs/swagger.json")))

	setupPublicRoutes(router)
	setupAuthRoutes(router)

	return router
}

func setEnvVariables(cfg *config.Config) {
	os.Setenv("DB_USER", cfg.DB.User)
	os.Setenv("DB_PASSWORD", cfg.DB.Password)
	os.Setenv("DB_HOST", cfg.DB.Host)
	os.Setenv("DB_PORT", cfg.DB.Port)
	os.Setenv("DB_NAME", cfg.DB.Name)

	os.Setenv("EMAIL_NAME", cfg.Email.Name)
	os.Setenv("EMAIL_PASSWORD", cfg.Email.Password)
	os.Setenv("SMTP_SERVER_HOST", cfg.SMTPServer.Host)
	os.Setenv("SMTP_SERVER_PORT", cfg.SMTPServer.Port)

	os.Setenv("REDIS_HOST", cfg.Redis.Host)
	os.Setenv("REDIS_PASSWORD", cfg.Redis.Password)
	os.Setenv("REDIS_PORT", cfg.Redis.Port)
	os.Setenv("REDIS_DB", cfg.Redis.DB)
}

func setupPublicRoutes(r *gin.Engine) {
	r.POST("/agent/register", login.Register)
	r.POST("/registertoken", e.SendVerificationCode) // 注册时发送验证码邮件
	r.POST("/agent/login", login.Login)

	r.GET("/agentscript", getscript.GetAgentScript)                         // 获取安装代理程序的脚本，?hostname=
	r.GET("/combinedscript", getscript.GetCombinedScript)                   // 获取合并后的脚本——包含安装代理程序和配置反向SSH隧道
	r.GET("/uninstallagentscript", getscript.GetAgentUninstallScript)       // 获取删除代理服务的脚本，?hostname=
	r.GET("/uninstallcombinedscript", getscript.GetCombinedUninstallScript) // 获取删除联合服务的脚本，?hostname=

	r.POST("/agent/addSystem_info", monitor.ReceiveAndStoreSystemMetrics) // 接收代理客户端发送过来的监控数据，并进行预警判断、处理
}

func setupAuthRoutes(r *gin.Engine) {
	auth := r.Group("/agent", middlewire.JWTAuthMiddleware())
	{
		// 用户信息
		auth.GET("/userInfo", info.GetUserInfo)                           // 当前登录用户的个人信息
		auth.GET("/allUserInfo", info.GetAllUserInfo)                     // 所有用户信息
		auth.POST("/updateUserInfo", update.UpdateUserInfo)               // 更新当前用户信息
		auth.POST("/reset_password", update.ResetPassword)                // 校验token，设置密码
		auth.POST("/request_reset_password", update.RequestResetPassword) // 请求设置密码，发送含校验码的邮件
		auth.GET("/info/recivelist", info.GetReceiveList)                 // 获取该用户作为接收者所接收到的所有信息
		auth.GET("/info/sendlist", info.GetSendList)                      // 获取该用户作为发送者所发送的所有信息
		auth.POST("/info/manage", notice.ManageNotice)                    // 处理通知，当前通知没有“同意”和“拒绝”，只有“未处理”、“已处理”、“已过期”3个状态，默认同意处理

		// 系统/公司管理员操作
		auth.POST("/registercompany", company.Register)       // 注册公司
		auth.POST("/addMember", admin.AddMember)              // 添加成员
		auth.POST("/deleteMembers", admin.DeleteMember)       // 批量删除成员
		auth.GET("/getmemberinfo", admin.GetMemberInfo)       // 获取公司成员信息
		auth.GET("/get-company-info", admin.GetCompanyInfo)   // 指定公司的信息（含成员信息）
		auth.GET("/get-company-list", company.GetCompanyList) // 公司列表
		auth.POST("/sshkey", admin.AddSShkey)                 // 添加SSH密钥
		auth.POST("/joincompany", admin.JoinCompany)          // 邀请成员加入公司
		auth.POST("/replaceadmin", admin.ReplaceAdmin)        // 更换管理员

		// 监控
		auth.POST("/install", install.InstallAgent)                        // 添加服务器
		auth.GET("/list", monitor.HostInfoList)                            // 获取服务器列表
		auth.GET("/monitor/:hostname", monitor.GetAgentInfo)               // 获取服务器的最新30个监控数据记录
		auth.GET("/monitor/status/:hostname", monitor.GetLatestSystemInfo) // 获取服务器最新的监控数据记录（实时信息）
		auth.POST("/delete", monitor.DeleteSystemInfo)                     // 删除服务器
		auth.GET("/getwarning", monitor.GetWarningRecordsByHostname)       // 获取预警记录

		// 脚本
		auth.GET("/agentscript", getscript.GetAgentScript)                         // 获取安装代理程序的脚本
		auth.GET("/sshscript", getscript.GetSSHScript)                             // 获取配置反向ssh的脚本
		auth.GET("/combinedscript", getscript.GetCombinedScript)                   // 获取合并后的脚本——包含安装代理程序和配置反向SSH隧道
		auth.GET("/uninstallagentscript", getscript.GetAgentUninstallScript)       // 获取删除代理服务的脚本，?hostname=
		auth.GET("/uninstallcombinedscript", getscript.GetCombinedUninstallScript) // 获取删除联合服务的脚本，?hostname=
		auth.GET("/port/get", pt.GetAvailablePort)                                 // 获取用于生成ssh脚本所需要的端口port

		// 文件传输
		auth.POST("/upload", transfer.CommonUpload)                // 本地文件上传
		auth.POST("/download", transfer.CommonDownload)            // 服务器文件下载（到本地）
		auth.POST("/transfer", transfer.TransferBetweenTwoServers) // 服务器间单文件传输

		// 邮件
		auth.POST("/sendemail", e.SendEmailHandler) // 发送指定信息的邮件

		// 日志
		auth.POST("/getuseroperationlogs", logs.GetUserOperationLogs) // 获取用户操作日志，支持按时间段、操作类型、按用户名筛选

		// 预警
		auth.POST("/setthreshold", threshold.UpdateThreshold) // 设置服务器内存/CPU使用率的阈值
	}
}
