package service

import (
	"context"
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

// CreateListing 创建服务能力（管理后台录入），默认直接上架。
func (s *ServiceListingService) CreateListing(ctx context.Context, providerID, providerName, title, category, description, region string, priceFen int64, unit, image string) (domain.ServiceListing, error) {
	now := time.Now()
	sl := domain.ServiceListing{
		ID:           nextID("service-listing"),
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
	return s.repo.Create(ctx, sl)
}

// CreateListingPending 用户自助发布服务能力：默认待审核（pending），
// 由协会在管理端审核通过后才进入公开列表（ListPublished 只返回 published）。
func (s *ServiceListingService) CreateListingPending(ctx context.Context, providerID, providerName, title, category, description, region string, priceFen int64, unit, image string) (domain.ServiceListing, error) {
	now := time.Now()
	sl := domain.ServiceListing{
		ID:           nextID("service-listing"),
		ProviderID:   providerID,
		ProviderName: providerName,
		Title:        title,
		Category:     category,
		Description:  description,
		Region:       region,
		PriceFen:     priceFen,
		Unit:         unit,
		Image:        image,
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return s.repo.Create(ctx, sl)
}

// ListPublished 公开列表：只返回上架中的服务能力。
func (s *ServiceListingService) ListPublished(ctx context.Context) ([]domain.ServiceListing, error) {
	all, err := s.repo.List(ctx)
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
func (s *ServiceListingService) Get(ctx context.Context, id string) (domain.ServiceListing, error) {
	return s.repo.FindByID(ctx, id)
}

// ListAdmin 管理端列表：返回全部（含下架），支持关键词（标题/服务商/描述）与分类过滤。
func (s *ServiceListingService) ListAdmin(ctx context.Context, keyword, category string) ([]domain.ServiceListing, error) {
	all, err := s.repo.List(ctx)
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
func (s *ServiceListingService) UpdateListing(ctx context.Context, sl domain.ServiceListing) (domain.ServiceListing, error) {
	return s.repo.Update(ctx, sl)
}

// DeleteListing 删除服务能力（管理后台用）。
func (s *ServiceListingService) DeleteListing(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
