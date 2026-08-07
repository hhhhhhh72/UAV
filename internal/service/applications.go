package service

import (
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

func (s *ApplicationService) Create(app domain.Application) (domain.Application, error) {
	return s.repo.Create(app)
}

func (s *ApplicationService) Get(id string) (domain.Application, error) {
	return s.repo.FindByID(id)
}

func (s *ApplicationService) ListMine(userID string, page, pageSize int) ([]domain.Application, int, error) {
	return s.repo.ListByUser(userID, (page-1)*pageSize, pageSize)
}

func (s *ApplicationService) ListAll(page, pageSize int) ([]domain.Application, int, error) {
	return s.repo.ListAll((page-1)*pageSize, pageSize)
}
