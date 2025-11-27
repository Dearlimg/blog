package config

type Config struct {
	App   AppConfig   `yaml:"app"`
	MySQL MySQLConfig `yaml:"mysql"`
	Redis RedisConfig `yaml:"redis"`
}

type AppConfig struct {
	Name      string `yaml:"name"`
	Version   string `yaml:"version"`
	StartTime string `yaml:"start_time"` // 使用字符串类型，避免解析问题
	Port      string `yaml:"port"`
}

type MySQLConfig struct {
	DSN string `yaml:"dns"` // 注意 yaml 标签对应配置中的 "dns"
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}
