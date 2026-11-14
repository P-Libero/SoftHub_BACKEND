package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		//允许哪些前端源（Origin，即 协议://域名:端口）跨域访问我。
		c.Header("Access-Control-Allow-Origin", "*")
		//允许前端跨域使用哪些 HTTP 请求方法
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		//允许前端在跨域请求里额外携带哪些自定义请求头（Request Headers）。
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		//这次“预检（询问）”的结果，浏览器可以在本地缓存多少秒。
		c.Header("Access-Control-Max-Age", "86400")

		// 放行浏览器的 OPTIONS 预检请求，不进入后续路由和鉴权
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
