package reply

import "time"

// ChatbotResponse 聊天机器人响应
type ChatbotResponse struct {
	ID           int32     `json:"id"`
	Name         string    `json:"name"`
	Personality  string    `json:"personality"`
	Background   string    `json:"background"`
	SystemPrompt string    `json:"system_prompt"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ChatbotListResponse 聊天机器人列表响应
type ChatbotListResponse struct {
	List []ChatbotResponse `json:"list"`
}

// ChatResponse 对话响应
type ChatResponse struct {
	Message   string    `json:"message"`   // AI 回复内容
	Duration  int64     `json:"duration"`  // 响应耗时（毫秒）
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// ChatHistoryResponse 对话历史响应
type ChatHistoryResponse struct {
	List []ChatHistoryItem `json:"list"`
}

// ChatHistoryItem 对话历史项
type ChatHistoryItem struct {
	ID        int32     `json:"id"`
	Role      string    `json:"role"`       // "user" 或 "assistant"
	Content   string    `json:"content"`    // 消息内容
	CreatedAt time.Time `json:"created_at"` // 创建时间
}
