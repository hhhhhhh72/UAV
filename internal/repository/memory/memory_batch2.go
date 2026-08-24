package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type transRepo struct {
	mu    sync.RWMutex
	items []domain.Transformation
}

func NewTransformationRepository() repository.TransformationRepository { return &transRepo{} }
func (r *transRepo) Create(ctx context.Context, t domain.Transformation) (domain.Transformation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, t)
	return t, nil
}
func (r *transRepo) FindByID(ctx context.Context, id string) (domain.Transformation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, t := range r.items {
		if t.ID == id {
			return t, nil
		}
	}
	return domain.Transformation{}, fmt.Errorf("not found")
}
func (r *transRepo) List(ctx context.Context, ownerID string) ([]domain.Transformation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Transformation, 0)
	for _, t := range r.items {
		if ownerID == "" || t.OwnerID == ownerID {
			out = append(out, t)
		}
	}
	return out, nil
}
func (r *transRepo) Update(ctx context.Context, t domain.Transformation) (domain.Transformation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == t.ID {
			t.UpdatedAt = time.Now()
			r.items[i] = t
			return t, nil
		}
	}
	return domain.Transformation{}, fmt.Errorf("not found")
}

func (r *transRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("not found")
}

type collegeRepo struct {
	mu    sync.RWMutex
	items []domain.College
}

func NewCollegeRepository() repository.CollegeRepository { return &collegeRepo{} }
func (r *collegeRepo) Create(ctx context.Context, c domain.College) (domain.College, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, c)
	return c, nil
}
func (r *collegeRepo) FindByID(ctx context.Context, id string) (domain.College, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.items {
		if c.ID == id {
			return c, nil
		}
	}
	return domain.College{}, fmt.Errorf("not found")
}
func (r *collegeRepo) List(ctx context.Context, region string) ([]domain.College, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.College, 0)
	for _, c := range r.items {
		// 与 PG 对齐：公开列表仅展示 active。
		if c.Status != "active" {
			continue
		}
		if region == "" || c.Region == region {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *collegeRepo) Update(ctx context.Context, c domain.College) (domain.College, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == c.ID {
			r.items[i] = c
			return c, nil
		}
	}
	return domain.College{}, fmt.Errorf("college %s not found", c.ID)
}

func (r *collegeRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("college %s not found", id)
}

// ---- StudyTour ----

type studyTourRepo struct {
	mu    sync.RWMutex
	items []domain.StudyTour
}

func NewStudyTourRepository() repository.StudyTourRepository { return &studyTourRepo{} }
func (r *studyTourRepo) Create(ctx context.Context, s domain.StudyTour) (domain.StudyTour, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, s)
	return s, nil
}
func (r *studyTourRepo) FindByID(ctx context.Context, id string) (domain.StudyTour, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.items {
		if s.ID == id {
			return s, nil
		}
	}
	return domain.StudyTour{}, fmt.Errorf("study %s not found", id)
}
func (r *studyTourRepo) List(ctx context.Context) ([]domain.StudyTour, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.StudyTour(nil), r.items...), nil
}
func (r *studyTourRepo) Update(ctx context.Context, s domain.StudyTour) (domain.StudyTour, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == s.ID {
			r.items[i] = s
			return s, nil
		}
	}
	return domain.StudyTour{}, fmt.Errorf("study %s not found", s.ID)
}
func (r *studyTourRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("study %s not found", id)
}

type coopRepo struct {
	mu    sync.RWMutex
	items []domain.CooperationProgram
}

func NewCooperationRepository() repository.CooperationRepository { return &coopRepo{} }
func (r *coopRepo) Create(ctx context.Context, cp domain.CooperationProgram) (domain.CooperationProgram, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, cp)
	return cp, nil
}
func (r *coopRepo) FindByID(ctx context.Context, id string) (domain.CooperationProgram, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, cp := range r.items {
		if cp.ID == id {
			return cp, nil
		}
	}
	return domain.CooperationProgram{}, fmt.Errorf("not found")
}
func (r *coopRepo) List(ctx context.Context, enterpriseID string) ([]domain.CooperationProgram, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.CooperationProgram, 0)
	for _, cp := range r.items {
		if enterpriseID == "" || cp.EnterpriseID == enterpriseID {
			out = append(out, cp)
		}
	}
	return out, nil
}
func (r *coopRepo) UpdateStatus(ctx context.Context, id, status string) (domain.CooperationProgram, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, cp := range r.items {
		if cp.ID == id {
			r.items[i].Status = status
			r.items[i].UpdatedAt = time.Now()
			return r.items[i], nil
		}
	}
	return domain.CooperationProgram{}, fmt.Errorf("not found")
}
