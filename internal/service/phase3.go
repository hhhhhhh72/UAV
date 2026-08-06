package service

import (
	"errors"
	"fmt"
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
func (s *EnrollmentService) All(offset, limit int) ([]domain.Enrollment, int, error) {
	return s.repo.ListAll(offset, limit)
}

func (s *EnrollmentService) Enroll(userID, courseID string, form EnrollmentForm) (domain.Enrollment, error) {
	if _, ok, _ := s.repo.FindByUserAndCourse(userID, courseID); ok {
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
	e := domain.Enrollment{ID: fmt.Sprintf("enroll-%d", now.UnixNano()), CourseID: courseID, UserID: userID,
		Name: form.Name, Phone: form.Phone, IDCard: form.IDCard, Gender: form.Gender, Birthday: birthday,
		Email: form.Email, Education: form.Education, Experience: form.Experience,
		PhotoURL: form.Photo, IDCardImage: form.IDCardImage, NoCrime: form.NoCrime,
		Status: "enrolled", CreatedAt: now}
	return s.repo.Create(e)
}

// validEnrollmentStatus 报名状态白名单（与前端 statusLabel 对齐）。
func validEnrollmentStatus(status string) bool {
	switch status {
	case "pending", "approved", "paid", "enrolled", "rejected":
		return true
	}
	return false
}

// Update 管理端编辑报名记录（基础信息 + 状态；全字段覆盖）。
// 状态校验：白名单 + 防回退（已缴费/已入学为定局状态，不允许改回待审核/驳回）。
func (s *EnrollmentService) Update(a domain.Actor, e domain.Enrollment) (domain.Enrollment, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Enrollment{}, errors.New("admin permission required")
	}
	if e.ID == "" {
		return domain.Enrollment{}, errors.New("enrollment id is required")
	}
	old, err := s.repo.FindByID(e.ID)
	if err != nil {
		return domain.Enrollment{}, err
	}
	if !validEnrollmentStatus(e.Status) {
		return domain.Enrollment{}, fmt.Errorf("invalid enrollment status %q", e.Status)
	}
	if (old.Status == "paid" || old.Status == "enrolled") && (e.Status == "pending" || e.Status == "rejected") {
		return domain.Enrollment{}, fmt.Errorf("cannot change enrollment status from %q to %q", old.Status, e.Status)
	}
	return s.repo.Update(e)
}

func (s *EnrollmentService) ListByCourse(courseID string) ([]domain.Enrollment, error) {
	return s.repo.ListByCourse(courseID)
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

func (s *TradeOrderService) Create(buyerID, productID, sellerID string, amountFen int64) (domain.TradeOrder, error) {
	now := time.Now()
	o := domain.TradeOrder{ID: fmt.Sprintf("torder-%d", now.UnixNano()), ProductID: productID, BuyerID: buyerID, SellerID: sellerID, AmountFen: amountFen, Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.repo.Create(o)
}

// orderFlow 订单合法状态流转（交易管理一期：pending → paid → shipped → completed / cancelled）
var orderFlow = map[string][]string{
	"pending":   {"paid", "cancelled"},
	"paid":      {"shipped", "cancelled"},
	"shipped":   {"completed", "cancelled"},
	"completed": {},
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

func (s *TradeOrderService) UpdateStatus(id, userID, newStatus string) (domain.TradeOrder, error) {
	o, err := s.repo.FindByID(id)
	if err != nil {
		return domain.TradeOrder{}, err
	}
	if o.BuyerID != userID && o.SellerID != userID {
		return domain.TradeOrder{}, fmt.Errorf("permission denied")
	}
	if err := checkOrderTransition(o.Status, newStatus); err != nil {
		return domain.TradeOrder{}, err
	}
	return s.repo.UpdateStatus(id, newStatus)
}

// UpdateStatusAdmin 管理端改单：跳过买卖双方校验，仍受状态机约束。
func (s *TradeOrderService) UpdateStatusAdmin(id, newStatus string) (domain.TradeOrder, error) {
	o, err := s.repo.FindByID(id)
	if err != nil {
		return domain.TradeOrder{}, err
	}
	if err := checkOrderTransition(o.Status, newStatus); err != nil {
		return domain.TradeOrder{}, err
	}
	return s.repo.UpdateStatus(id, newStatus)
}

// Delete 管理端删除订单（真删除，替代原假删除 stub）。
func (s *TradeOrderService) Delete(id string) error { return s.repo.Delete(id) }

func (s *TradeOrderService) ListMine(userID string) ([]domain.TradeOrder, error) {
	return s.repo.ListByUser(userID)
}

func (s *TradeOrderService) ListAll(offset, limit int) ([]domain.TradeOrder, int, error) {
	return s.repo.ListAll(offset, limit)
}

func (s *TradeOrderService) FindByID(id string) (domain.TradeOrder, error) {
	return s.repo.FindByID(id)
}
