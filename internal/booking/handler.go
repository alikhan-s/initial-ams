package booking

import (
	"encoding/json"
	"net/http"
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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req BookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ticket, err := h.service.BookTicket(r.Context(), req.FlightID, req.PassengerID)
	if err != nil {
		// Production Note: Don't leak internal DB errors.
		// Log the real error, return a generic message to user.
		http.Error(w, "Failed to book ticket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}
