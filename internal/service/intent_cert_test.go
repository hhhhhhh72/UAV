package service_test

import (
	"context"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
)

// seedEntCertRepo 向已有企业仓库预置用户的一条 approved 企业认证。
func seedEntCertRepo(t *testing.T, entRepo repository.EnterpriseRepository, userID string) {
	t.Helper()
	now := time.Now()
	if _, err := entRepo.Create(context.Background(), domain.Enterprise{
		ID: "ent-cert-" + userID, OwnerUserID: userID, Name: "认证企业-" + userID,
		Status: domain.EnterpriseApproved, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("seed enterprise cert: %v", err)
	}
}

// newCertifiedEntRepo 新建内存企业仓库并给 userID 预置一条已通过（approved）的企业认证
// ——登记对接的认证门槛（企业认证或飞手认证任一）所需的最小前置数据。
func newCertifiedEntRepo(t *testing.T, userID string) repository.EnterpriseRepository {
	t.Helper()
	entRepo := memory.NewEnterpriseRepository(nil)
	seedEntCertRepo(t, entRepo, userID)
	return entRepo
}
