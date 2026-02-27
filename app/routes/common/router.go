package common

import (
	apperrors "dungeons/app/errors"
	"dungeons/app/httpapi"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(httpapi.ErrorMiddleware())
	useCORS(router)
	noRoute(router)
	return router
}

func useCORS(r *gin.Engine) {
	r.Use(func(c *gin.Context) {
		allowOrigin := os.Getenv("ALLOW_ORIGIN")
		if allowOrigin == "" {
			allowOrigin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})
}

func noRoute(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		httpapi.JSONError(c, fmt.Errorf("resource not found: %w", apperrors.ErrNotFound))
	})
}
