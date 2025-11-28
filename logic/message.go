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
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type message struct{}

// GetMessage 获取评论列表（分页）
func (message) GetMessage(ctx *gin.Context, param *request.ParamGetMessage) (*reply.MessageListResponse, *reply.PageInfo, errcode.Err) {
	page := param.GetPage()
	pageSize := param.GetPageSize()

	logger.DebugWithCtx(ctx, "GetMessage with pagination",
		logger.Int("page", page),
		logger.Int("page_size", pageSize),
	)

	// 1. 查询数据列表
	messages, err := dao.Database.Message.GetMessage(ctx, page, pageSize)
	if err != nil {
		logger.ErrorWithCtx(ctx, "GetMessage failed",
			logger.ErrorField(err),
		)
		return nil, nil, errcode.ErrServer
	}

	// 2. 统计总数
	total, err := dao.Database.Message.CountMessage(ctx)
	if err != nil {
		logger.ErrorWithCtx(ctx, "CountMessage failed",
			logger.ErrorField(err),
		)
		return nil, nil, errcode.ErrServer
	}

	// 3. 转换数据格式
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

	// 4. 构建分页信息
	pageInfo := reply.NewPageInfo(page, pageSize, total)

	// 5. 构建响应
	result := &reply.MessageListResponse{
		List: list,
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
	// 评论列表缓存 key
	messageListCacheKey = "cache:message:list"
	// 缓存过期时间（5分钟）
	cacheExpiration = 5 * time.Minute
)

var cacheCtx = context.Background()

// deleteMessageListCache 删除评论列表缓存（缓存失效）
// 如果 Redis 不可用，静默失败（不影响主流程）
func deleteMessageListCache() error {
	if global.RedisClient == nil {
		// Redis 未初始化，静默失败
		return nil
	}

	err := global.RedisClient.Del(cacheCtx, messageListCacheKey).Err()
	if err != nil && err != redis.Nil {
		// Redis 删除失败，返回错误（但不会影响主流程，因为已经在 goroutine 中执行）
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}
