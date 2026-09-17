package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type TradingService struct {
	prodRepo   repository.ProductRepository
	repairRepo repository.RepairRepository
	// orderRepo 仅用于"删除商品前校验是否还有有效订单"这一条不变式。
	// 可为 nil（dev/单测）：此时软删除仍生效，只是少了这层拦截。
	orderRepo repository.TradeOrderRepository
	// userRepo 用于把卖家的**展示名**（昵称）落库，而不是落用户 ID。
	// 可为 nil：此时 SellerName 留空，前端回退到"认证商家"文案。
	userRepo repository.UserRepository
}

func NewTradingService(pr repository.ProductRepository, rr repository.RepairRepository, or repository.TradeOrderRepository, ur repository.UserRepository) *TradingService {
	return &TradingService{prodRepo: pr, repairRepo: rr, orderRepo: or, userRepo: ur}
}

// ErrProductInvalid 商品字段校验失败。Handler 据此映射 400（此前发布失败一律 403，
// 把"填错了"和"没权限"混为一谈，前端无法给出正确提示）。
var ErrProductInvalid = errors.New("商品信息不合法")

// ErrProductInTrade 商品仍有未取消的订单，不可删除。
// 删掉正在交易中的商品会让买家的订单详情/商品链接指向不存在的记录。
var ErrProductInTrade = errors.New("该商品有进行中的订单，无法删除")

// validProductTypes 商品类型白名单，与 domain 的 7 个 ProductType 常量一一对应。
// 此前 handler 直接 `domain.ProductType(in.ProdType)` 强转，空串/任意字符串都能落库。
var validProductTypes = map[domain.ProductType]bool{
	domain.ProductDrone:       true,
	domain.ProductPart:        true,
	domain.ProductRepair:      true,
	domain.ProductAerial:      true,
	domain.ProductTestFly:     true,
	domain.ProductCalibration: true,
	domain.ProductAirspace:    true,
}

const (
	maxProductTitleRunes = 100
	maxProductDescRunes  = 5000
	maxProductImages     = 9
	// maxProductCheckReasonRunes 驳回原因长度上限。
	maxProductCheckReasonRunes = 200
)

// ErrProductReviewInvalid 审核参数不合法（Handler 映射 400）。
var ErrProductReviewInvalid = errors.New("审核参数不合法")

// validDeliveries 交付方式白名单（空串 = 卖家未选，合法，下单时按商品类型兜底）。
var validDeliveries = map[string]bool{
	"":                        true,
	domain.DeliveryPickup:     true,
	domain.DeliveryCity:       true,
	domain.DeliveryLogistics:  true,
	domain.DeliveryNegotiable: true,
}

// OrderNeedsReceiver 判断某商品下单时是否必须提供收货信息。
//
// **优先看卖家选定的交付方式**——这才是这个字段存在的意义：
//
//	logistics(物流发货) / city(同城配送) → 必须给地址，否则卖家没法送
//	pickup(自提)                        → 不需要，线下自取
//	negotiable(可协商) / 未指定          → 按商品类型兜底（实物需要，服务类不需要）
//
// 此前交付方式是采集了却在提交时丢弃的，只能按商品类型猜：卖家选了「自提」，
// 买家仍被要求填收货地址。
func OrderNeedsReceiver(p domain.DroneProduct) bool {
	switch p.Delivery {
	case domain.DeliveryPickup:
		return false
	case domain.DeliveryCity, domain.DeliveryLogistics:
		return true
	}
	return p.ProdType == domain.ProductDrone || p.ProdType == domain.ProductPart
}

// ProductVisibleInHall 商品是否可出现在公开商城：**审核通过**且**在售**，缺一不可。
//
// 这个合取式是拆分审核维度的必要条件：拆开之后若仍只判 status，任何把 status 置成
// listed 的路径（管理端改状态、审核通过、卖家重新上架）都会让未审商品直接上架。
// 显式要求 check_status='passed' 才堵得住。
func ProductVisibleInHall(p domain.DroneProduct) bool {
	return p.CheckStatus == domain.ProductCheckPassed && p.Status == "listed"
}

// ReviewProduct 审核商品（管理端）：通过 → 上架；驳回 → 必须给原因且**不改变上架状态**。
//
// 拆分前"驳回"只能写 status='removed'，与"卖家自己下架"同值——卖家在小程序
// "我的发布"里分不清自己是被驳回还是主动下的，也没有原因可看。
func (s *TradingService) ReviewProduct(ctx context.Context, reviewer domain.Actor, id, checkStatus, reason string) (domain.DroneProduct, error) {
	reason = strings.TrimSpace(reason)
	switch checkStatus {
	case domain.ProductCheckPassed:
		// 通过不需要原因，同时清掉上一次的驳回理由（重新审核通过后不该再显示旧理由）。
		reason = ""
	case domain.ProductCheckRejected:
		if reason == "" {
			return domain.DroneProduct{}, fmt.Errorf("%w：驳回必须填写原因", ErrProductReviewInvalid)
		}
		if utf8.RuneCountInString(reason) > maxProductCheckReasonRunes {
			return domain.DroneProduct{}, fmt.Errorf("%w：驳回原因不能超过 %d 个字", ErrProductReviewInvalid, maxProductCheckReasonRunes)
		}
	default:
		return domain.DroneProduct{}, fmt.Errorf("%w：审核结果只能是 passed 或 rejected", ErrProductReviewInvalid)
	}
	return s.prodRepo.ReviewProduct(ctx, id, checkStatus, reason, reviewer.ID)
}

// sellerDisplayName 卖家展示名：取用户昵称。
// 此前 SellerName 直接落库 a.ID，公开接口为了不泄露手机号又不得不对它做哈希脱敏
// （httpapi 的 maskUserID），最终买家在小程序商品详情看到的"卖家"是一串哈希
// （mall/detail.vue 的卖家行 / 店铺卡片 / "该商品由 X 发布"）。展示名必须是可读文本，
// 脱敏只应作用于 seller_id。
func (s *TradingService) sellerDisplayName(ctx context.Context, userID string) string {
	if s.userRepo == nil {
		return ""
	}
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		slog.Warn("商品发布：查询卖家昵称失败，回退占位名", "user", userID, "err", err)
		return "平台用户"
	}
	if name := strings.TrimSpace(u.Name); name != "" {
		return name
	}
	return "平台用户"
}

// normalizeAndValidate 商品字段校验 + 归一。发布、卖家自助编辑、管理端编辑**共用同一套口径**
// ——分别写三份的话，迟早出现"发布时校验、编辑时绕过"。
func normalizeAndValidate(p *domain.DroneProduct) error {
	p.Title = strings.TrimSpace(p.Title)
	if p.Title == "" {
		return fmt.Errorf("%w：商品名称不能为空", ErrProductInvalid)
	}
	if utf8.RuneCountInString(p.Title) > maxProductTitleRunes {
		return fmt.Errorf("%w：商品名称不能超过 %d 个字", ErrProductInvalid, maxProductTitleRunes)
	}
	if !validProductTypes[p.ProdType] {
		return fmt.Errorf("%w：请选择有效的商品类型", ErrProductInvalid)
	}
	if utf8.RuneCountInString(p.Description) > maxProductDescRunes {
		return fmt.Errorf("%w：商品描述不能超过 %d 个字", ErrProductInvalid, maxProductDescRunes)
	}
	// 成色：空值归一为全新（DB 默认 'new'），非空则必须在新/旧之间。
	switch strings.TrimSpace(p.Condition) {
	case "":
		p.Condition = "new"
	case "new", "used":
		p.Condition = strings.TrimSpace(p.Condition)
	default:
		return fmt.Errorf("%w：成色只能是 new 或 used", ErrProductInvalid)
	}
	if p.PriceFen < 0 {
		return fmt.Errorf("%w：价格不能为负数", ErrProductInvalid)
	}
	if len(p.Images) > maxProductImages {
		return fmt.Errorf("%w：商品图片最多 %d 张", ErrProductInvalid, maxProductImages)
	}
	// 价格模式与金额必须自洽。未指定价格模式时按金额推断：
	// price_fen=0 → 面议（这正是它此前的展示语义），非 0 → 明码标价。
	// 这是给尚未传 price_mode 的旧调用方的兼容路径；小程序与管理端都会显式传。
	if strings.TrimSpace(p.PriceMode) == "" {
		if p.PriceFen == 0 {
			p.PriceMode = domain.PriceModeNegotiable
		} else {
			p.PriceMode = domain.PriceModeFixed
		}
	}
	switch p.PriceMode {
	case domain.PriceModeFixed:
		// 明码标价必须给出真实价格：0 元此前既表示"面议"又表示"填了 0"，必须堵死。
		if p.PriceFen <= 0 {
			return fmt.Errorf("%w：明码标价必须填写大于 0 的价格（面议请选择「面议」）", ErrProductInvalid)
		}
	case domain.PriceModeNegotiable:
		if p.PriceFen != 0 {
			return fmt.Errorf("%w：选择面议时价格必须留空", ErrProductInvalid)
		}
	default:
		return fmt.Errorf("%w：价格模式只能是 fixed 或 negotiable", ErrProductInvalid)
	}
	// 交付方式：空串合法（卖家未选，下单时按类型兜底），非空必须在白名单内。
	p.Delivery = strings.TrimSpace(p.Delivery)
	if !validDeliveries[p.Delivery] {
		return fmt.Errorf("%w：交付方式只能是自提 / 同城配送 / 物流发货 / 可协商", ErrProductInvalid)
	}
	return nil
}

// createProduct 用户发布与管理端新增共用的落库路径。
func (s *TradingService) createProduct(ctx context.Context, p domain.DroneProduct) (domain.DroneProduct, error) {
	if err := normalizeAndValidate(&p); err != nil {
		return domain.DroneProduct{}, err
	}
	now := time.Now()
	p.ID = nextID("product")
	p.Version = 1
	p.CreatedAt = now
	p.UpdatedAt = now
	// 状态由调用方决定：用户发布固定 pending（等审核），管理端新增可指定。
	if strings.TrimSpace(p.Status) == "" {
		p.Status = "pending"
	}
	return s.prodRepo.Create(ctx, p)
}

// CreateProduct 用户发布商品：固定进入待审核，卖家展示名取发布者昵称。
func (s *TradingService) CreateProduct(ctx context.Context, a domain.Actor, prodType domain.ProductType, title, desc, brand, model, condition, delivery, priceMode string, priceFen int64, images, detailImages []string) (domain.DroneProduct, error) {
	return s.createProduct(ctx, domain.DroneProduct{
		SellerID:    a.ID,
		SellerName:  s.sellerDisplayName(ctx, a.ID),
		ProdType:    prodType,
		Title:       title,
		Description: desc,
		PriceFen:    priceFen,
		Images:      images,
		Brand:       brand,
		Model:       model,
		Condition: condition,
		// 交付方式落库：它决定下单时是否强制收货地址（OrderNeedsReceiver）。
		Delivery:     delivery,
		PriceMode:    priceMode,
		DetailImages: detailImages,
		// 用户发布固定"待审核"：审核维度 pending、上架维度 pending（未上架）。
		// 管理端审核通过后才会变成 check_status=passed + status=listed。
		Status:      "pending",
		CheckStatus: domain.ProductCheckPending,
	})
}

// CreateProductByAdmin 管理端新增商品：卖家展示名由管理端指定（默认"平台自营"），
// 状态也由管理端指定（可 pending 也可直接 listed），其余校验与用户发布完全一致。
func (s *TradingService) CreateProductByAdmin(ctx context.Context, p domain.DroneProduct) (domain.DroneProduct, error) {
	if strings.TrimSpace(p.SellerName) == "" {
		p.SellerName = "平台自营"
	}
	// 管理端建的商品视为已审核（管理端本身就是审核方），但**仍然要求显式状态**：
	// 只有 check_status=passed + status=listed 才会出现在公开商城。
	if p.CheckStatus == "" {
		p.CheckStatus = domain.ProductCheckPassed
	}
	return s.createProduct(ctx, p)
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

// ToggleProductFavorite 收藏/取消收藏商品（登录用户可收藏任意存在商品）。
func (s *TradingService) ToggleProductFavorite(ctx context.Context, userID, productID string, favorite bool) error {
	if _, err := s.prodRepo.FindByID(ctx, productID); err != nil {
		return err
	}
	if favorite {
		return s.prodRepo.FavoriteProduct(ctx, userID, productID)
	}
	return s.prodRepo.UnfavoriteProduct(ctx, userID, productID)
}

// ListFavoriteProducts 当前用户收藏的商品列表（按收藏时间倒序）。
func (s *TradingService) ListFavoriteProducts(ctx context.Context, userID string) ([]domain.DroneProduct, error) {
	return s.prodRepo.ListFavoriteProducts(ctx, userID)
}

// GetProductAndCountView 详情访问：浏览量 +1 后返回（先读旧值再递增）
func (s *TradingService) GetProductAndCountView(ctx context.Context, id string) (domain.DroneProduct, error) {
	p, err := s.prodRepo.FindByID(ctx, id)
	if err != nil {
		return domain.DroneProduct{}, err
	}
	p.Views++
	if err := s.prodRepo.IncrementViews(ctx, id); err != nil {
		// 浏览量 +1 失败不阻断详情读取（返回陈旧计数），但必须记录，避免静默吞错。
		slog.Warn("increment product views failed", "product_id", id, "err", err)
	}
	return p, nil
}

// UpdateProduct 更新商品（管理后台用）：与发布共用同一套字段校验，
// 避免出现"发布时校验、管理端编辑时绕过"。
func (s *TradingService) UpdateProduct(ctx context.Context, p domain.DroneProduct) (domain.DroneProduct, error) {
	if err := normalizeAndValidate(&p); err != nil {
		return domain.DroneProduct{}, err
	}
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

// DeleteProduct 删除商品（管理后台用）：进回收站（软删除），不是物理删除。
//
// 两条不变式：
//  1. 仍有"有效订单"（未取消）时拒绝——商品正在交易中被删，买家的订单详情与商品链接
//     就会指向不存在的记录（trade_orders.product_id 无外键，DB 不会拦）。
//  2. 落到 deleted_at 而不是 DELETE：历史订单还要靠这一行显示商品名（见 repo.ListByIDs）。
func (s *TradingService) DeleteProduct(ctx context.Context, id string) error {
	if s.orderRepo != nil {
		live, err := s.orderRepo.HasLiveOrderForProduct(ctx, id)
		if err != nil {
			return fmt.Errorf("校验商品关联订单失败: %w", err)
		}
		if live {
			return ErrProductInTrade
		}
	}
	return s.prodRepo.SoftDelete(ctx, id)
}

// ProductEditInput 卖家自助编辑允许修改的字段集合。
// 归属、状态、审核结果都不在这里——那些由服务层决定，客户端说了不算。
type ProductEditInput struct {
	ProdType    domain.ProductType
	Title       string
	Description string
	Brand       string
	Model       string
	Condition   string
	PriceMode   string
	Delivery    string
	PriceFen    int64
	Images      []string
	// DetailImages 详情图（详情区长图），与 Images（顶部图集）分工不同
	DetailImages []string
}

// UpdateMyProduct 卖家自助编辑自己的商品。
//
// 归属校验在服务层（不是 Handler）：只能改自己的商品；改别人的按"不存在"处理，
// 不泄露"这件商品存在但不是你的"。
//
// **任何内容编辑都会把商品退回待审核**（check_status=pending, status=pending）。
// 理由：审核的对象是内容本身，内容变了原审核结论就失效了——否则卖家可以先发一件
// 合规商品过审、再改成违规内容，审核形同虚设。
func (s *TradingService) UpdateMyProduct(ctx context.Context, a domain.Actor, id string, in ProductEditInput) (domain.DroneProduct, error) {
	existing, err := s.prodRepo.FindByID(ctx, id)
	if err != nil {
		return domain.DroneProduct{}, err
	}
	if existing.SellerID != a.ID {
		return domain.DroneProduct{}, repository.ErrNotFound
	}
	if existing.Status == "sold" {
		return domain.DroneProduct{}, fmt.Errorf("%w：商品已售出，无法编辑", ErrProductInvalid)
	}
	existing.ProdType = in.ProdType
	existing.Title = in.Title
	existing.Description = in.Description
	existing.Brand = in.Brand
	existing.Model = in.Model
	existing.Condition = in.Condition
	existing.PriceMode = in.PriceMode
	existing.Delivery = in.Delivery
	existing.PriceFen = in.PriceFen
	existing.Images = in.Images
	existing.DetailImages = in.DetailImages
	// 退回待审核（见方法注释）。
	existing.CheckStatus = domain.ProductCheckPending
	existing.CheckReason = ""
	existing.ReviewedAt = nil
	existing.ReviewedBy = ""
	existing.Status = "pending"
	if err := normalizeAndValidate(&existing); err != nil {
		return domain.DroneProduct{}, err
	}
	return s.prodRepo.Update(ctx, existing)
}

// BulkSetProductStatus 管理端批量上下架：单条条件更新，返回实际改动行数。
//
// 与逐行 PUT 的差别不只是性能：逐行会把整行写回去，只改一个 status 却动了所有列，
// 并发编辑时后写覆盖先写（前端此前"传整行"正是接口契约缺单字段更新打的补丁）。
// 已售与回收站的行由仓储层条件跳过，不计入改动行数。
func (s *TradingService) BulkSetProductStatus(ctx context.Context, ids []string, status string) (int, error) {
	switch status {
	case "listed", "removed":
	default:
		return 0, fmt.Errorf("%w：只能批量上架或下架", ErrProductInvalid)
	}
	if len(ids) == 0 {
		return 0, fmt.Errorf("%w：请先选择商品", ErrProductInvalid)
	}
	return s.prodRepo.SetStatusBulk(ctx, ids, status)
}

// SetMyProductStatus 卖家自助上下架：只允许 listed / removed。
//
// sold 由订单系统占用，卖家不能自己改；未通过审核的商品也不能自己上架
// （两个约束都在 repository 的条件更新里，不是"先查后改"）。
func (s *TradingService) SetMyProductStatus(ctx context.Context, a domain.Actor, id, status string) (domain.DroneProduct, error) {
	switch status {
	case "listed", "removed":
	default:
		return domain.DroneProduct{}, fmt.Errorf("%w：只能上架或下架", ErrProductInvalid)
	}
	return s.prodRepo.SetProductStatus(ctx, id, a.ID, status)
}

// UndeleteProduct 回收站还原（管理后台用）。
func (s *TradingService) UndeleteProduct(ctx context.Context, id string) error {
	return s.prodRepo.Undelete(ctx, id)
}

// ListDeletedProducts 回收站列表（管理后台用）。
func (s *TradingService) ListDeletedProducts(ctx context.Context) ([]domain.DroneProduct, error) {
	return s.prodRepo.ListDeleted(ctx)
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
