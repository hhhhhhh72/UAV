package memory

import (
	"errors"
	"sync"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type shopRepo struct {
	mu   sync.RWMutex
	data map[string]domain.Shop
}

func NewShopRepository() repository.ShopRepository {
	return &shopRepo{data: make(map[string]domain.Shop)}
}

func (r *shopRepo) Create(s domain.Shop) (domain.Shop, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[s.ID] = s
	return s, nil
}

func (r *shopRepo) Update(s domain.Shop) (domain.Shop, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[s.ID]; !ok {
		return domain.Shop{}, errors.New("not found")
	}
	r.data[s.ID] = s
	return s, nil
}

func (r *shopRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

func (r *shopRepo) FindByID(id string) (domain.Shop, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.data[id]
	if !ok {
		return domain.Shop{}, errors.New("not found")
	}
	return s, nil
}

func (r *shopRepo) List(offset, limit int) ([]domain.Shop, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Shop, 0, len(r.data))
	for _, s := range r.data {
		out = append(out, s)
	}
	total := len(out)
	if offset >= total {
		return []domain.Shop{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return out[offset:end], total, nil
}
