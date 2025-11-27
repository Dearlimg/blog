package setting

type group struct {
	Dao     database
	Message message
	Conf    conf
	Redis   redisConfig
	Log     logConfig
}

var Group = new(group)

func Init() {
	Group.Conf.Init()
	Group.Log.Init()
	Group.Dao.Init()
	Group.Redis.Init()
}
