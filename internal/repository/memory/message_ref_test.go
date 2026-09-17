package memory_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
)

// DeleteByReference 是双条件删除：resource_id 命中 **且** resource_type 在列表里。
//
// 关键在"不误伤"——同 id 别的类型、同类型别的 id、以及管理端广播（两个字段都为空）
// 都必须留下。少一个条件就会把不相干的通知删掉。
func TestMessageRepoDeleteByReference(t *testing.T) {
	ctx := context.Background()
	r := memory.NewMessageRepository()
	seed := func(id, rt, rid string) {
		t.Helper()
		if _, err := r.Create(ctx, domain.Message{
			ID: id, ReceiverID: "u-1", SenderID: "sys", ResourceType: rt, ResourceID: rid,
		}); err != nil {
			t.Fatalf("seed %s: %v", id, err)
		}
	}
	seed("m-hit1", "demand", "d-1")
	seed("m-hit2", "demand_intent", "d-1")
	seed("m-other-type", "work_order", "d-1") // 同 resource_id，不同类型
	seed("m-other-id", "demand", "d-2")       // 同类型，不同 resource_id
	seed("m-broadcast", "", "")               // 管理端广播

	n, err := r.DeleteByReference(ctx, "d-1", []string{"demand", "demand_intent"})
	if err != nil {
		t.Fatalf("DeleteByReference: %v", err)
	}
	if n != 2 {
		t.Fatalf("删除条数=%d，want 2", n)
	}
	for _, keep := range []string{"m-other-type", "m-other-id", "m-broadcast"} {
		if _, err := r.FindByID(ctx, keep); err != nil {
			t.Fatalf("%s 不该被删：%v", keep, err)
		}
	}
	for _, gone := range []string{"m-hit1", "m-hit2"} {
		if _, err := r.FindByID(ctx, gone); err == nil {
			t.Fatalf("%s 应已被删除", gone)
		}
	}

	// 空条件不得删任何东西——否则会把管理端广播清光
	if n2, err := r.DeleteByReference(ctx, "", nil); err != nil || n2 != 0 {
		t.Fatalf("空条件：n=%d err=%v，want 0/nil", n2, err)
	}
	if _, err := r.FindByID(ctx, "m-broadcast"); err != nil {
		t.Fatal("空条件把管理端广播删掉了")
	}
}
