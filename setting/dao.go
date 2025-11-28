package setting

import (
	"blog/dao"
	"blog/dao/mysql"
	chatbotdao "blog/dao/mysql/chatbot"
	msgdao "blog/dao/mysql/message"
	"blog/global"
	"blog/pkg/logger"
	"fmt"
)

type database struct {
}

func (database) Init() {
	db, err := mysql.NewDB(global.Config.MySQL.DSN)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	dao.Database.DB = db
	dao.Database.Message = msgdao.NewDAO(db)
	dao.Database.Chatbot = chatbotdao.NewDAO(db)

	if logger.GetLogger() != nil {
		logger.Info("Database connected successfully with GORM")
	}
}
