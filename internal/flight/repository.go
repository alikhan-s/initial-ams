package flight

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound = errors.New("flight not found")
	ErrConflict = errors.New("flight version conflict") // For optimistic locking
)

// DBTX is an interface that matches both *pgxpool.Pool and pgx.Tx
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type Repository struct {
	db DBTX
}

func NewRepository(db DBTX) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, f *Flight) error {
	query := `
		INSERT INTO flights (flight_no, origin, destination, departure_time, arrival_time, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at, version
	`
	// Status defaults to SCHEDULED in DB, but good to be explicit
	if f.Status == "" {
		f.Status = StatusScheduled
	}

	return r.db.QueryRow(ctx, query,
		f.FlightNo, f.Origin, f.Destination, f.DepartureTime, f.ArrivalTime, f.Status,
	).Scan(&f.ID, &f.CreatedAt, &f.UpdatedAt, &f.Version)
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*Flight, error) {
	query := `
		SELECT id, flight_no, origin, destination, gate_id, departure_time, arrival_time, status, version, created_at, updated_at
		FROM flights
		WHERE id = $1
	`
	var f Flight
	err := r.db.QueryRow(ctx, query, id).Scan(
		&f.ID, &f.FlightNo, &f.Origin, &f.Destination, &f.GateID,
		&f.DepartureTime, &f.ArrivalTime, &f.Status, &f.Version, &f.CreatedAt, &f.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// UpdateStatus implements Optimistic Locking using the 'version' field
func (r *Repository) UpdateStatus(ctx context.Context, id int64, status Status, currentVersion int) error {
	query := `
		UPDATE flights 
		SET status = $1, version = version + 1, updated_at = NOW()
		WHERE id = $2 AND version = $3
	`
	tag, err := r.db.Exec(ctx, query, status, id, currentVersion)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		// If no rows were affected, it means either the ID doesn't exist
		// OR the version changed (someone else updated it first)
		return ErrConflict
	}
	return nil
}

// SearchParams holds filter criteria
type SearchParams struct {
	Origin      string
	Destination string
	Date        time.Time // We will search by day
}

func (r *Repository) Search(ctx context.Context, params SearchParams) ([]Flight, error) {
	// We search for flights on the specific day (00:00 to 23:59)
	startOfDay := time.Date(params.Date.Year(), params.Date.Month(), params.Date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
        SELECT id, flight_no, origin, destination, departure_time, arrival_time, status, gate_id
        FROM flights
        WHERE origin = $1 
        AND destination = $2 
        AND departure_time >= $3 
        AND departure_time < $4
        AND status != 'CANCELLED'
        ORDER BY departure_time ASC
    `

	rows, err := r.db.Query(ctx, query, params.Origin, params.Destination, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flights []Flight
	for rows.Next() {
		var f Flight
		// Note: We scan into a temp struct or directly into the slice
		if err := rows.Scan(&f.ID, &f.FlightNo, &f.Origin, &f.Destination, &f.DepartureTime, &f.ArrivalTime, &f.Status, &f.GateID); err != nil {
			return nil, err
		}
		flights = append(flights, f)
	}
	return flights, nil
}
