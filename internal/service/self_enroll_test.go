package service_test

import (
	"context"
	"strings"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 防自购自卖：课程发布者（org_id=本人）报名自己的课程必须被拒绝（返回明确中文错误）；
// 其他用户报名同一课程不受影响。
func TestEnrollSelfOwnedCourseRejected(t *testing.T) {
	courseRepo := memory.NewCourseRepository()
	if _, err := courseRepo.Create(context.Background(), domain.TrainingCourse{
		ID: "c-self", Title: "自营课", PriceFen: 0, OrgID: "u-1",
	}); err != nil {
		t.Fatalf("seed course: %v", err)
	}
	svc := service.NewEnrollmentService(memory.NewEnrollmentRepository(), courseRepo)

	// 发布者本人报名 → 拒绝（错误信息指明原因）
	_, err := svc.Enroll(context.Background(), "u-1", "c-self", service.EnrollmentForm{Name: "n"})
	if err == nil || !strings.Contains(err.Error(), "不能报名自己发布的课程") {
		t.Fatalf("owner enroll should be rejected with clear message, got: %v", err)
	}
	// 其他用户报名同一课程 → 正常放行
	if _, err := svc.Enroll(context.Background(), "u-2", "c-self", service.EnrollmentForm{Name: "n"}); err != nil {
		t.Fatalf("other user enroll should succeed: %v", err)
	}
}
