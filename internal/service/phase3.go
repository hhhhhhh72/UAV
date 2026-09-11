package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Enrollment ----

type EnrollmentService struct {
	repo       repository.EnrollmentRepository
	courseRepo repository.CourseRepository // 报名容量检查（course.MaxStudents）用
}

func NewEnrollmentService(repo repository.EnrollmentRepository, courseRepo repository.CourseRepository) *EnrollmentService {
	return &EnrollmentService{repo: repo, courseRepo: courseRepo}
}

// EnrollmentForm 培训报名表单数据（小程序 register.vue 12 字段）。
// PaidAmountFen 由服务端填充（payAndEnroll 冻结成功后写入），客户端不可见不可改。
type EnrollmentForm struct {
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	IDCard      string `json:"idCard"`
	Gender      string `json:"gender"`
	Birthday    string `json:"birthday"`
	Email       string `json:"email"`
	Education   string `json:"education"`
	Experience  string `json:"experience"`
	Photo       string `json:"photo"`
	IDCardImage string `json:"idCardImage"`
	IDCardBack  string `json:"idCardBack"`
	NoCrime     string `json:"noCrime"`

	// PaidAmountFen 报名时冻结的学费（分）。仅 payAndEnroll 在冻结成功后填充；
	// 免费/普通报名为 0。completeEnrollment 按此金额释放，与课程实时价格解耦。
	PaidAmountFen int64 `json:"-"`
}

// All 管理端全量报名记录（分页）。
func (s *EnrollmentService) All(ctx context.Context, offset, limit int) ([]domain.Enrollment, int, error) {
	return s.repo.ListAll(ctx, offset, limit)
}

// ListByCourseForActor 按课程查报名（含 PII）：课程归属者或管理员可查。
func (s *EnrollmentService) ListByCourseForActor(ctx context.Context, a domain.Actor, courseID string) ([]domain.Enrollment, error) {
	if s.courseRepo == nil {
		return nil, errors.New("course repository not available")
	}
	c, err := s.courseRepo.FindByID(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("course %s: %w", courseID, err)
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin && c.OrgID != a.ID {
		return nil, errors.New("only course owner or admin can view enrollments")
	}
	return s.repo.ListByCourse(ctx, courseID)
}

// Review 机构/管理员审核报名：enrolled/paid → approved / rejected（拒绝需原因）。
func (s *EnrollmentService) Review(ctx context.Context, a domain.Actor, id, action, reason string) (domain.Enrollment, error) {
	if s.courseRepo == nil {
		return domain.Enrollment{}, errors.New("course repository not available")
	}
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Enrollment{}, fmt.Errorf("enrollment %s: %w", id, err)
	}
	c, err := s.courseRepo.FindByID(ctx, e.CourseID)
	if err != nil {
		return domain.Enrollment{}, fmt.Errorf("course %s: %w", e.CourseID, err)
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin && c.OrgID != a.ID {
		return domain.Enrollment{}, errors.New("only course owner or admin can review enrollments")
	}
	if e.Status != "enrolled" && e.Status != "paid" {
		return domain.Enrollment{}, fmt.Errorf("only enrolled/paid enrollments can be reviewed (current %s)", e.Status)
	}
	if action != "approve" && action != "reject" {
		return domain.Enrollment{}, errors.New("invalid review action (approve/reject)")
	}
	if action == "reject" && strings.TrimSpace(reason) == "" {
		return domain.Enrollment{}, errors.New("reject reason is required")
	}
	to := "approved"
	if action == "reject" {
		to = "rejected"
	}
	ok, err := s.repo.UpdateStatusCas(ctx, id, e.Status, to)
	if err != nil {
		return domain.Enrollment{}, fmt.Errorf("update enrollment status: %w", err)
	}
	if !ok {
		return domain.Enrollment{}, errors.New("报名状态已变更，请刷新后重试")
	}
	e.Status = to
	e.ReviewNote = reason
	updated, err := s.repo.Update(ctx, e)
	if err != nil {
		return domain.Enrollment{}, fmt.Errorf("save review note: %w", err)
	}
	return updated, nil
}

// FindByID 按报名 ID 定位单条记录（管理端完成报名/编辑用，替代全表扫描）。
func (s *EnrollmentService) FindByID(ctx context.Context, id string) (domain.Enrollment, error) {
	return s.repo.FindByID(ctx, id)
}

// UpdateStatusCas 原子状态迁移（completed 终态 CAS）：并发完成报名时只有一方能置
// completed，另一方看到 false 直接 409——防重复释放学费（与仓储层 UPDATE WHERE 对齐）。
func (s *EnrollmentService) UpdateStatusCas(ctx context.Context, id, from, to string) (bool, error) {
	return s.repo.UpdateStatusCas(ctx, id, from, to)
}

// FindByUserAndCourse 查询用户对某课程是否已报名（课程详情"我的报名状态"标记用）。
func (s *EnrollmentService) FindByUserAndCourse(ctx context.Context, userID, courseID string) (domain.Enrollment, bool, error) {
	return s.repo.FindByUserAndCourse(ctx, userID, courseID)
}

func (s *EnrollmentService) Enroll(ctx context.Context, userID, courseID string, form EnrollmentForm) (domain.Enrollment, error) {
	// 并发防重复：check-then-insert 加进程内锁（双请求同时通过查重会重复报名，
	// 且付费报名会重复扣冻结金额）。
	unlock := lockByKey("enroll|" + userID + "|" + courseID)
	defer unlock()
	// 防超卖：容量检查（ListByCourse 计数 >= MaxStudents）是 check-then-act，
	// 不同用户并发报名同一课程会双双通过检查；课程维度锁串行化整个检查+创建。
	unlockCourse := lockByKey("enroll-course|" + courseID)
	defer unlockCourse()
	if _, ok, err := s.repo.FindByUserAndCourse(ctx, userID, courseID); err != nil {
		return domain.Enrollment{}, fmt.Errorf("check existing enrollment: %w", err)
	} else if ok {
		return domain.Enrollment{}, fmt.Errorf("already enrolled")
	}
	// 容量与状态门禁：full/upcoming 不可报名（与前端禁用一致，API 不再可绕过）；
	// 容量以 enrolled_count 列为真值（报名成功即 +1，remain 同步）——此前读列不维护，
	// 前端"仅剩 N/已满"全部是假数据，且 ListByCourse 全量计数把完成/驳回也算占座。
	if s.courseRepo != nil {
		if c, err := s.courseRepo.FindByID(ctx, courseID); err == nil {
			// 防自购自卖：课程发布者（OrgID=本人）不可报名自己的课程——机构自导自演报名会
			// 污染学员数据（刷报名数），且学费结算方向闭环（自己冻结-自己回收），违背托管金语义。
			if c.OrgID != "" && c.OrgID == userID {
				return domain.Enrollment{}, fmt.Errorf("不能报名自己发布的课程")
			}
			// 资金逻辑：付费课程禁止免费报名——此前 /enroll 不校验 price_fen，
			// 付费课可 0 元直接报名成功（绕过 pay-and-enroll 的学费冻结）。付费
			// 报名必须携带与课程价一致的 PaidAmountFen（payAndEnroll 冻结后固化写入）。
			if c.PriceFen > 0 && form.PaidAmountFen != c.PriceFen {
				return domain.Enrollment{}, fmt.Errorf("paid course requires payment (free enrollment not allowed)")
			}
			if c.Status == "full" || c.Status == "upcoming" {
				return domain.Enrollment{}, fmt.Errorf("course is not open for enrollment (status %s)", c.Status)
			}
			if c.MaxStudents > 0 && c.EnrolledCount >= c.MaxStudents {
				return domain.Enrollment{}, fmt.Errorf("course is full")
			}
		}
	}
	// 生日：前端提交 "YYYY-MM-DD"，解析为 DATE 语义；解析失败返回错误（不再静默丢弃）
	var birthday time.Time
	if form.Birthday != "" {
		bd, err := time.Parse("2006-01-02", form.Birthday)
		if err != nil {
			return domain.Enrollment{}, fmt.Errorf("invalid birthday format")
		}
		birthday = bd
	}
	now := time.Now()
	// 状态语义：付费报名（已冻结学费）→ paid（已缴费）；免费/待支付 → enrolled（已报名）。
	// 此前付费报名也记为 enrolled，管理端无法区分"已缴费待开班"与"仅占位报名"。
	status := "enrolled"
	if form.PaidAmountFen > 0 {
		status = "paid"
	}
	e := domain.Enrollment{ID: nextID("enroll"), CourseID: courseID, UserID: userID,
		Name: form.Name, Phone: form.Phone, IDCard: form.IDCard, Gender: form.Gender, Birthday: birthday,
		Email: form.Email, Education: form.Education, Experience: form.Experience,
		PhotoURL: form.Photo, IDCardImage: form.IDCardImage, IDCardBack: form.IDCardBack, NoCrime: form.NoCrime,
		Status: status, PaidAmountFen: form.PaidAmountFen, CreatedAt: now}
	// 先占名额（enrolled_count+1），再落报名记录；落库失败补偿 -1（学号不漂移）。
	// 仅当课程存在（FindByID 成功）才占位——兼容无课程仓储的测试与历史数据。
	bumpOK := false
	if s.courseRepo != nil {
		if c, err := s.courseRepo.FindByID(ctx, courseID); err == nil {
			_ = c
			if err := s.courseRepo.BumpEnrolled(ctx, courseID, 1); err == nil {
				bumpOK = true
			}
		}
	}
	if _, err := s.repo.Create(ctx, e); err != nil {
		if bumpOK && s.courseRepo != nil {
			_ = s.courseRepo.BumpEnrolled(ctx, courseID, -1)
		}
		return domain.Enrollment{}, err
	}
	return e, nil
}

// validEnrollmentStatus 报名状态白名单（与前端 statusLabel 对齐；completed 由管理端完成闭环写入）。
func validEnrollmentStatus(status string) bool {
	switch status {
	case "pending", "approved", "paid", "enrolled", "rejected", "completed":
		return true
	}
	return false
}

// Update 管理端编辑报名记录（基础信息 + 状态；全字段覆盖）。
// 状态校验：白名单 + 防回退（已缴费/已入学为定局状态，不允许改回待审核/驳回）。
func (s *EnrollmentService) Update(ctx context.Context, a domain.Actor, e domain.Enrollment) (domain.Enrollment, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Enrollment{}, errors.New("admin permission required")
	}
	if e.ID == "" {
		return domain.Enrollment{}, errors.New("enrollment id is required")
	}
	old, err := s.repo.FindByID(ctx, e.ID)
	if err != nil {
		return domain.Enrollment{}, err
	}
	if !validEnrollmentStatus(e.Status) {
		return domain.Enrollment{}, fmt.Errorf("invalid enrollment status %q", e.Status)
	}
	if (old.Status == "paid" || old.Status == "enrolled") && (e.Status == "pending" || e.Status == "rejected") {
		return domain.Enrollment{}, fmt.Errorf("cannot change enrollment status from %q to %q", old.Status, e.Status)
	}
	// completed 为终态：已完成（学费已释放/证书已发）的报名不可回退任何状态，防重复释放/发证
	if old.Status == "completed" && e.Status != "completed" {
		return domain.Enrollment{}, fmt.Errorf("cannot change completed enrollment status")
	}
	return s.repo.Update(ctx, e)
}

func (s *EnrollmentService) ListByCourse(ctx context.Context, courseID string) ([]domain.Enrollment, error) {
	return s.repo.ListByCourse(ctx, courseID)
}

// ListByUser 某用户全部报名（"我的报名"一次查询，避免按课程 N+1）。
func (s *EnrollmentService) ListByUser(ctx context.Context, userID string) ([]domain.Enrollment, error) {
	return s.repo.ListByUser(ctx, userID)
}

// ---- Expiry Checker ----

type ExpiryService struct{}

func NewExpiryService() *ExpiryService { return &ExpiryService{} }

// validExpireDate 排除无有效期的记录：time.Time 零值，或 PG NULL 经
// COALESCE 读成的 1970-01-01（IsZero 不命中，需用下限判断）。
func validExpireDate(t time.Time) bool {
	if t.IsZero() {
		return false
	}
	return t.After(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
}

func (s *ExpiryService) GetExpiringCerts(certs []domain.Certificate, withinDays int) []domain.Certificate {
	cutoff := time.Now().AddDate(0, 0, withinDays)
	out := []domain.Certificate{}
	for _, c := range certs {
		if c.Status == "approved" && validExpireDate(c.ExpireDate) && c.ExpireDate.Before(cutoff) {
			out = append(out, c)
		}
	}
	return out
}

func (s *ExpiryService) GetExpiringInspections(list []domain.AnnualInspection, withinDays int) []domain.AnnualInspection {
	cutoff := time.Now().AddDate(0, 0, withinDays)
	out := []domain.AnnualInspection{}
	for _, i := range list {
		if i.Status == "approved" && validExpireDate(i.ExpireDate) && i.ExpireDate.Before(cutoff) {
			out = append(out, i)
		}
	}
	return out
}

// ---- Trade Orders ----

// tradeRefType 订单资金的流水引用类型：冻结/放款/退款/转账统一用它，
// 幂等查询（HasFrozen/HasReleased/HasRefunded）与孤儿冻结补偿都按它匹配。
const tradeRefType = "trade_order"

type TradeOrderService struct {
	repo     repository.TradeOrderRepository
	prodRepo repository.ProductRepository // 订单取消时恢复商品为可售（可空）
	escrow   *EscrowService               // 资金托管（可空：未注入时退化为纯状态机，供 dev/测试）
}

func NewTradeOrderService(repo repository.TradeOrderRepository, prodRepo repository.ProductRepository) *TradeOrderService {
	return &TradeOrderService{repo: repo, prodRepo: prodRepo}
}

// SetEscrow 注入托管金服务：注入后商城订单具备真实资金闭环
// （付款冻结 → 确认收货放款 → 取消/售后退款），未注入则只走状态机。
func (s *TradeOrderService) SetEscrow(e *EscrowService) { s.escrow = e }

// freezeForOrder 付款冻结买家余额（订单金额）。余额不足返回 repository.ErrInsufficientBalance，
// Handler 映射 402 引导充值。已冻结过（重试）则跳过，不重复冻结。
func (s *TradeOrderService) freezeForOrder(ctx context.Context, o domain.TradeOrder) error {
	if s.escrow == nil || o.AmountFen <= 0 {
		return nil
	}
	if has, err := s.escrow.HasFrozen(ctx, o.BuyerID, tradeRefType, o.ID); err == nil && has {
		return nil
	}
	_, err := s.escrow.Freeze(ctx, o.BuyerID, o.AmountFen, tradeRefType, o.ID)
	return err
}

// settleToSeller 确认收货放款：买家冻结 → 卖家余额。
// EscrowService.Release 内部按 (买家, trade_order, 订单号) 幂等，重复调用不会双倍入账。
func (s *TradeOrderService) settleToSeller(ctx context.Context, o domain.TradeOrder) error {
	if s.escrow == nil || o.AmountFen <= 0 {
		return nil
	}
	if _, err := s.escrow.Release(ctx, o.BuyerID, o.SellerID, o.AmountFen, tradeRefType, o.ID); err != nil {
		return fmt.Errorf("订单 %s 放款失败: %w", o.ID, err)
	}
	return nil
}

// refundBuyerIfAny 退款给买家（幂等）：仅当这笔订单确实冻结过、且尚未退过款时执行。
// 无冻结（未付款订单）直接跳过——不需要也不应该产生退款流水。
func (s *TradeOrderService) refundBuyerIfAny(ctx context.Context, o domain.TradeOrder, amountFen int64) error {
	if s.escrow == nil || amountFen <= 0 {
		return nil
	}
	frozen, err := s.escrow.HasFrozen(ctx, o.BuyerID, tradeRefType, o.ID)
	if err != nil {
		return fmt.Errorf("查询订单冻结流水失败: %w", err)
	}
	if !frozen {
		return nil
	}
	refunded, err := s.escrow.HasRefunded(ctx, o.BuyerID, tradeRefType, o.ID)
	if err != nil {
		return fmt.Errorf("查询订单退款流水失败: %w", err)
	}
	if refunded {
		return nil
	}
	if _, err := s.escrow.Refund(ctx, o.BuyerID, amountFen, tradeRefType, o.ID); err != nil {
		return fmt.Errorf("订单 %s 退款失败: %w", o.ID, err)
	}
	return nil
}

func (s *TradeOrderService) Create(ctx context.Context, buyerID, productID, sellerID string, amountFen int64) (domain.TradeOrder, error) {
	// P3 修复：管理端建单金额/归属护栏——金额非负、三方必填、防自买自卖
	// （此前负数金额可建单，空 buyer/seller 可造残缺订单）。
	if amountFen < 0 {
		return domain.TradeOrder{}, errors.New("order amount cannot be negative")
	}
	if buyerID == "" || sellerID == "" || productID == "" {
		return domain.TradeOrder{}, errors.New("buyer, seller and product are required")
	}
	if buyerID == sellerID {
		return domain.TradeOrder{}, errors.New("buyer and seller must be different")
	}
	now := time.Now()
	// ID 含随机后缀：同纳秒并发下单会生成相同 UnixNano ID（内存 repo 不去重、PG 主键冲突）
	o := domain.TradeOrder{ID: fmt.Sprintf("torder-%d-%d", now.UnixNano(), rand.Intn(100000)), ProductID: productID, BuyerID: buyerID, SellerID: sellerID, AmountFen: amountFen, Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.repo.Create(ctx, o)
}

// orderFlow 订单合法状态流转（交易管理一期：pending → paid → shipped → completed / cancelled；
// 售后：paid/shipped/completed → aftersale（买家申请，paid=付款后未发货退款）→ completed（审核结案，售后记录留在 aftersale_* 字段））
var orderFlow = map[string][]string{
	"pending":   {"paid", "cancelled"},
	"paid":      {"shipped", "cancelled", "aftersale"},
	"shipped":   {"completed", "cancelled", "aftersale"},
	"completed": {"aftersale"},
	"aftersale": {"completed"},
	"cancelled": {},
}

// checkOrderTransition 校验状态流转是否合法。
func checkOrderTransition(current, next string) error {
	for _, ok := range orderFlow[current] {
		if ok == next {
			return nil
		}
	}
	return fmt.Errorf("非法订单状态流转: %s → %s", current, next)
}

func (s *TradeOrderService) UpdateStatus(ctx context.Context, id, userID, newStatus string) (domain.TradeOrder, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.TradeOrder{}, err
	}
	if o.BuyerID != userID && o.SellerID != userID {
		return domain.TradeOrder{}, fmt.Errorf("permission denied")
	}
	if err := checkOrderTransition(o.Status, newStatus); err != nil {
		return domain.TradeOrder{}, err
	}
	// 角色限定迁移：paid 仅管理端可设（UpdateStatusAdmin）；shipped 仅卖家可设；
	// completed 仅买家可设；cancelled 仅 pending 状态可取消（买卖双方在 pending 均可取消）。
	switch newStatus {
	case "paid":
		return domain.TradeOrder{}, fmt.Errorf("非法订单状态流转: %s → %s（paid 仅管理端可设置）", o.Status, newStatus)
	case "aftersale":
		// 售后必须走 ApplyAftersale（带售后字段写入）；经状态 PATCH 直达会形成
		// aftersale_status 为空的死状态且无法审核，一律拒绝
		return domain.TradeOrder{}, fmt.Errorf("非法订单状态流转: %s → %s（售后请走申请售后接口）", o.Status, newStatus)
	case "shipped":
		if o.SellerID != userID {
			return domain.TradeOrder{}, fmt.Errorf("permission denied: 仅卖家可标记发货")
		}
	case "completed":
		if o.BuyerID != userID {
			return domain.TradeOrder{}, fmt.Errorf("permission denied: 仅买家可确认收货")
		}
	case "cancelled":
		if o.Status != "pending" {
			return domain.TradeOrder{}, fmt.Errorf("非法订单状态流转: %s → %s（仅 pending 状态可取消）", o.Status, newStatus)
		}
	}
	// 原子迁移：WHERE status=当前读到的状态，并发改单时后写方失败（防 completed 被回退等非法覆盖）
	ok, updated, err := s.repo.CompareAndSetStatus(ctx, id, o.Status, newStatus)
	if err != nil {
		return domain.TradeOrder{}, err
	}
	if !ok {
		return domain.TradeOrder{}, fmt.Errorf("订单状态已变更，请刷新后重试")
	}
	// 资金钩子：确认收货放款给卖家 / 取消订单退款给买家（都幂等）。
	// 状态已落定后才动钱——失败时钱仍留在托管里，最坏是延迟到账，不会丢；
	// 因此这里记 Error 日志并返回错误，由人工/补偿处理。
	if err := s.settleMoney(ctx, updated, newStatus); err != nil {
		return domain.TradeOrder{}, err
	}
	// 订单取消：商品恢复为可售（sold → listed），重新出现在供给大厅。
	// 失败不再让接口报错——订单与资金都已落定，此时报错反而会让运营以为"取消失败"
	// 而重复操作（此前正是这个坑：商品被别的流程改过就返回错误，订单其实已取消）。
	if newStatus == "cancelled" && s.prodRepo != nil && o.ProductID != "" {
		if rerr := s.prodRepo.Restore(ctx, o.ProductID); rerr != nil {
			slog.Warn("订单已取消但商品状态未恢复，需人工确认", "order", o.ID, "product", o.ProductID, "error", rerr)
		}
	}
	return updated, nil
}

// settleMoney 按目标状态执行资金动作（completed 放款 / cancelled 退款），两者都幂等。
func (s *TradeOrderService) settleMoney(ctx context.Context, o domain.TradeOrder, newStatus string) error {
	if s.escrow == nil {
		return nil
	}
	switch newStatus {
	case "completed":
		if err := s.settleToSeller(ctx, o); err != nil {
			slog.Error("订单已完成但放款失败（资金仍在托管，需人工处理）", "order", o.ID, "error", err)
			return err
		}
	case "cancelled":
		if err := s.refundBuyerIfAny(ctx, o, o.AmountFen); err != nil {
			slog.Error("订单已取消但退款失败（资金仍在托管，需人工处理）", "order", o.ID, "error", err)
			return err
		}
	}
	return nil
}

// ApplyAftersale 买家申请售后：仅买家可申请；一次订单仅一份有效售后单
// （aftersale_status 非空即已有申请，不得重复）；状态机 paid/shipped/completed → aftersale。
func (s *TradeOrderService) ApplyAftersale(ctx context.Context, userID, orderID, aftType, reason, desc string, amountFen int64) (domain.TradeOrder, error) {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return domain.TradeOrder{}, err
	}
	if o.BuyerID != userID {
		return domain.TradeOrder{}, fmt.Errorf("permission denied")
	}
	if amountFen <= 0 || amountFen > o.AmountFen {
		return domain.TradeOrder{}, fmt.Errorf("售后金额必须在 0~订单金额之间（含 0 不可申请）")
	}
	// 判重口径：只有「待审核 / 已通过」才算已有有效售后。
	// 被驳回的（rejected）允许重新申请——此前用 AftersaleStatus != "" 一票否决，
	// 买家被驳回一次就永久失去售后权利（订单已回到 paid/shipped，却再也提不了）。
	if o.AftersaleStatus == "pending" || o.AftersaleStatus == "approved" {
		return domain.TradeOrder{}, fmt.Errorf("该订单已存在售后申请")
	}
	if err := checkOrderTransition(o.Status, "aftersale"); err != nil {
		return domain.TradeOrder{}, err
	}
	now := time.Now()
	// 记录售后前状态：驳回时恢复原状态（未发货已付款订单曾被迫 completed 卡死）。
	o.AftersaleFrom = o.Status
	o.Status = "aftersale"
	o.AftersaleType = aftType
	o.AftersaleReason = reason
	o.AftersaleDesc = desc
	o.AftersaleAmountFen = amountFen
	o.AftersaleStatus = "pending"
	o.AftersaleTime = now
	return s.repo.UpdateAftersale(ctx, o)
}

// refundForAftersale 售后同意后的资金处置（两种情形）：
//
//	A. 货款还在买家冻结里（订单 paid/shipped，尚未确认收货）：
//	   从冻结里退售后金额给买家，剩余部分放给卖家（部分退款 = 买卖双方达成降价成交）；
//	B. 货款已放给卖家（订单 completed）：钱已不在冻结里，只能从卖家余额扣回买家余额
//	   （Transfer）；卖家余额不足则返回错误，由平台人工介入，绝不静默吞掉。
//
//
//	C. 这笔订单从未冻结过资金（管理端建单/线下成交）：无钱可动，仅记 Warn 后结案。
//
// 幂等/可重试：A 分支的退款与放款各自幂等（refundBuyerIfAny / Release 内部查重），
// 因此「退了款但放款失败」重试时不会重复退款；B 由 AftersaleStatus 的
// pending→approved 单向流转保证只执行一次。
func (s *TradeOrderService) refundForAftersale(ctx context.Context, o domain.TradeOrder) error {
	if s.escrow == nil {
		return nil
	}
	refundAmt := o.AftersaleAmountFen
	if refundAmt <= 0 {
		return fmt.Errorf("售后金额异常: %d", refundAmt)
	}
	frozen, err := s.escrow.HasFrozen(ctx, o.BuyerID, tradeRefType, o.ID)
	if err != nil {
		return fmt.Errorf("查询订单冻结流水失败: %w", err)
	}
	if !frozen {
		// C：这笔订单从未冻结过资金（管理端建单/线下成交）：没有钱可动，
		// 记 Warn 后按状态结案，绝不凭空造一笔退款流水。
		slog.Warn("订单无托管资金，售后按状态结案（不产生退款流水）", "order", o.ID, "amount", refundAmt)
		return nil
	}
	released, err := s.escrow.HasReleased(ctx, o.BuyerID, tradeRefType, o.ID)
	if err != nil {
		return fmt.Errorf("查询订单放款流水失败: %w", err)
	}
	if released {
		// B：钱已放给卖家 → 卖家余额 → 买家余额
		if _, err := s.escrow.Transfer(ctx, o.SellerID, o.BuyerID, refundAmt, tradeRefType, o.ID); err != nil {
			if errors.Is(err, repository.ErrInsufficientBalance) {
				return fmt.Errorf("售后退款失败：卖家托管金余额不足，请联系平台处理（订单 %s）", o.ID)
			}
			return fmt.Errorf("售后退款失败: %w", err)
		}
		return nil
	}
	// A：钱还在冻结里 → 先退买家，再把剩余放给卖家
	if err := s.refundBuyerIfAny(ctx, o, refundAmt); err != nil {
		return err
	}
	if rest := o.AmountFen - refundAmt; rest > 0 {
		if _, err := s.escrow.Release(ctx, o.BuyerID, o.SellerID, rest, tradeRefType, o.ID); err != nil {
			return fmt.Errorf("售后部分退款后放款给卖家失败: %w", err)
		}
	}
	return nil
}

// ReviewAftersale 管理端审核售后单：同意 → aftersale_status=approved（退款完成）；
// 驳回 → aftersale_status=rejected。结案后订单状态回到 completed（交易结束态），
// 售后记录保留在 aftersale_* 字段供买家/后台查看。
func (s *TradeOrderService) ReviewAftersale(ctx context.Context, orderID string, approve bool) (domain.TradeOrder, error) {
	return s.reviewAftersale(ctx, orderID, approve)
}

// ReviewAftersaleAsSeller 卖家/管理员审核自己订单的售后单：
// 仅订单卖家（或平台/协会管理员）可审，其余逻辑与管理端审核一致。
func (s *TradeOrderService) ReviewAftersaleAsSeller(ctx context.Context, a domain.Actor, orderID string, approve bool) (domain.TradeOrder, error) {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return domain.TradeOrder{}, err
	}
	isAdmin := a.Role == domain.RolePlatformAdmin || a.Role == domain.RoleAssociationAdmin
	if !isAdmin && o.SellerID != a.ID {
		return domain.TradeOrder{}, fmt.Errorf("permission denied: 仅订单卖家或管理员可审核售后")
	}
	return s.reviewAftersale(ctx, orderID, approve)
}

func (s *TradeOrderService) reviewAftersale(ctx context.Context, orderID string, approve bool) (domain.TradeOrder, error) {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return domain.TradeOrder{}, err
	}
	if o.Status != "aftersale" || o.AftersaleStatus != "pending" {
		return domain.TradeOrder{}, fmt.Errorf("该订单不在售后待审核状态")
	}
	if approve {
		// 先退款再改状态：钱动不了（卖家余额不足等）就拒绝本次审批，状态保持待审核可重试，
		// 绝不允许出现"售后已批准、钱却没退"的账实不符。
		if err := s.refundForAftersale(ctx, o); err != nil {
			slog.Error("售后审批通过但退款失败", "order", o.ID, "error", err)
			return domain.TradeOrder{}, err
		}
		o.AftersaleStatus = "approved"
		o.Status = "completed"
	} else {
		o.AftersaleStatus = "rejected"
		// 驳回还原：恢复售后前状态（paid/shipped/completed）。
		// 此前无条件置 completed——已付款未发货订单被驳回后卖家无法发货、买家付钱拿不到货。
		if o.AftersaleFrom == "paid" || o.AftersaleFrom == "shipped" || o.AftersaleFrom == "completed" {
			o.Status = o.AftersaleFrom
		} else {
			o.Status = "completed"
		}
	}
	return s.repo.UpdateAftersale(ctx, o)
}

// PayOrder 买家模拟支付：仅订单买家可调，仅 pending → paid 迁移
// （真实微信支付接入后由服务端支付回调替代此接口，语义保持：买家确认付款）。
func (s *TradeOrderService) PayOrder(ctx context.Context, buyerID, orderID string) (domain.TradeOrder, error) {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return domain.TradeOrder{}, err
	}
	if o.BuyerID != buyerID {
		return domain.TradeOrder{}, fmt.Errorf("permission denied")
	}
	if err := checkOrderTransition(o.Status, "paid"); err != nil {
		return domain.TradeOrder{}, err
	}
	// 先冻结再改状态：钱不到位就不改状态。余额不足返回 ErrInsufficientBalance
	// （Handler 映射 402，前端引导去「我的托管金」充值）。
	if err := s.freezeForOrder(ctx, o); err != nil {
		return domain.TradeOrder{}, err
	}
	ok, updated, err := s.repo.CompareAndSetStatus(ctx, orderID, o.Status, "paid")
	if err != nil || !ok {
		// 状态没改成（并发改单/DB 错误）：把刚冻结的钱退回买家，避免资金滞留托管
		if rerr := s.refundBuyerIfAny(ctx, o, o.AmountFen); rerr != nil {
			slog.Error("订单支付状态迁移失败且退款失败（资金滞留，需人工处理）", "order", o.ID, "error", rerr)
		}
		if err != nil {
			return domain.TradeOrder{}, err
		}
		return domain.TradeOrder{}, fmt.Errorf("订单状态已变更，请刷新后重试")
	}
	return updated, nil
}

// UpdateStatusAdmin 管理端改单：跳过买卖双方校验，仍受状态机约束。
// 取消订单（pending/paid → cancelled）同步恢复商品为可售（sold → listed）。
func (s *TradeOrderService) UpdateStatusAdmin(ctx context.Context, id, newStatus string) (domain.TradeOrder, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.TradeOrder{}, err
	}
	// 售后必须走 ApplyAftersale（带售后字段写入）；管理端直置 aftersale 会形成
	// 无 aftersale_* 字段的死状态且无法审核（与 UpdateStatus 同款拦截）。
	if newStatus == "aftersale" {
		return domain.TradeOrder{}, fmt.Errorf("非法订单状态流转（售后请走申请售后接口）")
	}
	if err := checkOrderTransition(o.Status, newStatus); err != nil {
		return domain.TradeOrder{}, err
	}
	ok, updated, err := s.repo.CompareAndSetStatus(ctx, id, o.Status, newStatus)
	if err != nil {
		return domain.TradeOrder{}, err
	}
	if !ok {
		return domain.TradeOrder{}, fmt.Errorf("订单状态已变更，请刷新后重试")
	}
	// 管理端改单同样走资金钩子（取消 → 退款给买家；完成 → 放款给卖家）
	if err := s.settleMoney(ctx, updated, newStatus); err != nil {
		return domain.TradeOrder{}, err
	}
	if newStatus == "cancelled" && s.prodRepo != nil && o.ProductID != "" {
		if rerr := s.prodRepo.Restore(ctx, o.ProductID); rerr != nil {
			slog.Warn("管理端取消订单后商品状态未恢复，需人工确认", "order", o.ID, "product", o.ProductID, "error", rerr)
		}
	}
	return updated, nil
}

// Delete 管理端删除订单（真删除）：被删除的未完成订单对应商品恢复可售。
func (s *TradeOrderService) Delete(ctx context.Context, id string) error {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return s.repo.Delete(ctx, id)
	}
	// 未完成订单（pending/paid）：先把买家的钱退回去，再删单；
	// 退款失败则不删（钱不能因为删单而丢失）；商品恢复失败不阻塞删除（可人工确认）。
	if o.Status == "pending" || o.Status == "paid" {
		if err := s.refundBuyerIfAny(ctx, o, o.AmountFen); err != nil {
			return fmt.Errorf("删除订单前退款失败: %w", err)
		}
		if o.ProductID != "" && s.prodRepo != nil {
			if err := s.prodRepo.Restore(ctx, o.ProductID); err != nil {
				slog.Warn("删除订单后商品状态未恢复，需人工确认", "order", o.ID, "product", o.ProductID, "error", err)
			}
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *TradeOrderService) ListMine(ctx context.Context, userID string) ([]domain.TradeOrder, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *TradeOrderService) ListAll(ctx context.Context, offset, limit int) ([]domain.TradeOrder, int, error) {
	return s.repo.ListAll(ctx, offset, limit)
}

// ListAllFiltered 管理端订单列表：过滤 + 分页下沉 SQL（不再全量拉取后内存过滤）。
func (s *TradeOrderService) ListAllFiltered(ctx context.Context, f repository.TradeOrderFilter) ([]domain.TradeOrder, int, error) {
	return s.repo.ListFiltered(ctx, f)
}

func (s *TradeOrderService) FindByID(ctx context.Context, id string) (domain.TradeOrder, error) {
	return s.repo.FindByID(ctx, id)
}
