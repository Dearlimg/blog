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

// ParamGetChatHistory 获取对话历史的请求参数（分页）
type ParamGetChatHistory struct {
	Page     int `form:"page" binding:"omitempty,min=1"`              // 页码，从1开始，默认1
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"` // 每页数量，默认10，最大100
}

// GetPage 获取页码，默认1
func (p *ParamGetChatHistory) GetPage() int {
	if p.Page <= 0 {
		return 1
	}
	return p.Page
}

// GetPageSize 获取每页数量，默认10
func (p *ParamGetChatHistory) GetPageSize() int {
	if p.PageSize <= 0 {
		return 10
	}
	if p.PageSize > 100 {
		return 100
	}
	return p.PageSize
}

// GetOffset 获取偏移量
func (p *ParamGetChatHistory) GetOffset() int {
	return (p.GetPage() - 1) * p.GetPageSize()
}
