package flight

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Service struct {
	repo   *Repository
	logger *slog.Logger
}

func NewService(repo *Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

type CreateFlightParams struct {
	FlightNo      string `json:"flight_no"`
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureTime string `json:"departure_time"` // Maps JSON "departure_time" to this field
	ArrivalTime   string `json:"arrival_time"`   // Maps JSON "arrival_time" to this field
}

func (s *Service) CreateFlight(ctx context.Context, params CreateFlightParams) (*Flight, error) {
	// Parse Times
	dep, err := time.Parse(time.RFC3339, params.DepartureTime)
	if err != nil {
		return nil, fmt.Errorf("invalid departure time format: %w", err)
	}
	arr, err := time.Parse(time.RFC3339, params.ArrivalTime)
	if err != nil {
		return nil, fmt.Errorf("invalid arrival time format: %w", err)
	}

	// Construct Domain Model
	flight := &Flight{
		FlightNo:      params.FlightNo,
		Origin:        params.Origin,
		Destination:   params.Destination,
		DepartureTime: dep,
		ArrivalTime:   arr,
		Status:        StatusScheduled,
	}

	// Validate Business Rules
	if err := flight.Validate(); err != nil {
		s.logger.Warn("flight validation failed", "error", err)
		return nil, err
	}

	// Persist
	if err := s.repo.Create(ctx, flight); err != nil {
		s.logger.Error("failed to create flight", "error", err)
		return nil, err
	}

	s.logger.Info("flight created", "flight_no", flight.FlightNo, "id", flight.ID)
	return flight, nil
}

// UpdateStatus updates the flight status while respecting optimistic locking.
func (s *Service) UpdateStatus(ctx context.Context, id int64, status Status, version int) error {
	if status == StatusScheduled {
		return fmt.Errorf("cannot revert to SCHEDULED status once changed")
	}

	// Call Repo (which handles the version check)
	if err := s.repo.UpdateStatus(ctx, id, status, version); err != nil {
		if err == ErrConflict {
			return fmt.Errorf("conflict: flight has been modified by another user (refresh data)")
		}
		s.logger.Error("failed to update flight status", "id", id, "error", err)
		return err
	}

	s.logger.Info("flight status updated", "flight_id", id, "new_status", status)
	return nil
}
