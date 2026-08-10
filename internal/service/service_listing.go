package service

import (
	"fmt"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ServiceListingService manages enterprise service capability listings (PRD ②-2 供给能力展示).
type ServiceListingService struct {
	repo repository.ServiceListingRepository
}

func NewServiceListingService(r repository.ServiceListingRepository) *ServiceListingService {
	return &ServiceListingService{repo: r}
}

// CreateListing 创建服务能力（管理后台录入 / 企业发布），默认上架状态。
func (s *ServiceListingService) CreateListing(providerID, providerName, title, category, description, region string, priceFen int64, unit, image string) (domain.ServiceListing, error) {
	now := time.Now()
	sl := domain.ServiceListing{
		ID:           fmt.Sprintf("service-listing-%d", now.UnixNano()),
		ProviderID:   providerID,
		ProviderName: providerName,
		Title:        title,
		Category:     category,
		Description:  description,
		Region:       region,
		PriceFen:     priceFen,
		Unit:         unit,
		Image:        image,
		Status:       "published",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return s.repo.Create(sl)
}

// ListPublished 公开列表：只返回上架中的服务能力。
func (s *ServiceListingService) ListPublished() ([]domain.ServiceListing, error) {
	all, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	out := make([]domain.ServiceListing, 0, len(all))
	for _, sl := range all {
		if sl.Status == "" || sl.Status == "published" {
			out = append(out, sl)
		}
	}
	return out, nil
}

// Get 按 ID 查询（管理后台编辑用）。
func (s *ServiceListingService) Get(id string) (domain.ServiceListing, error) {
	return s.repo.FindByID(id)
}

// ListAdmin 管理端列表：返回全部（含下架），支持关键词（标题/服务商/描述）与分类过滤。
func (s *ServiceListingService) ListAdmin(keyword, category string) ([]domain.ServiceListing, error) {
	all, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	kw := strings.TrimSpace(keyword)
	out := make([]domain.ServiceListing, 0, len(all))
	for _, sl := range all {
		if category != "" && sl.Category != category {
			continue
		}
		if kw != "" &&
			!strings.Contains(sl.Title, kw) &&
			!strings.Contains(sl.ProviderName, kw) &&
			!strings.Contains(sl.Description, kw) {
			continue
		}
		out = append(out, sl)
	}
	return out, nil
}

// UpdateListing 更新服务能力（管理后台用）。
func (s *ServiceListingService) UpdateListing(sl domain.ServiceListing) (domain.ServiceListing, error) {
	return s.repo.Update(sl)
}

// DeleteListing 删除服务能力（管理后台用）。
func (s *ServiceListingService) DeleteListing(id string) error {
	return s.repo.Delete(id)
}
