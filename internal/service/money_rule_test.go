package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 资金规则回归：付费课程禁止免费报名（/enroll 此前不校验 price_fen，
// 付费课可 0 元直接报名成功，绕过 pay-and-enroll 的学费冻结）。
func TestPaidCourseFreeEnrollRejected(t *testing.T) {
	courseRepo := memory.NewCourseRepository()
	if _, err := courseRepo.Create(context.Background(), domain.TrainingCourse{
		ID: "c-paid", Title: "付费课", PriceFen: 50000,
	}); err != nil {
		t.Fatalf("seed course: %v", err)
	}
	if _, err := courseRepo.Create(context.Background(), domain.TrainingCourse{
		ID: "c-free", Title: "免费课", PriceFen: 0,
	}); err != nil {
		t.Fatalf("seed free course: %v", err)
	}
	svc := service.NewEnrollmentService(memory.NewEnrollmentRepository(), courseRepo)

	// 付费课程不带金额 → 拒绝
	if _, err := svc.Enroll(context.Background(), "u-1", "c-paid", service.EnrollmentForm{Name: "n"}); err == nil {
		t.Fatal("paid course free enroll should be rejected")
	}
	// 付费课程带错误金额 → 拒绝
	if _, err := svc.Enroll(context.Background(), "u-1", "c-paid", service.EnrollmentForm{Name: "n", PaidAmountFen: 1}); err == nil {
		t.Fatal("paid course with wrong amount should be rejected")
	}
	// 付费课程带正确金额 → 成功
	if _, err := svc.Enroll(context.Background(), "u-2", "c-paid", service.EnrollmentForm{Name: "n", PaidAmountFen: 50000}); err != nil {
		t.Fatalf("paid enroll with correct amount should succeed: %v", err)
	}
	// 免费课程不带金额 → 成功（不受影响）
	if _, err := svc.Enroll(context.Background(), "u-3", "c-free", service.EnrollmentForm{Name: "n"}); err != nil {
		t.Fatalf("free course enroll should succeed: %v", err)
	}
}
