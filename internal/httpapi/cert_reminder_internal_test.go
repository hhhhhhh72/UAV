package httpapi

import (
	"context"
	"strings"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 证书到期提醒扫描：窗口内发一条、重复扫描幂等、窗口外/久远过期/未审核通过都不发。
func TestCertExpiryReminderScan(t *testing.T) {
	certRepo := memory.NewCertificateRepository()
	msgRepo := memory.NewMessageRepository()
	srv := &Server{
		trainingSvc: service.NewTrainingService(certRepo, nil, nil, nil),
		expirySvc:   service.NewExpiryService(),
		msgSvc:      service.NewMessageService(msgRepo),
	}
	ctx := context.Background()
	now := time.Now()

	seed := []domain.Certificate{
		{ID: "cert-soon", UserID: "u-soon", CertType: domain.CertCAAC, CertNumber: "A-1", ExpireDate: now.AddDate(0, 0, 10), Status: "approved"},
		{ID: "cert-expired", UserID: "u-expired", CertType: domain.CertUTCDJI, CertNumber: "A-5", ExpireDate: now.AddDate(0, 0, -5), Status: "approved"},
		{ID: "cert-later", UserID: "u-later", CertType: domain.CertCAAC, CertNumber: "A-2", ExpireDate: now.AddDate(0, 0, 200), Status: "approved"},
		{ID: "cert-ancient", UserID: "u-ancient", CertType: domain.CertCAAC, CertNumber: "A-3", ExpireDate: now.AddDate(0, 0, -400), Status: "approved"},
		{ID: "cert-pending", UserID: "u-pending", CertType: domain.CertCAAC, CertNumber: "A-4", ExpireDate: now.AddDate(0, 0, 5), Status: "pending"},
	}
	for _, c := range seed {
		if _, err := certRepo.Create(ctx, c); err != nil {
			t.Fatalf("seed cert %s: %v", c.ID, err)
		}
	}

	countFor := func(userID string) int {
		t.Helper()
		list, err := msgRepo.ListByUser(ctx, userID, false)
		if err != nil {
			t.Fatalf("list messages for %s: %v", userID, err)
		}
		return len(list)
	}

	srv.scanExpiringCertificates()

	// 1) 30 天内到期 → 恰好一条，且带上证书 ID
	soon, err := msgRepo.ListByUser(ctx, "u-soon", false)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(soon) != 1 {
		t.Fatalf("窗口内证书应提醒 1 条，实际 %d 条", len(soon))
	}
	if soon[0].Title != "证书到期提醒" || soon[0].ResourceType != "certificate" || soon[0].ResourceID != "cert-soon" {
		t.Fatalf("提醒挂载字段不对: %+v", soon[0])
	}
	if !strings.Contains(soon[0].Content, "到期") {
		t.Fatalf("文案应说明到期: %q", soon[0].Content)
	}

	// 2) 刚过期 → 也提醒，但文案是"已过期"
	expired, _ := msgRepo.ListByUser(ctx, "u-expired", false)
	if len(expired) != 1 {
		t.Fatalf("刚过期证书应提醒 1 条，实际 %d 条", len(expired))
	}
	if !strings.Contains(expired[0].Content, "过期") {
		t.Fatalf("已过期文案不对: %q", expired[0].Content)
	}

	// 3) 窗口外 / 过期过久 / 未审核通过 → 一条都不发
	for _, u := range []string{"u-later", "u-ancient", "u-pending"} {
		if n := countFor(u); n != 0 {
			t.Fatalf("%s 不该收到提醒，实际 %d 条", u, n)
		}
	}

	// 4) 幂等：再扫一遍不能重复轰炸
	srv.scanExpiringCertificates()
	if n := countFor("u-soon"); n != 1 {
		t.Fatalf("重复扫描后应仍为 1 条，实际 %d 条", n)
	}
	if n := countFor("u-expired"); n != 1 {
		t.Fatalf("重复扫描后应仍为 1 条，实际 %d 条", n)
	}
}
