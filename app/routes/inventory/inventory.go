package inventory

import (
	controller "dungeons/app/controllers/inventory"

	"github.com/gin-gonic/gin"
)

func SetupRouter(v1 *gin.RouterGroup, handler *controller.Handler, authMiddleware gin.HandlerFunc) {
	v1.GET("/inventory", authMiddleware, handler.GetInventory)
}
