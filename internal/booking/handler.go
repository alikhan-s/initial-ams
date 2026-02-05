package booking

import (
	"initial-airport-management-system/internal/flight"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type BookRequest struct {
	FlightID    int64 `json:"flight_id"`
	PassengerID int64 `json:"passenger_id"`
}

// Create handles POST /api/v1/bookings
func (h *Handler) Create(c *gin.Context) {
	var req BookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ticket, err := h.service.BookTicket(c.Request.Context(), req.FlightID, req.PassengerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

// Cancel handles POST /api/v1/bookings/:id/cancel
func (h *Handler) Cancel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.service.CancelTicket(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

// BookingDetailsResponse aggregates data
type BookingDetailsResponse struct {
	Ticket *Ticket        `json:"ticket"`
	Flight *flight.Flight `json:"flight,omitempty"`
}

// GetDetails handles GET /api/v1/bookings/:id
func (h *Handler) GetDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	ctx := c.Request.Context()

	ticket, err := h.service.GetTicket(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	resp := BookingDetailsResponse{Ticket: ticket}

	flightInf, err := h.service.flightRepo.FindByID(ctx, ticket.FlightID)
	if err != nil {
		slog.Warn("failed to fetch flight details", "err", err)
	} else {
		resp.Flight = flightInf
	}

	c.JSON(http.StatusOK, resp)
}
