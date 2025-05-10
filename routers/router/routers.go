package router

import (
	"blog/middleware"
	"blog/routers"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), middleware.CorsMiddleware())
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
	return r
}
