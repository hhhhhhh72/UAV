package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// TestUpdateApplicationStatusOwnershipFirst: 回归 C5——投递状态更新的归属校验必须先于落库。
// 场景：企业 A 的职位收到 user-1 的投递，无关者（ent-2）尝试改状态必须被拒绝，
// 且拒绝后状态保持原值（越权写入不得生效）。
func TestUpdateApplicationStatusOwnershipFirst(t *testing.T) {
	svc := service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository())

	// Arrange: 企业 A 建职位并发布，user-1 投递
	job, err := svc.CreateJob(context.Background(), entActor(), "飞手招聘", "描述", "重庆", 100000)
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if _, err := svc.PublishJob(context.Background(), entActor(), job.ID); err != nil {
		t.Fatalf("publish job: %v", err)
	}
	resume, err := svc.CreateResume(context.Background(), indActor(), "简历", "张三", "13800000000", "a@b.com", "本科", "经验", nil, "", "内容", "public")
	if err != nil {
		t.Fatalf("create resume: %v", err)
	}
	app, err := svc.Apply(context.Background(), indActor(), job.ID, resume.ID)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	stranger := domain.Actor{ID: "ent-2", Role: domain.RoleEnterprise}

	// Act: 无关企业尝试把投递改为"面试中"
	if _, err := svc.UpdateApplicationStatus(context.Background(), stranger, app.ID, domain.AppInterviewing); err == nil {
		t.Fatal("stranger update should be rejected")
	}

	// Assert: 状态未被越权修改
	apps, err := svc.ListMyApplications(context.Background(), indActor())
	if err != nil {
		t.Fatalf("list my apps: %v", err)
	}
	if len(apps) != 1 || apps[0].Status != domain.AppSubmitted {
		t.Fatalf("status changed by stranger: %+v", apps)
	}

	// Act/Assert: 合法操作者——职位所属企业按状态机推进（submitted→viewed→interviewing）
	if _, err := svc.UpdateApplicationStatus(context.Background(), entActor(), app.ID, domain.AppViewed); err != nil {
		t.Fatalf("job owner update viewed: %v", err)
	}
	if _, err := svc.UpdateApplicationStatus(context.Background(), entActor(), app.ID, domain.AppInterviewing); err != nil {
		t.Fatalf("job owner update interviewing: %v", err)
	}
	// 非法跳变（interviewing→withdrawn 求职者撤回不允许；submitted→offered 直接录用不允许）应被拒绝
	if _, err := svc.UpdateApplicationStatus(context.Background(), indActor(), app.ID, domain.AppWithdrawn); err == nil {
		t.Fatal("applicant withdraw after interviewing should be rejected")
	}
	apps, _ = svc.ListMyApplications(context.Background(), indActor())
	if apps[0].Status != domain.AppInterviewing {
		t.Fatalf("status after owner updates: %s", apps[0].Status)
	}
}
