package service

import (
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

func (s *EnrollmentService) Enroll(userID, courseID string, form EnrollmentForm) (domain.Enrollment, error) {
	if _, ok, _ := s.repo.FindByUserAndCourse(userID, courseID); ok {
		return domain.Enrollment{}, fmt.Errorf("already enrolled")
	}
	now := time.Now()
	e := domain.Enrollment{ID: fmt.Sprintf("enroll-%d", now.UnixNano()), CourseID: courseID, UserID: userID,
		Name: form.Name, Phone: form.Phone, IDCard: form.IDCard, Gender: form.Gender, Birthday: form.Birthday,
		Email: form.Email, Education: form.Education, Experience: form.Experience,
		PhotoURL: form.Photo, IDCardImage: form.IDCardImage, NoCrime: form.NoCrime,
		Status: "enrolled", CreatedAt: now}
	return s.repo.Create(e)
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

func (s *TradeOrderService) UpdateStatus(id, userID, newStatus string) (domain.TradeOrder, error) {
	o, err := s.repo.FindByID(id)
	if err != nil {
		return domain.TradeOrder{}, err
	}
	if o.BuyerID != userID && o.SellerID != userID {
		return domain.TradeOrder{}, fmt.Errorf("permission denied")
	}
	return s.repo.UpdateStatus(id, newStatus)
}

func (s *TradeOrderService) ListMine(userID string) ([]domain.TradeOrder, error) {
	return s.repo.ListByUser(userID)
}

func (s *TradeOrderService) ListAll(offset, limit int) ([]domain.TradeOrder, int, error) {
	return s.repo.ListAll(offset, limit)
}

func (s *TradeOrderService) FindByID(id string) (domain.TradeOrder, error) {
	return s.repo.FindByID(id)
}
