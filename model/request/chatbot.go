package request

// ParamCreateChatbot 创建聊天机器人请求参数
type ParamCreateChatbot struct {
	Name        string `json:"name" binding:"required"`        // 机器人名称
	Personality string `json:"personality" binding:"required"` // 性格设定
	Background  string `json:"background"`                     // 背景设定（可选）
}

// ParamUpdateChatbot 更新聊天机器人请求参数
type ParamUpdateChatbot struct {
	Name        string `json:"name,omitempty"`        // 机器人名称
	Personality string `json:"personality,omitempty"` // 性格设定
	Background  string `json:"background,omitempty"`  // 背景设定
}

// ParamChatbotChat 聊天机器人对话请求参数
type ParamChatbotChat struct {
	Message string `json:"message" binding:"required"` // 用户消息
	Stream  bool   `json:"stream,omitempty"`           // 是否使用流式响应，默认false
}
