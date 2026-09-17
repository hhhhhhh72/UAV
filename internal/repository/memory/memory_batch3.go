package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type rescueCaseRepo struct {
	mu    sync.RWMutex
	items []domain.RescueCase
}

func NewRescueCaseRepository() repository.RescueCaseRepository { return &rescueCaseRepo{} }
func (r *rescueCaseRepo) Create(ctx context.Context, rc domain.RescueCase) (domain.RescueCase, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, rc)
	return rc, nil
}
func (r *rescueCaseRepo) FindByID(ctx context.Context, id string) (domain.RescueCase, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rc := range r.items {
		if rc.ID == id {
			return rc, nil
		}
	}
	return domain.RescueCase{}, fmt.Errorf("not found")
}
func (r *rescueCaseRepo) List(ctx context.Context, eventType, q string, offset, limit int) ([]domain.RescueCase, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	query := strings.ToLower(strings.TrimSpace(q))
	filtered := make([]domain.RescueCase, 0)
	for _, rc := range r.items {
		// 与 PG 对齐：仅已发布案例公开展示
		if rc.Status != "published" {
			continue
		}
		if eventType != "" && rc.EventType != eventType {
			continue
		}
		if query != "" && !matchAnyFold(query, rc.Title, rc.Location, rc.Summary, rc.TeamName, rc.DroneModel) {
			continue
		}
		filtered = append(filtered, rc)
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	return paginateSlice(filtered, offset, limit)
}

// matchAnyFold reports whether any field contains query (case-insensitive).
func matchAnyFold(query string, fields ...string) bool {
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), query) {
			return true
		}
	}
	return false
}

type emergDeptRepo struct {
	mu     sync.RWMutex
	depts  []domain.EmergencyDept
	drills []domain.EmergencyDrill
}

func NewEmergencyDeptRepository() repository.EmergencyDeptRepository { return &emergDeptRepo{} }
func (r *emergDeptRepo) CreateDept(ctx context.Context, d domain.EmergencyDept) (domain.EmergencyDept, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.depts = append(r.depts, d)
	return d, nil
}
func (r *emergDeptRepo) ListDepts(ctx context.Context) ([]domain.EmergencyDept, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := append([]domain.EmergencyDept(nil), r.depts...)
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (r *emergDeptRepo) CreateDrill(ctx context.Context, d domain.EmergencyDrill) (domain.EmergencyDrill, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drills = append(r.drills, d)
	return d, nil
}
func (r *emergDeptRepo) ListDrills(ctx context.Context, deptID string) ([]domain.EmergencyDrill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.EmergencyDrill, 0)
	for _, d := range r.drills {
		if deptID == "" || d.DeptID == deptID {
			out = append(out, d)
		}
	}
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// assocMemberRepo 已移除：协会 8 级角色从未在生产使用（0 行数据、前端零调用）。
// 见 internal/domain/models_batch3.go 顶部说明与迁移 000113。
