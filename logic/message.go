package logic

import (
	"blog/dao"
	"blog/model/entity"
	"blog/model/reply"
	"blog/model/request"
	"blog/pkg/errcode"
	"blog/pkg/logger"
	"blog/service"
	"github.com/gin-gonic/gin"
	"time"
)

type message struct{}

// GetMessage 获取评论列表（实现旁路缓存策略）
func (message) GetMessage(ctx *gin.Context) ([]reply.ReplyMessage, errcode.Err) {
	cacheService := service.GetCacheService()

	// 1. 先查缓存（Cache-Aside Pattern）
	cachedMessages, err := cacheService.GetMessageList()
	if err == nil && cachedMessages != nil && len(cachedMessages) > 0 {
		// 缓存命中，直接返回
		logger.DebugWithCtx(ctx, "Cache hit for message list",
			logger.Int("count", len(cachedMessages)),
		)
		return cachedMessages, nil
	}

	// 2. 缓存未命中或出错，查数据库
	logger.DebugWithCtx(ctx, "Cache miss, querying database")
	rly, err := dao.Database.Message.GetMessage(ctx)
	if err != nil {
		return nil, errcode.ErrServer
	}

	// 3. 转换数据格式
	result := make([]reply.ReplyMessage, 0, len(rly))
	for _, msg := range rly {
		result = append(result, reply.ReplyMessage{
			Id:        msg.ID,
			Name:      msg.Name,
			Email:     msg.Email,
			Content:   msg.Content,
			Create_at: msg.CreateAt,
		})
	}

	// 4. 异步写入缓存（不阻塞返回，即使失败也不影响主流程）
	// 注意：goroutine 中无法直接使用 ctx，需要获取 request_id
	requestID := logger.GetRequestID(ctx)
	go func() {
		if err := cacheService.SetMessageList(result); err != nil {
			logger.Warn("Failed to set cache (non-blocking)",
				logger.ErrorField(err),
				logger.String("request_id", requestID),
			)
		} else {
			logger.Debug("Cache updated successfully",
				logger.Int("message_count", len(result)),
				logger.String("request_id", requestID),
			)
		}
	}()

	return result, nil
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
	cacheService := service.GetCacheService()
	// 注意：goroutine 中无法直接使用 ctx，需要获取 request_id
	requestID := logger.GetRequestID(ctx)
	go func() {
		if err := cacheService.DeleteMessageList(); err != nil {
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
