package message

import (
	"blog/model/entity"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DAO Message 数据访问对象
type DAO struct {
	db *gorm.DB
}

// NewDAO 创建 Message DAO 实例
func NewDAO(db *gorm.DB) *DAO {
	return &DAO{db: db}
}

// GetMessage 获取最新的5条评论
func (d *DAO) GetMessage(ctx context.Context) ([]*entity.Message, error) {
	var messages []*entity.Message

	err := d.db.WithContext(ctx).
		Where("id > ?", 0).
		Order("create_at DESC").
		Limit(5).
		Find(&messages).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

// CreateMessage 创建新评论
func (d *DAO) CreateMessage(ctx context.Context, message *entity.Message) error {
	// 设置创建时间
	if message.CreateAt.IsZero() {
		message.CreateAt = time.Now()
	}

	err := d.db.WithContext(ctx).Create(message).Error
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}
