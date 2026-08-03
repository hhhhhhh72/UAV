package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"drone-platform/internal/domain"
)

// ---- Trading ----

func (s *Server) createProduct(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title, Description, Brand, Model, Condition string
		ProdType                                    string `json:"prod_type"`
		PriceFen                                    int64  `json:"price_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	p, err := s.tradingSvc.CreateProduct(a, domain.ProductType(in.ProdType), in.Title, in.Description, in.Brand, in.Model, in.Condition, in.PriceFen)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, p)
}

func (s *Server) listProducts(w http.ResponseWriter, r *http.Request) {
	products, err := s.tradingSvc.ListProducts(r.URL.Query().Get("prod_type"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// 公开列表只显示在售商品（下架/已售不展示）
	listed := make([]domain.DroneProduct, 0, len(products))
	for _, p := range products {
		if p.Status == "" || p.Status == "listed" {
			listed = append(listed, p)
		}
	}
	// 关键词过滤（标题/品牌/型号，内存过滤——商品量级小）
	if kw := strings.TrimSpace(r.URL.Query().Get("keyword")); kw != "" {
		filtered := make([]domain.DroneProduct, 0, len(listed))
		for _, p := range listed {
			if strings.Contains(p.Title, kw) || strings.Contains(p.Brand, kw) || strings.Contains(p.Model, kw) {
				filtered = append(filtered, p)
			}
		}
		listed = filtered
	}
	respond(w, r, http.StatusOK, listed)
}

// GET /api/v1/products/{id} — 商品详情（公开）
func (s *Server) getProductDetail(w http.ResponseWriter, r *http.Request) {
	p, err := s.tradingSvc.GetProduct(r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, p)
}

// POST /api/v1/admin/products — 管理后台创建商品
func (s *Server) adminCreateProduct(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ID          string   `json:"id"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		ProdType    string   `json:"prod_type"`
		Brand       string   `json:"brand"`
		Model       string   `json:"model"`
		Condition   string   `json:"condition"`
		PriceFen    int64    `json:"price_fen"`
		Status      string   `json:"status"`
		Images      []string `json:"images"`
		SellerName  string   `json:"seller_name"`
	}
	if err := decode(r, &in); err != nil || in.Title == "" {
		fail(w, r, http.StatusBadRequest, errors.New("title required"))
		return
	}
	now := time.Now()
	p := domain.DroneProduct{
		ID:          in.ID,
		Title:       in.Title,
		Description: in.Description,
		ProdType:    domain.ProductType(in.ProdType),
		Brand:       in.Brand,
		Model:       in.Model,
		Condition:   in.Condition,
		PriceFen:    in.PriceFen,
		Status:      in.Status,
		Images:      in.Images,
		SellerName:  in.SellerName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if p.ID == "" {
		p.ID = fmt.Sprintf("prod-%d", now.UnixNano())
	}
	if p.Status == "" {
		p.Status = "listed"
	}
	if p.ProdType == "" {
		p.ProdType = domain.ProductDrone
	}
	if p.Condition == "" {
		p.Condition = "new"
	}
	if p.SellerName == "" {
		p.SellerName = "平台自营"
	}
	created, err := s.tradingSvc.CreateProduct(domain.Actor{Role: domain.RolePlatformAdmin}, p.ProdType, p.Title, p.Description, p.Brand, p.Model, p.Condition, p.PriceFen)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// CreateProduct 内部固定 status=listed，这里回写真实状态与展示字段
	created.Status = p.Status
	created.Images = in.Images
	created.SellerName = p.SellerName
	if _, err := s.tradingSvc.UpdateProduct(created); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, created)
}

// PUT /api/v1/admin/products/{id} — 管理后台更新商品
func (s *Server) adminUpdateProduct(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		ProdType    string   `json:"prod_type"`
		Brand       string   `json:"brand"`
		Model       string   `json:"model"`
		Condition   string   `json:"condition"`
		PriceFen    int64    `json:"price_fen"`
		Status      string   `json:"status"`
		Images      []string `json:"images"`
		SellerName  string   `json:"seller_name"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	existing, err := s.tradingSvc.GetProduct(r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	if in.Title != "" {
		existing.Title = in.Title
	}
	if in.Description != "" {
		existing.Description = in.Description
	}
	if in.ProdType != "" {
		existing.ProdType = domain.ProductType(in.ProdType)
	}
	if in.Brand != "" {
		existing.Brand = in.Brand
	}
	if in.Model != "" {
		existing.Model = in.Model
	}
	if in.Condition != "" {
		existing.Condition = in.Condition
	}
	if in.PriceFen > 0 {
		existing.PriceFen = in.PriceFen
	}
	if in.Status != "" {
		existing.Status = in.Status
	}
	if in.Images != nil {
		existing.Images = in.Images
	}
	if in.SellerName != "" {
		existing.SellerName = in.SellerName
	}
	updated, err := s.tradingSvc.UpdateProduct(existing)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, updated)
}

// DELETE /api/v1/admin/products/{id} — 管理后台删除商品
func (s *Server) adminDeleteProduct(w http.ResponseWriter, r *http.Request) {
	if err := s.tradingSvc.DeleteProduct(r.PathValue("id")); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

// GET /api/v1/admin/products — 管理后台商品列表
func (s *Server) listAdminProducts(w http.ResponseWriter, r *http.Request) {
	products, err := s.tradingSvc.ListProducts(r.URL.Query().Get("prod_type"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, products, len(products))
}

func (s *Server) createRepair(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		ProductDesc string `json:"product_desc"`
		FaultDesc   string `json:"fault_desc"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	rp, err := s.tradingSvc.CreateRepair(a, in.ProductDesc, in.FaultDesc)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, rp)
}

func (s *Server) listMyRepairs(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	repairs, err := s.tradingSvc.ListMyRepairs(a)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, repairs)
}

// ---- Insurance ----

func (s *Server) createPolicy(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		DroneModel  string    `json:"drone_model"`
		DroneSN     string    `json:"drone_sn"`
		PolicyType  string    `json:"policy_type"`
		PremiumFen  int64     `json:"premium_fen"`
		CoverageFen int64     `json:"coverage_fen"`
		StartDate   time.Time `json:"start_date"`
		EndDate     time.Time `json:"end_date"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	p, err := s.insuranceSvc.CreatePolicy(a, in.DroneModel, in.DroneSN, in.PolicyType, in.PremiumFen, in.CoverageFen, in.StartDate, in.EndDate)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, p)
}

func (s *Server) listMyPolicies(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	policies, err := s.insuranceSvc.ListMyPolicies(a)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, policies)
}

func (s *Server) createInspection(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		DroneModel  string    `json:"drone_model"`
		DroneSN     string    `json:"drone_sn"`
		InspectDate time.Time `json:"inspect_date"`
		ExpireDate  time.Time `json:"expire_date"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	i, err := s.insuranceSvc.CreateInspection(a, in.DroneModel, in.DroneSN, in.InspectDate, in.ExpireDate)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, i)
}

func (s *Server) listMyInspections(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	inspections, err := s.insuranceSvc.ListMyInspections(a)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, inspections)
}

// ---- Finance ----

func (s *Server) applyLoan(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		AmountFen  int64  `json:"amount_fen"`
		TermMonths int    `json:"term_months"`
		Purpose    string `json:"purpose"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	l, err := s.financeSvc.ApplyLoan(a, in.AmountFen, in.TermMonths, in.Purpose)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, l)
}

func (s *Server) listMyLoans(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	loans, err := s.financeSvc.ListMyLoans(a)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, loans)
}
