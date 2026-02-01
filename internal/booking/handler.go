package booking

import (
	"encoding/json"
	"net/http"
	"strconv"
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

// GetFullDetails fetches Ticket + Flight + Passenger data.
func (h *Handler) GetFullDetails(w http.ResponseWriter, r *http.Request) {
	// Parse Ticket ID
	idStr := r.URL.Query().Get("id")
	// ... simple parsing logic ...
	id, _ := strconv.ParseInt(idStr, 10, 64)

	ctx := r.Context()

	// Fetch Ticket (We need this first to know WHICH flight/passenger to fetch)
	ticket, err := h.service.GetTicket(ctx, id)
	if err != nil {
		http.Error(w, "ticket not found", http.StatusNotFound)
		return
	}

	// Define the Result Structure
	type Response struct {
		Ticket    *Ticket  `json:"ticket"`
		Flight    any      `json:"flight"`
		Passenger any      `json:"passenger"`
		Errors    []string `json:"errors,omitempty"`
	}
	resp := Response{Ticket: ticket}

	// Parallel Fetching using Channels
	// We create buffered channels to prevent blocking if we exit early
	flightChan := make(chan any, 1)
	passChan := make(chan any, 1)
	errChan := make(chan error, 2)

	// Goroutine A: Fetch Flight
	go func() {
		// Ideally, we'd use h.service.flightService.GetFlight(ticket.FlightID)
		// Assuming we can access repo:
		f, err := h.service.flightRepo.FindByID(ctx, ticket.FlightID)
		if err != nil {
			errChan <- err
			return
		}
		flightChan <- f
	}()

	// Goroutine B: Fetch Passenger (Mocking the fetch here, normally calling PassengerService)
	go func() {
		// In a real monolith, you might inject PassengerService into BookingHandler
		// For now, we simulate a struct
		p := map[string]any{
			"id":     ticket.PassengerID,
			"status": "fetched_via_goroutine",
		}
		passChan <- p
	}()

	// Synchronization (Wait for both)
	// We wait for 2 results (flight + passenger) or timeout/errors.
	// Simple for-loop implementation:
	for i := 0; i < 2; i++ {
		select {
		case f := <-flightChan:
			resp.Flight = f
		case p := <-passChan:
			resp.Passenger = p
		case err := <-errChan:
			// Log error but maybe return partial content?
			resp.Errors = append(resp.Errors, err.Error())
		}
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
