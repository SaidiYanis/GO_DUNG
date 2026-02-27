package dungeon

import (
	"dungeons/app/auth"
	"dungeons/app/httpapi"
	"dungeons/app/models"
	service "dungeons/app/services/dungeon"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CreateDungeon(c *gin.Context) {
	var req models.CreateDungeonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.JSONError(c, err)
		return
	}
	mjID := auth.PlayerID(c)
	d, err := h.service.CreateDungeon(c.Request.Context(), mjID, req)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusCreated, d)
}

func (h *Handler) UpdateDungeon(c *gin.Context) {
	dungeonID, err := httpapi.ParseID(c, "id")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	var req models.UpdateDungeonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.JSONError(c, err)
		return
	}
	d, err := h.service.UpdateDungeon(c.Request.Context(), auth.PlayerID(c), dungeonID, req)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, d)
}

func (h *Handler) PublishDungeon(c *gin.Context) {
	dungeonID, err := httpapi.ParseID(c, "id")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	d, err := h.service.PublishDungeon(c.Request.Context(), auth.PlayerID(c), dungeonID)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, d)
}

func (h *Handler) CreateStep(c *gin.Context) {
	dungeonID, err := httpapi.ParseID(c, "id")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	var req models.CreateBossStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.JSONError(c, err)
		return
	}
	step, err := h.service.CreateStep(c.Request.Context(), auth.PlayerID(c), dungeonID, req)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusCreated, step)
}

func (h *Handler) UpdateStep(c *gin.Context) {
	dungeonID, err := httpapi.ParseID(c, "id")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	stepID, err := httpapi.ParseID(c, "stepId")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	var req models.UpdateBossStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.JSONError(c, err)
		return
	}
	step, err := h.service.UpdateStep(c.Request.Context(), auth.PlayerID(c), dungeonID, stepID, req)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, step)
}

func (h *Handler) ReorderSteps(c *gin.Context) {
	dungeonID, err := httpapi.ParseID(c, "id")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	var req models.ReorderBossStepsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.JSONError(c, err)
		return
	}
	steps, err := h.service.ReorderSteps(c.Request.Context(), auth.PlayerID(c), dungeonID, req)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, steps)
}

func (h *Handler) ListPublished(c *gin.Context) {
	params := httpapi.ParsePagination(c)
	out, err := h.service.ListPublished(c.Request.Context(), params)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, models.ListResponse[models.Dungeon]{
		Data: out,
		Pagination: models.Pagination{
			Page:  params.Page,
			Limit: params.Limit,
		},
	})
}

func (h *Handler) GetPublished(c *gin.Context) {
	dungeonID, err := httpapi.ParseID(c, "id")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	d, steps, err := h.service.GetPublishedByID(c.Request.Context(), dungeonID)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, gin.H{"dungeon": d, "steps": steps})
}
