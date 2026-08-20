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
	repo repository.EnrollmentRepository
}

func NewEnrollmentService(repo repository.EnrollmentRepository) *EnrollmentService {
	return &EnrollmentService{repo: repo}
}

// EnrollmentForm 培训报名表单数据（小程序 register.vue 12 字段）。
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
	// 生日：前端提交 "YYYY-MM-DD"，解析为 DATE 语义
	var birthday time.Time
	if form.Birthday != "" {
		if bd, err := time.Parse("2006-01-02", form.Birthday); err == nil {
			birthday = bd
		}
	}
	now := time.Now()
	e := domain.Enrollment{ID: nextID("enroll"), CourseID: courseID, UserID: userID,
		Name: form.Name, Phone: form.Phone, IDCard: form.IDCard, Gender: form.Gender, Birthday: birthday,
		Email: form.Email, Education: form.Education, Experience: form.Experience,
		PhotoURL: form.Photo, IDCardImage: form.IDCardImage, NoCrime: form.NoCrime,
		Status: "enrolled", CreatedAt: now}
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
	return s.repo.Update(ctx, e)
}

func (s *EnrollmentService) ListByCourse(ctx context.Context, courseID string) ([]domain.Enrollment, error) {
	return s.repo.ListByCourse(ctx, courseID)
}

// ---- Expiry Checker ----

type ExpiryService struct{}

func NewExpiryService() *ExpiryService { return &ExpiryService{} }

func (s *ExpiryService) GetExpiringCerts(certs []domain.Certificate, withinDays int) []domain.Certificate {
	cutoff := time.Now().AddDate(0, 0, withinDays)
	out := []domain.Certificate{}
	for _, c := range certs {
		if c.ExpireDate.Before(cutoff) && c.Status == "approved" {
			out = append(out, c)
		}
	}
	return out
}

func (s *ExpiryService) GetExpiringInspections(list []domain.AnnualInspection, withinDays int) []domain.AnnualInspection {
	cutoff := time.Now().AddDate(0, 0, withinDays)
	out := []domain.AnnualInspection{}
	for _, i := range list {
		if i.ExpireDate.Before(cutoff) && i.Status == "approved" {
			out = append(out, i)
		}
	}
	return out
}

// ---- Trade Orders ----

type TradeOrderService struct {
	repo repository.TradeOrderRepository
}

func NewTradeOrderService(repo repository.TradeOrderRepository) *TradeOrderService {
	return &TradeOrderService{repo: repo}
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
	return s.repo.UpdateStatus(ctx, id, newStatus)
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
	if amountFen < 0 || amountFen > o.AmountFen {
		return domain.TradeOrder{}, fmt.Errorf("售后金额必须在 0~订单金额之间")
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

// UpdateStatusAdmin 管理端改单：跳过买卖双方校验，仍受状态机约束。
func (s *TradeOrderService) UpdateStatusAdmin(ctx context.Context, id, newStatus string) (domain.TradeOrder, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.TradeOrder{}, err
	}
	if err := checkOrderTransition(o.Status, newStatus); err != nil {
		return domain.TradeOrder{}, err
	}
	return s.repo.UpdateStatus(ctx, id, newStatus)
}

// Delete 管理端删除订单（真删除，替代原假删除 stub）。
func (s *TradeOrderService) Delete(ctx context.Context, id string) error {
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
