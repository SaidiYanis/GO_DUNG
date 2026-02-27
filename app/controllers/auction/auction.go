package auction

import (
	"dungeons/app/auth"
	"dungeons/app/httpapi"
	"dungeons/app/models"
	service "dungeons/app/services/auction"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CreateListing(c *gin.Context) {
	var req models.CreateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.JSONError(c, err)
		return
	}
	listing, err := h.service.CreateListing(c.Request.Context(), auth.PlayerID(c), req)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusCreated, listing)
}

func (h *Handler) ListActive(c *gin.Context) {
	params := httpapi.ParsePagination(c)
	listings, err := h.service.ListActive(c.Request.Context(), params)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, models.ListResponse[models.Listing]{
		Data: listings,
		Pagination: models.Pagination{
			Page:  params.Page,
			Limit: params.Limit,
		},
	})
}

func (h *Handler) Buy(c *gin.Context) {
	listingID, err := httpapi.ParseID(c, "id")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	var req models.BuyListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.JSONError(c, err)
		return
	}
	listing, err := h.service.Buy(c.Request.Context(), auth.PlayerID(c), listingID, req)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, listing)
}

func (h *Handler) Cancel(c *gin.Context) {
	listingID, err := httpapi.ParseID(c, "id")
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	listing, err := h.service.Cancel(c.Request.Context(), auth.PlayerID(c), listingID)
	if err != nil {
		httpapi.JSONError(c, err)
		return
	}
	httpapi.JSON(c, http.StatusOK, listing)
}
