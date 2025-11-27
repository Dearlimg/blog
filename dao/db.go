package dao

import (
	"blog/dao/mysql/message"

	"gorm.io/gorm"
)

type database struct {
	DB      *gorm.DB
	Message *message.DAO
}

var Database = new(database)
