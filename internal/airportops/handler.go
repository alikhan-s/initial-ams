package airportops

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateGateParams defines the JSON payload for creating a gate
type CreateGateParams struct {
	TerminalID int64  `json:"terminal_id"`
	Code       string `json:"code"`
}

// CreateGate handles POST /gates
func (h *Handler) CreateGate(w http.ResponseWriter, r *http.Request) {
	var req CreateGateParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	gate, err := h.service.CreateGate(r.Context(), req.TerminalID, req.Code)
	if err != nil {
		if errors.Is(err, ErrGateAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict) // Return 409
			return
		}
		http.Error(w, "Failed to create gate: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(gate)
}

// AssignGateParams defines the JSON payload for assigning a gate
type AssignGateParams struct {
	FlightID int64 `json:"flight_id"`
	GateID   int64 `json:"gate_id"`
}

// AssignGate handles POST /flights/gate
func (h *Handler) AssignGate(w http.ResponseWriter, r *http.Request) {
	var req AssignGateParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.AssignGate(r.Context(), req.FlightID, req.GateID); err != nil {
		http.Error(w, "Failed to assign gate: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"gate assigned"}`))
}
