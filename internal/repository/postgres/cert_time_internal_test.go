package postgres

import (
	"testing"
	"time"
)

// BUG-001 回归：证书"没有有效期"必须以 NULL 落库。
//
// 直接写 Go 零值时间会落成 0001-01-01 08:05:43+08:05（当年 LMT 偏移），它不是 NULL，
// 读侧的 COALESCE 兜不住，前端拿到这个假日期就判"已过期"、卡片显示"至 0001-01-01"。
func TestCertTimeOrNull(t *testing.T) {
	if got := certTimeOrNull(time.Time{}); got != nil {
		t.Fatalf("零值时间应写成 NULL，实际 %v", got)
	}
	if got := certTimeOrNull(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)); got != nil {
		t.Fatalf("1970 哨兵值应写成 NULL，实际 %v", got)
	}
	if got := certTimeOrNull(time.Date(1999, 12, 31, 0, 0, 0, 0, time.UTC)); got != nil {
		t.Fatalf("2000 年之前的哨兵值应写成 NULL，实际 %v", got)
	}
	real := time.Date(2029, 1, 14, 0, 0, 0, 0, time.UTC)
	got := certTimeOrNull(real)
	if got == nil || !got.Equal(real) {
		t.Fatalf("真实有效期必须原样落库，实际 %v", got)
	}
}

// BUG-001 读侧：NULL 与历史遗留的 0001/1970 哨兵值都要还原成"无有效期"（Go 零值时间），
// 这样 service 侧 certValid 才按"长期有效"放行、validExpireDate 才不把它算进到期台账。
func TestCertTimeFromNull(t *testing.T) {
	if got := certTimeFromNull(nil); !got.IsZero() {
		t.Fatalf("NULL 应还原成零值时间，实际 %v", got)
	}
	legacy := time.Date(1, 1, 1, 8, 5, 43, 0, time.FixedZone("LMT", 8*3600+343))
	if got := certTimeFromNull(&legacy); !got.IsZero() {
		t.Fatalf("历史遗留的 0001 哨兵值应还原成零值时间，实际 %v", got)
	}
	old := time.Date(1970, 1, 1, 8, 0, 0, 0, time.FixedZone("CST", 8*3600))
	if got := certTimeFromNull(&old); !got.IsZero() {
		t.Fatalf("1970 哨兵值应还原成零值时间，实际 %v", got)
	}
	real := time.Date(2029, 1, 14, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	if got := certTimeFromNull(&real); got.IsZero() || !got.Equal(real) {
		t.Fatalf("真实有效期应原样还原，实际 %v", got)
	}
}
