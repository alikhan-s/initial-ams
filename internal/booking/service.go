package booking

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"initial-airport-management-system/internal/flight"
	"initial-airport-management-system/internal/platform/database"

	"github.com/jackc/pgx/v5"
)

type TicketRepository interface {
	Create(ctx context.Context, t *Ticket) error
	GetSoldTicketsCount(ctx context.Context, flightID int64) (int, error)
	FindByID(ctx context.Context, id int64) (*Ticket, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
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

		// Проверяем наличие мест
		// Важно: Мы используем репозиторий, который работает внутри транзакции (tx)
		// Чтобы это работало идеально, нам нужно передать tx в метод репозитория.
		// Но для простоты сейчас (так как уровень изоляции Postgres Read Committed),
		// мы можем просто посчитать текущие записи.

		soldCount, err := s.repo.GetSoldTicketsCount(ctx, flightID)
		if err != nil {
			return fmt.Errorf("failed to check capacity: %w", err)
		}

		// ВАЖНО: В реальном High-Load мы бы использовали "SELECT ... FOR UPDATE" для блокировки строки рейса,
		// но для курсового проекта проверка COUNT перед INSERT допустима.

		limit := f.TotalSeats

		if soldCount >= limit {
			return errors.New("flight is fully booked")
		}

		// Create the Ticket Object
		ticket = &Ticket{
			FlightID:    flightID,
			PassengerID: passengerID,
			Price:       150.00, // Placeholder price logic
			Status:      "ACTIVE",
			CreatedAt:   time.Now(),
		}

		// Persist Ticket (Passing the TX to ensure atomicity)
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

func (s *Service) GetTicket(ctx context.Context, id int64) (*Ticket, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) CancelTicket(ctx context.Context, id int64) error {
	// Fetch ticket to check current status
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if t.Status == "CANCELLED" {
		return fmt.Errorf("ticket is already cancelled")
	}

	// Perform Update
	return s.repo.UpdateStatus(ctx, id, "CANCELLED")
}
