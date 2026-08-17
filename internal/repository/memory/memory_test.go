package memory_test

import (
	"context"
	"sync"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
)

// TestEnterpriseRepoConcurrency verifies BLK-12 fix: no data race under concurrent reads.
func TestEnterpriseRepoConcurrency(t *testing.T) {
	repo := memory.NewEnterpriseRepository(nil)
	var wg sync.WaitGroup
	const goroutines = 50

	// Concurrent reads should not race.
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.Pending(context.Background())
			if err != nil {
				t.Errorf("Pending() error: %v", err)
			}
		}()
	}
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.Search(context.Background(), "重庆")
			if err != nil {
				t.Errorf("Search() error: %v", err)
			}
		}()
	}
	wg.Wait()
}

// TestDemandRepoConcurrency verifies existing lock works under concurrent reads+writes.
func TestDemandRepoConcurrency(t *testing.T) {
	repo := memory.NewDemandRepository(nil)
	var wg sync.WaitGroup
	const goroutines = 20

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _ = repo.List(context.Background(), repository.DemandFilter{})
		}(i)
	}
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _ = repo.Create(context.Background(), domain.Demand{ID: "test-" + string(rune(idx))})
		}(i)
	}
	wg.Wait()
}
