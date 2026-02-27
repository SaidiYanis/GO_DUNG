package dungeon

import (
	"dungeons/app/auth"
	controller "dungeons/app/controllers/dungeon"

	"github.com/gin-gonic/gin"
)

func SetupRouter(v1 *gin.RouterGroup, handler *controller.Handler, authMiddleware gin.HandlerFunc) {
	mj := v1.Group("/mj")
	mj.Use(authMiddleware, auth.RequireRole("mj"))
	{
		dungeons := mj.Group("/dungeons")
		{
			dungeons.POST("", handler.CreateDungeon)
			dungeons.PUT("/:id", handler.UpdateDungeon)
			dungeons.POST("/:id/publish", handler.PublishDungeon)
			dungeons.POST("/:id/steps", handler.CreateStep)
			dungeons.PUT("/:id/steps/:stepId", handler.UpdateStep)
			dungeons.PUT("/:id/steps/reorder", handler.ReorderSteps)
		}
	}

	v1.GET("/dungeons", handler.ListPublished)
	v1.GET("/dungeons/:id", handler.GetPublished)
}
