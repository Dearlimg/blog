package entity

import (
	"time"
)

// Message GORM 模型，对应数据库 message 表
type Message struct {
	ID       int32     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name     string    `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Email    string    `gorm:"column:email;type:varchar(100);not null" json:"email"`
	Content  string    `gorm:"column:content;type:varchar(1024)" json:"content"`
	CreateAt time.Time `gorm:"column:create_at;type:timestamp" json:"create_at"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "message"
}
