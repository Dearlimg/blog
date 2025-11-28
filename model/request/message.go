package request

// ParamCreateMessage 创建消息的请求参数
type ParamCreateMessage struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Content string `json:"content" binding:"required"`
}

// ParamGetMessage 获取消息列表的请求参数（分页）
type ParamGetMessage struct {
	Page     int `form:"page" binding:"omitempty,min=1"`              // 页码，从1开始，默认1
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"` // 每页数量，默认10，最大100
}

// GetPage 获取页码，默认1
func (p *ParamGetMessage) GetPage() int {
	if p.Page <= 0 {
		return 1
	}
	return p.Page
}

// GetPageSize 获取每页数量，默认10
func (p *ParamGetMessage) GetPageSize() int {
	if p.PageSize <= 0 {
		return 10
	}
	if p.PageSize > 100 {
		return 100
	}
	return p.PageSize
}

// GetOffset 获取偏移量
func (p *ParamGetMessage) GetOffset() int {
	return (p.GetPage() - 1) * p.GetPageSize()
}
