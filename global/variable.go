package global

import (
	"blog/model/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	Config      = new(config.Config)
	RedisClient *redis.Client
	DB          *gorm.DB
)
