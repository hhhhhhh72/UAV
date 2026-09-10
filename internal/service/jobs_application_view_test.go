package service_test

import (
	"context"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// jobFixture 我的投递富化的测试夹具：企业建职位并发布 → 求职者建简历并投递。
// ents 可选：传入企业仓库时额外建一家企业（owner_user_id = 企业账号 ID，与生产一致）。
type jobFixture struct {
	svc      *service.JobService
	jobRepo  repository.JobRepository
	ent      domain.Enterprise
	jobID    string
	resumeID string
}

func newJobFixture(t *testing.T, ents ...repository.EnterpriseRepository) *jobFixture {
	t.Helper()
	ctx := context.Background()
	jobRepo := memory.NewJobRepository()
	svc := service.NewJobService(jobRepo, memory.NewResumeRepository(), memory.NewJobApplicationRepository(), ents...)
	f := &jobFixture{svc: svc, jobRepo: jobRepo}
	if len(ents) > 0 {
		ent, err := ents[0].Create(ctx, domain.Enterprise{
			ID: "ent-real-1", OwnerUserID: entActor().ID, Name: "渝航智能科技有限公司", Status: domain.EnterpriseApproved,
		})
		if err != nil {
			t.Fatalf("create enterprise: %v", err)
		}
		f.ent = ent
	}
	job, err := svc.CreateJob(ctx, entActor(), "无人机飞手（巡检方向）", "线路巡检", "重庆·渝北", 800000, "全职")
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if _, err := svc.PublishJob(ctx, entActor(), job.ID); err != nil {
		t.Fatalf("publish job: %v", err)
	}
	resume, err := svc.CreateResume(ctx, indActor(), "简历", "张三", "13800000000", "a@b.com", "本科", "两年", nil, "", "内容", "public")
	if err != nil {
		t.Fatalf("create resume: %v", err)
	}
	if _, err := svc.Apply(ctx, indActor(), job.ID, resume.ID); err != nil {
		t.Fatalf("apply: %v", err)
	}
	f.jobID = job.ID
	f.resumeID = resume.ID
	return f
}

// 我的投递富化（求职者视角）：投递记录只存 job_id，服务端必须补齐职位快照与发布单位名称，
// 否则前端只能显示一串裸 ID（BUG-007）。
// 单位归属两种来源都要能解析：发布者用户 ID（CreateJob 以 actor.ID 归属）与企业实体 ID（管理端指定）。
func TestListMyApplicationViewsEnrichesJobAndEnterprise(t *testing.T) {
	ctx := context.Background()
	entRepo := memory.NewEnterpriseRepository(nil)
	f := newJobFixture(t, entRepo)

	views, err := f.svc.ListMyApplicationViews(ctx, indActor())
	if err != nil {
		t.Fatalf("list views: %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("expect 1 view, got %d", len(views))
	}
	v := views[0]
	// 顶层字段保持原样：既有页面（招聘大厅/职位详情）按 job_id 判断"是否已投递"
	if v.JobID != f.jobID || v.Status != domain.AppSubmitted || v.ID == "" {
		t.Fatalf("top-level fields changed: %+v", v.JobApplication)
	}
	if v.Job == nil {
		t.Fatal("job snapshot missing")
	}
	if v.Job.Title != "无人机飞手（巡检方向）" || v.Job.Location != "重庆·渝北" || v.Job.JobType != "全职" || v.Job.SalaryFen != 800000 {
		t.Fatalf("job snapshot mismatch: %+v", *v.Job)
	}
	if v.Job.Status != domain.JobPublished {
		t.Fatalf("job status: %s", v.Job.Status)
	}
	// job.enterprise_id = ent-1（发布者用户 ID）→ 经 FindByOwner 解析出单位名
	if v.EnterpriseName != "渝航智能科技有限公司" {
		t.Fatalf("enterprise name via owner: %q", v.EnterpriseName)
	}

	// 第二种来源：job.enterprise_id 直接是企业实体 ID（管理端建/转移职位）
	direct := domain.Job{ID: "job-admin-owned", EnterpriseID: f.ent.ID, Title: "无人机运维工程师",
		Location: "重庆·南岸", JobType: "实习", SalaryFen: 1000000, Status: domain.JobPublished,
		Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if _, err := f.jobRepo.Create(ctx, direct); err != nil {
		t.Fatalf("create direct job: %v", err)
	}
	if _, err := f.svc.Apply(ctx, indActor(), direct.ID, f.resumeID); err != nil {
		t.Fatalf("apply direct: %v", err)
	}
	views, err = f.svc.ListMyApplicationViews(ctx, indActor())
	if err != nil {
		t.Fatalf("list views 2: %v", err)
	}
	var found *service.MyApplicationView
	for i := range views {
		if views[i].JobID == direct.ID {
			found = &views[i]
		}
	}
	if found == nil || found.Job == nil || found.Job.Title != "无人机运维工程师" {
		t.Fatalf("direct job snapshot missing: %+v", found)
	}
	if found.EnterpriseName != "渝航智能科技有限公司" {
		t.Fatalf("enterprise name via id: %q", found.EnterpriseName)
	}

	// 职位被删除：Job 为 nil（前端显示"职位已删除"），单位名留空，且不报错
	if err := f.jobRepo.Delete(ctx, direct.ID); err != nil {
		t.Fatalf("delete job: %v", err)
	}
	views, err = f.svc.ListMyApplicationViews(ctx, indActor())
	if err != nil {
		t.Fatalf("list views after delete: %v", err)
	}
	for _, it := range views {
		if it.JobID != direct.ID {
			continue
		}
		if it.Job != nil || it.EnterpriseName != "" {
			t.Fatalf("deleted job should yield nil job and empty name: %+v", it)
		}
		if it.Status != domain.AppSubmitted {
			t.Fatalf("application status must survive missing job: %s", it.Status)
		}
	}
}

// 未注入企业仓库（3 参构造，测试/嵌入场景）时不得 panic：职位快照照常，单位名留空。
func TestListMyApplicationViewsWithoutEnterpriseRepo(t *testing.T) {
	f := newJobFixture(t)
	views, err := f.svc.ListMyApplicationViews(context.Background(), indActor())
	if err != nil {
		t.Fatalf("list views: %v", err)
	}
	if len(views) != 1 || views[0].Job == nil || views[0].Job.Title == "" {
		t.Fatalf("job snapshot missing without enterprise repo: %+v", views)
	}
	if views[0].EnterpriseName != "" {
		t.Fatalf("enterprise name should be empty without repo, got %q", views[0].EnterpriseName)
	}
}

// 空投递返回 [] 而不是 nil：响应体里的 {"data":null} 会逼每个前端调用点各自兜底。
func TestListMyApplicationViewsEmptyIsSlice(t *testing.T) {
	svc := service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository())
	views, err := svc.ListMyApplicationViews(context.Background(), indActor())
	if err != nil {
		t.Fatalf("list views: %v", err)
	}
	if views == nil || len(views) != 0 {
		t.Fatalf("expect empty non-nil slice, got %#v", views)
	}
}
