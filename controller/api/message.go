package api

import (
	"blog/logic"
	"blog/model/request"
	"blog/pkg/errcode"
	"blog/pkg/logger"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

type message struct{}

func (message) GetMessage(ctx *gin.Context) {
	rly := response.NewResponse(ctx)

	// 绑定分页参数
	param := &request.ParamGetMessage{}
	if err := ctx.ShouldBindQuery(param); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}

	// 调用逻辑层获取分页数据
	res, pageInfo, err := logic.Logics.Message.GetMessage(ctx, param)
	if err != nil {
		rly.Reply(err)
		return
	}

	// 返回带分页信息的响应
	rly.SuccessWithPage(res, pageInfo)
}

func (message) PostMessage(ctx *gin.Context) {
	rly := response.NewResponse(ctx)

	param := &request.ParamCreateMessage{}
	if err := ctx.ShouldBindJSON(param); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}
	logger.DebugWithCtx(ctx, "Received message creation request",
		logger.String("name", param.Name),
		logger.String("email", param.Email),
	)
	// 构造业务参数
	logicParam := &request.ParamCreateMessage{
		Name:    param.Name,
		Email:   param.Email,
		Content: param.Content,
	}

	res, err := logic.Logics.Message.PostMessage(ctx, logicParam)
	if err != nil {
		logger.ErrorWithCtx(ctx, "PostMessage failed",
			logger.ErrorField(err),
			logger.String("name", logicParam.Name),
		)
		rly.Reply(err)
		return
	}

	rly.Reply(nil, res)
}

//func (message) PostMessage(ctx *gin.Context) {
//	rly := app.NewResponse(ctx)
//	param := &request.ParamCreateMessage{}
//	if err := ctx.ShouldBindJSON(param); err != nil {
//		log.Println(err)
//		return
//	}
//	err := logic.Logics.Message.PostMessage(ctx, param)
//	if err != nil {
//		return
//	}
//	rly.Reply(err)
//}
