package airportops

import (
	"context"
	"errors"
	"initial-airport-management-system/internal/flight"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrGateAlreadyExists = errors.New("gate with this code already exists in the terminal")

type Repository struct {
	db flight.DBTX
}

func NewRepository(db flight.DBTX) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateGate(ctx context.Context, g *Gate) error {
	query := `INSERT INTO gates (terminal_id, code, status) VALUES ($1, $2, $3) RETURNING id`

	err := r.db.QueryRow(ctx, query, g.TerminalID, g.Code, g.Status).Scan(&g.ID)
	if err != nil {
		// Check for Postgres Unique Violation (23505)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrGateAlreadyExists
		}
		return err
	}
	return nil
}

// AssignGateToFlight updates the flight table.
func (r *Repository) AssignGateToFlight(ctx context.Context, flightID, gateID int64) error {
	query := `UPDATE flights SET gate_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, gateID, flightID)
	return err
}
