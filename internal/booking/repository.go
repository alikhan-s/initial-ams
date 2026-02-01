package booking

import (
	"context"
	"initial-airport-management-system/internal/flight" // importing for the DBTX interface
)

type Repository struct {
	db flight.DBTX // Reusing the DBTX interface for transaction support
}

func NewRepository(db flight.DBTX) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, t *Ticket) error {
	query := `
		INSERT INTO tickets (flight_id, passenger_id, price, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	// t.SeatNo is empty initially, so we don't insert it yet
	return r.db.QueryRow(ctx, query,
		t.FlightID, t.PassengerID, t.Price, t.Status, t.CreatedAt,
	).Scan(&t.ID)
}

// GetSoldTicketsCount returns the number of active tickets for a flight
func (r *Repository) GetSoldTicketsCount(ctx context.Context, flightID int64) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM tickets 
		WHERE flight_id = $1 AND status = 'ACTIVE'
	`
	var count int
	err := r.db.QueryRow(ctx, query, flightID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
