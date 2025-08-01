package cors

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware 是一个中间件，用于配置CORS
func CORSMiddleware() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // 允许的源，可以根据需要修改http://localhost:8081
	// config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	// config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowMethods = []string{"*"}
	config.AllowHeaders = []string{"*"}

	return cors.New(config)
}