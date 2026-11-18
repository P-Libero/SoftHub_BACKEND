package router

import (
	"softhub/handler"
	"softhub/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func SetupRouter(total *handler.TotalHandler, rdb *redis.Client) *gin.Engine {
	gin.ForceConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(middleware.CORS())
	r.GET("/health", healthCheck)

	api := r.Group("/api")
	v1 := api.Group("/v1")
	softwareRouterGroup(v1, total, rdb)
	authRouterGroup(v1, total, rdb)
	userRouterGroup(v1, total, rdb)
	return r
}

func softwareRouterGroup(v1 *gin.RouterGroup, total *handler.TotalHandler, rdb *redis.Client) {
	h := total.Software
	// 公开接口
	v1.GET("/software", h.List)
	v1.GET("/software/:id", h.GetByID)
	v1.GET("/software/:id/versions", h.ListVersion)
	v1.GET("/software/:id/versions/latest", h.GetLatestVersion)
	v1.GET("/software/:id/versions/:versionID/download", h.DownloadVersionFile)

	// 需要登录的管理员接口
	authed := v1.Group("")
	authed.Use(middleware.JWTAuth(rdb), middleware.RoleMiddleware("admin")) //限定admin才能进
	{
		authed.POST("/software", h.Create)
		authed.PUT("/software/:id", h.Update)
		authed.DELETE("/software/:id", h.Delete)
		authed.POST("/software/:id/versions", h.CreateVersion)
		authed.POST("/software/:id/versions/:versionID/upload", h.UploadVersionFile)
	}
}

func authRouterGroup(v1 *gin.RouterGroup, total *handler.TotalHandler, rdb *redis.Client) {
	auth := v1.Group("/auth")
	auth.POST("/register", middleware.RateLimit(rdb, 5, time.Minute), total.User.Register)
	auth.POST("/login", middleware.RateLimit(rdb, 5, time.Minute), total.User.Login)
	auth.POST("/logout", middleware.JWTAuth(rdb), total.User.Logout)
}

func userRouterGroup(v1 *gin.RouterGroup, total *handler.TotalHandler, rdb *redis.Client) {
	user := v1.Group("/user")
	user.Use(middleware.JWTAuth(rdb))
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
