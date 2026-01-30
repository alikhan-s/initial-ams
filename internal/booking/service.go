package booking

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"initial-airport-management-system/internal/flight"
	"initial-airport-management-system/internal/platform/database"

	"github.com/jackc/pgx/v5"
)

type TicketRepository interface {
	Create(ctx context.Context, t *Ticket) error
}

type Service struct {
	repo       TicketRepository
	flightRepo *flight.Repository // We need flight data
	txManager  *database.TxManager
	logger     *slog.Logger
}

func NewService(repo TicketRepository, flightRepo *flight.Repository, tx *database.TxManager, logger *slog.Logger) *Service {
	return &Service{
		repo:       repo,
		flightRepo: flightRepo,
		txManager:  tx,
		logger:     logger,
	}
}

func (s *Service) BookTicket(ctx context.Context, flightID, passengerID int64) (*Ticket, error) {
	var ticket *Ticket

	// Execute business logic inside a transaction
	err := s.txManager.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Check if Flight Exists (using the Flight Repo inside the TX)
		// Ideally, flightRepo methods should accept the 'tx' object.
		// For this milestone, we often query normally, but for strict consistency:
		// We'd pass 'tx' to flightRepo. But finding by ID is usually safe to do outside lock unless checking capacity.

		f, err := s.flightRepo.FindByID(ctx, flightID)
		if err != nil {
			return errors.New("flight not found")
		}

		if f.Status == flight.StatusCancelled || f.Status == flight.StatusDeparted {
			return errors.New("cannot book closed flight")
		}

		// 2. Create the Ticket Object
		ticket = &Ticket{
			FlightID:    flightID,
			PassengerID: passengerID,
			Price:       150.00, // Placeholder price logic
			Status:      "ACTIVE",
			CreatedAt:   time.Now(),
		}

		// 3. Persist Ticket (Passing the TX to ensure atomicity)
		// We need to cast our generic 'tx' to the specific interface the repo expects
		// This requires the Repo to accept the DBTX interface we defined earlier.
		if err := s.repo.Create(ctx, ticket); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("booking failed", "error", err)
		return nil, err
	}

	s.logger.Info("ticket created", "id", ticket.ID)
	return ticket, nil
}
