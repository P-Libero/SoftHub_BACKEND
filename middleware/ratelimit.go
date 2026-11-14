package middleware

import (
	"softhub/pkg/response"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count     int
	expiresAt time.Time
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	//内层函数引用了外层函数的局部变量 mu 和 visitors
	//而内层函数被 return 出去长期存活了。于是 Go 编译器自动把 mu 和 visitors 从栈（Stack）移到了堆（Heap）上
	//只要后端进程不关，visitors 这张表就一直活在堆内存里，所有 HTTP 请求读写的都是同一块堆内存。
	var mu sync.Mutex
	var visitors = make(map[string]*visitor)
	return func(c *gin.Context) {
		ip := c.ClientIP() //获取客户端ip
		now := time.Now()
		mu.Lock()
		// exists表示是否查到ip所对应的值 v就是取到的count
		v, exists := visitors[ip]
		if !exists || now.After(v.expiresAt) {
			visitors[ip] = &visitor{
				count:     1,
				expiresAt: now.Add(window),
			}
			mu.Unlock()
			c.Next()
			return
		}
		v.count++
		if v.count > limit {
			mu.Unlock()
			response.TooManyRequests(c, "请求过于频繁，请稍后重试")
			c.Abort()
			return
		}
		mu.Unlock()
		c.Next()
	}
}
