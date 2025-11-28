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

// GetMessage 获取评论列表（分页）
func (d *DAO) GetMessage(ctx context.Context, page, pageSize int) ([]*entity.Message, error) {
	var messages []*entity.Message

	offset := (page - 1) * pageSize
	err := d.db.WithContext(ctx).
		Where("id > ?", 0).
		Order("create_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

// CountMessage 统计评论总数
func (d *DAO) CountMessage(ctx context.Context) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).
		Model(&entity.Message{}).
		Where("id > ?", 0).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
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
