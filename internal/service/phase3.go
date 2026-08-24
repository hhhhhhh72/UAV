package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
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
	NoCrime     string `json:"noCrime"`

	// PaidAmountFen 报名时冻结的学费（分）。仅 payAndEnroll 在冻结成功后填充；
	// 免费/普通报名为 0。completeEnrollment 按此金额释放，与课程实时价格解耦。
	PaidAmountFen int64 `json:"-"`
}

// All 管理端全量报名记录（分页）。
func (s *EnrollmentService) All(ctx context.Context, offset, limit int) ([]domain.Enrollment, int, error) {
	return s.repo.ListAll(ctx, offset, limit)
}

// FindByID 按报名 ID 定位单条记录（管理端完成报名/编辑用，替代全表扫描）。
func (s *EnrollmentService) FindByID(ctx context.Context, id string) (domain.Enrollment, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *EnrollmentService) Enroll(ctx context.Context, userID, courseID string, form EnrollmentForm) (domain.Enrollment, error) {
	// 并发防重复：check-then-insert 加进程内锁（双请求同时通过查重会重复报名，
	// 且付费报名会重复扣冻结金额）。
	unlock := lockByKey("enroll|" + userID + "|" + courseID)
	defer unlock()
	if _, ok, _ := s.repo.FindByUserAndCourse(ctx, userID, courseID); ok {
		return domain.Enrollment{}, fmt.Errorf("already enrolled")
	}
	// 容量：course.MaxStudents > 0 时，该课已报名数 >= MaxStudents 拒绝。
	// 课程不存在/查询失败时跳过容量检查（兼容无课程仓储的测试与历史数据）。
	if s.courseRepo != nil {
		if c, err := s.courseRepo.FindByID(ctx, courseID); err == nil && c.MaxStudents > 0 {
			if enrolls, err := s.repo.ListByCourse(ctx, courseID); err == nil && len(enrolls) >= c.MaxStudents {
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
	e := domain.Enrollment{ID: nextID("enroll"), CourseID: courseID, UserID: userID,
		Name: form.Name, Phone: form.Phone, IDCard: form.IDCard, Gender: form.Gender, Birthday: birthday,
		Email: form.Email, Education: form.Education, Experience: form.Experience,
		PhotoURL: form.Photo, IDCardImage: form.IDCardImage, NoCrime: form.NoCrime,
		Status: "enrolled", PaidAmountFen: form.PaidAmountFen, CreatedAt: now}
	return s.repo.Create(ctx, e)
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

type TradeOrderService struct {
	repo     repository.TradeOrderRepository
	prodRepo repository.ProductRepository // 订单取消时恢复商品为可售（可空）
}

func NewTradeOrderService(repo repository.TradeOrderRepository, prodRepo repository.ProductRepository) *TradeOrderService {
	return &TradeOrderService{repo: repo, prodRepo: prodRepo}
}

func (s *TradeOrderService) Create(ctx context.Context, buyerID, productID, sellerID string, amountFen int64) (domain.TradeOrder, error) {
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
	// 订单取消：商品恢复为可售（sold → listed），重新出现在供给大厅
	if newStatus == "cancelled" && s.prodRepo != nil && o.ProductID != "" {
		if rerr := s.prodRepo.Restore(ctx, o.ProductID); rerr != nil {
			return domain.TradeOrder{}, fmt.Errorf("订单已取消但商品恢复失败: %w", rerr)
		}
	}
	return updated, nil
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
	if o.AftersaleStatus != "" {
		return domain.TradeOrder{}, fmt.Errorf("该订单已存在售后申请")
	}
	if err := checkOrderTransition(o.Status, "aftersale"); err != nil {
		return domain.TradeOrder{}, err
	}
	now := time.Now()
	o.Status = "aftersale"
	o.AftersaleType = aftType
	o.AftersaleReason = reason
	o.AftersaleDesc = desc
	o.AftersaleAmountFen = amountFen
	o.AftersaleStatus = "pending"
	o.AftersaleTime = now
	return s.repo.UpdateAftersale(ctx, o)
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
		o.AftersaleStatus = "approved"
	} else {
		o.AftersaleStatus = "rejected"
	}
	o.Status = "completed"
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
	ok, updated, err := s.repo.CompareAndSetStatus(ctx, orderID, o.Status, "paid")
	if err != nil {
		return domain.TradeOrder{}, err
	}
	if !ok {
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
	if newStatus == "cancelled" && s.prodRepo != nil && o.ProductID != "" {
		if rerr := s.prodRepo.Restore(ctx, o.ProductID); rerr != nil {
			return domain.TradeOrder{}, fmt.Errorf("订单已取消但商品恢复失败: %w", rerr)
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
	// 未完成订单（pending/paid）删除后商品恢复；已完成/售后单不恢复（交易已终结）
	if o.ProductID != "" && (o.Status == "pending" || o.Status == "paid") && s.prodRepo != nil {
		_ = s.prodRepo.Restore(ctx, o.ProductID)
	}
	return s.repo.Delete(ctx, id)
}

func (s *TradeOrderService) ListMine(ctx context.Context, userID string) ([]domain.TradeOrder, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *TradeOrderService) ListAll(ctx context.Context, offset, limit int) ([]domain.TradeOrder, int, error) {
	return s.repo.ListAll(ctx, offset, limit)
}

func (s *TradeOrderService) FindByID(ctx context.Context, id string) (domain.TradeOrder, error) {
	return s.repo.FindByID(ctx, id)
}
