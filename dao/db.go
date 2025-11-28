package dao

import (
	"blog/dao/mysql/chatbot"
	"blog/dao/mysql/message"

	"gorm.io/gorm"
)

type database struct {
	DB      *gorm.DB
	Message *message.DAO
	Chatbot *chatbot.DAO
}

var Database = new(database)
