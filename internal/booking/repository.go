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

func (r *Repository) FindByID(ctx context.Context, id int64) (*Ticket, error) {
	query := `
		SELECT id, flight_id, passenger_id, seat_no, price, status, created_at
		FROM tickets
		WHERE id = $1
	`
	var t Ticket
	// Note: seat_no can be NULL in DB. Ideally handle with sql.NullString,
	// but assuming for this milestone it handles empty strings or is set.
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.FlightID, &t.PassengerID, &t.SeatNo, &t.Price, &t.Status, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE tickets SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
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
