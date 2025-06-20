package main

import (
	"serverM/server/config"
	"serverM/server/handle/admin"
	"serverM/server/handle/agent/getscript"
	"serverM/server/handle/agent/install"
	pt "serverM/server/handle/agent/port"
	"serverM/server/handle/agent/threshold"
	"serverM/server/handle/company"
	e "serverM/server/handle/email"
	"serverM/server/handle/monitor" // 引入 monitor 包
	"serverM/server/handle/transfer"
	g "serverM/server/handle/transfer/global"
	trans "serverM/server/handle/transfer/trans-init"
	"serverM/server/handle/notice"
	"serverM/server/handle/user/info"
	"serverM/server/handle/user/login"
	"serverM/server/handle/user/update"
	"serverM/server/logs"
	"serverM/server/middlewire"
	"serverM/server/middlewire/cors"
	db "serverM/server/model/init"
	"serverM/server/redis"
	"time"

	"log"
	"os"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func main() {

	//ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	//defer stop()

	//go func() {
	//	<-ctx.Done()
	//	// 执行连接关闭
	//	if err := model.DB.Close(); err != nil {
	//		log.Printf("Failed to close database connection: %v", err)
	//	}
	//	if err := model.TDengine.Close(); err != nil {
	//		log.Printf("Failed to close TDengine connection: %v", err)
	//	}
	//	os.Exit(0)
	//}()

	logs.InitZapSugarDefault()

	//读取DBConfig.yaml文件
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	//设置数据库连接的环境变量
	os.Setenv("DB_USER", config.DB.User)
	os.Setenv("DB_PASSWORD", config.DB.Password)
	os.Setenv("DB_HOST", config.DB.Host)
	os.Setenv("DB_PORT", config.DB.Port)
	os.Setenv("DB_NAME", config.DB.Name)
	//设置TDengine数据库连接的环境变量
	os.Setenv("TDENGINE_USER", config.TDengine.User)
	os.Setenv("TDENGINE_PASSWORD", config.TDengine.Password)
	os.Setenv("TDENGINE_HOST", config.TDengine.Host)
	os.Setenv("TDENGINE_PORT", config.TDengine.Port)
	os.Setenv("TDENGINE_NAME", config.TDengine.Name)
	// OSS服务
	os.Setenv("OSS_REGION", config.OSS.OSS_REGION)
	os.Setenv("OSS_ACCESS_KEY_ID", config.OSS.OSS_ACCESS_KEY_ID)
	os.Setenv("OSS_ACCESS_KEY_SECRET", config.OSS.OSS_ACCESS_KEY_SECRET)
	os.Setenv("OSS_BUCKET", config.OSS.OSS_BUCKET)
	// 邮箱服务
	os.Setenv("EMAIL_NAME", config.Email.Name)
	os.Setenv("EMAIL_PASSWORD", config.Email.Password)
	os.Setenv("BASE_URL", config.Email.Url)
	os.Setenv("SMTP_SERVER_HOST", config.SMTPServer.Host)
	os.Setenv("SMTP_SERVER_PORT", config.SMTPServer.Port)
	// Redis服务
	os.Setenv("REDIS_HOST", config.Redis.Host)
	os.Setenv("REDIS_PASSWORD", config.Redis.Password)
	os.Setenv("REDIS_PORT", config.Redis.Port)
	os.Setenv("REDIS_DB", config.Redis.DB)

	router := gin.Default()
	router.Use(cors.CORSMiddleware())
	// 初始化 Redis 连接
	if err := redis.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// 连接数据库
	if err := db.ConnectDatabase(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	//连接TDengine数据库
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
	// 初始化redis
	// if err := db.InitRedis(); err!= nil {
	// 	log.Fatalf("Failed to connect to redis: %v", err)
	// }
	// 初始化SSH连接池及文件传输服务
	g.Pool = trans.NewSSHConnectionPool(10, 10*time.Minute)
	stopChan := make(chan struct{})
	defer close(stopChan)
	go g.Pool.Cleanup(stopChan)          // 启动清理协程
	trans.NewFileTransferService(g.Pool) // 初始化文件传输服务
	g.FTS = trans.NewFileTransferService(g.Pool)

	go monitor.CheckServerStatus()
	router.Static("/static", "./static")

	// 注册 Swagger 路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/docs/swagger.json")))

	router.POST("/agent/register", login.Register)
	router.POST("/registertoken", e.SendVerificationCode) // 注册时发送验证码邮件
	router.POST("/agent/login", login.Login)
	router.GET("/defaultagentscript", getscript.GetAgentScript) // 获取安装代理程序的脚本

	// 需要 JWT 认证的路由
	auth := router.Group("/agent", middlewire.JWTAuthMiddleware())
	{
		// 用户信息
		auth.GET("/userInfo", info.GetUserInfo)
		auth.GET("/allUserInfo", info.GetAllUserInfo)
		auth.POST("/updateUserInfo", update.UpdateUserInfo)
		router.POST("/reset_password", update.ResetPassword)
		auth.POST("/request_reset_password", update.RequestResetPassword)
		auth.GET("/info/recivelist", info.GetReceiveList) //获取该用户作为接收者所接收到的所有信息
		auth.GET("/info/sendlist", info.GetSendList)      //获取该用户作为发送者所发送的所有信息
		auth.POST("/info/manage", notice.ManageNotice)      //处理通知状态

		// 系统/公司管理员操作
		auth.POST("/registercompany", company.Register)       // 注册公司
		auth.POST("/addMember", admin.AddMember)              // 添加成员
		auth.POST("/deleteMembers", admin.DeleteMember)       // 批量删除成员
		auth.GET("/getmemberinfo", admin.GetMemberInfo)       // 获取公司成员信息
		auth.GET("/get-company-info", admin.GetCompanyInfo)   // 指定公司的信息（含成员信息）
		auth.GET("/get-company-list", company.GetCompanyList) // 公司列表
		auth.POST("/sshkey", admin.AddSShkey)                 // 添加SSH密钥
		auth.POST("/joincompany", admin.JoinCompany)          // 邀请成员加入公司
		auth.POST("/replaceadmin", admin.ReplaceAdmin)        // 更换管 3理员

		// 监控
		auth.POST("/install", install.InstallAgent)
		auth.GET("/list", monitor.HostInfoList)
		auth.GET("/monitor/:hostname", monitor.GetAgentInfo)
		auth.GET("/monitor/status/:hostname", monitor.GetLatestSystemInfo)
		auth.POST("/delete", monitor.DeleteSystemInfo)
		auth.GET("/getwarning", monitor.GetWarningRecordsByHostname)

		// 脚本
		auth.GET("/agentscript", getscript.GetAgentScript)                         // 获取安装代理程序的脚本
		auth.GET("/sshscript", getscript.GetSSHScript)                             // 获取配置反向ssh的脚本
		auth.GET("/combinedscript", getscript.GetCombinedScript)                   // 获取合并后的脚本——包含安装代理程序和配置反向SSH隧道
		auth.GET("/uninstallagentscript", getscript.GetAgentUninstallScript)       // 获取删除代理服务的脚本，?hostname=
		auth.GET("/uninstallcombinedscript", getscript.GetCombinedUninstallScript) // 获取删除联合服务的脚本，?hostname=
		auth.GET("/port/get", pt.GetAvailablePort)                                 // 获取用于生成ssh脚本所需要的端口port

		// 文件传输
		auth.POST("/upload", transfer.CommonUpload)
		auth.POST("/download", transfer.CommonDownload)
		auth.POST("/transfer", transfer.TransferBetweenTwoServers)

		// 邮件
		auth.POST("/sendemail", e.SendEmailHandler)

		// 日志
		auth.POST("/getuseroperationlogs", logs.GetUserOperationLogs) // 获取用户操作日志，支持按时间段、操作类型、按用户名筛选

		// 预警
		auth.POST("/setthreshold", threshold.UpdateThreshold) // 设置阈值
	}

	router.GET("/agentscript", getscript.GetAgentScript)                         // 获取安装代理程序的脚本，?hostname=
	router.GET("/combinedscript", getscript.GetCombinedScript)                   // 获取安装代理程序的脚本，?hostname=
	router.GET("/uninstallagentscript", getscript.GetAgentUninstallScript)       // 获取删除代理服务的脚本，?hostname=
	router.GET("/uninstallcombinedscript", getscript.GetCombinedUninstallScript) // 获取删除联合服务的脚本，?hostname=

	router.POST("/agent/addSystem_info", monitor.ReceiveAndStoreSystemMetrics)

	// 预警的消费
	err = monitor.StartWorkerPool(100, 10000) // 100个消费者，管道长度10000
	if err != nil {
		log.Fatalf("Failed to start worker pool: %v", err)
	}

	// 通知notice的过期判断和处理：每天2点执行一次判断和修改
	go notice.ExpireOldNotices()

	router.Run("0.0.0.0:8080")
}
