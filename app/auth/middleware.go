package auth

import (
	apperrors "dungeons/app/errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CtxPlayerID = "playerID"
	CtxRole     = "role"
)

func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		head := c.GetHeader("Authorization")
		if !strings.HasPrefix(head, "Bearer ") {
			c.Error(fmt.Errorf("missing bearer token: %w", apperrors.ErrUnauthorized))
			c.Abort()
			return
		}

		claims, err := Parse(secret, strings.TrimPrefix(head, "Bearer "))
		if err != nil {
			c.Error(fmt.Errorf("invalid token: %w", apperrors.ErrUnauthorized))
			c.Abort()
			return
		}
		c.Set(CtxPlayerID, claims.Sub)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get(CtxRole)
		if !ok || v.(string) != role {
			c.Error(fmt.Errorf("role %s required: %w", role, apperrors.ErrForbidden))
			c.Abort()
			return
		}
		c.Next()
	}
}

func PlayerID(c *gin.Context) string {
	v, _ := c.Get(CtxPlayerID)
	if id, ok := v.(string); ok {
		return id
	}
	return ""
}
