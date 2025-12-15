package logic

import (
	"blog/dao"
	"blog/global"
	"blog/model/entity"
	"blog/model/reply"
	"blog/model/request"
	"blog/pkg/cache"
	"blog/pkg/errcode"
	"blog/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type message struct{}

// GetMessage 获取评论列表（分页，带缓存）
func (m message) GetMessage(ctx *gin.Context, param *request.ParamGetMessage) (*reply.MessageListResponse, *reply.PageInfo, errcode.Err) {
	page := param.GetPage()
	pageSize := param.GetPageSize()

	logger.DebugWithCtx(ctx, "GetMessage with pagination",
		logger.Int("page", page),
		logger.Int("page_size", pageSize),
	)

	// 尝试使用Redis有序集合实现高效分页
	if cache.IsCacheEnabled() {
		if result, pageInfo, err := m.getMessageFromRedis(ctx, page, pageSize); err == nil {
			return result, pageInfo, nil
		}
	}

	// 缓存未命中或Redis不可用，从数据库查询
	messages, err := dao.Database.Message.GetMessage(ctx, page, pageSize)
	if err != nil {
		logger.ErrorWithCtx(ctx, "GetMessage failed",
			logger.ErrorField(err),
		)
		return nil, nil, errcode.ErrServer
	}

	// 统计总数
	total, err := dao.Database.Message.CountMessage(ctx)
	if err != nil {
		logger.ErrorWithCtx(ctx, "CountMessage failed",
			logger.ErrorField(err),
		)
		return nil, nil, errcode.ErrServer
	}

	// 转换数据格式
	list := make([]reply.ReplyMessage, 0, len(messages))
	for _, msg := range messages {
		list = append(list, reply.ReplyMessage{
			Id:        msg.ID,
			Name:      msg.Name,
			Email:     msg.Email,
			Content:   msg.Content,
			Create_at: msg.CreateAt,
		})

		// 异步将单条消息写入缓存
		if cache.IsCacheEnabled() {
			go m.cacheMessage(ctx, msg)
		}
	}

	// 构建分页信息
	pageInfo := reply.NewPageInfo(page, pageSize, total)

	// 构建响应
	result := &reply.MessageListResponse{
		List: list,
	}

	// 更新Redis有序集合（异步）
	if cache.IsCacheEnabled() {
		go m.updateMessageIdsCache(ctx, messages)
	}

	logger.DebugWithCtx(ctx, "GetMessage success",
		logger.Int("count", len(list)),
		logger.Int64("total", total),
	)

	return result, pageInfo, nil
}

// getMessageFromRedis 从Redis获取评论列表
func (m message) getMessageFromRedis(ctx *gin.Context, page, pageSize int) (*reply.MessageListResponse, *reply.PageInfo, error) {
	// 1. 获取消息ID列表（使用有序集合的分页功能）
	idKey := getMessageIdsCacheKey()
	start := int64((page - 1) * pageSize)
	stop := int64(page*pageSize - 1)

	// ZRevRange：按时间倒序获取消息ID
	messageIDs, err := global.RedisClient.ZRevRange(ctx, idKey, start, stop).Result()
	if err != nil {
		logger.WarnWithCtx(ctx, "Failed to get message IDs from Redis",
			logger.ErrorField(err),
			logger.String("key", idKey),
		)
		return nil, nil, err
	}

	if len(messageIDs) == 0 {
		// Redis中没有消息ID列表，让主函数从数据库查询
		return nil, nil, fmt.Errorf("no message IDs in cache")
	}

	// 2. 获取消息总数
	total, err := global.RedisClient.ZCard(ctx, idKey).Result()
	if err != nil {
		logger.WarnWithCtx(ctx, "Failed to get message count from Redis",
			logger.ErrorField(err),
			logger.String("key", idKey),
		)
		// 总数获取失败，从数据库查询
		if total, err = dao.Database.Message.CountMessage(ctx); err != nil {
			return nil, nil, err
		}
	}

	// 3. 批量获取消息内容（使用Pipeline减少网络开销）
	pipeline := global.RedisClient.Pipeline()
	cmds := make([]*redis.StringCmd, len(messageIDs))

	for i, id := range messageIDs {
		messageKey := fmt.Sprintf("%s:%s", messageCacheKeyPrefix, id)
		cmds[i] = pipeline.Get(ctx, messageKey)
	}

	_, err = pipeline.Exec(ctx)
	if err != nil {
		logger.WarnWithCtx(ctx, "Failed to get messages from Redis pipeline",
			logger.ErrorField(err),
		)
		return nil, nil, err
	}

	// 4. 组装结果
	list := make([]reply.ReplyMessage, 0, len(messageIDs))
	for i, cmd := range cmds {
		if data, err := cmd.Result(); err == nil {
			var message entity.Message
			if err := json.Unmarshal([]byte(data), &message); err == nil {
				list = append(list, reply.ReplyMessage{
					Id:        message.ID,
					Name:      message.Name,
					Email:     message.Email,
					Content:   message.Content,
					Create_at: message.CreateAt,
				})
			}
		} else if err != redis.Nil {
			logger.WarnWithCtx(ctx, "Failed to get message from Redis",
				logger.ErrorField(err),
				logger.String("id", messageIDs[i]),
			)
		}
	}

	// 如果缓存的消息数量不足，返回失败，让主函数从数据库查询
	if len(list) < len(messageIDs) {
		return nil, nil, fmt.Errorf("insufficient cached messages")
	}

	// 构建分页信息
	pageInfo := reply.NewPageInfo(page, pageSize, total)

	logger.DebugWithCtx(ctx, "GetMessage from Redis",
		logger.Int("count", len(list)),
		logger.Int64("total", total),
	)

	return &reply.MessageListResponse{List: list}, pageInfo, nil
}

// cacheMessage 缓存单条评论消息
func (m message) cacheMessage(ctx context.Context, message *entity.Message) {
	if !cache.IsCacheEnabled() {
		return
	}

	messageKey := getMessageItemCacheKey(message.ID)
	jsonData, err := json.Marshal(message)
	if err != nil {
		logger.Warn("Failed to marshal message",
			logger.ErrorField(err),
			logger.Int32("message_id", message.ID),
		)
		return
	}

	// 使用随机过期时间
	expire := cache.GetRandomExpiration()
	if err := global.RedisClient.Set(ctx, messageKey, jsonData, expire).Err(); err != nil {
		logger.Warn("Failed to cache message",
			logger.ErrorField(err),
			logger.String("cache_key", messageKey),
		)
	}
}

// updateMessageIdsCache 更新评论消息ID列表缓存
func (m message) updateMessageIdsCache(ctx context.Context, messages []*entity.Message) {
	if !cache.IsCacheEnabled() || len(messages) == 0 {
		return
	}

	idKey := getMessageIdsCacheKey()
	pipe := global.RedisClient.Pipeline()

	for _, msg := range messages {
		// 分数使用消息ID（假设ID是递增的）或创建时间戳
		score := float64(msg.ID)
		pipe.ZAdd(ctx, idKey, redis.Z{Score: score, Member: msg.ID})
	}

	// 设置过期时间
	pipe.Expire(ctx, idKey, cache.GetHotDataExpiration())

	if _, err := pipe.Exec(ctx); err != nil {
		logger.Warn("Failed to update message IDs cache",
			logger.ErrorField(err),
			logger.String("cache_key", idKey),
		)
	}
}

// PostMessage 创建新评论（实现旁路缓存策略：写入后删除缓存）
func (m *message) PostMessage(ctx *gin.Context, param *request.ParamCreateMessage) (*reply.ReplyMessage, errcode.Err) {
	// 校验输入参数
	if param == nil {
		return nil, errcode.ErrParamsNotValid.WithDetails("nil parameter")
	}

	// 创建 GORM 模型
	message := &entity.Message{
		Name:     param.Name,
		Email:    param.Email,
		Content:  param.Content,
		CreateAt: time.Now(),
	}

	// 1. 先写入数据库
	if err := dao.Database.Message.CreateMessage(ctx, message); err != nil {
		logger.ErrorWithCtx(ctx, "CreateMessage failed",
			logger.ErrorField(err),
			logger.String("name", param.Name),
			logger.String("email", param.Email),
		)
		return nil, errcode.ErrServer
	}

	logger.InfoWithCtx(ctx, "Message created successfully",
		logger.String("name", param.Name),
		logger.String("email", param.Email),
		logger.Int32("id", message.ID),
	)

	// 2. 删除缓存（使缓存失效，下次查询时会重新从数据库加载）
	// 注意：goroutine 中无法直接使用 ctx，需要获取 request_id
	requestID := logger.GetRequestID(ctx)
	go func() {
		if err := deleteMessageListCache(); err != nil {
			logger.Warn("Failed to delete cache",
				logger.ErrorField(err),
				logger.String("request_id", requestID),
			)
		} else {
			logger.Debug("Cache invalidated successfully after creating new message",
				logger.String("request_id", requestID),
			)
		}
	}()

	// 3. 转换为响应格式并返回
	result := &reply.ReplyMessage{
		Id:        message.ID,
		Name:      message.Name,
		Email:     message.Email,
		Content:   message.Content,
		Create_at: message.CreateAt,
	}

	return result, nil
}

// 缓存相关常量和方法
const (
	// 评论列表缓存 key 前缀
	messageListCacheKeyPrefix = "cache:message:list"
	// 单条消息缓存 key 前缀
	messageCacheKeyPrefix = "cache:message:item"
	// 消息ID列表缓存 key 前缀
	messageIdsCacheKeyPrefix = "cache:message:ids"
)

var cacheCtx = context.Background()

// getMessageCacheKey 获取消息列表缓存 key（保留原方法兼容）
func getMessageCacheKey(page, pageSize int) string {
	return fmt.Sprintf("%s:page:%d:size:%d", messageListCacheKeyPrefix, page, pageSize)
}

// getMessageIdsCacheKey 获取消息ID列表缓存 key
func getMessageIdsCacheKey() string {
	return messageIdsCacheKeyPrefix
}

// getMessageItemCacheKey 获取单条消息缓存 key
func getMessageItemCacheKey(messageID int32) string {
	return fmt.Sprintf("%s:%d", messageCacheKeyPrefix, messageID)
}

// deleteMessageListCache 删除评论列表缓存（缓存失效，删除所有相关缓存）
// 如果 Redis 不可用，静默失败（不影响主流程）
func deleteMessageListCache() error {
	if !cache.IsCacheEnabled() {
		return nil
	}

	// 使用缓存工具包删除所有相关缓存
	patterns := []string{
		messageListCacheKeyPrefix + ":*",
		messageIdsCacheKeyPrefix,
		messageCacheKeyPrefix + ":*",
	}

	for _, pattern := range patterns {
		if err := cache.DeleteCacheByPattern(cacheCtx, pattern); err != nil {
			return fmt.Errorf("failed to delete cache: %w", err)
		}
	}

	return nil
}
