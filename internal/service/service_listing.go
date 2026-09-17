package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ServiceListingService 服务能力的**适配层**。
//
// 合并背景：同一个"航拍服务"，卖家此前可以走两条路发布——service_listings（纯展示、
// 不能下单）或 drone_products + prod_type=aerial（可下单、走托管金结算）。两条路字段不同、
// 状态机不同、前端入口不同，用户在「发布」页看到两个入口做同一件事。
//
// 现在数据统一落在商品表（migration 000110 已把旧行迁过去）。本类型保留原来的方法签名，
// 内部全部读写商品——**HTTP 层与路由一行都不用改**，小程序与管理端零改动切到新数据；
// 出问题时把这里换回旧仓储即可回退。等验证无回归，下一步才删 service_listings 表。
//
// 真实来源：商品表里 prod_type ∈ {repair, aerial, test_fly, calibration, airspace} 的行。
type ServiceListingService struct {
	prodRepo repository.ProductRepository
}

func NewServiceListingService(pr repository.ProductRepository) *ServiceListingService {
	return &ServiceListingService{prodRepo: pr}
}

// serviceProdTypes 属于「服务能力」的商品类型。
// 实物商品（整机/配件）不属于服务能力，绝不能出现在服务列表里。
var serviceProdTypes = map[domain.ProductType]bool{
	domain.ProductRepair:      true,
	domain.ProductAerial:      true,
	domain.ProductTestFly:     true,
	domain.ProductCalibration: true,
	domain.ProductAirspace:    true,
}

// prodTypeFromCategory 服务分类 → 商品类型。与 migration 000110 的回填规则保持一致
// （航拍→aerial / 试飞→test_fly / 检测、标定→calibration / 空域→airspace / 其余→repair）。
// 推断不出时归维修服务：服务类兜底，绝不会误判成实物商品。
func prodTypeFromCategory(category string) domain.ProductType {
	c := strings.TrimSpace(category)
	switch {
	case strings.Contains(c, "航拍"):
		return domain.ProductAerial
	case strings.Contains(c, "试飞"):
		return domain.ProductTestFly
	case strings.Contains(c, "检测"), strings.Contains(c, "标定"):
		return domain.ProductCalibration
	case strings.Contains(c, "空域"):
		return domain.ProductAirspace
	default:
		return domain.ProductRepair
	}
}

// listingStatusOf 商品的两个维度 → 服务能力对外状态（保持旧 API 的取值）。
//
//	check_status=pending   → pending（待审核）
//	check_status=rejected  → rejected（新增取值：此前驳回与下架不可区分）
//	passed + status=listed → published
//	其余（sold/removed/未上架）→ offline
func listingStatusOf(p domain.DroneProduct) string {
	switch p.CheckStatus {
	case domain.ProductCheckPending:
		return "pending"
	case domain.ProductCheckRejected:
		return "rejected"
	}
	if p.Status == "listed" {
		return "published"
	}
	return "offline"
}

// toListing 商品 → 服务能力 DTO（HTTP 响应形状保持不变）。
func toListing(p domain.DroneProduct) domain.ServiceListing {
	img := ""
	if len(p.Images) > 0 {
		img = p.Images[0]
	}
	return domain.ServiceListing{
		ID:           p.ID,
		ProviderID:   p.SellerID,
		ProviderName: p.SellerName,
		Title:        p.Title,
		Category:     p.Category,
		Description:  p.Description,
		Region:       p.Region,
		PriceFen:     p.PriceFen,
		Unit:         p.Unit,
		Image:        img,
		Status:       listingStatusOf(p),
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

// fromListing 服务能力 DTO → 商品（写入方向：创建/更新）。
func fromListing(sl domain.ServiceListing) domain.DroneProduct {
	images := []string{}
	if strings.TrimSpace(sl.Image) != "" {
		images = append(images, sl.Image)
	}
	pm := domain.PriceModeFixed
	if sl.PriceFen == 0 {
		pm = domain.PriceModeNegotiable
	}
	return domain.DroneProduct{
		ID:          sl.ID,
		SellerID:    sl.ProviderID,
		SellerName:  sl.ProviderName,
		ProdType:    prodTypeFromCategory(sl.Category),
		Title:       sl.Title,
		Description: sl.Description,
		PriceFen:    sl.PriceFen,
		PriceMode:   pm,
		Delivery:    domain.DeliveryNegotiable, // 服务不涉及寄送，也不强制收货地址
		Images:      images,
		Condition:   "new",
		Category:    sl.Category,
		Region:      sl.Region,
		Unit:        sl.Unit,
		CheckStatus: domain.ProductCheckPassed,
		Status:      "listed",
		Version:     1,
		CreatedAt:   sl.CreatedAt,
		UpdatedAt:   sl.UpdatedAt,
	}
}

// listServiceProducts 取全部服务类商品（不含回收站，由仓储层保证）。
func (s *ServiceListingService) listServiceProducts(ctx context.Context) ([]domain.DroneProduct, error) {
	all, err := s.prodRepo.List(ctx, "")
	if err != nil {
		return nil, err
	}
	out := make([]domain.DroneProduct, 0, len(all))
	for _, p := range all {
		if serviceProdTypes[p.ProdType] {
			out = append(out, p)
		}
	}
	return out, nil
}

// CreateListing 创建服务能力（管理后台录入），默认直接上架。
func (s *ServiceListingService) CreateListing(ctx context.Context, providerID, providerName, title, category, description, region string, priceFen int64, unit, image, status string) (domain.ServiceListing, error) {
	return s.create(ctx, providerID, providerName, title, category, description, region, priceFen, unit, image, status, true)
}

// CreateListingPending 用户自助发布服务能力：默认待审核，协会审核通过后才进入公开列表。
func (s *ServiceListingService) CreateListingPending(ctx context.Context, providerID, providerName, title, category, description, region string, priceFen int64, unit, image string) (domain.ServiceListing, error) {
	return s.create(ctx, providerID, providerName, title, category, description, region, priceFen, unit, image, "pending", false)
}

func (s *ServiceListingService) create(ctx context.Context, providerID, providerName, title, category, description, region string, priceFen int64, unit, image, status string, adminCreate bool) (domain.ServiceListing, error) {
	if priceFen < 0 {
		return domain.ServiceListing{}, errors.New("price cannot be negative")
	}
	if strings.TrimSpace(title) == "" {
		return domain.ServiceListing{}, errors.New("服务标题不能为空")
	}
	if strings.TrimSpace(providerName) == "" {
		// seller_name 是买家看到的卖家名，空串会让卡片显示不出服务商
		providerName = "平台自营"
	}
	if strings.TrimSpace(providerID) == "" {
		// 空 seller_id 会让后续订单归属与售后全断
		providerID = "platform"
	}
	now := time.Now()
	p := fromListing(domain.ServiceListing{
		ID: nextID("service-listing"), ProviderID: providerID, ProviderName: providerName,
		Title: title, Category: category, Description: description, Region: region,
		PriceFen: priceFen, Unit: unit, Image: image, CreatedAt: now, UpdatedAt: now,
	})
	if adminCreate {
		// 管理端录入视为已审核；status 语义仍沿用旧 API：published / offline
		p.CheckStatus = domain.ProductCheckPassed
		if status == "offline" {
			p.Status = "removed"
		}
	} else {
		// 用户自助发布：两个维度都进待审
		p.CheckStatus = domain.ProductCheckPending
		p.Status = "pending"
	}
	created, err := s.prodRepo.Create(ctx, p)
	if err != nil {
		return domain.ServiceListing{}, err
	}
	return toListing(created), nil
}

// ListPublished 公开列表：只返回上架中（审核通过 + 在售）的服务能力。
func (s *ServiceListingService) ListPublished(ctx context.Context) ([]domain.ServiceListing, error) {
	items, err := s.listServiceProducts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ServiceListing, 0, len(items))
	for _, p := range items {
		if ProductVisibleInHall(p) {
			out = append(out, toListing(p))
		}
	}
	return out, nil
}

// Get 按 ID 查询（管理后台编辑 / 公开详情用）。
func (s *ServiceListingService) Get(ctx context.Context, id string) (domain.ServiceListing, error) {
	p, err := s.prodRepo.FindByID(ctx, id)
	if err != nil {
		return domain.ServiceListing{}, err
	}
	// 实物商品不属于服务能力：按"不存在"处理，避免 /service-listings/{id} 泄漏商品
	if !serviceProdTypes[p.ProdType] {
		return domain.ServiceListing{}, fmt.Errorf("service listing %s: %w", id, repository.ErrNotFound)
	}
	return toListing(p), nil
}

// ListAdmin 管理端列表：返回全部（含待审/下架），支持关键词与分类过滤。
func (s *ServiceListingService) ListAdmin(ctx context.Context, keyword, category string) ([]domain.ServiceListing, error) {
	items, err := s.listServiceProducts(ctx)
	if err != nil {
		return nil, err
	}
	kw := strings.TrimSpace(keyword)
	out := make([]domain.ServiceListing, 0, len(items))
	for _, p := range items {
		if category != "" && p.Category != category {
			continue
		}
		if kw != "" &&
			!strings.Contains(p.Title, kw) &&
			!strings.Contains(p.SellerName, kw) &&
			!strings.Contains(p.Description, kw) {
			continue
		}
		out = append(out, toListing(p))
	}
	return out, nil
}

// UpdateListing 更新服务能力（管理后台用）。
func (s *ServiceListingService) UpdateListing(ctx context.Context, sl domain.ServiceListing) (domain.ServiceListing, error) {
	existing, err := s.prodRepo.FindByID(ctx, sl.ID)
	if err != nil {
		return domain.ServiceListing{}, err
	}
	if !serviceProdTypes[existing.ProdType] {
		return domain.ServiceListing{}, fmt.Errorf("service listing %s: %w", sl.ID, repository.ErrNotFound)
	}
	p := fromListing(sl)
	// 保留不该由本入口改动的字段
	p.Version = existing.Version
	p.CheckStatus = existing.CheckStatus
	p.CheckReason = existing.CheckReason
	p.ReviewedAt = existing.ReviewedAt
	p.ReviewedBy = existing.ReviewedBy
	p.Views = existing.Views
	// status 语义：旧 API 只给 published / offline
	if sl.Status == "published" {
		p.Status = "listed"
	} else {
		p.Status = "removed"
	}
	updated, err := s.prodRepo.Update(ctx, p)
	if err != nil {
		return domain.ServiceListing{}, err
	}
	return toListing(updated), nil
}

// DeleteListing 删除服务能力（管理后台用）：与商品一致走回收站，不是物理删除。
func (s *ServiceListingService) DeleteListing(ctx context.Context, id string) error {
	return s.prodRepo.SoftDelete(ctx, id)
}

// ToggleFavorite 收藏/取消收藏服务能力。收藏关系与商品收藏共用一张表——
// migration 000110 已把旧的 service_listing_favorites 迁进 product_favorites。
func (s *ServiceListingService) ToggleFavorite(ctx context.Context, userID, listingID string, favorite bool) error {
	if _, err := s.Get(ctx, listingID); err != nil {
		return err
	}
	if favorite {
		return s.prodRepo.FavoriteProduct(ctx, userID, listingID)
	}
	return s.prodRepo.UnfavoriteProduct(ctx, userID, listingID)
}

// ListFavorites 当前用户收藏的服务能力列表（按收藏时间倒序）。
func (s *ServiceListingService) ListFavorites(ctx context.Context, userID string) ([]domain.ServiceListing, error) {
	all, err := s.prodRepo.ListFavoriteProducts(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ServiceListing, 0, len(all))
	for _, p := range all {
		if serviceProdTypes[p.ProdType] {
			out = append(out, toListing(p))
		}
	}
	return out, nil
}
