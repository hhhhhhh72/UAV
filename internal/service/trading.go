package service

import (
	"context"
	"errors"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type TradingService struct {
	prodRepo   repository.ProductRepository
	repairRepo repository.RepairRepository
}

func NewTradingService(pr repository.ProductRepository, rr repository.RepairRepository) *TradingService {
	return &TradingService{prodRepo: pr, repairRepo: rr}
}

func (s *TradingService) CreateProduct(ctx context.Context, a domain.Actor, prodType domain.ProductType, title, desc, brand, model, condition string, priceFen int64, images []string) (domain.DroneProduct, error) {
	if priceFen < 0 {
		return domain.DroneProduct{}, errors.New("price cannot be negative")
	}
	now := time.Now()
	// 用户发布商品默认"待审核"，管理后台通过后才上架（公开列表只展示 listed）
	p := domain.DroneProduct{ID: nextID("product"), SellerID: a.ID, SellerName: a.ID,
		ProdType: prodType, Title: title, Description: desc, PriceFen: priceFen,
		Brand: brand, Model: model, Condition: condition, Images: images, Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.prodRepo.Create(ctx, p)
}

func (s *TradingService) ListProducts(ctx context.Context, prodType string) ([]domain.DroneProduct, error) {
	return s.prodRepo.List(ctx, prodType)
}

// ListTopProducts 首页 Top-N 商品（透传 repo.ListTop，SQL 端 LIMIT 不整表）。
func (s *TradingService) ListTopProducts(ctx context.Context, prodType string, limit int) ([]domain.DroneProduct, error) {
	return s.prodRepo.ListTop(ctx, prodType, limit)
}

// SumProductViews 商品浏览量总和（首页 stats.views，聚合查询）。
func (s *TradingService) SumProductViews(ctx context.Context, prodType string) (int, error) {
	return s.prodRepo.SumViews(ctx, prodType)
}

// ListProductsByIDs 批量按 ID 取商品（订单列表补商品名，防 N+1）。
func (s *TradingService) ListProductsByIDs(ctx context.Context, ids []string) ([]domain.DroneProduct, error) {
	return s.prodRepo.ListByIDs(ctx, ids)
}

func (s *TradingService) GetProduct(ctx context.Context, id string) (domain.DroneProduct, error) {
	return s.prodRepo.FindByID(ctx, id)
}

// GetProductAndCountView 详情访问：浏览量 +1 后返回（先读旧值再递增）
func (s *TradingService) GetProductAndCountView(ctx context.Context, id string) (domain.DroneProduct, error) {
	p, err := s.prodRepo.FindByID(ctx, id)
	if err != nil {
		return domain.DroneProduct{}, err
	}
	p.Views++
	_ = s.prodRepo.IncrementViews(ctx, id)
	return p, nil
}

// UpdateProduct 更新商品（管理后台用）
func (s *TradingService) UpdateProduct(ctx context.Context, p domain.DroneProduct) (domain.DroneProduct, error) {
	return s.prodRepo.Update(ctx, p)
}

// MarkProductSold 下单抢占：仅 listed 商品可标记 sold（防一物多卖/超卖）。
func (s *TradingService) MarkProductSold(ctx context.Context, id string) error {
	return s.prodRepo.MarkSold(ctx, id)
}

// RestoreProduct 订单创建失败时回滚：sold → listed。
func (s *TradingService) RestoreProduct(ctx context.Context, id string) error {
	return s.prodRepo.Restore(ctx, id)
}

// DeleteProduct 删除商品（管理后台用）
func (s *TradingService) DeleteProduct(ctx context.Context, id string) error {
	return s.prodRepo.Delete(ctx, id)
}

func (s *TradingService) CreateRepair(ctx context.Context, a domain.Actor, productDesc, faultDesc string) (domain.RepairOrder, error) {
	now := time.Now()
	r := domain.RepairOrder{ID: nextID("repair"), CustomerID: a.ID,
		ProductDesc: productDesc, FaultDesc: faultDesc, Status: "submitted", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.repairRepo.Create(ctx, r)
}

func (s *TradingService) ListMyRepairs(ctx context.Context, a domain.Actor) ([]domain.RepairOrder, error) {
	return s.repairRepo.ListByUser(ctx, a.ID)
}

func (s *TradingService) ListAllRepairs(ctx context.Context, offset, limit int) ([]domain.RepairOrder, int, error) {
	return s.repairRepo.ListAll(ctx, offset, limit)
}
