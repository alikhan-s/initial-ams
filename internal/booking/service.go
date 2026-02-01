package booking

import (
	"context"
	"errors"
	"fmt"
	"initial-airport-management-system/internal/flight"
	"initial-airport-management-system/internal/platform/database"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type TicketRepository interface {
	Create(ctx context.Context, t *Ticket) error
	GetSoldTicketsCount(ctx context.Context, flightID int64) (int, error)
	FindByID(ctx context.Context, id int64) (*Ticket, error)         // Added
	UpdateStatus(ctx context.Context, id int64, status string) error // Added
}

type Service struct {
	repo       TicketRepository
	flightRepo *flight.Repository
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

// BookTicket creates a ticket if seats are available
func (s *Service) BookTicket(ctx context.Context, flightID, passengerID int64) (*Ticket, error) {
	var ticket *Ticket

	// The 'tx' variable here represents the active transaction
	err := s.txManager.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Initialize repos that use THIS transaction
		// Note: We create lightweight repo instances that share the transaction connection
		txFlightRepo := flight.NewRepository(tx)
		txBookingRepo := NewRepository(tx) // Assumes NewRepository is exported in booking package

		// 2. Use the transactional repos for all logic
		f, err := txFlightRepo.FindByID(ctx, flightID)
		if err != nil {
			return errors.New("flight not found")
		}

		if f.Status == flight.StatusCancelled || f.Status == flight.StatusDeparted {
			return errors.New("cannot book closed flight")
		}

		// Check capacity using the TX repo (ensures consistency)
		soldCount, err := txBookingRepo.GetSoldTicketsCount(ctx, flightID)
		if err != nil {
			return fmt.Errorf("failed to check capacity: %w", err)
		}

		if soldCount >= f.TotalSeats {
			return errors.New("flight is fully booked")
		}

		ticket = &Ticket{
			FlightID:    flightID,
			PassengerID: passengerID,
			Price:       150.00,
			Status:      "ACTIVE",
			CreatedAt:   time.Now(),
		}

		// Create ticket using the TX repo
		return txBookingRepo.Create(ctx, ticket)
	})

	if err != nil {
		s.logger.Error("booking failed", "error", err)
		return nil, err
	}

	s.logger.Info("ticket created", "id", ticket.ID)
	return ticket, nil
}

// GetTicket retrieves a single ticket
func (s *Service) GetTicket(ctx context.Context, id int64) (*Ticket, error) {
	return s.repo.FindByID(ctx, id)
}

// CancelTicket cancels a booking
func (s *Service) CancelTicket(ctx context.Context, id int64) error {
	// In a real app, you would check if it's too close to departure time here.
	s.logger.Info("cancelling ticket", "id", id)
	return s.repo.UpdateStatus(ctx, id, "CANCELLED")
}
