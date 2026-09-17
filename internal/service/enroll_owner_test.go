package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 无主课程（org_id 为空）不允许报名。
//
// 回归背景：course.OrgID 决定学费结算方向（completeEnrollment 用
// Release(学员, course.OrgID, 学费)）。为空时 Release 直接 fail-closed 报错，
// 机构侧「新的报名待审核」也无人可发——学员交了钱却永远毕不了业，学费永久
// 冻结在托管金里（孤儿补偿要求业务行不存在，而报名行是存在的）。
// 管理端建课路径历史上不写 org_id，生产上确实存在这种 published 带价课程。
func TestEnrollRejectsOwnerlessCourse(t *testing.T) {
	ctx := context.Background()
	courseRepo := memory.NewCourseRepository()
	now := time.Now()
	if _, err := courseRepo.Create(ctx, domain.TrainingCourse{
		ID: "course-orphan", Title: "无主课程", Status: "published",
		MaxStudents: 10, PriceFen: 680000, OrgID: "", OrgName: "",
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("seed ownerless course: %v", err)
	}
	svc := service.NewEnrollmentService(memory.NewEnrollmentRepository(), courseRepo)

	// 免费报名路径
	_, err := svc.Enroll(ctx, "student-x", "course-orphan", service.EnrollmentForm{
		Name: "学员甲", Phone: "13800000001",
	})
	if err == nil {
		t.Fatal("无主课程不该允许报名（免费路径）")
	}
	if !strings.Contains(err.Error(), "开课机构") {
		t.Fatalf("错误信息应说明原因，实际：%v", err)
	}

	// 付费报名路径：携带正确学费也不放行（payAndEnroll 会据此回滚已冻结的学费）
	_, err = svc.Enroll(ctx, "student-y", "course-orphan", service.EnrollmentForm{
		Name: "学员乙", Phone: "13800000002", PaidAmountFen: 680000,
	})
	if err == nil {
		t.Fatal("无主课程不该允许报名（付费路径）")
	}

	// 有主课程仍然正常放行——护栏不能误伤
	if _, err := courseRepo.Create(ctx, domain.TrainingCourse{
		ID: "course-owned", Title: "有主课程", Status: "published",
		MaxStudents: 10, OrgID: "org-1", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("seed owned course: %v", err)
	}
	if _, err := svc.Enroll(ctx, "student-x", "course-owned", service.EnrollmentForm{Name: "学员甲"}); err != nil {
		t.Fatalf("有主课程应可报名，实际：%v", err)
	}
}
