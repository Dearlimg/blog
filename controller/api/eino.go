package api

import (
	"blog/global"
	"blog/logic"
	"blog/model/request"
	"blog/pkg/errcode"
	"blog/pkg/logger"
	"blog/pkg/response"
	"context"
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type eino struct{}

// CreateChatbot 创建聊天机器人
func (e *eino) CreateChatbot(ctx *gin.Context) {
	rly := response.NewResponse(ctx)

	param := &request.ParamCreateChatbot{}
	if err := ctx.ShouldBindJSON(param); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	logger.DebugWithCtx(ctx, "Received create chatbot request",
		logger.String("name", param.Name),
	)

	res, err := logic.Logics.Eino.CreateChatbot(ctx, param)
	if err != nil {
		logger.ErrorWithCtx(ctx, "CreateChatbot failed",
			logger.ErrorField(err),
		)
		rly.Reply(err)
		return
	}

	rly.Reply(nil, res)
}

// GetChatbot 获取聊天机器人
func (e *eino) GetChatbot(ctx *gin.Context) {
	rly := response.NewResponse(ctx)

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid chatbot id"))
		return
	}

	res, errCode := logic.Logics.Eino.GetChatbot(ctx, int32(id))
	if errCode != nil {
		rly.Reply(errCode)
		return
	}

	rly.Reply(nil, res)
}

// ListChatbots 获取所有聊天机器人
func (e *eino) ListChatbots(ctx *gin.Context) {
	rly := response.NewResponse(ctx)

	res, err := logic.Logics.Eino.ListChatbots(ctx)
	if err != nil {
		logger.ErrorWithCtx(ctx, "ListChatbots failed",
			logger.ErrorField(err),
		)
		rly.Reply(err)
		return
	}

	rly.Reply(nil, res)
}

// UpdateChatbot 更新聊天机器人
func (e *eino) UpdateChatbot(ctx *gin.Context) {
	rly := response.NewResponse(ctx)

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid chatbot id"))
		return
	}

	param := &request.ParamUpdateChatbot{}
	if err := ctx.ShouldBindJSON(param); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	res, errCode := logic.Logics.Eino.UpdateChatbot(ctx, int32(id), param)
	if errCode != nil {
		rly.Reply(errCode)
		return
	}

	rly.Reply(nil, res)
}

// DeleteChatbot 删除聊天机器人
func (e *eino) DeleteChatbot(ctx *gin.Context) {
	rly := response.NewResponse(ctx)

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid chatbot id"))
		return
	}

	errCode := logic.Logics.Eino.DeleteChatbot(ctx, int32(id))
	if errCode != nil {
		rly.Reply(errCode)
		return
	}

	rly.SuccessWithMessage("chatbot deleted successfully", nil)
}

// ChatbotChat 聊天机器人对话
func (e *eino) ChatbotChat(ctx *gin.Context) {
	rly := response.NewResponse(ctx)

	idStr := ctx.Param("id")
	chatbotID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid chatbot id"))
		return
	}

	param := &request.ParamChatbotChat{}
	if err := ctx.ShouldBindJSON(param); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	logger.DebugWithCtx(ctx, "Received chatbot chat request",
		logger.Int32("chatbot_id", int32(chatbotID)),
		logger.String("message", param.Message),
		logger.Bool("stream", param.Stream),
	)

	// 如果使用流式响应
	if param.Stream {
		e.handleChatbotStreamChat(ctx, int32(chatbotID), param)
		return
	}

	// 非流式响应
	res, errCode := logic.Logics.Eino.ChatbotChat(ctx, int32(chatbotID), param)
	if errCode != nil {
		logger.ErrorWithCtx(ctx, "ChatbotChat failed",
			logger.ErrorField(errCode),
		)
		rly.Reply(errCode)
		return
	}

	rly.Reply(nil, res)
}

// GetChatHistory 获取对话历史
func (e *eino) GetChatHistory(ctx *gin.Context) {
	rly := response.NewResponse(ctx)

	idStr := ctx.Param("id")
	chatbotID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails("invalid chatbot id"))
		return
	}

	// 绑定分页参数
	param := &request.ParamGetChatHistory{}
	if err := ctx.ShouldBindQuery(param); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	// 调用逻辑层获取分页数据
	res, pageInfo, errCode := logic.Logics.Eino.GetChatHistory(ctx, int32(chatbotID), param)
	if errCode != nil {
		rly.Reply(errCode)
		return
	}

	// 返回带分页信息的响应
	rly.SuccessWithPage(res, pageInfo)
}

// handleChatbotStreamChat 处理流式聊天
func (e *eino) handleChatbotStreamChat(ctx *gin.Context, chatbotID int32, param *request.ParamChatbotChat) {
	// 设置流式响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	// 获取 Eino StreamReader
	streamReader, err := logic.Logics.Eino.ChatbotChatStream(ctx, chatbotID, param)
	if err != nil {
		logger.ErrorWithCtx(ctx, "Failed to create stream reader",
			logger.ErrorField(err),
		)
		ctx.SSEvent("error", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	defer streamReader.Close()

	// 收集完整回复用于保存历史记录
	var fullResponse strings.Builder

	ctx.Stream(func(w io.Writer) bool {
		msg, err := streamReader.Recv()
		if err != nil {
			if err.Error() != "stream closed" && err.Error() != "EOF" {
				logger.ErrorWithCtx(ctx, "Stream read error",
					logger.ErrorField(err),
				)
				ctx.SSEvent("error", map[string]interface{}{
					"error": err.Error(),
				})
			}
			return false
		}

		// 提取消息内容
		content := ""
		if msg != nil {
			content = msg.Content
			fullResponse.WriteString(content)
		}

		if content != "" {
			// 获取模型名称
			modelName := global.Config.Ollama.Model
			if modelName == "" {
				modelName = "deepseek-r1:8b"
			}

			// 发送数据块
			ctx.SSEvent("message", map[string]interface{}{
				"chunk": content,
				"done":  false,
				"model": modelName,
			})
			return true
		}

		return false
	})

	// 获取模型名称
	modelName := global.Config.Ollama.Model
	if modelName == "" {
		modelName = "deepseek-r1:8b"
	}

	// 发送完成标记
	ctx.SSEvent("message", map[string]interface{}{
		"chunk": "",
		"done":  true,
		"model": modelName,
	})

	// 异步保存对话历史
	go func() {
		// 保存用户消息
		logic.Logics.Eino.SaveChatHistory(context.Background(), chatbotID, "user", param.Message)

		// 保存AI回复
		if fullResponse.Len() > 0 {
			logic.Logics.Eino.SaveChatHistory(context.Background(), chatbotID, "assistant", fullResponse.String())
		}
	}()
}
