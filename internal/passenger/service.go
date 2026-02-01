package passenger

import (
	"context"
	"log/slog"
)

type Service struct {
	repo   *Repository
	logger *slog.Logger
}

func NewService(repo *Repository, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) Create(ctx context.Context, p *Passenger) error {
	if err := s.repo.Create(ctx, p); err != nil {
		s.logger.Error("failed to create passenger", "err", err)
		return err
	}
	return nil
}

func (s *Service) GetProfile(ctx context.Context, id int64) (*Passenger, error) {
	return s.repo.FindByID(ctx, id)
}
