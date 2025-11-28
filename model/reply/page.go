package reply

// PageInfo 分页信息
type PageInfo struct {
	Page       int   `json:"page"`        // 当前页码
	PageSize   int   `json:"page_size"`   // 每页数量
	Total      int64 `json:"total"`       // 总记录数
	TotalPages int   `json:"total_pages"` // 总页数
}

// CalculateTotalPages 计算总页数
func (p *PageInfo) CalculateTotalPages() {
	if p.Total > 0 && p.PageSize > 0 {
		p.TotalPages = int((p.Total + int64(p.PageSize) - 1) / int64(p.PageSize))
	} else {
		p.TotalPages = 0
	}
}

// NewPageInfo 创建分页信息
func NewPageInfo(page, pageSize int, total int64) *PageInfo {
	pageInfo := &PageInfo{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
	pageInfo.CalculateTotalPages()
	return pageInfo
}
