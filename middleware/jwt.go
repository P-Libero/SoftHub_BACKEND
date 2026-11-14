package middleware

import (
	"softhub/pkg/response"
	"softhub/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// 上下文里存用户信息的 key
const (
	CtxUserID   = "user_id"
	CtxUsername = "username"
	CtxRole     = "role"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "缺少 Authorization 头")
			c.Abort()
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Authorization 格式错误")
			c.Abort()
			return
		}
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "token无效或已过期")
			c.Abort()
			return
		}
		c.Set(CtxUsername, claims.Username)
		c.Set(CtxRole, claims.Role)
		c.Set(CtxUserID, claims.UserID)
		c.Next()
	}

}
