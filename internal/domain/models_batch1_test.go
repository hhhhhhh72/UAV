package domain

import (
	"testing"
	"time"
)

// ParseTime 只接受 RFC3339 格式；其余输入回退到当前时间。
func TestParseTimeRFC3339(t *testing.T) {
	in := "2024-06-01T12:00:00Z"
	want := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	got := ParseTime(in)
	if !got.Equal(want) {
		t.Errorf("ParseTime(%q) = %v, want %v", in, got, want)
	}
}

func TestParseTimeFallback(t *testing.T) {
	before := time.Now()
	// "2006-01-02" 与 "2024-06-01" 是日期串（非 RFC3339），空串/非法串同样触发回退。
	for _, in := range []string{"2006-01-02", "2024-06-01", "", "not-a-time"} {
		got := ParseTime(in)
		if got.IsZero() {
			t.Errorf("ParseTime(%q) returned zero time; want fallback to now", in)
		}
		if got.Before(before.Add(-time.Minute)) || got.After(before.Add(time.Minute)) {
			t.Errorf("ParseTime(%q) = %v; want within 1 minute of now (%v)", in, got, before)
		}
	}
}
