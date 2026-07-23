package service_test

import (
	"testing"

	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

func TestMatchingService_Recommend(t *testing.T) {
	repo := memory.NewDemandRepository(nil)
	svc := service.NewMatchingService(repo)

	results, err := svc.Recommend("user-1", 29.5, 106.5, "cable_inspection", "", 10)
	if err != nil {
		t.Fatalf("recommend failed: %v", err)
	}
	t.Logf("got %d recommendations", len(results))
	for _, r := range results {
		t.Logf("  score=%.4f reasons=%v title=%s", r.Score, r.Reasons, r.Demand.Title)
	}
}

func TestMatchingService_SearchAndMatch(t *testing.T) {
	repo := memory.NewDemandRepository(nil)
	svc := service.NewMatchingService(repo)

	results, err := svc.SearchAndMatch("巡检", 29.5, 106.5, "cable_inspection", 10)
	if err != nil {
		t.Fatalf("search+match failed: %v", err)
	}
	t.Logf("found %d results for q=巡检", len(results))
}
