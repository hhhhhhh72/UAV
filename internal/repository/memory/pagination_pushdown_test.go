package memory_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
)

// 列表分页下沉的回归测试（2026-09-18）。
//
// 这四处列表此前是「拉全量 N 行 → 上层内存过滤 → 再切片」，毛病有两个，且都只有数据量
// 上来才会发作、光看代码「能跑」：
//   ① total 报的是**当前这一批**的条数，不是过滤后的总数 —— 前端分页器据此算总页数，
//      于是用户只看到第一页就以为「平台就这几条」；
//   ② 拉取上限（2000 / 100000）之外的数据被静默截断；
//   ③ 排序发生在切片之前的那一批内部，翻页后整体顺序是乱的（sort=views/video 尤其明显）。
//
// 所以断言的核心只有一条：**total 必须是过滤后的真实总数，与 limit 无关**；
// 外加「跨页仍保持全局有序」。只断言「本页返回了几条」是抓不到这三个 bug 的。

func TestJobListPublishedFiltersAndCountsInStore(t *testing.T) {
	r := memory.NewJobRepository()
	ctx := context.Background()
	for _, j := range []domain.Job{
		{ID: "j-1", EnterpriseID: "e-1", Title: "无人机飞手", Location: "重庆", JobType: "全职", Status: domain.JobPublished},
		{ID: "j-2", EnterpriseID: "e-1", Title: "测绘工程师", Location: "成都", JobType: "兼职", Status: domain.JobPublished},
		{ID: "j-3", EnterpriseID: "e-1", Title: "无人机维修", Location: "重庆", JobType: "全职", Status: domain.JobPublished},
		{ID: "j-4", EnterpriseID: "e-1", Title: "已关闭职位", Location: "重庆", JobType: "全职", Status: domain.JobClosed},
	} {
		if _, err := r.Create(ctx, j); err != nil {
			t.Fatalf("create %s: %v", j.ID, err)
		}
	}

	items, total, err := r.ListPublished(ctx, "", "", 0, 100)
	if err != nil || total != 3 || len(items) != 3 {
		t.Fatalf("无过滤：total=%d len=%d err=%v，期望 3/3（closed 不该出现）", total, len(items), err)
	}

	// ★ 核心断言：limit=1 时 total 仍必须是过滤后的总数 3，而不是本页的 1 条
	items, total, err = r.ListPublished(ctx, "", "", 0, 1)
	if err != nil || total != 3 || len(items) != 1 {
		t.Fatalf("分页：total=%d len=%d err=%v，期望 total=3 len=1（total 是总数，不是本页条数）", total, len(items), err)
	}

	items, total, _ = r.ListPublished(ctx, "", "全职", 0, 100)
	if total != 2 || len(items) != 2 {
		t.Fatalf("按类型过滤：total=%d len=%d，期望 2/2", total, len(items))
	}
	for _, j := range items {
		if j.JobType != "全职" {
			t.Fatalf("按类型过滤混入了 %q", j.JobType)
		}
	}

	if _, total, _ = r.ListPublished(ctx, "重庆", "", 0, 100); total != 2 {
		t.Fatalf("按地点搜索：total=%d，期望 2", total)
	}
	if _, total, _ = r.ListPublished(ctx, "无人机", "", 0, 100); total != 2 {
		t.Fatalf("按标题搜索：total=%d，期望 2", total)
	}

	// 组合过滤 + 翻页：两页都必须是同一次过滤的结果，total 不变
	first, total, _ := r.ListPublished(ctx, "无人机", "全职", 0, 1)
	second, total2, _ := r.ListPublished(ctx, "无人机", "全职", 1, 1)
	if total != 2 || total2 != 2 {
		t.Fatalf("组合过滤 total 应恒为 2，实际 %d / %d", total, total2)
	}
	if len(first) != 1 || len(second) != 1 || first[0].ID == second[0].ID {
		t.Fatalf("翻页结果不对：first=%v second=%v", first, second)
	}
}

func TestArticleListByCategoryFiltersStatusInStore(t *testing.T) {
	r := memory.NewArticleRepository()
	ctx := context.Background()
	for _, a := range []domain.Article{
		{ID: "a-1", Title: "政策一", Category: "policy", Status: "published"},
		{ID: "a-2", Title: "政策二（草稿）", Category: "policy", Status: "draft"},
		{ID: "a-3", Title: "案例一", Category: "case", Status: "published"},
	} {
		if _, err := r.Create(ctx, a); err != nil {
			t.Fatalf("create %s: %v", a.ID, err)
		}
	}

	items, total, err := r.ListByCategory(ctx, "", "published", 0, 100)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("公开端只看 published：total=%d len=%d err=%v，期望 2/2", total, len(items), err)
	}
	for _, a := range items {
		if a.Status != "published" {
			t.Fatalf("公开端混入了 %q 状态的文章", a.Status)
		}
	}

	// ★ 核心断言：limit=1 时 total 仍是 2
	items, total, _ = r.ListByCategory(ctx, "", "published", 0, 1)
	if total != 2 || len(items) != 1 {
		t.Fatalf("分页：total=%d len=%d，期望 total=2 len=1", total, len(items))
	}

	// 管理端：status 传空 = 全量（含草稿）
	if _, total, _ = r.ListByCategory(ctx, "", "", 0, 100); total != 3 {
		t.Fatalf("管理端全量：total=%d，期望 3", total)
	}

	items, total, _ = r.ListByCategory(ctx, "policy", "published", 0, 100)
	if total != 1 || len(items) != 1 || items[0].ID != "a-1" {
		t.Fatalf("category+status 组合：total=%d items=%v，期望只剩 a-1", total, items)
	}
}

func TestPortfolioListPublishedFiltersAndSortsInStore(t *testing.T) {
	r := memory.NewPortfolioRepository()
	ctx := context.Background()
	for _, p := range []domain.MemberPortfolio{
		{ID: "p-1", Name: "渝航科技", Description: "植保无人机服务", Category: "整机", Industry: "农业", Status: "published", Views: 5, Featured: true},
		{ID: "p-2", Name: "山城测绘", Description: "测绘与巡检", Category: "服务", Industry: "测绘", Status: "published", Views: 9},
		{ID: "p-3", Name: "两江智控", Description: "飞控研发", Category: "整机", Industry: "制造", Status: "published", Views: 1, VideoURL: "http://x/v.mp4", VideoCount: 2},
		{ID: "p-4", Name: "草稿品牌", Description: "不该出现", Category: "整机", Industry: "制造", Status: "draft", Views: 99},
	} {
		if _, err := r.Create(ctx, p); err != nil {
			t.Fatalf("create %s: %v", p.ID, err)
		}
	}

	if _, total, err := r.ListPublished(ctx, "", "", "", false, 0, 100); err != nil || total != 3 {
		t.Fatalf("无过滤：total=%d err=%v，期望 3（draft 不该出现）", total, err)
	}

	// category 同时匹配 category 与 industry：整机 2 条、测绘 1 条
	if _, total, _ := r.ListPublished(ctx, "", "整机", "", false, 0, 100); total != 2 {
		t.Fatalf("按 category 过滤：total=%d，期望 2", total)
	}
	if _, total, _ := r.ListPublished(ctx, "", "测绘", "", false, 0, 100); total != 1 {
		t.Fatalf("按 industry 过滤：total=%d，期望 1", total)
	}

	// q 命中描述而不是名称
	if _, total, _ := r.ListPublished(ctx, "植保", "", "", false, 0, 100); total != 1 {
		t.Fatalf("按描述搜索：total=%d，期望 1", total)
	}

	// featuredOnly
	items, total, _ := r.ListPublished(ctx, "", "", "", true, 0, 100)
	if total != 1 || len(items) != 1 || items[0].ID != "p-1" {
		t.Fatalf("只看精选：total=%d items=%v，期望只剩 p-1", total, items)
	}

	// ★ 排序必须全局有序，而不是「每页各自有序」：
	// views 分别是 5/9/1，逐页取一条应当得到 9 → 5 → 1；且每页的 total 恒为 3。
	want := []string{"p-2", "p-1", "p-3"}
	for i, id := range want {
		page, total, err := r.ListPublished(ctx, "", "", "views", false, i, 1)
		if err != nil || total != 3 || len(page) != 1 {
			t.Fatalf("sort=views 第 %d 页：total=%d len=%d err=%v，期望 total=3 len=1", i, total, len(page), err)
		}
		if page[0].ID != id {
			t.Fatalf("sort=views 第 %d 页取到 %s，期望 %s（排序只作用于当前批就说明没下沉到存储层）",
				i, page[0].ID, id)
		}
	}

	// sort=video：有视频的排最前
	if page, _, _ := r.ListPublished(ctx, "", "", "video", false, 0, 1); len(page) != 1 || page[0].ID != "p-3" {
		t.Fatalf("sort=video 首页应是 p-3，实际 %v", page)
	}
}

func TestEmergencyListDispatchesFiltersStatusInStore(t *testing.T) {
	r := memory.NewEmergencyRepository()
	ctx := context.Background()
	for _, d := range []domain.EmergencyDispatch{
		{ID: "d-1", ResourceID: "r-1", EventDesc: "火情侦察", Status: "completed"},
		{ID: "d-2", ResourceID: "r-1", EventDesc: "物资投送", Status: "pending"},
		{ID: "d-3", ResourceID: "r-2", EventDesc: "灾情巡查", Status: "pending"},
	} {
		if _, err := r.CreateDispatch(ctx, d); err != nil {
			t.Fatalf("create %s: %v", d.ID, err)
		}
	}

	if _, total, err := r.ListDispatches(ctx, "", "", 0, 100); err != nil || total != 3 {
		t.Fatalf("无过滤：total=%d err=%v，期望 3", total, err)
	}

	items, total, err := r.ListDispatches(ctx, "", "pending", 0, 100)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("按状态过滤：total=%d len=%d err=%v，期望 2/2", total, len(items), err)
	}
	for _, d := range items {
		if d.Status != "pending" {
			t.Fatalf("按状态过滤混入了 %q", d.Status)
		}
	}

	// ★ 核心断言：limit=1 时 total 仍是过滤后的 2
	items, total, _ = r.ListDispatches(ctx, "", "pending", 0, 1)
	if total != 2 || len(items) != 1 {
		t.Fatalf("分页：total=%d len=%d，期望 total=2 len=1", total, len(items))
	}

	// 资源 + 状态组合
	if _, total, _ := r.ListDispatches(ctx, "r-1", "pending", 0, 100); total != 1 {
		t.Fatalf("资源+状态组合：total=%d，期望 1", total)
	}
}