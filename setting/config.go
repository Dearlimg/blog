package setting

import (
	"blog/global"
)

type conf struct {
}

func (conf) Init() {
	global.Config.MySQL.DSN = "root:sta_go@tcp(47.118.19.28:3307)/blog?charset=utf8mb4&parseTime=True&loc=Local"
	global.Config.App.Name = "Blog"
	global.Config.App.Port = "0.0.0.0:8002"
	global.Config.App.Version = "v0.0.0"
	global.Config.Redis.Addr = "47.118.19.28:6379"
	global.Config.Redis.Password = "sta_go"
	global.Config.Redis.DB = 0
}
