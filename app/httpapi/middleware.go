package httpapi

import "github.com/gin-gonic/gin"

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Writer.Written() {
			return
		}
		if len(c.Errors) > 0 {
			JSONError(c, c.Errors.Last().Err)
		}
	}
}
