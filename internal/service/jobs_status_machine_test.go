package service_test

import (
	"context"
	"errors"
	"testing"

	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

func newJobSvc() *service.JobService {
	return service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository())
}

// TestUpdateJobStatusWhitelist: 回归 HIGH——管理端更新职位状态必须落在
// draft/published/closed 白名单内，旧实现任意字符串直写。
func TestUpdateJobStatusWhitelist(t *testing.T) {
	svc := newJobSvc()
	j, err := svc.CreateJob(context.Background(), entActor(), "飞手招聘", "描述", "重庆", 100000)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"draft", "published", "closed"} {
		if _, err := svc.UpdateJob(context.Background(), j.ID, "飞手招聘", "描述", "重庆", "全职", 100000, s); err != nil {
			t.Errorf("UpdateJob(%q): %v, want ok", s, err)
		}
	}
	for _, s := range []string{"", "banana", "active", "OPEN"} {
		if _, err := svc.UpdateJob(context.Background(), j.ID, "飞手招聘", "描述", "重庆", "全职", 100000, s); !errors.Is(err, service.ErrInvalidJobStatus) {
			t.Errorf("UpdateJob(%q): err=%v, want ErrInvalidJobStatus", s, err)
		}
	}
	// 非法状态不得落库
	got, _ := svc.GetJob(context.Background(), j.ID)
	if got.Status != "closed" {
		t.Errorf("status after rejected update: %q, want closed (unchanged)", got.Status)
	}
}

// TestPublishCloseStateGuards: 回归 HIGH——发布/关闭必须走状态机：
// draft→published→closed，重复或跳步一律拒绝。
func TestPublishCloseStateGuards(t *testing.T) {
	svc := newJobSvc()
	j, _ := svc.CreateJob(context.Background(), entActor(), "飞手招聘", "描述", "重庆", 100000)

	// 草稿不能关闭
	if _, err := svc.CloseJob(context.Background(), entActor(), j.ID); !errors.Is(err, service.ErrInvalidJobTransition) {
		t.Fatalf("close draft: err=%v, want ErrInvalidJobTransition", err)
	}
	// draft → published
	if _, err := svc.PublishJob(context.Background(), entActor(), j.ID); err != nil {
		t.Fatalf("publish draft: %v", err)
	}
	// 重复发布拒绝
	if _, err := svc.PublishJob(context.Background(), entActor(), j.ID); !errors.Is(err, service.ErrInvalidJobTransition) {
		t.Fatalf("re-publish: err=%v, want ErrInvalidJobTransition", err)
	}
	// published → closed
	if _, err := svc.CloseJob(context.Background(), entActor(), j.ID); err != nil {
		t.Fatalf("close published: %v", err)
	}
	// 已关闭不可再发布/关闭
	if _, err := svc.PublishJob(context.Background(), entActor(), j.ID); !errors.Is(err, service.ErrInvalidJobTransition) {
		t.Fatalf("publish closed: err=%v, want ErrInvalidJobTransition", err)
	}
	if _, err := svc.CloseJob(context.Background(), entActor(), j.ID); !errors.Is(err, service.ErrInvalidJobTransition) {
		t.Fatalf("re-close: err=%v, want ErrInvalidJobTransition", err)
	}
}

// TestApplyRejectsNonPublishedJobs: 回归 HIGH——草稿/已关闭职位不可投递，
// 旧实现只拦"投自己的职位"。
func TestApplyRejectsNonPublishedJobs(t *testing.T) {
	svc := newJobSvc()
	draft, _ := svc.CreateJob(context.Background(), entActor(), "草稿职位", "描述", "重庆", 100000)
	r, _ := svc.CreateResume(context.Background(), indActor(), "简历", "张三", "138", "a@b.com", "本科", "", nil, "", "", "public")

	// 草稿不可投递
	if _, err := svc.Apply(context.Background(), indActor(), draft.ID, r.ID); err == nil {
		t.Fatal("apply to draft job accepted")
	}
	// 招聘中可投递
	if _, err := svc.PublishJob(context.Background(), entActor(), draft.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Apply(context.Background(), indActor(), draft.ID, r.ID); err != nil {
		t.Fatalf("apply to published job: %v", err)
	}
	// 关闭后新投递被拒
	svc.CloseJob(context.Background(), entActor(), draft.ID)
	if _, err := svc.Apply(context.Background(), indActor(), draft.ID, r.ID); err == nil {
		t.Fatal("apply to closed job accepted")
	}
	// 已关闭但未撤回的旧投递仍存在，不影响状态机
	apps, _ := svc.ListMyApplications(context.Background(), indActor())
	if len(apps) != 1 {
		t.Fatalf("existing applications: %d, want 1", len(apps))
	}
}
