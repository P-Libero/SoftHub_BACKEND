package config

import (
	"errors"
	"os"
)

type Config struct {
	Port          string
	AdminUsername string
	AdminPassword string
	JWTSecret     string
	DB            Database
	MinIO         MinIO
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

func Load() (*Config, error) {
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
	return cfg, nil
}
