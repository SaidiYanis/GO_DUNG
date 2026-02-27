package player

import (
	"dungeons/app/auth"
	"dungeons/app/httpapi"
	"dungeons/app/models"
	service "dungeons/app/services/player"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.JSONError(c, err)
		return
	}
	resp, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusCreated, resp)
}

func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.JSONError(c, err)
		return
	}
	resp, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, resp)
}

func (h *Handler) Me(c *gin.Context) {
	playerID := auth.PlayerID(c)
	resp, err := h.service.Me(c.Request.Context(), playerID)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, resp)
}

func (h *Handler) List(c *gin.Context) {
	params := httpapi.ParsePagination(c)
	players, err := h.service.List(c.Request.Context(), params)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, models.ListResponse[models.PlayerResponse]{
		Data: players,
		Pagination: models.Pagination{
			Page:  params.Page,
			Limit: params.Limit,
		},
	})
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := httpapi.ParseID(c, "id")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	player, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, player)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := httpapi.ParseID(c, "id")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	var req models.UpdatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.JSONError(c, err)
		return
	}
	player, err := h.service.UpdateDisplayName(c.Request.Context(), id, req)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, player)
}
