package memory

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type rescueCaseRepo struct{ mu sync.RWMutex; items []domain.RescueCase }
func NewRescueCaseRepository() repository.RescueCaseRepository { return &rescueCaseRepo{} }
func (r *rescueCaseRepo) Create(rc domain.RescueCase) (domain.RescueCase, error) { r.mu.Lock(); defer r.mu.Unlock(); r.items = append(r.items, rc); return rc, nil }
func (r *rescueCaseRepo) FindByID(id string) (domain.RescueCase, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, rc := range r.items { if rc.ID == id { return rc, nil } }
	return domain.RescueCase{}, fmt.Errorf("not found")
}
func (r *rescueCaseRepo) List(eventType, q string, offset, limit int) ([]domain.RescueCase, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	query := strings.ToLower(strings.TrimSpace(q))
	filtered := make([]domain.RescueCase, 0)
	for _, rc := range r.items {
		if eventType != "" && rc.EventType != eventType {
			continue
		}
		if query != "" && !matchAnyFold(query, rc.Title, rc.Location, rc.Summary, rc.TeamName, rc.DroneModel) {
			continue
		}
		filtered = append(filtered, rc)
	}
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

type emergDeptRepo struct{ mu sync.RWMutex; depts []domain.EmergencyDept; drills []domain.EmergencyDrill }
func NewEmergencyDeptRepository() repository.EmergencyDeptRepository { return &emergDeptRepo{} }
func (r *emergDeptRepo) CreateDept(d domain.EmergencyDept) (domain.EmergencyDept, error) { r.mu.Lock(); defer r.mu.Unlock(); r.depts = append(r.depts, d); return d, nil }
func (r *emergDeptRepo) ListDepts() ([]domain.EmergencyDept, error) { r.mu.RLock(); defer r.mu.RUnlock(); return append([]domain.EmergencyDept(nil), r.depts...), nil }
func (r *emergDeptRepo) CreateDrill(d domain.EmergencyDrill) (domain.EmergencyDrill, error) { r.mu.Lock(); defer r.mu.Unlock(); r.drills = append(r.drills, d); return d, nil }
func (r *emergDeptRepo) ListDrills(deptID string) ([]domain.EmergencyDrill, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := make([]domain.EmergencyDrill, 0)
	for _, d := range r.drills { if deptID == "" || d.DeptID == deptID { out = append(out, d) } }
	return out, nil
}

type assocMemberRepo struct{ mu sync.RWMutex; items []domain.AssociationMember }
func NewAssociationMemberRepository() repository.AssociationMemberRepository { return &assocMemberRepo{} }
func (r *assocMemberRepo) Create(m domain.AssociationMember) (domain.AssociationMember, error) { r.mu.Lock(); defer r.mu.Unlock(); r.items = append(r.items, m); return m, nil }
func (r *assocMemberRepo) FindByUserID(userID string) (domain.AssociationMember, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, m := range r.items { if m.UserID == userID { return m, nil } }
	return domain.AssociationMember{}, fmt.Errorf("not found")
}
func (r *assocMemberRepo) List(role string, offset, limit int) ([]domain.AssociationMember, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	filtered := make([]domain.AssociationMember, 0)
	for _, m := range r.items { if role == "" || string(m.Role) == role { filtered = append(filtered, m) } }
	return paginateSlice(filtered, offset, limit)
}
func (r *assocMemberRepo) UpdateRole(id string, role domain.AssociationRole) (domain.AssociationMember, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i, m := range r.items { if m.ID == id { r.items[i].Role = role; r.items[i].UpdatedAt = time.Now(); return r.items[i], nil } }
	return domain.AssociationMember{}, fmt.Errorf("not found")
}
