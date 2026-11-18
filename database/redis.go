package database

import (
	"context"
	"fmt"
	"softhub/config"
	"time"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg config.Redis) (*redis.Client, error) {
	//调用 redis.NewClient 时不会发起任何 TCP 三次握手。
	//即使填了一个不存在的 IP 或错误的密码，redis.NewClient 依然会成功返回指针。
	//因此必须紧接着执行一次 Ping 探活。
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
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return rdb, nil
}
