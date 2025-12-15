package cache

import (
	"blog/global"
	"context"
	"math/rand"
	"time"
)

// Init 初始化缓存系统
func Init() {
	// 初始化随机数生成器
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// GetRandomExpiration 获取随机过期时间
func GetRandomExpiration() time.Duration {
	minExpire, err1 := time.ParseDuration(global.Config.Cache.MinExpire)
	maxExpire, err2 := time.ParseDuration(global.Config.Cache.MaxExpire)

	// 如果解析失败，使用默认值
	if err1 != nil || err2 != nil || minExpire >= maxExpire {
		minExpire = 5 * time.Minute
		maxExpire = 10 * time.Minute
	}

	randomExpire := minExpire + time.Duration(rand.Int63n(int64(maxExpire-minExpire)))
	return randomExpire
}

// GetDefaultExpiration 获取默认过期时间
func GetDefaultExpiration() time.Duration {
	expire, err := time.ParseDuration(global.Config.Cache.DefaultExpire)
	if err != nil {
		expire = 5 * time.Minute
	}
	return expire
}

// GetHotDataExpiration 获取热点数据过期时间
func GetHotDataExpiration() time.Duration {
	expire, err := time.ParseDuration(global.Config.Cache.HotDataExpire)
	if err != nil {
		expire = 30 * time.Minute
	}
	return expire
}

// GetCacheVersion 获取缓存版本号
func GetCacheVersion(ctx context.Context, keyPrefix string) (int64, error) {
	if !global.Config.Cache.CacheVersionEnabled || global.RedisClient == nil {
		return 0, nil
	}

	versionKey := keyPrefix + ":version"
	version, err := global.RedisClient.Incr(ctx, versionKey).Result()
	if err != nil {
		return 0, err
	}

	// 设置版本号的过期时间，确保版本号不会一直存在
	global.RedisClient.Expire(ctx, versionKey, 24*time.Hour)

	return version, nil
}

// DeleteCacheByPattern 通过模式删除缓存
// 如果 Redis 不可用，静默失败
func DeleteCacheByPattern(ctx context.Context, pattern string) error {
	if global.RedisClient == nil {
		return nil
	}

	// 使用 Scan 命令查找匹配的键
	iter := global.RedisClient.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}

	// 如果有匹配的键，批量删除
	if len(keys) > 0 {
		if err := global.RedisClient.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}

	return nil
}

// IsCacheEnabled 检查缓存是否启用
func IsCacheEnabled() bool {
	return global.Config.Cache.Enabled && global.RedisClient != nil
}
