package common

import (
	"dungeons/app/httpapi"
	"dungeons/app/server"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	httpapi.JSON(c, http.StatusOK, gin.H{"status": "pong"})
}

func Version(c *gin.Context) {
	srv := server.GetServer()
	httpapi.JSON(c, http.StatusOK, gin.H{"version": srv.Version})
}
