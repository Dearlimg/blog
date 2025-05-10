package routers

import (
	"blog/controller/api"
	"github.com/gin-gonic/gin"
)

type message struct{}

func (message) Init(routers *gin.RouterGroup) {
	r := routers.Group("/message")
	{
		r.GET("/getmessage", api.Apis.Message.GetMessage)
		r.POST("/postmessage", api.Apis.Message.PostMessage)
	}
}
