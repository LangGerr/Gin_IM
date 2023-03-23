package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 表示可以访问所有的域名
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
		// 设置最大访问时间
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		// 设置允许访问的方法 get post...
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
		// 设置访问头
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}
