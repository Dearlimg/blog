package routers

import (
	"blog/controller/api"
	"github.com/gin-gonic/gin"
)

type Eino struct {
}

func (e *Eino) Init(routers *gin.RouterGroup) {
	r := routers.Group("/chatbots")
	{
		// 聊天机器人管理
		r.POST("", api.Apis.Eino.CreateChatbot)       // 创建聊天机器人
		r.GET("", api.Apis.Eino.ListChatbots)         // 获取所有聊天机器人
		r.GET("/:id", api.Apis.Eino.GetChatbot)       // 获取指定聊天机器人
		r.PUT("/:id", api.Apis.Eino.UpdateChatbot)    // 更新聊天机器人
		r.DELETE("/:id", api.Apis.Eino.DeleteChatbot) // 删除聊天机器人

		// 对话相关
		r.POST("/:id/chat", api.Apis.Eino.ChatbotChat)      // 与聊天机器人对话
		r.GET("/:id/history", api.Apis.Eino.GetChatHistory) // 获取对话历史
	}
}
