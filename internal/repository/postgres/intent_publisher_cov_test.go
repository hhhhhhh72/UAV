package postgres_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
)

// ListByPublisher 的 PG 实现走 JOIN demands，是发布方消息页「一键同意」的判定依据。
//
// 关键点：归属必须按 **demands.publisher_id**，而不是按 intentor_id——
// 后者会把"我投出去的意向"当成"我收到的意向"，一键同意就会认错人、替别人成单。
// 本用例特意让同一个意向方在两条需求上都投意向，专门卡这个区分。
func TestCoverage_IntentRepoListByPublisher(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	ctx := context.Background()
	demandRepo := store.NewDemandRepository()
	intentRepo := store.NewIntentRepository()

	pubID := ug("cov-ipub")
	otherPubID := ug("cov-iother")
	dMine := ug("cov-idemand1")
	dOther := ug("cov-idemand2")

	for _, d := range []domain.Demand{
		{
			ID: dMine, PublisherID: pubID, PublisherName: "发布方甲", Contact: "13800000000",
			BizType: domain.BizCableInspection, District: "渝北区", CityCode: "500112",
			Title: "我的需求", Description: "归属测试", Status: domain.DemandPublished,
		},
		{
			ID: dOther, PublisherID: otherPubID, PublisherName: "发布方乙", Contact: "13800000001",
			BizType: domain.BizCableInspection, District: "渝北区", CityCode: "500112",
			Title: "别人的需求", Description: "归属测试", Status: domain.DemandPublished,
		},
	} {
		if _, err := demandRepo.Create(ctx, d); err != nil {
			t.Fatalf("seed demand %s: %v", d.ID, err)
		}
	}

	// 同一个意向方在两条需求上都投了意向
	intentorID := ug("cov-iintentor")
	iMine := ug("cov-iintent1")
	iOther := ug("cov-iintent2")
	for _, it := range []domain.DemandIntent{
		{ID: iMine, DemandID: dMine, IntentorID: intentorID, IntentorName: "飞手小李", Contact: "13900000000", Status: "pending"},
		{ID: iOther, DemandID: dOther, IntentorID: intentorID, IntentorName: "飞手小李", Contact: "13900000000", Status: "pending"},
	} {
		if _, err := intentRepo.Create(ctx, it); err != nil {
			t.Fatalf("seed intent %s: %v", it.ID, err)
		}
	}

	got, err := intentRepo.ListByPublisher(ctx, pubID)
	if err != nil {
		t.Fatalf("ListByPublisher: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListByPublisher(%s): len=%d, want 1（只含本人需求上的意向）", pubID, len(got))
	}
	if got[0].ID != iMine {
		t.Fatalf("ListByPublisher 取到 %s，want %s", got[0].ID, iMine)
	}
	if got[0].Status != "pending" || got[0].IntentorName != "飞手小李" {
		t.Fatalf("字段未正确映射: %+v", got[0])
	}
	if got[0].DemandTitle != "我的需求" {
		t.Fatalf("DemandTitle=%q, want 我的需求（JOIN d.title 未带出）", got[0].DemandTitle)
	}

	// 反向：意向方本人没有发布过需求 → 空（若错按 intentor_id 归属，这里会拿到 2 条）
	if none, err := intentRepo.ListByPublisher(ctx, intentorID); err != nil || len(none) != 0 {
		t.Fatalf("ListByPublisher(意向方): len=%d err=%v, want 0/nil", len(none), err)
	}
	// 另一个发布者只看到自己那条
	if otherGot, err := intentRepo.ListByPublisher(ctx, otherPubID); err != nil || len(otherGot) != 1 || otherGot[0].ID != iOther {
		t.Fatalf("ListByPublisher(另一发布者): len=%d err=%v", len(otherGot), err)
	}
}
