package airportops

import (
	"context"
	"initial-airport-management-system/internal/flight"
)

type Repository struct {
	db flight.DBTX
}

func NewRepository(db flight.DBTX) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateGate(ctx context.Context, g *Gate) error {
	query := `INSERT INTO gates (terminal_id, code, status) VALUES ($1, $2, $3) RETURNING id`
	return r.db.QueryRow(ctx, query, g.TerminalID, g.Code, g.Status).Scan(&g.ID)
}

// AssignGateToFlight updates the flight table.
func (r *Repository) AssignGateToFlight(ctx context.Context, flightID, gateID int64) error {
	query := `UPDATE flights SET gate_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, gateID, flightID)
	return err
}
