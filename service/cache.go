package service

import (
	"blog/global"
	"blog/model/reply"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// 评论列表缓存 key
	messageListCacheKey = "cache:message:list"
	// 缓存过期时间（5分钟）
	cacheExpiration = 5 * time.Minute
)

var ctx = context.Background()

// CacheService 缓存服务
type CacheService struct {
	client *redis.Client
}

// NewCacheService 创建缓存服务实例
func NewCacheService() *CacheService {
	return &CacheService{
		client: global.RedisClient,
	}
}

// GetMessageList 从缓存获取评论列表
// 返回 (messages, error)
// - 如果缓存命中，返回 (messages, nil)
// - 如果缓存未命中，返回 (nil, nil)
// - 如果发生错误，返回 (nil, error)
func (c *CacheService) GetMessageList() ([]reply.ReplyMessage, error) {
	if c.client == nil {
		// Redis 未初始化，返回未命中（不返回错误，允许降级到数据库查询）
		return nil, nil
	}

	// 从 Redis 获取缓存
	data, err := c.client.Get(ctx, messageListCacheKey).Result()
	if err == redis.Nil {
		// 缓存不存在，返回 nil, nil（表示未命中，但不算错误）
		return nil, nil
	}
	if err != nil {
		// Redis 错误，返回 nil, nil（允许降级到数据库查询）
		return nil, nil
	}

	// 反序列化 JSON
	var messages []reply.ReplyMessage
	if err := json.Unmarshal([]byte(data), &messages); err != nil {
		// 反序列化失败，返回 nil, nil（允许降级到数据库查询）
		return nil, nil
	}

	return messages, nil
}

// SetMessageList 将评论列表写入缓存
// 如果 Redis 不可用，静默失败（不影响主流程）
func (c *CacheService) SetMessageList(messages []reply.ReplyMessage) error {
	if c.client == nil {
		// Redis 未初始化，静默失败
		return nil
	}

	// 序列化为 JSON
	data, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}

	// 写入 Redis，设置过期时间
	err = c.client.Set(ctx, messageListCacheKey, data, cacheExpiration).Err()
	if err != nil {
		// Redis 写入失败，返回错误（但不会影响主流程，因为已经在 goroutine 中执行）
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// DeleteMessageList 删除评论列表缓存（缓存失效）
// 如果 Redis 不可用，静默失败（不影响主流程）
func (c *CacheService) DeleteMessageList() error {
	if c.client == nil {
		// Redis 未初始化，静默失败
		return nil
	}

	err := c.client.Del(ctx, messageListCacheKey).Err()
	if err != nil && err != redis.Nil {
		// Redis 删除失败，返回错误（但不会影响主流程，因为已经在 goroutine 中执行）
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}

// GetCacheService 获取缓存服务实例（单例）
var cacheService *CacheService

func GetCacheService() *CacheService {
	if cacheService == nil {
		cacheService = NewCacheService()
	}
	return cacheService
}
