package entity

import (
	"time"
)

// Chatbot 聊天机器人实体
type Chatbot struct {
	ID           int32     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Personality  string    `gorm:"column:personality;type:text" json:"personality"`     // 性格设定
	Background   string    `gorm:"column:background;type:text" json:"background"`       // 背景设定
	SystemPrompt string    `gorm:"column:system_prompt;type:text" json:"system_prompt"` // 系统提示词（自动生成）
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamp" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamp" json:"updated_at"`
}

// TableName 指定表名
func (Chatbot) TableName() string {
	return "chatbot"
}

// ChatHistory 对话历史实体
type ChatHistory struct {
	ID        int32     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ChatbotID int32     `gorm:"column:chatbot_id;not null;index" json:"chatbot_id"`
	Role      string    `gorm:"column:role;type:varchar(20);not null" json:"role"` // "user" 或 "assistant"
	Content   string    `gorm:"column:content;type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp" json:"created_at"`
}

// TableName 指定表名
func (ChatHistory) TableName() string {
	return "chat_history"
}
