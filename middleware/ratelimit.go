package middleware

import (
	"softhub/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimit(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	//旧方法的注释:
	//内层函数引用了外层函数的局部变量 mu 和 visitors
	//而内层函数被 return 出去长期存活了。于是 Go 编译器自动把 mu 和 visitors 从栈（Stack）移到了堆（Heap）上
	//只要后端进程不关，visitors 这张表就一直活在堆内存里，所有 HTTP 请求读写的都是同一块堆内存。
	//var mu sync.Mutex
	//var visitors = make(map[string]*visitor)
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// 构造当前客户端 IP 在 Redis 中的唯一限流 Key，例如 "ratelimit:127.0.0.1"
		key := "ratelimit:" + c.ClientIP()
		// TODO:之后修改为令牌桶 因为这里有边界双倍突刺的风险
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			_ = c.Error(err)
			response.ServerError(c, "")
			c.Abort()
			return
		}
		if count == 1 {
			if err := rdb.Expire(ctx, key, window).Err(); err != nil {
				_ = c.Error(err)
				response.ServerError(c, "")
				c.Abort()
				return
			}
		}
		// 3. 超过窗口内最大允许次数，拦截并返回 HTTP 429
		if count > int64(limit) {
			response.TooManyRequests(c, "请求过于频繁，请稍后重试")
			c.Abort()
			return
		}

		c.Next()
	}
}
