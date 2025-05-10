package setting

import (
	"blog/global"
)

type conf struct {
}

func (conf) Init() {
	global.Config.MySQL.DSN = "root:Gao12345@tcp(123.249.32.125:3307)/blog?charset=utf8mb4&parseTime=True&loc=Local"
	global.Config.App.Name = "Blog"
	global.Config.App.Port = "0.0.0.0:8002"
	global.Config.App.Version = "v0.0.0"
}
