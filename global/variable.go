package global

import (
	"blog/model/config"
	"github.com/redis/go-redis/v9"
)

var (
	Config      = new(config.Config)
	RedisClient *redis.Client
)
