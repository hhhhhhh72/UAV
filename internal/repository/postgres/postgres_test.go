package postgres_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/postgres"
)

func databaseURL() string {
	u := os.Getenv("DATABASE_URL")
	if u == "" {
		u = "postgres://drone:drone_secret@127.0.0.1:5433/drone_platform?sslmode=disable"
	}
	return u
}
func ug(prefix string) string { return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()) }

// Tests share one database and run in parallel; running the full migration
// set from every test interleaves non-transactional multi-statement files.
// migrateOnce serializes migration to a single execution.
var migrateOnce sync.Once
var migrateErr error

func setupStore(t *testing.T) *postgres.Store {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	store, err := postgres.NewStore(ctx, databaseURL(), nil)
	if err != nil {
		t.Skipf("no PG: %v", err)
		return nil
	}
	migrateOnce.Do(func() {
		migrateErr = store.RunMigrationsFromDir(ctx, postgres.MigrationsDir())
	})
	if migrateErr != nil {
		t.Fatalf("migration: %v", migrateErr)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestDemandRepo_CreateAndFind(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	repo := store.NewDemandRepository()
	id := ug("td")
	d := domain.Demand{ID: id, PublisherID: "u-1", Title: "PG测试", Contact: "138", BizType: domain.BizCableInspection, Status: domain.DemandPending, Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	c, err := repo.Create(d)
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != id {
		t.Errorf("ID mismatch")
	}
	f, err := repo.FindByID(id)
	if err != nil || f.ID != id {
		t.Fatal("find failed")
	}
}

func TestDemandRepo_SetStatus(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	repo := store.NewDemandRepository()
	id := ug("td2")
	repo.Create(domain.Demand{ID: id, PublisherID: "u-2", Status: domain.DemandPending, Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	d, err := repo.SetStatus(id, domain.DemandPublished)
	if err != nil || d.Status != domain.DemandPublished {
		t.Fatalf("set status: %v", err)
	}
}

func TestDemandRepo_CompareAndSetStatus(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	repo := store.NewDemandRepository()
	id := ug("td3")
	repo.Create(domain.Demand{ID: id, PublisherID: "u-3", Status: domain.DemandPending, Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	ok, d, err := repo.CompareAndSetStatus(id, domain.DemandPending, domain.DemandPublished)
	if err != nil || !ok || d.Status != domain.DemandPublished {
		t.Fatal("CAS failed")
	}
	ok2, _, _ := repo.CompareAndSetStatus(id, domain.DemandCancelled, domain.DemandPublished)
	if ok2 {
		t.Fatal("CAS should fail with wrong old status")
	}
}

func TestBidRepo_CreateAndList(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	demandID := ug("dp")
	store.NewDemandRepository().Create(domain.Demand{ID: demandID, PublisherID: "u-1", Status: domain.DemandPublished, Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	repo := store.NewBidRepository()
	bidID := ug("tb")
	b, err := repo.Create(domain.DemandBid{ID: bidID, DemandID: demandID, BidderID: "u-2", AmountFen: 50000, Status: "pending", Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	list, _ := repo.ListByDemand(demandID)
	if len(list) == 0 {
		t.Fatal("list empty")
	}
	_ = b
}

func TestBidRepo_UpdateStatus(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	did := ug("dp2")
	store.NewDemandRepository().Create(domain.Demand{ID: did, PublisherID: "u-1", Status: domain.DemandPublished, Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	repo := store.NewBidRepository()
	id := ug("tb2")
	repo.Create(domain.DemandBid{ID: id, DemandID: did, BidderID: "u-3", Status: "pending", Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	b, err := repo.UpdateStatus(id, "accepted")
	if err != nil || b.Status != "accepted" {
		t.Fatalf("update: %v", err)
	}
}

func TestContractRepo_CreateAndUpdateStatus(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	repo := store.NewContractRepository()
	id := ug("tc")
	repo.Create(domain.Contract{ID: id, EnterpriseID: "e-1", TemplateID: "tpl-1", Status: domain.ContractDraft, Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	c, err := repo.UpdateStatus(id, domain.ContractSent)
	if err != nil || c.Status != domain.ContractSent {
		t.Fatalf("update status: %v", err)
	}
}

func TestContractRepo_FindByID(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	repo := store.NewContractRepository()
	id := ug("tc2")
	repo.Create(domain.Contract{ID: id, EnterpriseID: "e-2", Status: domain.ContractDraft, Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	c, err := repo.FindByID(id)
	if err != nil || c.ID != id {
		t.Fatal("find failed")
	}
}

func TestEmploymentRepo_ListWithPagination(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	repo := store.NewEmploymentRepository()
	id := ug("te")
	repo.Create(domain.EmploymentRequest{ID: id, EnterpriseID: "e-1", Position: "飞手", Headcount: 5, Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	list, total, err := repo.ListByEnterprise("e-1", 0, 20)
	if err != nil || total == 0 {
		t.Fatalf("list: %v total=%d", err, total)
	}
	_ = list
}
