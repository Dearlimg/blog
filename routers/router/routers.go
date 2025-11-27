package router

import (
	"blog/middleware"
	"blog/routers"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), middleware.CorsMiddleware())

	// 初始化 Redis 布隆过滤器
	if err := middleware.InitRedisBloomFilter(); err != nil {
		fmt.Printf("Failed to init Redis Bloom Filter: %v\n", err)
	}

	root := r.Group("api")
	{
		root.GET("ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
		rg := routers.Routers
		rg.Message.Init(root)
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
				fmt.Printf("Failed to add path to bloom filter: %s, error: %v\n", fullPath, err)
			} else {
				fmt.Printf("Added path to bloom filter: %s\n", fullPath)
			}
		}
	}
	fmt.Printf("Total paths added to bloom filter: %d\n", len(paths))
}
