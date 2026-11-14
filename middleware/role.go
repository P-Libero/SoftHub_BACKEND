package middleware

import (
	"slices"
	"softhub/pkg/response"

	"github.com/gin-gonic/gin"
)

func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(CtxRole)
		if slices.Contains(roles, role) {
			c.Next()
			return
		}
		response.Forbidden(c, "权限不足")
		c.Abort()
	}
}
