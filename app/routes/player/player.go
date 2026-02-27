package player

import (
	"dungeons/app/auth"
	controller "dungeons/app/controllers/player"

	"github.com/gin-gonic/gin"
)

func SetupRouter(v1 *gin.RouterGroup, handler *controller.Handler, authMiddleware gin.HandlerFunc) {
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
	}

	v1.GET("/me", authMiddleware, handler.Me)

	players := v1.Group("/players")
	players.Use(authMiddleware, auth.RequireRole("mj"))
	{
		players.GET("", handler.List)
		players.GET("/:id", handler.GetByID)
		players.PUT("/:id", handler.Update)
	}
}
