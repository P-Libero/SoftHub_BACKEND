package router

import (
	"softhub/handler"
	"softhub/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRouter(total *handler.TotalHandler) *gin.Engine {
	gin.ForceConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(middleware.CORS())
	r.GET("/health", healthCheck)

	api := r.Group("/api")
	v1 := api.Group("/v1")
	softwareRouterGroup(v1, total)
	authRouterGroup(v1, total)
	userRouterGroup(v1, total)
	return r
}

func softwareRouterGroup(v1 *gin.RouterGroup, total *handler.TotalHandler) {
	h := total.Software
	// 公开接口
	v1.GET("/software", h.List)
	v1.GET("/software/:id", h.GetByID)
	v1.GET("/software/:id/versions", h.ListVersion)
	v1.GET("/software/:id/versions/latest", h.GetLatestVersion)
	v1.GET("/software/:id/versions/:versionID/download", h.DownloadVersionFile)

	// 需要登录的管理员接口
	authed := v1.Group("")
	authed.Use(middleware.JWTAuth(), middleware.RoleMiddleware("admin"))
	{
		authed.POST("/software", h.Create)
		authed.PUT("/software/:id", h.Update)
		authed.DELETE("/software/:id", h.Delete)
		authed.POST("/software/:id/versions", h.CreateVersion)
		authed.POST("/software/:id/versions/:versionID/upload", h.UploadVersionFile)
	}
}

func authRouterGroup(v1 *gin.RouterGroup, total *handler.TotalHandler) {
	auth := v1.Group("/auth")
	auth.POST("/register", middleware.RateLimit(5, time.Minute), total.User.Register)
	auth.POST("/login", middleware.RateLimit(5, time.Minute), total.User.Login)
}

func userRouterGroup(v1 *gin.RouterGroup, total *handler.TotalHandler) {
	user := v1.Group("/user")
	user.Use(middleware.JWTAuth())
	user.GET("/me", total.User.FindMe)
}

// 全局测试路由
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"status": "ok",
		},
	})
}
