package service_test

import (
	"context"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// WorkOrderService.ListAll：管理端全量工单查询（仪表盘业务趋势用）。
// 权限：仅平台/协会管理员；分页与排序：创建时间倒序。
func TestWorkOrderListAllPermissionsAndPaging(t *testing.T) {
	orderRepo := memory.NewWorkOrderRepository()
	svc := service.NewWorkOrderService(orderRepo, memory.NewDemandRepository(nil), memory.NewIntentRepository())
	ctx := context.Background()

	// 非管理员一律拒绝
	for _, role := range []domain.Role{domain.RoleIndividual, domain.RoleEnterprise} {
		if _, _, err := svc.ListAll(ctx, domain.Actor{ID: "u1", Role: role}, 0, 10); err == nil {
			t.Fatalf("role %s 应被拒绝", role)
		}
	}

	// 空库：管理员可查，返回 0 条
	for _, role := range []domain.Role{domain.RolePlatformAdmin, domain.RoleAssociationAdmin} {
		items, total, err := svc.ListAll(ctx, domain.Actor{ID: "admin-1", Role: role}, 0, 10)
		if err != nil {
			t.Fatalf("role %s: %v", role, err)
		}
		if total != 0 || len(items) != 0 {
			t.Fatalf("role %s: 空库应为 0 条, got total=%d len=%d", role, total, len(items))
		}
	}

	// 播种 3 单（时间递增）：ListAll 应按创建时间倒序返回
	base := time.Now().Add(-3 * time.Hour)
	for i := 0; i < 3; i++ {
		if _, err := orderRepo.Create(ctx, domain.WorkOrder{
			ID:        "wo-" + string(rune('a'+i)),
			OrderNo:   "NO-" + string(rune('a'+i)),
			Status:    domain.WorkOrderPending,
			CreatedAt: base.Add(time.Duration(i) * time.Hour),
			UpdatedAt: base.Add(time.Duration(i) * time.Hour),
		}); err != nil {
			t.Fatalf("seed work order %d: %v", i, err)
		}
	}

	admin := domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}
	items, total, err := svc.ListAll(ctx, admin, 0, 2)
	if err != nil {
		t.Fatalf("list page 1: %v", err)
	}
	if total != 3 || len(items) != 2 {
		t.Fatalf("page 1: total=%d len=%d, want 3/2", total, len(items))
	}
	if items[0].ID != "wo-c" {
		t.Fatalf("page 1 首条 = %s, want wo-c（创建时间倒序）", items[0].ID)
	}
	page2, total2, err := svc.ListAll(ctx, admin, 2, 2)
	if err != nil {
		t.Fatalf("list page 2: %v", err)
	}
	if total2 != 3 || len(page2) != 1 || page2[0].ID != "wo-a" {
		t.Fatalf("page 2: total=%d len=%d first=%v, want 3/1/wo-a", total2, len(page2), page2[0].ID)
	}
	// 越界偏移：返回空页但 total 仍正确
	empty, total3, err := svc.ListAll(ctx, admin, 99, 10)
	if err != nil {
		t.Fatalf("list offset 99: %v", err)
	}
	if total3 != 3 || len(empty) != 0 {
		t.Fatalf("offset 99: total=%d len=%d, want 3/0", total3, len(empty))
	}
}
