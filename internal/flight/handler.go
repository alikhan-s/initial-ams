package flight

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create handles POST /api/v1/flights
func (h *Handler) Create(c *gin.Context) {
	var req CreateFlightParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	flight, err := h.service.CreateFlight(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, flight)
}

// Search handles GET /api/v1/flights?origin=GUW&destination=TSE&date=2025-10-25
func (h *Handler) Search(c *gin.Context) {
	origin := c.Query("origin")
	dest := c.Query("destination")
	dateStr := c.Query("date")

	if origin == "" || dest == "" || dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing origin, destination, or date"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format (use YYYY-MM-DD)"})
		return
	}

	flights, err := h.service.repo.Search(c.Request.Context(), SearchParams{
		Origin:      origin,
		Destination: dest,
		Date:        date,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, flights)
}

type UpdateStatusRequest struct {
	Status  Status `json:"status"`
	Version int    `json:"version"`
}

// UpdateStatus handles PATCH /api/v1/flights/:id/status
func (h *Handler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid flight ID"})
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), id, req.Status, req.Version); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
