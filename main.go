package main

import (
	"context"
	"log"
	"softhub/config"
	"softhub/database"
	"softhub/handler"
	"softhub/model"
	"softhub/repository"
	"softhub/router"
	"softhub/service"
	"softhub/storage"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	//打开数据库
	db, err := database.Connect(cfg.DB)
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	// 开发阶段自动创建或补充数据表
	if err := db.AutoMigrate(&model.SoftwareVersion{}, &model.Software{}, &model.User{}); err != nil {
		//检查数据库里是否已经存在 Software 对应的表（默认表名通常是 softwares 或根据命名策略转换）。
		//如果不存在，就创建它；如果存在，就对比结构体定义和现有表结构，补充缺失的列、索引或外键约束，让表结构跟上代码里的定义。
		log.Fatal("failed to migrate database: ", err)
	}

	minioStorage, err := storage.NewMinIO(cfg.MinIO)
	if err != nil {
		log.Fatal("failed to connect minio storage: ", err)
	}

	//依赖注入
	softwareRepo := repository.NewSoftwareRepository(db)
	userRepo := repository.NewUserRepository(db)


	softwareService := service.NewSoftwareService(softwareRepo, minioStorage)
	userService := service.NewUserService(userRepo)

	//确保管理员账号存在
	if cfg.AdminUsername != "" && cfg.AdminPassword != "" {
		if err := userService.EnsureAdmin(context.Background(), cfg.AdminUsername, cfg.AdminPassword); err != nil {
			log.Fatal("failed to ensure admin: ", err)
		}
	}

	softwareHandler := handler.NewSoftwareHandler(softwareService)
	userHandler := handler.NewUserHandler(userService)

	totalHandler := handler.NewTotalHandler(softwareHandler, userHandler)

	//	main只负责初始化和启动服务。
	r := router.SetupRouter(totalHandler)
	addr := ":" + cfg.Port
	log.Println("Starting server starting on " + addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("failed to start server: ", err)
	}

}
