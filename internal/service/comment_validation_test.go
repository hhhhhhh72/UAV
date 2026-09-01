package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 评论校验回归：空内容拒绝、不存在的帖子拒绝（防孤儿评论）。
func TestCommentValidation(t *testing.T) {
	svc := service.NewCommunityService(memory.NewPostRepository(), memory.NewCommentRepository(), memory.NewReportRepository())
	actor := domain.Actor{ID: "u-1", Role: domain.RoleIndividual}

	// 空内容 → 拒绝
	if _, err := svc.CreateComment(context.Background(), actor, "post-1", "   "); err == nil {
		t.Fatal("empty comment must be rejected")
	}
	// 不存在的帖子 → 拒绝
	if _, err := svc.CreateComment(context.Background(), actor, "post-nope", "内容"); err == nil {
		t.Fatal("comment on missing post must be rejected")
	}
	// 正常评论 → 成功：帖子须先经管理端发布（评论仅限已发布帖——审计 P3 修复）
	p, err := svc.CreatePost(context.Background(), actor, "标题", "内容", nil)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}
	admin := domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}
	// 未发布的帖子禁止评论（审计 P3：仅允许对已发布帖评论）
	if _, err := svc.CreateComment(context.Background(), actor, p.ID, "评论未发布帖"); err == nil {
		t.Fatal("comment on pending post must be rejected")
	}
	if _, err := svc.PublishPost(context.Background(), admin, p.ID); err != nil {
		t.Fatalf("publish post: %v", err)
	}
	c, err := svc.CreateComment(context.Background(), actor, p.ID, "正常评论")
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	if c.Status != "active" {
		t.Fatalf("comment status: %q", c.Status)
	}
}
