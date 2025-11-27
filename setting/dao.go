package setting

import (
	"blog/dao"
	"blog/dao/mysql"
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
	dao.Database.DB = db
	fmt.Println("Database connected successfully with GORM")
}
