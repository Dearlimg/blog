package middleware

import (
	"blog/global"
	"context"
	"hash/fnv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	// 布隆过滤器使用的 Redis key 前缀
	bloomFilterKey = "bloom:path:filter"
	// 路径精确匹配的 Redis set key
	pathSetKey = "bloom:path:set"
	// 位数组大小（建议根据路径数量调整，这里使用 100000 位）
	bloomSize = 100000
	// 哈希函数数量
	hashCount = 3
)

var (
	pathFilter *RedisBloomFilter
	once       sync.Once
	ctx        = context.Background()
)

// RedisBloomFilter 基于 Redis 的布隆过滤器
type RedisBloomFilter struct {
	client   *redis.Client
	size     uint
	hashFunc []hashFunc
}

// hashFunc 哈希函数类型
type hashFunc func([]byte) uint

// InitRedisBloomFilter 初始化 Redis 布隆过滤器
func InitRedisBloomFilter() error {
	var err error
	once.Do(func() {
		if global.RedisClient == nil {
			// 如果 Redis 客户端未初始化，创建新的连接
			global.RedisClient = redis.NewClient(&redis.Options{
				Addr:     global.Config.Redis.Addr,
				Password: global.Config.Redis.Password,
				DB:       global.Config.Redis.DB,
			})

			// 测试连接
			_, err = global.RedisClient.Ping(ctx).Result()
			if err != nil {
				return
			}
		}

		pathFilter = &RedisBloomFilter{
			client:   global.RedisClient,
			size:     bloomSize,
			hashFunc: make([]hashFunc, hashCount),
		}

		// 初始化哈希函数
		for i := 0; i < hashCount; i++ {
			seed := uint32(i + 1)
			// 使用立即执行函数避免闭包变量共享问题
			pathFilter.hashFunc[i] = func(s uint32) hashFunc {
				return func(data []byte) uint {
					h := fnv.New32a()
					h.Write(data)
					h.Write([]byte{byte(s)})
					return uint(h.Sum32()) % bloomSize
				}
			}(seed)
		}
	})
	return err
}

// AddPath 添加路径到 Redis 布隆过滤器
func AddPath(path string) error {
	if pathFilter == nil {
		if err := InitRedisBloomFilter(); err != nil {
			return err
		}
	}

	// 1. 添加到 Redis Set（用于精确匹配，避免误判）
	err := pathFilter.client.SAdd(ctx, pathSetKey, path).Err()
	if err != nil {
		return err
	}

	// 2. 添加到布隆过滤器（使用 Redis 位图）
	data := []byte(path)
	for _, fn := range pathFilter.hashFunc {
		index := fn(data)
		// 使用 SETBIT 设置位
		err = pathFilter.client.SetBit(ctx, bloomFilterKey, int64(index), 1).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// ContainsPath 检查路径是否存在
// 先检查 Redis Set（精确匹配），如果不存在再检查布隆过滤器
func ContainsPath(path string) (bool, error) {
	if pathFilter == nil {
		if err := InitRedisBloomFilter(); err != nil {
			return false, err
		}
	}

	// 1. 先检查 Redis Set（精确匹配，避免误判）
	exists, err := pathFilter.client.SIsMember(ctx, pathSetKey, path).Result()
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	// 2. 检查布隆过滤器
	// 如果布隆过滤器返回 false，说明一定不存在
	// 如果返回 true，可能存在（可能有误判）
	data := []byte(path)
	for _, fn := range pathFilter.hashFunc {
		index := fn(data)
		bit, err := pathFilter.client.GetBit(ctx, bloomFilterKey, int64(index)).Result()
		if err != nil {
			return false, err
		}
		if bit == 0 {
			// 如果任何一位为 0，说明一定不存在
			return false, nil
		}
	}

	// 所有位都为 1，可能存在（但可能有误判）
	// 由于我们已经用 Set 做了精确匹配，这里返回 true 表示可能存在
	return true, nil
}

// BloomFilterMiddleware 布隆过滤器中间件
// 用于快速过滤不存在的路径请求
func BloomFilterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 确保布隆过滤器已初始化
		if pathFilter == nil {
			if err := InitRedisBloomFilter(); err != nil {
				// 如果初始化失败，记录错误但继续处理请求
				c.Next()
				return
			}
		}

		path := c.Request.URL.Path
		method := c.Request.Method

		// 组合方法和路径作为完整路径标识
		fullPath := method + ":" + path

		// 检查路径是否在布隆过滤器中
		exists, err := ContainsPath(fullPath)
		if err != nil {
			// 如果 Redis 操作出错，记录错误但继续处理请求
			// 避免因为 Redis 故障导致整个服务不可用
			c.Next()
			return
		}

		if !exists {
			// 如果布隆过滤器判断不存在，直接返回404
			// 注意：由于我们使用了 Set 做精确匹配，这里返回 false 说明一定不存在
			c.JSON(404, gin.H{
				"error": "path not found",
				"path":  path,
			})
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next()
	}
}

// GetPathFilter 获取路径过滤器实例（用于测试或调试）
func GetPathFilter() *RedisBloomFilter {
	return pathFilter
}
