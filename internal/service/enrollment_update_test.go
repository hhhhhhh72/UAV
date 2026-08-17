package service_test

import (
	"context"
	"strings"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 管理端编辑报名记录：正常更新 / 权限 / 状态白名单 / 防回退 / 404
func TestEnrollmentUpdate(t *testing.T) {
	repo := memory.NewEnrollmentRepository()
	svc := service.NewEnrollmentService(repo)

	created, err := svc.Enroll(context.Background(), "u-1", "crs-1", service.EnrollmentForm{Name: "张三", Phone: "13800000000", Birthday: "2000-01-02"})
	if err != nil {
		t.Fatalf("enroll: %v", err)
	}
	if created.Status != "enrolled" {
		t.Fatalf("initial status should be enrolled, got %q", created.Status)
	}

	t.Run("admin can update basic info and status", func(t *testing.T) {
		upd, err := svc.Update(context.Background(), admActor(), domain.Enrollment{
			ID: created.ID, Name: "李四", Phone: "13900000000", IDCard: "110101199001011234",
			Gender: "女", Email: "a@b.com", Education: "本科", Experience: "3年",
			CourseID: "crs-2", Status: "paid",
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if upd.Name != "李四" || upd.Phone != "13900000000" || upd.Status != "paid" {
			t.Fatalf("updated fields mismatch: %+v", upd)
		}
		got, _, err := svc.All(context.Background(), 0, 10)
		if err != nil || len(got) != 1 || got[0].Name != "李四" {
			t.Fatalf("repo not updated: %+v err=%v", got, err)
		}
	})

	t.Run("non-admin is rejected", func(t *testing.T) {
		_, err := svc.Update(context.Background(), domain.Actor{ID: "u-1", Role: domain.RoleIndividual}, domain.Enrollment{ID: created.ID, Status: "pending"})
		if err == nil || !strings.Contains(err.Error(), "permission") {
			t.Fatalf("want permission error, got %v", err)
		}
	})

	t.Run("invalid status value is rejected", func(t *testing.T) {
		_, err := svc.Update(context.Background(), admActor(), domain.Enrollment{ID: created.ID, Status: "bogus"})
		if err == nil || !strings.Contains(err.Error(), "invalid enrollment status") {
			t.Fatalf("want invalid status error, got %v", err)
		}
	})

	t.Run("cannot roll back paid/enrolled to pending or rejected", func(t *testing.T) {
		// 当前已 paid（上一步更新）→ 改回 pending 应拒绝
		_, err := svc.Update(context.Background(), admActor(), domain.Enrollment{ID: created.ID, Status: "pending"})
		if err == nil || !strings.Contains(err.Error(), "cannot change") {
			t.Fatalf("want rollback error, got %v", err)
		}
		_, err = svc.Update(context.Background(), admActor(), domain.Enrollment{ID: created.ID, Status: "rejected"})
		if err == nil || !strings.Contains(err.Error(), "cannot change") {
			t.Fatalf("want rollback error for rejected, got %v", err)
		}
		// paid → enrolled 为前向流转，允许
		if _, err := svc.Update(context.Background(), admActor(), domain.Enrollment{ID: created.ID, Status: "enrolled"}); err != nil {
			t.Fatalf("paid -> enrolled should be allowed: %v", err)
		}
	})

	t.Run("unknown id returns not found", func(t *testing.T) {
		_, err := svc.Update(context.Background(), admActor(), domain.Enrollment{ID: "enroll-missing", Status: "pending"})
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Fatalf("want not found error, got %v", err)
		}
	})
}
