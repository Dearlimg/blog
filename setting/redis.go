package setting

import (
	"blog/global"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type redisConfig struct{}

var ctx = context.Background()

func (redisConfig) Init() {
	global.RedisClient = redis.NewClient(&redis.Options{
		Addr:     global.Config.Redis.Addr,
		Password: global.Config.Redis.Password,
		DB:       global.Config.Redis.DB,
	})

	// 测试连接
	_, err := global.RedisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Redis connection failed: %v\n", err)
		// 不 panic，允许服务继续运行（布隆过滤器会在中间件中处理错误）
	} else {
		fmt.Println("Redis connected successfully")
	}
}
