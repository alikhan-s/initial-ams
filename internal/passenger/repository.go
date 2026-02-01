package passenger

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

func (r *Repository) Create(ctx context.Context, p *Passenger) error {
	query := `
		INSERT INTO passengers (user_id, passport_no, phone)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query, p.UserID, p.PassportNo, p.Phone).Scan(&p.ID)
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*Passenger, error) {
	query := `SELECT id, user_id, passport_no, phone FROM passengers WHERE id = $1`
	var p Passenger
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.UserID, &p.PassportNo, &p.Phone)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
