package logic

import (
	"blog/dao"
	mysql2 "blog/dao/mysql/sqlc"
	"blog/model/reply"
	"blog/model/request"
	"database/sql"
	"github.com/Dearlimg/Goutils/pkg/app/errcode"
	"github.com/gin-gonic/gin"
	"log"
)

type message struct{}

func (message) GetMessage(ctx *gin.Context) ([]reply.ReplyMessage, errcode.Err) {
	rly, err := dao.Database.DB.GetMessage(ctx)
	if err != nil {
		return nil, errcode.ErrServer
	}
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
	return result, nil
}

type CreateMessageParams struct {
	Name    string
	Email   string
	Content sql.NullString
}

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

	if err := dao.Database.DB.CreateMessage(ctx, &dbParams); err != nil {
		log.Printf("CreateMessage failed: %v", err)
		return errcode.ErrServer
	}

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
