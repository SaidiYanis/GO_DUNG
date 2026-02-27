package main

import (
	controllers "dungeons/app/controllers/common"
	routes "dungeons/app/routes/common"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	router := routes.SetupRouter()
	router.GET("/ping", controllers.Ping)
	router.GET("/version", controllers.Version)
	return router
}
