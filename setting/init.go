package setting

type group struct {
	Dao     database
	Message message
	Conf    conf
}

var Group = new(group)

func Init() {
	Group.Conf.Init()
	Group.Dao.Init()
}
