package inventory

import (
	"dungeons/app/auth"
	"dungeons/app/httpapi"
	service "dungeons/app/services/inventory"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) GetInventory(c *gin.Context) {
	inv, err := h.service.GetInventory(c.Request.Context(), auth.PlayerID(c))
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, inv)
}
