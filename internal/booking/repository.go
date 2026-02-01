package booking

import (
	"context"
	"database/sql"
	"errors"
	"initial-airport-management-system/internal/flight"
)

type Repository struct {
	db flight.DBTX
}

func NewRepository(db flight.DBTX) *Repository {
	return &Repository{db: db}
}

// Create persists a new ticket
func (r *Repository) Create(ctx context.Context, t *Ticket) error {
	query := `
		INSERT INTO tickets (flight_id, passenger_id, price, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		t.FlightID, t.PassengerID, t.Price, t.Status, t.CreatedAt,
	).Scan(&t.ID)
}

// GetSoldTicketsCount returns the number of active tickets for a flight
func (r *Repository) GetSoldTicketsCount(ctx context.Context, flightID int64) (int, error) {
	query := `SELECT COUNT(*) FROM tickets WHERE flight_id = $1 AND status = 'ACTIVE'`
	var count int
	err := r.db.QueryRow(ctx, query, flightID).Scan(&count)
	return count, err
}

// FindByID retrieves a ticket by its ID
func (r *Repository) FindByID(ctx context.Context, id int64) (*Ticket, error) {
	query := `
		SELECT id, flight_id, passenger_id, seat_no, price, status, created_at
		FROM tickets
		WHERE id = $1
	`
	var t Ticket
	var seatNo sql.NullString // Handle nullable seat_no

	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.FlightID, &t.PassengerID, &seatNo, &t.Price, &t.Status, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if seatNo.Valid {
		t.SeatNo = seatNo.String
	}
	return &t, nil
}

// UpdateStatus changes the status of a ticket (e.g., "CANCELLED")
func (r *Repository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE tickets SET status = $1 WHERE id = $2`
	result, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return errors.New("ticket not found or status unchanged")
	}
	return nil
}
