package flight

import (
	"encoding/json"
	"net/http"
	"time"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create handles POST /flights
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateFlightParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	flight, err := h.service.CreateFlight(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(flight)
}

// Search handles GET /flights?origin=ALA&destination=TSE&date=2025-10-25
// This function uses the "time" package, which fixes your import error
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	origin := r.URL.Query().Get("origin")
	dest := r.URL.Query().Get("destination")
	dateStr := r.URL.Query().Get("date")

	if origin == "" || dest == "" || dateStr == "" {
		http.Error(w, "Missing origin, destination, or date", http.StatusBadRequest)
		return
	}

	// Usage of "time" package here
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	flights, err := h.service.repo.Search(r.Context(), SearchParams{
		Origin:      origin,
		Destination: dest,
		Date:        date,
	})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flights)
}
