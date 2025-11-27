package setting

type group struct {
	Dao     database
	Message message
	Conf    conf
	Redis   redisConfig
}

var Group = new(group)

func Init() {
	Group.Conf.Init()
	Group.Dao.Init()
	Group.Redis.Init()
}
