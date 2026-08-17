package service

import (
	"context"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ApplicationService manages service applications (miniprogram /api/submit).
type ApplicationService struct {
	repo repository.ApplicationRepository
}

func NewApplicationService(repo repository.ApplicationRepository) *ApplicationService {
	return &ApplicationService{repo: repo}
}

func (s *ApplicationService) Create(ctx context.Context, app domain.Application) (domain.Application, error) {
	return s.repo.Create(ctx, app)
}

func (s *ApplicationService) Get(ctx context.Context, id string) (domain.Application, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ApplicationService) ListMine(ctx context.Context, userID string, page, pageSize int) ([]domain.Application, int, error) {
	return s.repo.ListByUser(ctx, userID, (page-1)*pageSize, pageSize)
}

func (s *ApplicationService) ListAll(ctx context.Context, page, pageSize int) ([]domain.Application, int, error) {
	return s.repo.ListAll(ctx, (page-1)*pageSize, pageSize)
}
