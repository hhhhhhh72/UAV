package service_test

import (
	"context"
	"errors"
	"testing"

	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// TestComplianceDocCategoryNormalization: 回归 C11——合规文档英文分类传值归一为中文规范值
// （种子数据与小程序 knowledge.vue tabs 均为 政策/法规/标准/指南）。
func TestComplianceDocCategoryNormalization(t *testing.T) {
	svc := service.NewComplianceService(memory.NewComplianceRepository())
	d, err := svc.CreateDoc(context.Background(), "适航条例", "policy", "民航局", "2026-01-01", "published", "内容", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if d.Category != "政策" {
		t.Fatalf("stored category %q, want 政策", d.Category)
	}
	if d2, _ := svc.CreateDoc(context.Background(), "飞行规范", "法规", "民航局", "2026-02-01", "published", "内容", "", nil); d2.Category != "法规" {
		t.Fatalf("stored category %q, want 法规", d2.Category)
	}
	// 英文筛选与中文筛选等价
	for _, cat := range []string{"policy", "政策"} {
		if _, total, err := svc.ListDocs(context.Background(), cat, 1, 20); err != nil || total != 1 {
			t.Errorf("ListDocs(%q): total=%d err=%v, want 1", cat, total, err)
		}
	}
	if _, total, _ := svc.ListDocs(context.Background(), "法规", 1, 20); total != 1 {
		t.Errorf("ListDocs(法规): total=%d, want 1", total)
	}
	if _, total, _ := svc.ListDocs(context.Background(), "", 1, 20); total != 2 {
		t.Errorf("ListDocs(all): total=%d, want 2", total)
	}
}

// TestComplianceStandardCategoryNormalization: 回归 C11——标准分类列。
// 旧实现 standard_docs 无 category 列、内存 ListStandards 忽略过滤、PG 侧 WHERE category=$1 报错。
func TestComplianceStandardCategoryNormalization(t *testing.T) {
	svc := service.NewComplianceService(memory.NewComplianceRepository())
	s, err := svc.CreateStandard(context.Background(), "低空飞行标准", "group", "T/CDA-001", "协会", "2026-07-01", "published", "范围", "")
	if err != nil {
		t.Fatal(err)
	}
	if s.Category != "团体标准" {
		t.Fatalf("stored category %q, want 团体标准", s.Category)
	}
	if _, err := svc.CreateStandard(context.Background(), "国标测试", "national", "GB/T-001", "标委", "2026-06-01", "published", "范围", ""); err != nil {
		t.Fatal(err)
	}
	// 中文/英文筛选等价，各命中各自分类
	for _, cat := range []string{"团体标准", "group"} {
		items, total, err := svc.ListStandards(context.Background(), cat, 1, 20)
		if err != nil || total != 1 {
			t.Fatalf("ListStandards(%q): total=%d err=%v, want 1", cat, total, err)
		}
		if len(items) != 1 || items[0].Category != "团体标准" {
			t.Fatalf("ListStandards(%q): items=%+v, want 1 团体标准", cat, items)
		}
	}
	if _, total, _ := svc.ListStandards(context.Background(), "国家标准", 1, 20); total != 1 {
		t.Errorf("ListStandards(国家标准): total=%d, want 1", total)
	}
	if _, total, _ := svc.ListStandards(context.Background(), "", 1, 20); total != 2 {
		t.Errorf("ListStandards(all): total=%d, want 2", total)
	}
	// 更新时同样归一
	upd, err := svc.UpdateStandard(context.Background(), s.ID, "低空飞行标准v2", "enterprise", "T/CDA-002", "协会", "2026-08-01", "范围2", "published", "")
	if err != nil {
		t.Fatal(err)
	}
	if upd.Category != "企业标准" {
		t.Fatalf("updated category %q, want 企业标准", upd.Category)
	}
}

// TestResourceBookingService: 回归 C11——预约的资源必须存在，成功后落库 pending。
func TestResourceBookingService(t *testing.T) {
	svc := service.NewResourceService(memory.NewResourceRepository())
	if _, err := svc.Book(context.Background(), "u-1", "res-missing", "2026-08-20", "测试", "张三", "13800000000"); !errors.Is(err, service.ErrResourceNotFound) {
		t.Fatalf("book missing resource: err=%v, want ErrResourceNotFound", err)
	}
	res, err := svc.Create(context.Background(), "u-admin", "M300 航测机", "drone", "M300", "RTK", "重庆", "预约说明", 50000, "public")
	if err != nil {
		t.Fatal(err)
	}
	bk, err := svc.Book(context.Background(), "u-1", res.ID, "2026-08-20", "航测项目", "张三", "13800000000")
	if err != nil {
		t.Fatal(err)
	}
	if bk.Status != "pending" || bk.UserID != "u-1" || bk.BookingDate != "2026-08-20" {
		t.Fatalf("booking=%+v, want pending/u-1/2026-08-20", bk)
	}
	bookings, err := svc.ListBookingsByResource(context.Background(), res.ID)
	if err != nil || len(bookings) != 1 {
		t.Fatalf("bookings by resource: %d, err=%v, want 1", len(bookings), err)
	}
}
