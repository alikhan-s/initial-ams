package booking

import (
	"encoding/json"
	"initial-airport-management-system/internal/flight"
	"net/http"
	"strconv"
	"sync"
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

// Create handles POST /bookings
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req BookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ticket, err := h.service.BookTicket(r.Context(), req.FlightID, req.PassengerID)
	if err != nil {
		http.Error(w, "Failed to book ticket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

// Cancel handles POST /bookings/cancel?id=1
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.service.CancelTicket(r.Context(), id); err != nil {
		http.Error(w, "Failed to cancel ticket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"cancelled"}`))
}

// BookingDetailsResponse aggregates data from different sources
type BookingDetailsResponse struct {
	Ticket *Ticket        `json:"ticket"`
	Flight *flight.Flight `json:"flight,omitempty"`
}

// GetDetails handles GET /bookings/details?id=1
// It demonstrates using Goroutines to fetch related data (Flight) in parallel.
func (h *Handler) GetDetails(w http.ResponseWriter, r *http.Request) {
	// Parse ID
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Fetch Ticket (Synchronous - we need this to know WHICH flight to fetch)
	ticket, err := h.service.GetTicket(ctx, id)
	if err != nil {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	resp := BookingDetailsResponse{Ticket: ticket}

	// Parallel Fetching: Fetch Flight data in a Goroutine
	// Since BookingService has access to FlightRepo, we can use it.
	var wg sync.WaitGroup
	wg.Add(1)

	// Channel to capture errors from goroutines
	errChan := make(chan error, 1)

	go func() {
		defer wg.Done()
		// We access flightRepo via the service.
		// Note: In strict architecture, we might prefer a Service method like s.service.GetFlightInfo(...)
		// but accessing the repo here works for this structure.
		f, err := h.service.flightRepo.FindByID(ctx, ticket.FlightID)
		if err != nil {
			errChan <- err
			return
		}
		resp.Flight = f
	}()

	wg.Wait()
	close(errChan)

	// Check if the goroutine reported an error (optional: could just return partial data)
	if err := <-errChan; err != nil {
		// Log error in production, maybe return partial response
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
