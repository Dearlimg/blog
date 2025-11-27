package mysql

import (
	"blog/global"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDB 创建新的数据库连接
func NewDB(dsn string) (*gorm.DB, error) {
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

	return db, nil
}
