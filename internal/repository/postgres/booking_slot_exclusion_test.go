package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// TestPG_BookingSlotExclusion 场地预约的时段重叠由**数据库**兜底（migration 000120）。
//
// 为什么需要：冲突判定此前只在 service 层，而 lockByKey 是进程内锁 —— 多一个 API
// 实例就各锁各的，同一场地同一时段会被双重占用；check-then-insert 本身也没有库级兜底
// （唯一索引表达不了区间重叠）。约束用「生成列 + EXCLUDE」表达，因为 PG 的排他约束
// 不支持 WHERE 部分约束：非占用状态时生成列为 NULL，NULL 之间不冲突。
//
// 这里直接调仓储层、绕过 service 的键锁，测的就是库那一层。
func TestPG_BookingSlotExclusion(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	ctx := context.Background()
	repo := store.NewTestSiteRepository()

	siteID := ug("site")
	base := time.Now().Add(72 * time.Hour).Truncate(time.Second)
	mk := func(id, status string, startOffset, endOffset time.Duration) domain.TestSiteBooking {
		return domain.TestSiteBooking{
			ID: id, SiteID: siteID, UserID: "u-1", Status: status,
			StartTime: base.Add(startOffset), EndTime: base.Add(endOffset),
			Purpose: "test", ContactName: "n", ContactPhone: "13800000000",
		}
	}

	// 第一笔 approved 占住 10:00~12:00（相对 base）
	if _, err := repo.CreateBooking(ctx, mk(ug("bk"), "approved", 0, 2*time.Hour)); err != nil {
		t.Fatalf("第一笔 approved 应能落库: %v", err)
	}
	// 同一场地完全重叠 → 必须被库拒绝，且翻译成 ErrSlotTaken
	_, err := repo.CreateBooking(ctx, mk(ug("bk"), "approved", time.Hour, 3*time.Hour))
	if !errors.Is(err, repository.ErrSlotTaken) {
		t.Fatalf("重叠的 approved 应被拒并翻译为 ErrSlotTaken，实得 %v", err)
	}
	// 紧邻（后一场正好从 12:00 开始，闭区间语义）→ 也算冲突
	_, err = repo.CreateBooking(ctx, mk(ug("bk"), "approved", 2*time.Hour, 4*time.Hour))
	if !errors.Is(err, repository.ErrSlotTaken) {
		t.Fatalf("紧邻时段（闭区间）应被拒，实得 %v", err)
	}
	// 不重叠 → 允许
	if _, err := repo.CreateBooking(ctx, mk(ug("bk"), "approved", 3*time.Hour, 4*time.Hour)); err != nil {
		t.Fatalf("不重叠的 approved 应能落库: %v", err)
	}
	// 取消态与已通过时段重叠 → 不占位，允许（生成列为 NULL）
	if _, err := repo.CreateBooking(ctx, mk(ug("bk"), "cancelled", 0, 2*time.Hour)); err != nil {
		t.Fatalf("cancelled 与 approved 重叠应被允许: %v", err)
	}
}
