package httpapi

import (
	"testing"
	"time"
)

// parseDashboardRange：合法值映射窗口/粒度，缺省与非法值回落 12 个月。
func TestParseDashboardRange(t *testing.T) {
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	cases := []struct {
		raw        string
		wantKey    string
		wantBucket string
		wantPoints int
		wantSince  time.Time
	}{
		{"7d", "7d", "day", 7, dayStart.AddDate(0, 0, -6)},
		{"30d", "30d", "day", 30, dayStart.AddDate(0, 0, -29)},
		{"90d", "90d", "day", 90, dayStart.AddDate(0, 0, -89)},
		{"", "12m", "month", 12, time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -11, 0)},
		{"bogus", "12m", "month", 12, time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -11, 0)},
	}
	for _, c := range cases {
		got := parseDashboardRange(c.raw)
		if got.Key != c.wantKey || got.Bucket != c.wantBucket || got.Points != c.wantPoints {
			t.Errorf("parseDashboardRange(%q) = {%s %s %d}, want {%s %s %d}",
				c.raw, got.Key, got.Bucket, got.Points, c.wantKey, c.wantBucket, c.wantPoints)
		}
		if !got.Since.Equal(c.wantSince) {
			t.Errorf("parseDashboardRange(%q).Since = %v, want %v", c.raw, got.Since, c.wantSince)
		}
	}
}

// buildWindowTrends：窗口内补零铺满、窗口外忽略、同一天累加。
func TestBuildWindowTrendsDaily(t *testing.T) {
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	rng := dashboardRange{Key: "7d", Since: dayStart.AddDate(0, 0, -6), Bucket: "day", Points: 7}
	items := []time.Time{
		dayStart,                     // 今天
		dayStart.AddDate(0, 0, -1),   // 昨天
		dayStart.AddDate(0, 0, -1),   // 昨天（第 2 条）
		dayStart.AddDate(0, 0, -10),  // 窗口外，应忽略
	}
	out := buildWindowTrends(items, func(v time.Time) time.Time { return v }, rng)
	if len(out) != 7 {
		t.Fatalf("points = %d, want 7", len(out))
	}
	if out[0]["date"] != dayStart.AddDate(0, 0, -6).Format("2006-01-02") {
		t.Errorf("first bucket = %v", out[0]["date"])
	}
	if out[6]["count"] != 1 {
		t.Errorf("today count = %v, want 1", out[6]["count"])
	}
	if out[5]["count"] != 2 {
		t.Errorf("yesterday count = %v, want 2", out[5]["count"])
	}
	if out[0]["count"] != 0 {
		t.Errorf("oldest bucket count = %v, want 0（窗口内无数据补零）", out[0]["count"])
	}
}

// buildWindowTrends：月粒度按自然月分桶。
func TestBuildWindowTrendsMonthly(t *testing.T) {
	now := time.Now()
	rng := dashboardRange{
		Key: "12m", Bucket: "month", Points: 12,
		Since: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -11, 0),
	}
	items := []time.Time{now, now.AddDate(0, -3, 0), now.AddDate(0, -20, 0)}
	out := buildWindowTrends(items, func(v time.Time) time.Time { return v }, rng)
	if len(out) != 12 {
		t.Fatalf("points = %d, want 12", len(out))
	}
	if out[11]["count"] != 1 {
		t.Errorf("当前月 count = %v, want 1", out[11]["count"])
	}
	if out[8]["count"] != 1 {
		t.Errorf("3 个月前 count = %v, want 1", out[8]["count"])
	}
	if out[0]["count"] != 0 {
		t.Errorf("最早月 count = %v, want 0", out[0]["count"])
	}
}
