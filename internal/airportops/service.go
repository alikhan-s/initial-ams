package airportops

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type Service struct {
	repo   *Repository
	logger *slog.Logger
}

func NewService(repo *Repository, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// CreateGate creates a new gate in a specific terminal.
func (s *Service) CreateGate(ctx context.Context, terminalID int64, code string) (*Gate, error) {
	// 1. Input Validation
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("gate code cannot be empty")
	}
	if terminalID <= 0 {
		return nil, fmt.Errorf("invalid terminal ID")
	}

	// Construct Domain Object
	// Status defaults to "OPEN" as per database schema default
	gate := &Gate{
		TerminalID: terminalID,
		Code:       code,
		Status:     "OPEN",
	}

	// 3. Persist
	if err := s.repo.CreateGate(ctx, gate); err != nil {
		s.logger.Error("failed to create gate", "terminal_id", terminalID, "code", code, "error", err)
		return nil, err // Repo should handle unique constraint violation (duplicate code)
	}

	s.logger.Info("gate created", "id", gate.ID, "code", gate.Code, "terminal_id", gate.TerminalID)
	return gate, nil
}

func (s *Service) AssignGate(ctx context.Context, flightID, gateID int64) error {
	// Future: Check if gate is actually free at that time
	if err := s.repo.AssignGateToFlight(ctx, flightID, gateID); err != nil {
		s.logger.Error("failed to assign gate", "flight_id", flightID, "err", err)
		return err
	}
	s.logger.Info("gate assigned", "flight_id", flightID, "gate_id", gateID)
	return nil
}
