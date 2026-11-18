package middleware

import (
	"softhub/pkg/response"
	"softhub/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 上下文里存用户信息的 key
const (
	CtxUserID    = "user_id"
	CtxUsername  = "username"
	CtxRole      = "role"
	CtxToken     = "token"
	CtxExpiresAt = "expires_at"
)

func JWTAuth(rdb *redis.Client) gin.HandlerFunc {
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
		tokenString := parts[1]
		//先查一下本地的
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "token无效或已过期")
			c.Abort()
			return
		}
		//再查黑名单
		exists, err := rdb.Exists(c.Request.Context(), "jwt:blacklist:"+tokenString).Result()
		if err != nil {
			_ = c.Error(err)
			response.ServerError(c, "服务器内部错误")
			c.Abort()
			return
		}
		if exists > 0 {
			response.Unauthorized(c, "token已失效，请重新登录")
			c.Abort()
			return
		}

		c.Set(CtxUsername, claims.Username)
		c.Set(CtxRole, claims.Role)
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxToken, tokenString)
		c.Set(CtxExpiresAt, claims.ExpiresAt.Time)
		c.Next()
	}

}
