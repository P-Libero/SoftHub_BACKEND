package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	Port          string
	AdminUsername string
	AdminPassword string
	JWTSecret     string
	DB            Database
	MinIO         MinIO
	Redis         Redis
}

type Database struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

type MinIO struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type Redis struct {
	Addr     string //服务程序地址
	Password string
	DB       int //逻辑数据库默认编号0
}

func Load() (*Config, error) {
	redisDB := 0
	if dbstr := os.Getenv("REDIS_DB"); dbstr != "" {
		var err error
		redisDB, err = strconv.Atoi(dbstr) //转换为整型
		if err != nil {
			return nil, errors.New("REDIS_DB必须为整型")
		}
	}

	cfg := &Config{
		Port:          os.Getenv("PORT"),
		AdminUsername: os.Getenv("ADMIN_USERNAME"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		DB: Database{
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Name:     os.Getenv("DB_NAME"),
		},
		MinIO: MinIO{
			Endpoint:  os.Getenv("MINIO_ENDPOINT"),
			AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
			SecretKey: os.Getenv("MINIO_SECRET_KEY"),
			Bucket:    os.Getenv("MINIO_BUCKET"),
			UseSSL:    os.Getenv("MINIO_USE_SSL") == "true",
		},
		Redis: Redis{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       redisDB,
		},
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET 未配置")
	}
	if cfg.DB.User == "" {
		cfg.DB.User = "root"
	}
	if cfg.DB.Host == "" {
		cfg.DB.Host = "127.0.0.1"
	}
	if cfg.DB.Port == "" {
		cfg.DB.Port = "3306"
	}
	if cfg.DB.Name == "" {
		cfg.DB.Name = "SoftHub"
	}
	if cfg.MinIO.Endpoint == "" {
		cfg.MinIO.Endpoint = "127.0.0.1:9000"
	}
	if cfg.MinIO.AccessKey == "" || cfg.MinIO.SecretKey == "" {
		return nil, errors.New("MINIO密钥未配置")
	}
	if cfg.MinIO.Bucket == "" {
		cfg.MinIO.Bucket = "softhub"
	}

	if cfg.Redis.Addr == "" {
		cfg.Redis.Addr = "127.0.0.1:6379"
	}

	return cfg, nil
}
