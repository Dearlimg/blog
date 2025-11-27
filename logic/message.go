package logic

import (
	"blog/dao"
	mysql2 "blog/dao/mysql/sqlc"
	"blog/model/reply"
	"blog/model/request"
	"blog/service"
	"database/sql"
	"github.com/Dearlimg/Goutils/pkg/app/errcode"
	"github.com/gin-gonic/gin"
	"log"
)

type message struct{}

// GetMessage 获取评论列表（实现旁路缓存策略）
func (message) GetMessage(ctx *gin.Context) ([]reply.ReplyMessage, errcode.Err) {
	cacheService := service.GetCacheService()

	// 1. 先查缓存（Cache-Aside Pattern）
	cachedMessages, err := cacheService.GetMessageList()
	if err == nil && cachedMessages != nil && len(cachedMessages) > 0 {
		// 缓存命中，直接返回
		log.Printf("Cache hit for message list, returning %d messages", len(cachedMessages))
		return cachedMessages, nil
	}

	// 2. 缓存未命中或出错，查数据库
	log.Printf("Cache miss or error, querying database")
	rly, err := dao.Database.DB.GetMessage(ctx)
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
			Content:   msg.Content.String,
			Create_at: msg.CreateAt.Time,
		})
	}

	// 4. 异步写入缓存（不阻塞返回，即使失败也不影响主流程）
	go func() {
		if err := cacheService.SetMessageList(result); err != nil {
			log.Printf("Failed to set cache (non-blocking): %v", err)
		} else {
			log.Printf("Cache updated successfully with %d messages", len(result))
		}
	}()

	return result, nil
}

type CreateMessageParams struct {
	Name    string
	Email   string
	Content sql.NullString
}

// PostMessage 创建新评论（实现旁路缓存策略：写入后删除缓存）
func (m *message) PostMessage(ctx *gin.Context, param *request.ParamCreateMessage) errcode.Err {
	// 校验输入参数
	if param == nil {
		return errcode.ErrParamsNotValid.WithDetails("nil parameter")
	}

	// 正确初始化结构体
	params := &CreateMessageParams{
		Name:    param.Name,
		Email:   param.Email,
		Content: sql.NullString{Valid: true, String: param.Content},
	}

	// 转换数据库参数
	dbParams := mysql2.CreateMessageParams{
		Name:    params.Name,
		Email:   params.Email,
		Content: params.Content,
	}

	// 1. 先写入数据库
	if err := dao.Database.DB.CreateMessage(ctx, &dbParams); err != nil {
		log.Printf("CreateMessage failed: %v", err)
		return errcode.ErrServer
	}

	// 2. 删除缓存（使缓存失效，下次查询时会重新从数据库加载）
	cacheService := service.GetCacheService()
	go func() {
		if err := cacheService.DeleteMessageList(); err != nil {
			log.Printf("Failed to delete cache: %v", err)
		} else {
			log.Printf("Cache invalidated successfully after creating new message")
		}
	}()

	return nil
}

//func (message) PostMessage(ctx *gin.Context, param *request.ParamCreateMessage) errcode.Err {
//	var params *CreateMessageParams
//	fmt.Println(param, param.Content)
//	params.Name = param.Name
//	params.Email = param.Email
//	params.Content = sql.NullString{
//		Valid:  true,
//		String: param.Content,
//	}
//	fmt.Println(params)
//	err := dao.Database.DB.CreateMessage(ctx, (*mysql2.CreateMessageParams)(params))
//	if err != nil {
//		return errcode.ErrServer
//	}
//	return nil
//}
