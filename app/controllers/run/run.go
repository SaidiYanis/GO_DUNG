package run

import (
	"dungeons/app/auth"
	"dungeons/app/httpapi"
	"dungeons/app/models"
	service "dungeons/app/services/run"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Start(c *gin.Context) {
	var req models.StartRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.JSONError(c, err)
		return
	}
	run, err := h.service.Start(c.Request.Context(), auth.PlayerID(c), req)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusCreated, run)
}

func (h *Handler) List(c *gin.Context) {
	params := httpapi.ParsePagination(c)
	runs, err := h.service.List(c.Request.Context(), auth.PlayerID(c), params)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, models.ListResponse[models.Run]{
		Data: runs,
		Pagination: models.Pagination{
			Page:  params.Page,
			Limit: params.Limit,
		},
	})
}

func (h *Handler) Get(c *gin.Context) {
	runID, err := httpapi.ParseID(c, "id")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	run, err := h.service.Get(c.Request.Context(), auth.PlayerID(c), runID)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, run)
}

func (h *Handler) Attempt(c *gin.Context) {
	runID, err := httpapi.ParseID(c, "id")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	stepID, err := httpapi.ParseID(c, "stepId")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	var req models.AttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.JSONError(c, err)
		return
	}
	attempt, err := h.service.Attempt(c.Request.Context(), auth.PlayerID(c), runID, stepID, req)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, attempt)
}
