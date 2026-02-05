package airportops

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type CreateGateParams struct {
	TerminalID int64  `json:"terminal_id"`
	Code       string `json:"code"`
}

// CreateGate handles POST /api/v1/gates
func (h *Handler) CreateGate(c *gin.Context) {
	var req CreateGateParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	gate, err := h.service.CreateGate(c.Request.Context(), req.TerminalID, req.Code)
	if err != nil {
		if errors.Is(err, ErrGateAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gate)
}

type AssignGateParams struct {
	FlightID int64 `json:"flight_id"`
	GateID   int64 `json:"gate_id"`
}

// AssignGate handles POST /api/v1/gates/assign
func (h *Handler) AssignGate(c *gin.Context) {
	var req AssignGateParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.service.AssignGate(c.Request.Context(), req.FlightID, req.GateID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "gate assigned"})
}
