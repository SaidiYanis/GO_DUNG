package run

import (
	controller "dungeons/app/controllers/run"

	"github.com/gin-gonic/gin"
)

func SetupRouter(v1 *gin.RouterGroup, handler *controller.Handler, authMiddleware gin.HandlerFunc) {
	runs := v1.Group("/runs")
	runs.Use(authMiddleware)
	{
		runs.POST("", handler.Start)
		runs.GET("", handler.List)
		runs.GET("/:id", handler.Get)
		runs.POST("/:id/steps/:stepId/attempt", handler.Attempt)
	}
}
