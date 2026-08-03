package service

import (
	"fmt"
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

func (s *TradingService) CreateProduct(a domain.Actor, prodType domain.ProductType, title, desc, brand, model, condition string, priceFen int64) (domain.DroneProduct, error) {
	now := time.Now()
	p := domain.DroneProduct{ID: fmt.Sprintf("product-%d", now.UnixNano()), SellerID: a.ID, SellerName: a.ID,
		ProdType: prodType, Title: title, Description: desc, PriceFen: priceFen,
		Brand: brand, Model: model, Condition: condition, Status: "listed", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.prodRepo.Create(p)
}

func (s *TradingService) ListProducts(prodType string) ([]domain.DroneProduct, error) {
	return s.prodRepo.List(prodType)
}

func (s *TradingService) GetProduct(id string) (domain.DroneProduct, error) {
	return s.prodRepo.FindByID(id)
}

// GetProductAndCountView 详情访问：浏览量 +1 后返回（先读旧值再递增）
func (s *TradingService) GetProductAndCountView(id string) (domain.DroneProduct, error) {
	p, err := s.prodRepo.FindByID(id)
	if err != nil {
		return domain.DroneProduct{}, err
	}
	p.Views++
	_ = s.prodRepo.IncrementViews(id)
	return p, nil
}

// UpdateProduct 更新商品（管理后台用）
func (s *TradingService) UpdateProduct(p domain.DroneProduct) (domain.DroneProduct, error) {
	return s.prodRepo.Update(p)
}

// DeleteProduct 删除商品（管理后台用）
func (s *TradingService) DeleteProduct(id string) error {
	return s.prodRepo.Delete(id)
}

func (s *TradingService) CreateRepair(a domain.Actor, productDesc, faultDesc string) (domain.RepairOrder, error) {
	now := time.Now()
	r := domain.RepairOrder{ID: fmt.Sprintf("repair-%d", now.UnixNano()), CustomerID: a.ID,
		ProductDesc: productDesc, FaultDesc: faultDesc, Status: "submitted", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.repairRepo.Create(r)
}

func (s *TradingService) ListMyRepairs(a domain.Actor) ([]domain.RepairOrder, error) {
	return s.repairRepo.ListByUser(a.ID)
}
