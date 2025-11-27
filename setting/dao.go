package setting

import (
	"blog/dao"
	"blog/dao/mysql"
	msgdao "blog/dao/mysql/message"
	"blog/global"
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

	fmt.Println("Database connected successfully with GORM")
}
