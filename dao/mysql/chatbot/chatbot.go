package chatbot

import (
	"blog/model/entity"
	"context"
	"fmt"

	"gorm.io/gorm"
)

// DAO Chatbot 数据访问对象
type DAO struct {
	db *gorm.DB
}

// NewDAO 创建 Chatbot DAO 实例
func NewDAO(db *gorm.DB) *DAO {
	return &DAO{db: db}
}

// CreateChatbot 创建聊天机器人
func (d *DAO) CreateChatbot(ctx context.Context, chatbot *entity.Chatbot) error {
	err := d.db.WithContext(ctx).Create(chatbot).Error
	if err != nil {
		return fmt.Errorf("failed to create chatbot: %w", err)
	}
	return nil
}

// GetChatbot 根据ID获取聊天机器人
func (d *DAO) GetChatbot(ctx context.Context, id int32) (*entity.Chatbot, error) {
	var chatbot entity.Chatbot
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&chatbot).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get chatbot: %w", err)
	}
	return &chatbot, nil
}

// ListChatbots 获取所有聊天机器人
func (d *DAO) ListChatbots(ctx context.Context) ([]*entity.Chatbot, error) {
	var chatbots []*entity.Chatbot
	err := d.db.WithContext(ctx).Order("created_at DESC").Find(&chatbots).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list chatbots: %w", err)
	}
	return chatbots, nil
}

// UpdateChatbot 更新聊天机器人
func (d *DAO) UpdateChatbot(ctx context.Context, chatbot *entity.Chatbot) error {
	err := d.db.WithContext(ctx).Model(chatbot).Updates(map[string]interface{}{
		"name":          chatbot.Name,
		"personality":   chatbot.Personality,
		"background":    chatbot.Background,
		"system_prompt": chatbot.SystemPrompt,
		"updated_at":    chatbot.UpdatedAt,
	}).Error
	if err != nil {
		return fmt.Errorf("failed to update chatbot: %w", err)
	}
	return nil
}

// DeleteChatbot 删除聊天机器人
func (d *DAO) DeleteChatbot(ctx context.Context, id int32) error {
	// 先删除关联的对话历史
	if err := d.db.WithContext(ctx).Where("chatbot_id = ?", id).Delete(&entity.ChatHistory{}).Error; err != nil {
		return fmt.Errorf("failed to delete chat history: %w", err)
	}

	// 删除聊天机器人
	err := d.db.WithContext(ctx).Delete(&entity.Chatbot{}, id).Error
	if err != nil {
		return fmt.Errorf("failed to delete chatbot: %w", err)
	}
	return nil
}

// AddChatHistory 添加对话历史
func (d *DAO) AddChatHistory(ctx context.Context, history *entity.ChatHistory) error {
	err := d.db.WithContext(ctx).Create(history).Error
	if err != nil {
		return fmt.Errorf("failed to add chat history: %w", err)
	}
	return nil
}

// GetChatHistory 获取对话历史（按时间升序，用于列表展示，支持分页）
func (d *DAO) GetChatHistory(ctx context.Context, chatbotID int32, page, pageSize int) ([]*entity.ChatHistory, error) {
	var history []*entity.ChatHistory
	offset := (page - 1) * pageSize
	query := d.db.WithContext(ctx).
		Where("chatbot_id = ?", chatbotID).
		Order("created_at ASC").
		Offset(offset).
		Limit(pageSize)
	err := query.Find(&history).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}
	return history, nil
}

// CountChatHistory 统计对话历史总数
func (d *DAO) CountChatHistory(ctx context.Context, chatbotID int32) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).
		Model(&entity.ChatHistory{}).
		Where("chatbot_id = ?", chatbotID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count chat history: %w", err)
	}
	return count, nil
}

// GetLatestChatHistory 获取最新的对话历史（按时间降序取前N条，然后反转以保持时间顺序，用于聊天上下文）
func (d *DAO) GetLatestChatHistory(ctx context.Context, chatbotID int32, limit int) ([]*entity.ChatHistory, error) {
	var history []*entity.ChatHistory
	query := d.db.WithContext(ctx).Where("chatbot_id = ?", chatbotID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&history).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get latest chat history: %w", err)
	}

	// 反转切片，使其按时间升序排列（从旧到新）
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

// DeleteChatHistory 删除对话历史
func (d *DAO) DeleteChatHistory(ctx context.Context, chatbotID int32) error {
	err := d.db.WithContext(ctx).Where("chatbot_id = ?", chatbotID).Delete(&entity.ChatHistory{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete chat history: %w", err)
	}
	return nil
}
