package setting

import (
	"blog/dao"
	"blog/dao/mysql"
	"blog/global"
)

type database struct {
}

func (database) Init() {
	dao.Database.DB = mysql.Init(global.Config.MySQL.DSN)
}
