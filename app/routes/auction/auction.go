package auction

import (
	controller "dungeons/app/controllers/auction"

	"github.com/gin-gonic/gin"
)

func SetupRouter(v1 *gin.RouterGroup, handler *controller.Handler, authMiddleware gin.HandlerFunc) {
	group := v1.Group("/auction")
	{
		group.GET("/listings", handler.ListActive)
		group.POST("/listings", authMiddleware, handler.CreateListing)
		group.POST("/listings/:id/buy", authMiddleware, handler.Buy)
		group.POST("/listings/:id/cancel", authMiddleware, handler.Cancel)
	}
}
