package logic

import (
	"blog/dao"
	"blog/global"
	"blog/model/entity"
	"blog/model/reply"
	"blog/model/request"
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

	// 构建缓存 key
	cacheKey := getMessageCacheKey(page, pageSize)

	// 1. 先尝试从缓存获取
	if global.RedisClient != nil {
		cachedData, err := global.RedisClient.Get(cacheCtx, cacheKey).Result()
		if err == nil {
			// 缓存命中
			var cachedResult struct {
				Result   *reply.MessageListResponse `json:"result"`
				PageInfo *reply.PageInfo            `json:"page_info"`
			}
			if err := json.Unmarshal([]byte(cachedData), &cachedResult); err == nil {
				logger.DebugWithCtx(ctx, "GetMessage cache hit",
					logger.String("cache_key", cacheKey),
				)
				return cachedResult.Result, cachedResult.PageInfo, nil
			}
			logger.WarnWithCtx(ctx, "Failed to unmarshal cached data",
				logger.ErrorField(err),
			)
		} else if err != redis.Nil {
			logger.WarnWithCtx(ctx, "Failed to get cache",
				logger.ErrorField(err),
			)
		}
	}

	// 2. 缓存未命中，从数据库查询
	messages, err := dao.Database.Message.GetMessage(ctx, page, pageSize)
	if err != nil {
		logger.ErrorWithCtx(ctx, "GetMessage failed",
			logger.ErrorField(err),
		)
		return nil, nil, errcode.ErrServer
	}

	// 3. 统计总数
	total, err := dao.Database.Message.CountMessage(ctx)
	if err != nil {
		logger.ErrorWithCtx(ctx, "CountMessage failed",
			logger.ErrorField(err),
		)
		return nil, nil, errcode.ErrServer
	}

	// 4. 转换数据格式
	list := make([]reply.ReplyMessage, 0, len(messages))
	for _, msg := range messages {
		list = append(list, reply.ReplyMessage{
			Id:        msg.ID,
			Name:      msg.Name,
			Email:     msg.Email,
			Content:   msg.Content,
			Create_at: msg.CreateAt,
		})
	}

	// 5. 构建分页信息
	pageInfo := reply.NewPageInfo(page, pageSize, total)

	// 6. 构建响应
	result := &reply.MessageListResponse{
		List: list,
	}

	// 7. 写入缓存（异步，不影响主流程）
	if global.RedisClient != nil {
		go func() {
			cachedData := struct {
				Result   *reply.MessageListResponse `json:"result"`
				PageInfo *reply.PageInfo            `json:"page_info"`
			}{
				Result:   result,
				PageInfo: pageInfo,
			}
			jsonData, err := json.Marshal(cachedData)
			if err == nil {
				if err := global.RedisClient.Set(cacheCtx, cacheKey, jsonData, cacheExpiration).Err(); err != nil {
					logger.Warn("Failed to set cache",
						logger.ErrorField(err),
						logger.String("cache_key", cacheKey),
					)
				} else {
					logger.Debug("Cache set successfully",
						logger.String("cache_key", cacheKey),
					)
				}
			}
		}()
	}

	logger.DebugWithCtx(ctx, "GetMessage success",
		logger.Int("count", len(list)),
		logger.Int64("total", total),
	)

	return result, pageInfo, nil
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
	// 缓存过期时间（5分钟）
	cacheExpiration = 5 * time.Minute
)

var cacheCtx = context.Background()

// getMessageCacheKey 获取消息列表缓存 key
func getMessageCacheKey(page, pageSize int) string {
	return fmt.Sprintf("%s:page:%d:size:%d", messageListCacheKeyPrefix, page, pageSize)
}

// deleteMessageListCache 删除评论列表缓存（缓存失效，删除所有分页缓存）
// 如果 Redis 不可用，静默失败（不影响主流程）
func deleteMessageListCache() error {
	if global.RedisClient == nil {
		// Redis 未初始化，静默失败
		return nil
	}

	// 使用 pattern 删除所有相关的分页缓存
	pattern := messageListCacheKeyPrefix + ":*"
	iter := global.RedisClient.Scan(cacheCtx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(cacheCtx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan cache keys: %w", err)
	}

	if len(keys) > 0 {
		if err := global.RedisClient.Del(cacheCtx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete cache: %w", err)
		}
	}

	return nil
}
