package router

import (
	"blog/middleware"
	"blog/pkg/logger"
	"blog/pkg/response"
	"blog/routers"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(gin.Recovery(), gin.Logger(), middleware.CorsMiddleware())

	if err := middleware.InitRedisBloomFilter(); err != nil {
		logger.Warn("Failed to init Redis Bloom Filter",
			logger.ErrorField(err),
		)
	}

	root := r.Group("api")
	{
		root.GET("ping", func(c *gin.Context) {
			rly := response.NewResponse(c)
			rly.SuccessWithMessage("pong", nil)
		})
		rg := routers.Routers
		rg.Message.Init(root)
		rg.Eino.Init(root)
	}
	collectRoutesAndAddToBloom(r)
	r.Use(middleware.BloomFilterMiddleware())
	return r
}

// collectRoutesAndAddToBloom 收集所有路由并添加到布隆过滤器
func collectRoutesAndAddToBloom(r *gin.Engine) {
	paths := make(map[string]bool)

	for _, route := range r.Routes() {
		fullPath := route.Method + ":" + route.Path
		fullPath = strings.ReplaceAll(fullPath, "//", "/")
		if !paths[fullPath] {
			paths[fullPath] = true
			if err := middleware.AddPath(fullPath); err != nil {
				logger.Warn("Failed to add path to bloom filter",
					logger.String("path", fullPath),
					logger.ErrorField(err),
				)
			} else {
				logger.Debug("Added path to bloom filter",
					logger.String("path", fullPath),
				)
			}
		}
	}
	logger.Info("Bloom filter initialized",
		logger.Int("total_paths", len(paths)),
	)
}
