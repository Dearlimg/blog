package config

import "time"

type Config struct {
	App   AppConfig   `yaml:"app"`
	MySQL MySQLConfig `yaml:"mysql"`
}

type AppConfig struct {
	Name      string    `yaml:"name"`
	Version   string    `yaml:"version"`
	StartTime time.Time `yaml:"start_time"`
	Port      string    `yaml:"port"`
}

type MySQLConfig struct {
	DSN string `yaml:"dns"` // 注意 yaml 标签对应配置中的 "dns"
}
