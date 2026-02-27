package httpapi

import (
	apperrors "dungeons/app/errors"
	"dungeons/app/models"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParseID(c *gin.Context, key string) (string, error) {
	id := c.Param(key)
	if id == "" {
		return "", fmt.Errorf("missing path param %s: %w", key, apperrors.ErrValidation)
	}
	return id, nil
}

func ParsePagination(c *gin.Context) models.QueryParams {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
	return models.QueryParams{Page: page, Limit: limit}.Normalize()
}
