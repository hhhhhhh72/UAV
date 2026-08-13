package service_test

import (
	"testing"
	"time"

	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// TestRescueCaseEventTypeNormalization: 回归 C12——英文键创建/筛选与中文存量数据互认。
// 小程序 tab 传中文（山火/洪水/…），API 客户端可能传 domain 注释中的英文键。
func TestRescueCaseEventTypeNormalization(t *testing.T) {
	svc := service.NewRescueCaseService(memory.NewRescueCaseRepository())

	// 英文键创建 → 存储中文规范值
	rc, err := svc.Create("山火案例", "mountain_fire", "南山", "M300", "应急队", "摘要", "结果", "教训", "协会", time.Now())
	if err != nil || rc.EventType != "山火" {
		t.Fatalf("english create: type=%q err=%v, want 山火", rc.EventType, err)
	}
	svc.Create("洪水案例", "洪水", "嘉陵江", "M300", "应急队", "", "", "", "", time.Now())
	svc.Create("滑坡案例", "landslide", "武隆", "M300", "应急队", "", "", "", "", time.Now())

	// 中文与英文筛选都能命中中文存量
	if _, total, _ := svc.List("山火", "", 1, 20); total != 1 {
		t.Fatalf("中文筛选 total=%d, want 1", total)
	}
	if _, total, _ := svc.List("mountain_fire", "", 1, 20); total != 1 {
		t.Fatalf("英文筛选 total=%d, want 1", total)
	}
	// 未知类型原样传递，仍按精确匹配
	if _, total, _ := svc.List("landslide", "", 1, 20); total != 1 {
		t.Fatalf("未知类型筛选 total=%d, want 1", total)
	}
}

// TestRescueCaseListSearch: 回归 C12——q 参数按标题/地点/摘要/队伍/机型子串搜索，
// 可与 event_type 组合。
func TestRescueCaseListSearch(t *testing.T) {
	svc := service.NewRescueCaseService(memory.NewRescueCaseRepository())
	svc.Create("南山森林火情侦察", "山火", "重庆市南岸区", "M300", "应急队", "热成像侦察", "", "", "", time.Now())
	svc.Create("嘉陵江巡查", "洪水", "北碚", "M300", "应急队", "", "", "", "", time.Now())

	if _, total, _ := svc.List("", "南岸", 1, 20); total != 1 {
		t.Fatalf("q=南岸 total=%d, want 1", total)
	}
	if _, total, _ := svc.List("", "热成像", 1, 20); total != 1 {
		t.Fatalf("q=热成像 total=%d, want 1", total)
	}
	if _, total, _ := svc.List("", "不存在", 1, 20); total != 0 {
		t.Fatalf("q=不存在 total=%d, want 0", total)
	}
	// 组合筛选：类型 + q
	if _, total, _ := svc.List("洪水", "嘉陵江", 1, 20); total != 1 {
		t.Fatalf("组合筛选 total=%d, want 1", total)
	}
	if _, total, _ := svc.List("山火", "嘉陵江", 1, 20); total != 0 {
		t.Fatalf("错配组合筛选 total=%d, want 0", total)
	}
}

// TestEmergencyResourceTypeNormalization: 回归 C12——中文 res_type 创建归一为英文规范值，
// res_type（中英皆可）/q 筛选生效。
func TestEmergencyResourceTypeNormalization(t *testing.T) {
	svc := service.NewEmergencyService(memory.NewEmergencyRepository())
	r1, err := svc.CreateResource("u-1", "应急无人机01", "无人机", "M300RTK+热成像", "南岸", "138", 2)
	if err != nil || r1.ResType != "drone" {
		t.Fatalf("中文 res_type 创建: type=%q err=%v, want drone", r1.ResType, err)
	}
	svc.CreateResource("u-1", "通讯指挥车", "comm", "卫星链路", "渝中", "139", 1)

	if _, total, _ := svc.ListResources("drone", "", 1, 20); total != 1 {
		t.Fatalf("res_type=drone total=%d, want 1", total)
	}
	// 中文筛选同样归一
	if _, total, _ := svc.ListResources("无人机", "", 1, 20); total != 1 {
		t.Fatalf("中文 res_type 筛选 total=%d, want 1", total)
	}
	if _, total, _ := svc.ListResources("", "热成像", 1, 20); total != 1 {
		t.Fatalf("q=热成像 total=%d, want 1", total)
	}
	if _, total, _ := svc.ListResources("comm", "指挥车", 1, 20); total != 1 {
		t.Fatalf("组合筛选 total=%d, want 1", total)
	}
}
