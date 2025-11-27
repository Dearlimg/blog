package logic

import (
	"blog/dao"
	"blog/model/entity"
	"blog/model/reply"
	"blog/model/request"
	"blog/pkg/logger"
	"blog/service"
	"github.com/Dearlimg/Goutils/pkg/app/errcode"
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
		logger.Debug("Cache hit for message list",
			logger.Int("count", len(cachedMessages)),
		)
		return cachedMessages, nil
	}

	// 2. 缓存未命中或出错，查数据库
	logger.Debug("Cache miss, querying database")
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
	go func() {
		if err := cacheService.SetMessageList(result); err != nil {
			logger.Warn("Failed to set cache (non-blocking)",
				logger.ErrorField(err),
			)
		} else {
			logger.Debug("Cache updated successfully",
				logger.Int("message_count", len(result)),
			)
		}
	}()

	return result, nil
}

// PostMessage 创建新评论（实现旁路缓存策略：写入后删除缓存）
func (m *message) PostMessage(ctx *gin.Context, param *request.ParamCreateMessage) errcode.Err {
	// 校验输入参数
	if param == nil {
		return errcode.ErrParamsNotValid.WithDetails("nil parameter")
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
		logger.Error("CreateMessage failed",
			logger.ErrorField(err),
			logger.String("name", param.Name),
			logger.String("email", param.Email),
		)
		return errcode.ErrServer
	}

	logger.Info("Message created successfully",
		logger.String("name", param.Name),
		logger.String("email", param.Email),
	)

	// 2. 删除缓存（使缓存失效，下次查询时会重新从数据库加载）
	cacheService := service.GetCacheService()
	go func() {
		if err := cacheService.DeleteMessageList(); err != nil {
			logger.Warn("Failed to delete cache",
				logger.ErrorField(err),
			)
		} else {
			logger.Debug("Cache invalidated successfully after creating new message")
		}
	}()

	return nil
}
