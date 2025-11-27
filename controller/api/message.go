package api

import (
	"blog/logic"
	"blog/model/request"
	"blog/pkg/logger"
	"github.com/Dearlimg/Goutils/pkg/app"
	"github.com/Dearlimg/Goutils/pkg/app/errcode"
	"github.com/gin-gonic/gin"
)

type message struct{}

func (message) GetMessage(ctx *gin.Context) {
	rly := app.NewResponse(ctx)
	res, err := logic.Logics.Message.GetMessage(ctx)
	if err != nil {
		return
	}
	rly.Reply(err, res)
}

func (message) PostMessage(ctx *gin.Context) {
	rly := app.NewResponse(ctx)

	param := &request.ParamCreateMessage{}
	if err := ctx.ShouldBindJSON(param); err != nil {
		rly.Reply(errcode.ErrParamsNotValid.WithDetails(err.Error()))
		return
	}
	logger.Debug("Received message creation request",
		logger.String("name", param.Name),
		logger.String("email", param.Email),
	)
	// 构造业务参数
	logicParam := &request.ParamCreateMessage{
		Name:    param.Name,
		Email:   param.Email,
		Content: param.Content,
	}

	if err := logic.Logics.Message.PostMessage(ctx, logicParam); err != nil {
		logger.Error("PostMessage failed",
			logger.ErrorField(err),
			logger.String("name", logicParam.Name),
		)
		rly.Reply(err)
		return
	}

	rly.Reply(nil)
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
