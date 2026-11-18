package database

import (
	"context"
	"softhub/config"
	"time"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg config.Redis) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	//创建 5 秒超时上下文，主动发送 PING 指令验证网络连通性与密码是否正确
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close() //不活就关闭
		return nil, err
	}
	return rdb, nil
}
