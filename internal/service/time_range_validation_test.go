package service_test

import (
	"context"
	"testing"
	"time"

	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 时间顺序回归：结束时间早于开始时间拒绝（end 零值放行——应急进行中场景）。
func TestTimeRangeValidation(t *testing.T) {
	start := time.Now()
	end := start.Add(-time.Hour)

	// 活动
	eSvc := service.NewEventService(memory.NewEventRepository())
	if _, err := eSvc.Create(context.Background(), "t", "exhibition", "d", "l", "c", start, end, 10); err == nil {
		t.Fatal("event with end < start must be rejected")
	}
	// 应急调度
	emSvc := service.NewEmergencyService(memory.NewEmergencyRepository())
	if _, err := emSvc.CreateDispatch(context.Background(), "r-1", "d", "l", "c", "", "", start, end); err == nil {
		t.Fatal("dispatch with end < start must be rejected")
	}
	// 正常 + end 零值（进行中）放行
	ev, err := eSvc.Create(context.Background(), "t2", "exhibition", "d", "l", "c", start, time.Time{}, 10)
	if err != nil {
		t.Fatalf("event with zero end should be allowed: %v", err)
	}
	if ev.StartTime.IsZero() {
		t.Fatal("event start time lost")
	}
}
