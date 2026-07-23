package memory

import (
	"fmt"
	"sync"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type transRepo struct{ mu sync.RWMutex; items []domain.Transformation }
func NewTransformationRepository() repository.TransformationRepository { return &transRepo{} }
func (r *transRepo) Create(t domain.Transformation) (domain.Transformation, error) { r.mu.Lock(); defer r.mu.Unlock(); r.items = append(r.items, t); return t, nil }
func (r *transRepo) FindByID(id string) (domain.Transformation, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, t := range r.items { if t.ID == id { return t, nil } }
	return domain.Transformation{}, fmt.Errorf("not found")
}
func (r *transRepo) List(ownerID string) ([]domain.Transformation, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := make([]domain.Transformation, 0)
	for _, t := range r.items { if ownerID == "" || t.OwnerID == ownerID { out = append(out, t) } }
	return out, nil
}
func (r *transRepo) Update(t domain.Transformation) (domain.Transformation, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i, v := range r.items { if v.ID == t.ID { t.UpdatedAt = time.Now(); r.items[i] = t; return t, nil } }
	return domain.Transformation{}, fmt.Errorf("not found")
}

type collegeRepo struct{ mu sync.RWMutex; items []domain.College }
func NewCollegeRepository() repository.CollegeRepository { return &collegeRepo{} }
func (r *collegeRepo) Create(c domain.College) (domain.College, error) { r.mu.Lock(); defer r.mu.Unlock(); r.items = append(r.items, c); return c, nil }
func (r *collegeRepo) FindByID(id string) (domain.College, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, c := range r.items { if c.ID == id { return c, nil } }
	return domain.College{}, fmt.Errorf("not found")
}
func (r *collegeRepo) List(region string) ([]domain.College, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := make([]domain.College, 0)
	for _, c := range r.items { if region == "" || c.Region == region { out = append(out, c) } }
	return out, nil
}

type coopRepo struct{ mu sync.RWMutex; items []domain.CooperationProgram }
func NewCooperationRepository() repository.CooperationRepository { return &coopRepo{} }
func (r *coopRepo) Create(cp domain.CooperationProgram) (domain.CooperationProgram, error) { r.mu.Lock(); defer r.mu.Unlock(); r.items = append(r.items, cp); return cp, nil }
func (r *coopRepo) FindByID(id string) (domain.CooperationProgram, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, cp := range r.items { if cp.ID == id { return cp, nil } }
	return domain.CooperationProgram{}, fmt.Errorf("not found")
}
func (r *coopRepo) List(enterpriseID string) ([]domain.CooperationProgram, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := make([]domain.CooperationProgram, 0)
	for _, cp := range r.items { if enterpriseID == "" || cp.EnterpriseID == enterpriseID { out = append(out, cp) } }
	return out, nil
}
func (r *coopRepo) UpdateStatus(id, status string) (domain.CooperationProgram, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i, cp := range r.items { if cp.ID == id { r.items[i].Status = status; r.items[i].UpdatedAt = time.Now(); return r.items[i], nil } }
	return domain.CooperationProgram{}, fmt.Errorf("not found")
}
