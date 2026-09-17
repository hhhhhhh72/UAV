package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/service"
)

// ---- Trading ----

// platformSellerID 平台自营商品的固定卖家 ID（管理端建商品时无真实用户 ID，
// 空 SellerID 会导致下单后订单 seller_id 为空串——订单归属与售后均依赖卖家）。
const platformSellerID = "platform"

func (s *Server) createProduct(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	// 发布商品的认证门禁：商品详情页对买家白纸黑字承诺「平台认证商家」，
	// 而此前只要登录就能发商品——标识对每一件在售商品都是假话。
	// 现在只有完成**企业认证（approved）**的账号能上架，标识随之变成真话。
	//
	// 为什么放在 handler 而不是 service.CreateProduct：这是**路径策略**，不是
	// "创建商品"的固有不变式——管理端代建（adminCreateProduct）要能替任意卖家建商品，
	// 不该被卖家的认证状态卡住。与 createCompetitionByEnterprise（企业自助发布）同一形状。
	ents, err := s.enterpriseSvc.ListMine(r.Context(), a)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	certified := false
	for _, e := range ents {
		if e.Status == domain.EnterpriseApproved {
			certified = true
			break
		}
	}
	if !certified {
		fail(w, r, http.StatusForbidden, errors.New("请先完成企业认证后再发布商品"))
		return
	}
	var in struct {
		Title, Description, Brand, Model, Condition string
		ProdType                                    string   `json:"prod_type"`
		Delivery                                    string   `json:"delivery"`
		PriceMode                                   string   `json:"price_mode"`
		PriceFen                                    int64    `json:"price_fen"`
		Images                                      []string `json:"images"`
		DetailImages                                []string `json:"detail_images"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	p, err := s.tradingSvc.CreateProduct(r.Context(), a, domain.ProductType(in.ProdType), in.Title, in.Description, in.Brand, in.Model, in.Condition, in.Delivery, in.PriceMode, in.PriceFen, in.Images, in.DetailImages)
	if err != nil {
		// 字段校验失败是 400（用户能改），其余才是 500（服务故障）。
		// 此前一律 403，把"填错了"显示成"没权限"。
		if errors.Is(err, service.ErrProductInvalid) {
			fail(w, r, http.StatusBadRequest, err)
			return
		}
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, p)
}

// PATCH /api/v1/products/{id} — 卖家自助编辑自己的商品
//
// 任何内容编辑都会把商品退回待审核（服务层决定），编辑完成后要等协会重新审核通过。
func (s *Server) updateMyProduct(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		ProdType    string   `json:"prod_type"`
		Brand       string   `json:"brand"`
		Model       string   `json:"model"`
		Condition   string   `json:"condition"`
		PriceMode   string   `json:"price_mode"`
		Delivery    string   `json:"delivery"`
		PriceFen     int64    `json:"price_fen"`
		Images       []string `json:"images"`
		DetailImages []string `json:"detail_images"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	p, err := s.tradingSvc.UpdateMyProduct(r.Context(), a, r.PathValue("id"), service.ProductEditInput{
		ProdType:    domain.ProductType(in.ProdType),
		Title:       in.Title,
		Description: in.Description,
		Brand:       in.Brand,
		Model:       in.Model,
		Condition:   in.Condition,
		PriceMode:   in.PriceMode,
		Delivery:    in.Delivery,
		PriceFen:     in.PriceFen,
		Images:       in.Images,
		DetailImages: in.DetailImages,
	})
	if err != nil {
		// 字段不合法 → 400；不属于自己的商品 → 404（服务层返回 ErrNotFound）。
		if errors.Is(err, service.ErrProductInvalid) {
			fail(w, r, http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			fail(w, r, http.StatusNotFound, errors.New("product not found"))
			return
		}
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, p)
}

// POST /api/v1/products/{id}/status — 卖家自助上下架（listed / removed）
func (s *Server) setMyProductStatus(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	p, err := s.tradingSvc.SetMyProductStatus(r.Context(), a, r.PathValue("id"), in.Status)
	if err != nil {
		if errors.Is(err, service.ErrProductInvalid) {
			fail(w, r, http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			// 覆盖三种情况：不存在 / 不是自己的 / 未过审却想上架 / 已售。
			// 统一 404 是有意的——不向调用方区分"没有"和"不是你的"。
			fail(w, r, http.StatusNotFound, errors.New("商品不存在、不属于你，或当前状态不允许该操作"))
			return
		}
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, p)
}

func (s *Server) listProducts(w http.ResponseWriter, r *http.Request) {
	products, err := s.tradingSvc.ListProducts(r.Context(), r.URL.Query().Get("prod_type"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// mine=1：只看当前用户发布的商品（含已下架/已售，未登录返回空）
	if r.URL.Query().Get("mine") == "1" {
		a, ok := authenticatedActor(r)
		if !ok {
			paginatedRespond(w, r, []domain.DroneProduct{}, 0)
			return
		}
		mine := make([]domain.DroneProduct, 0, len(products))
		for _, p := range products {
			if p.SellerID == a.ID {
				mine = append(mine, p)
			}
		}
		s.fillCoverDimensions(r.Context(), mine)
		paginatedRespond(w, r, mine, len(mine))
		return
	}
	// 公开列表口径：审核通过 + 在售，两个维度缺一不可（service.ProductVisibleInHall）。
	// ⚠ 这是拆分审核维度时必须同步改的那一处：只判 status 的话，任何把 status 置成
	// listed 的路径（管理端 PUT、审核通过、重新上架）都会让未审核商品直接出现在商城。
	listed := make([]domain.DroneProduct, 0, len(products))
	for _, p := range products {
		if service.ProductVisibleInHall(p) {
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
	// P1 脱敏：公开列表返回前把手机号注册用户的 seller_id（user-<手机号>）替换为 user-***，
	// 防止公开接口批量抓取手机号；mine=1 分支返回本人商品不脱敏。
	// P1 脱敏：只作用于 seller_id（手机号注册用户的 ID 形如 user-<手机号>）。
	// seller_name 现在是卖家昵称，不是 ID，不再需要脱敏——此前因为它落库即 a.ID
	// 才被迫一起哈希，结果买家看到的"卖家"是一串哈希。
	for i := range listed {
		listed[i].SellerID = maskUserID(listed[i].SellerID)
	}
	s.fillCoverDimensions(r.Context(), listed)
	paginatedRespond(w, r, listed, len(listed))
}

// GET /api/v1/products/{id} — 商品详情（公开，浏览量+1）
// 本 handler 同时服务管理端路由 GET /api/v1/admin/products/{id}（admin_list_routes.go），
// 管理端需要真实 seller_id，不脱敏。
func (s *Server) getProductDetail(w http.ResponseWriter, r *http.Request) {
	p, err := s.tradingSvc.GetProductAndCountView(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	// 待审核/下架/已售商品不公开：非卖家本人且非管理端请求一律 404。
	// 卖家本人可看自己的待审核商品（发布管理需要），管理端路由（/api/v1/admin/products/{id}）
	// 复用本 handler 但 isAdminRequest 放行，管理后台可看全部状态。
	// 与公开列表同一口径（ProductVisibleInHall），避免出现"列表里搜不到、直链能打开"。
	// 卖家本人可看自己的待审/被驳回商品（发布管理需要），管理端路由放行全部状态。
	if !service.ProductVisibleInHall(p) && !isAdminRequest(r) {
		if a, ok := authenticatedActor(r); !ok || a.ID != p.SellerID {
			fail(w, r, http.StatusNotFound, errors.New("product not found"))
			return
		}
	}
	// P1 脱敏：公开请求返回前替换手机号注册用户的 seller_id/seller_name，防止手机号泄露。
	// 商品本人（卖家）查看自己的商品时保留真实 ID——前端 isOwnProduct 依赖它
	// 隐藏"立即购买"按钮（自买后端同样拒绝）。
	if !isAdminRequest(r) {
		if a, ok := authenticatedActor(r); !ok || a.ID != p.SellerID {
			// 只脱敏 seller_id；seller_name 是卖家昵称（可读），脱敏它等于让买家看不到卖家。
			p.SellerID = maskUserID(p.SellerID)
		}
	}
	// 详情页同样需要封面比例（顶部大图按原比例显示，不裁）
	one := []domain.DroneProduct{p}
	s.fillCoverDimensions(r.Context(), one)
	p = one[0]
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
		PriceMode   string   `json:"price_mode"`
		Delivery    string   `json:"delivery"`
		PriceFen    int64    `json:"price_fen"`
		Status      string   `json:"status"`
		Images      []string `json:"images"`
		DetailImages []string `json:"detail_images"`
		SellerName  string   `json:"seller_name"`
	}
	if err := decode(r, &in); err != nil || in.Title == "" {
		fail(w, r, http.StatusBadRequest, errors.New("title required"))
		return
	}
	now := time.Now()
	p := domain.DroneProduct{
		// 管理端建商品没有真实用户，用固定平台卖家 ID：空 SellerID 会让下单后的订单
		// seller_id 为空串，订单归属与售后流程都依赖它。
		SellerID:    platformSellerID,
		Title:       in.Title,
		Description: in.Description,
		ProdType:    domain.ProductType(in.ProdType),
		Brand:       in.Brand,
		Model:       in.Model,
		Condition:   in.Condition,
		PriceMode:   in.PriceMode,
		Delivery:    in.Delivery,
		PriceFen:    in.PriceFen,
		Status:       in.Status,
		Images:       in.Images,
		DetailImages: in.DetailImages,
		SellerName:   in.SellerName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if p.Status == "" {
		p.Status = "listed"
	}
	if p.ProdType == "" {
		p.ProdType = domain.ProductDrone
	}
	// 单次落库。此前是 CreateProduct（内部固定 status=listed）+ UpdateProduct 回写真实状态，
	// 同一个商品写两遍、中间态还会被并发读到；ID 也由服务端生成（管理端传的 id 一律忽略，
	// 避免客户端指定主键）。
	created, err := s.tradingSvc.CreateProductByAdmin(r.Context(), p)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
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
		PriceFen    *int64   `json:"price_fen"` // 指针区分"未传"与"传 0"（改面议价）
		PriceMode   string   `json:"price_mode"`
		Delivery    string   `json:"delivery"`
		Status       string   `json:"status"`
		Images       []string `json:"images"`
		DetailImages []string `json:"detail_images"`
		SellerName   string   `json:"seller_name"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	existing, err := s.tradingSvc.GetProduct(r.Context(), r.PathValue("id"))
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
	if in.PriceFen != nil {
		existing.PriceFen = *in.PriceFen
	}
	if strings.TrimSpace(in.PriceMode) != "" {
		existing.PriceMode = strings.TrimSpace(in.PriceMode)
	}
	if strings.TrimSpace(in.Delivery) != "" {
		existing.Delivery = strings.TrimSpace(in.Delivery)
	}
	if in.Status != "" {
		existing.Status = in.Status
	}
	if in.Images != nil {
		existing.Images = in.Images
	}
	// 详情图：nil = 未传（保持原值），空数组 = 显式清空——与 images 同语义
	if in.DetailImages != nil {
		existing.DetailImages = in.DetailImages
	}
	if in.SellerName != "" {
		existing.SellerName = in.SellerName
	}
	updated, err := s.tradingSvc.UpdateProduct(r.Context(), existing)
	if err != nil {
		// 走统一映射：字段不合法 → 400（此前一律 500，前端分不清"填错了"和"服务挂了"）。
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, updated)
}

// DELETE /api/v1/admin/products/{id} — 管理后台删除商品（进回收站，非物理删除）
// 有进行中订单时返回 409，前端应提示先处理订单。
func (s *Server) adminDeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.tradingSvc.DeleteProduct(r.Context(), id); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok", "recycled": "true"})
}

// POST /api/v1/admin/products/batch-status — 批量上下架
//
// 替代"前端逐行 PUT 整行"：一条 UPDATE，只动 status 列，不碰其他字段。
// 返回实际改动行数——已售与回收站的行会被跳过，前端据此提示"部分未处理"。
func (s *Server) batchSetProductStatus(w http.ResponseWriter, r *http.Request) {
	var in struct {
		IDs    []string `json:"ids"`
		Status string   `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	n, err := s.tradingSvc.BulkSetProductStatus(r.Context(), in.IDs, in.Status)
	if err != nil {
		if errors.Is(err, service.ErrProductInvalid) {
			fail(w, r, http.StatusBadRequest, err)
			return
		}
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	adminID := ""
	if a, ok := authenticatedActor(r); ok {
		adminID = a.ID
	}
	s.audit(r.Context(), adminID, "batch_product_status", "product", strings.Join(in.IDs, ","), in.Status)
	respond(w, r, http.StatusOK, map[string]any{"updated": n, "requested": len(in.IDs)})
}

// POST /api/v1/admin/products/{id}/review — 商品审核（通过 / 驳回）
//
// 审核是**独立动作**，不是"改 status"：驳回必须给原因、留审核人/时间，并写审计。
// 此前管理端的"通过/驳回"只是给 PUT 传一个 status，驳回写的是 removed——
// 与"卖家主动下架"同值，卖家分不清，也没有原因和审计。
func (s *Server) adminReviewProduct(w http.ResponseWriter, r *http.Request) {
	var in struct {
		CheckStatus string `json:"check_status"` // passed=通过 / rejected=驳回
		CheckReason string `json:"check_reason"` // 驳回必填
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	adminID := ""
	if a, ok := authenticatedActor(r); ok {
		adminID = a.ID
	}
	p, err := s.tradingSvc.ReviewProduct(r.Context(), domain.Actor{ID: adminID}, r.PathValue("id"), in.CheckStatus, in.CheckReason)
	if err != nil {
		if errors.Is(err, service.ErrProductReviewInvalid) {
			fail(w, r, http.StatusBadRequest, err)
			return
		}
		adminFail(w, r, err)
		return
	}
	// 商品路径此前**完全没有审计**（对比订单路径有）。审核结果 + 驳回原因一并留痕，
	// 事后可复盘"谁在什么时候以什么理由驳回了哪件商品"。
	s.audit(r.Context(), adminID, "review_product", "product", p.ID, p.CheckStatus+":"+p.CheckReason)
	respond(w, r, http.StatusOK, p)
}

// POST /api/v1/admin/products/{id}/restore — 回收站还原
func (s *Server) adminRestoreProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.tradingSvc.UndeleteProduct(r.Context(), id); err != nil {
		adminFail(w, r, err)
		return
	}
	p, err := s.tradingSvc.GetProduct(r.Context(), id)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, p)
}

// GET /api/v1/admin/products — 管理后台商品列表
// deleted=1 时返回回收站（仅已软删除的行），其余情况只返回未删除的行。
func (s *Server) listAdminProducts(w http.ResponseWriter, r *http.Request) {
	var (
		products []domain.DroneProduct
		err      error
	)
	if r.URL.Query().Get("deleted") == "1" {
		products, err = s.tradingSvc.ListDeletedProducts(r.Context())
	} else {
		products, err = s.tradingSvc.ListProducts(r.Context(), r.URL.Query().Get("prod_type"))
	}
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	filtered, total := adminListFilter(products, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(p domain.DroneProduct) string { return p.Title },
		func(p domain.DroneProduct) string { return p.Status })
	// 审核状态是**独立维度**：与 status 同时给定时两者是 AND 关系
	//（如 status=listed & check_status=passed = 真正在售的商品）。
	if cs := strings.TrimSpace(r.URL.Query().Get("check_status")); cs != "" {
		byCheck := make([]domain.DroneProduct, 0, len(filtered))
		for _, p := range filtered {
			if p.CheckStatus == cs {
				byCheck = append(byCheck, p)
			}
		}
		filtered, total = byCheck, len(byCheck)
	}
	paginatedRespond(w, r, filtered, total)
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
	rp, err := s.tradingSvc.CreateRepair(r.Context(), a, in.ProductDesc, in.FaultDesc)
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
	repairs, err := s.tradingSvc.ListMyRepairs(r.Context(), a)
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
	p, err := s.insuranceSvc.CreatePolicy(r.Context(), a, in.DroneModel, in.DroneSN, in.PolicyType, in.PremiumFen, in.CoverageFen, in.StartDate, in.EndDate)
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
	policies, err := s.insuranceSvc.ListMyPolicies(r.Context(), a)
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
	i, err := s.insuranceSvc.CreateInspection(r.Context(), a, in.DroneModel, in.DroneSN, in.InspectDate, in.ExpireDate)
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
	inspections, err := s.insuranceSvc.ListMyInspections(r.Context(), a)
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
	l, err := s.financeSvc.ApplyLoan(r.Context(), a, in.AmountFen, in.TermMonths, in.Purpose)
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
	loans, err := s.financeSvc.ListMyLoans(r.Context(), a)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, loans)
}
