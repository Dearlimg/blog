package mysql

import (
	"blog/global"
	"blog/model/entity"
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 数据库接口
type DB interface {
	GetMessage(ctx context.Context) ([]*entity.Message, error)
	CreateMessage(ctx context.Context, message *entity.Message) error
}

// gormDB GORM 数据库实现
type gormDB struct {
	db *gorm.DB
}

// NewDB 创建新的数据库实例
func NewDB(dsn string) (DB, error) {
	// 配置 GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 可以改为 logger.Silent 关闭日志
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 获取底层 sql.DB 进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(100)                // 最大打开连接数
	sqlDB.SetMaxIdleConns(20)                 // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // 连接最大存活时间
	sqlDB.SetConnMaxIdleTime(2 * time.Minute) // 空闲连接最大存活时间

	// 测试连接
	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database connection verification failed: %w", err)
	}

	// 保存到 global
	global.DB = db

	return &gormDB{db: db}, nil
}

// GetMessage 获取最新的5条评论
func (g *gormDB) GetMessage(ctx context.Context) ([]*entity.Message, error) {
	var messages []*entity.Message

	err := g.db.WithContext(ctx).
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
func (g *gormDB) CreateMessage(ctx context.Context, message *entity.Message) error {
	// 设置创建时间
	if message.CreateAt.IsZero() {
		message.CreateAt = time.Now()
	}

	err := g.db.WithContext(ctx).Create(message).Error
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}
