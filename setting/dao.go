package setting

import (
	"blog/dao"
	"blog/dao/mysql"
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

	// 初始化 DAO
	dao.Database.DB = db
	dao.Database.Message = msgdao.NewDAO(db)

	// 使用 logger 记录（如果已初始化）
	// 注意：这里 logger 可能还没初始化，所以先检查
	if logger.GetLogger() != nil {
		logger.Info("Database connected successfully with GORM")
	}
}
