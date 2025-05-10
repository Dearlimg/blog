package api

import (
	"blog/logic"
	"blog/model/request"
	"fmt"
	"github.com/Dearlimg/Goutils/pkg/app"
	"github.com/Dearlimg/Goutils/pkg/app/errcode"
	"github.com/gin-gonic/gin"
)

type message struct{}

func (message) GetMessage(ctx *gin.Context) () {
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
	fmt.Println(param)
	// 构造业务参数
	logicParam := &request.ParamCreateMessage{
		Name:    param.Name,
		Email:   param.Email,
		Content: param.Content,
	}

	if err := logic.Logics.Message.PostMessage(ctx, logicParam); err != nil {
		rly.Reply(err)
		fmt.Println("postmessage logic", err)
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
