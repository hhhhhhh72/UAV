package service_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// ─────────────────────────────────────────────────────────────────────────────
// biz_modules.go — Compliance / Portfolio 补充分支
// ─────────────────────────────────────────────────────────────────────────────

func TestComplianceService_FindAndDelete(t *testing.T) {
	svc := service.NewComplianceService(memory.NewComplianceRepository())

	doc, err := svc.CreateDoc(context.Background(), "政策", "政策", "协会", "2026-01-01", "published", "摘要", "http://f", nil)
	if err != nil {
		t.Fatalf("CreateDoc: %v", err)
	}
	// FindDocByID
	if _, err := svc.FindDocByID(context.Background(), doc.ID); err != nil {
		t.Fatalf("FindDocByID: %v", err)
	}
	if _, err := svc.FindDocByID(context.Background(), "nope"); err == nil {
		t.Fatal("FindDocByID: expected error for unknown id")
	}

	std, err := svc.CreateStandard(context.Background(), "标准", "国家标准", "GB/T", "协会", "2026-01-01", "published", "范围", "http://f")
	if err != nil {
		t.Fatalf("CreateStandard: %v", err)
	}
	// FindStandardByID
	if _, err := svc.FindStandardByID(context.Background(), std.ID); err != nil {
		t.Fatalf("FindStandardByID: %v", err)
	}
	if _, err := svc.FindStandardByID(context.Background(), "nope"); err == nil {
		t.Fatal("FindStandardByID: expected error for unknown id")
	}
	// DeleteStandard
	if err := svc.DeleteStandard(context.Background(), std.ID); err != nil {
		t.Fatalf("DeleteStandard: %v", err)
	}
	if err := svc.DeleteStandard(context.Background(), "nope"); err == nil {
		t.Fatal("DeleteStandard: expected error for unknown id")
	}
}

func TestPortfolioService_ListAndDelete(t *testing.T) {
	svc := service.NewPortfolioService(memory.NewPortfolioRepository())

	p, err := svc.Create(context.Background(), "ent-1", "品牌", "logo", "cover", "描述", "联系人", nil, nil)
	if err != nil {
		t.Fatalf("PortfolioService.Create: %v", err)
	}
	list, total, err := svc.List(context.Background(), 1, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("PortfolioService.List: total=%d len=%d err=%v", total, len(list), err)
	}
	if err := svc.Delete(context.Background(), domain.Actor{ID: "ent-1", Role: domain.RoleEnterprise}, p.ID); err != nil {
		t.Fatalf("PortfolioService.Delete: %v", err)
	}
	if err := svc.Delete(context.Background(), domain.Actor{ID: "ent-1", Role: domain.RoleEnterprise}, "nope"); err == nil {
		t.Fatal("PortfolioService.Delete: expected error for unknown id")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// enterprise.go — AttachDocument / ListDocuments
// ─────────────────────────────────────────────────────────────────────────────

func TestEnterpriseSvc_AttachDocument(t *testing.T) {
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), memory.NewUserRepository(nil))
	owner := domain.Actor{ID: "ent-owner", Role: domain.RoleEnterprise}

	e, err := svc.Create(context.Background(), owner, service.CreateEnterpriseInput{Name: "测试企业"})
	if err != nil {
		t.Fatalf("EnterpriseSvc.Create: %v", err)
	}

	// not found
	if _, err := svc.AttachDocument(context.Background(), owner, "nope", "f1", "license"); err == nil {
		t.Fatal("AttachDocument: expected error for unknown enterprise")
	}
	// 归属校验：无关用户被拒
	stranger := domain.Actor{ID: "stranger", Role: domain.RoleIndividual}
	if _, err := svc.AttachDocument(context.Background(), stranger, e.ID, "f1", "license"); err == nil {
		t.Fatal("AttachDocument: expected permission denied for non-owner non-admin")
	}
	// owner 成功
	doc, err := svc.AttachDocument(context.Background(), owner, e.ID, "f1", "license")
	if err != nil {
		t.Fatalf("AttachDocument(owner): %v", err)
	}
	if doc.EnterpriseID != e.ID || doc.ReviewStatus != "pending" {
		t.Fatalf("AttachDocument: got %+v", doc)
	}
	// admin 成功
	admin := domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}
	if _, err := svc.AttachDocument(context.Background(), admin, e.ID, "f2", "idcard"); err != nil {
		t.Fatalf("AttachDocument(admin): %v", err)
	}
}

func TestEnterpriseSvc_ListDocuments(t *testing.T) {
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), memory.NewUserRepository(nil))
	owner := domain.Actor{ID: "ent-owner", Role: domain.RoleEnterprise}

	e, _ := svc.Create(context.Background(), owner, service.CreateEnterpriseInput{Name: "测试企业"})
	if _, err := svc.AttachDocument(context.Background(), owner, e.ID, "f1", "license"); err != nil {
		t.Fatalf("AttachDocument: %v", err)
	}

	// not found
	if _, err := svc.ListDocuments(context.Background(), owner, "nope"); err == nil {
		t.Fatal("ListDocuments: expected error for unknown enterprise")
	}
	// 非 owner 非 admin → 403
	if _, err := svc.ListDocuments(context.Background(), domain.Actor{ID: "other", Role: domain.RoleEnterprise}, e.ID); err == nil {
		t.Fatal("ListDocuments: expected permission denied for non-owner non-admin")
	}
	// owner 成功
	docs, err := svc.ListDocuments(context.Background(), owner, e.ID)
	if err != nil || len(docs) != 1 {
		t.Fatalf("ListDocuments(owner): len=%d err=%v", len(docs), err)
	}
	// admin 成功
	docs, err = svc.ListDocuments(context.Background(), domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, e.ID)
	if err != nil || len(docs) != 1 {
		t.Fatalf("ListDocuments(admin): len=%d err=%v", len(docs), err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// files.go — Upload / UploadPrivate / uploadTo
// ─────────────────────────────────────────────────────────────────────────────

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("simulated read failure") }

func TestFileService_UploadPaths(t *testing.T) {
	svc := service.NewFileService(t.TempDir())

	// 正常写入
	rec, err := svc.Upload(context.Background(), "owner-1", "a.txt", "text/plain", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if rec.ID == "" || rec.SizeBytes != 5 || len(rec.SHA256) != 64 || rec.OwnerID != "owner-1" {
		t.Fatalf("Upload: unexpected record %+v", rec)
	}

	// 私有目录
	rec2, err := svc.UploadPrivate(context.Background(), "owner-1", "b.txt", "text/plain", strings.NewReader("world"))
	if err != nil {
		t.Fatalf("UploadPrivate: %v", err)
	}
	if rec2.ID == "" || rec2.SizeBytes != 5 {
		t.Fatalf("UploadPrivate: unexpected record %+v", rec2)
	}

	// 写失败 → 清理已创建文件并报错
	if _, err := svc.Upload(context.Background(), "owner-1", "c.txt", "text/plain", failingReader{}); err == nil {
		t.Fatal("Upload: expected error when reader fails")
	}

	// 上传目录是文件 → MkdirAll 失败
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	svcBlock := service.NewFileService(blocker)
	if _, err := svcBlock.Upload(context.Background(), "o", "f", "text/plain", strings.NewReader("x")); err == nil {
		t.Fatal("Upload: expected error when upload dir cannot be created")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// home.go — sortByDistance / sanitizeDemand / 截断
// ─────────────────────────────────────────────────────────────────────────────

func TestHomeService_SortSanitize(t *testing.T) {
	demandRepo := memory.NewDemandRepository(nil)
	seed := func(id, title string, lat, lng float64) {
		if _, err := demandRepo.Create(context.Background(), domain.Demand{
			ID: id, Status: domain.DemandPublished, Title: title,
			PublisherID: "pub-1", Contact: "13800000000",
			Latitude: lat, Longitude: lng, CreatedAt: time.Now(),
		}); err != nil {
			t.Fatalf("seed %s: %v", id, err)
		}
	}
	seed("d-far", "远", 29.9, 106.5)
	seed("d-near", "近", 29.5, 106.5)
	seed("d-mid", "中", 29.7, 106.5)

	entRepo := memory.NewEnterpriseRepository(nil)
	if _, err := entRepo.Create(context.Background(), domain.Enterprise{ID: "ent-1", OwnerUserID: "o", Name: "商家", Status: domain.EnterpriseApproved}); err != nil {
		t.Fatal(err)
	}
	svc := service.NewHomeService(demandRepo, entRepo)

	hd := svc.GetHome(context.Background(), "重庆", 29.5, 106.5)
	if len(hd.HotDemands) != 3 {
		t.Fatalf("GetHome: len=%d, want 3", len(hd.HotDemands))
	}
	// 距离升序：近 → 中 → 远
	if hd.HotDemands[0].Title != "近" || hd.HotDemands[1].Title != "中" || hd.HotDemands[2].Title != "远" {
		t.Fatalf("GetHome: sort order = %q,%q,%q", hd.HotDemands[0].Title, hd.HotDemands[1].Title, hd.HotDemands[2].Title)
	}
	// sanitizeDemand 脱敏
	for _, d := range hd.HotDemands {
		if d.PublisherID != "" || d.Contact != "" || d.Latitude != 0 || d.Longitude != 0 {
			t.Fatalf("GetHome: sanitize failed: %+v", d)
		}
	}
	if len(hd.Shops) != 1 {
		t.Fatalf("GetHome: shops len=%d, want 1", len(hd.Shops))
	}
}

func TestHomeService_CityDefaultAndTruncate(t *testing.T) {
	demandRepo := memory.NewDemandRepository(nil)
	for i := 0; i < 12; i++ {
		if _, err := demandRepo.Create(context.Background(), domain.Demand{
			ID: fmt.Sprintf("d-%d", i), Status: domain.DemandPublished, Title: "x", CreatedAt: time.Now(),
		}); err != nil {
			t.Fatal(err)
		}
	}
	// enterpriseRepo 为 nil → 跳过商家列表分支
	svc := service.NewHomeService(demandRepo, nil)
	hd := svc.GetHome(context.Background(), "", 0, 0)
	if hd.City != "重庆" {
		t.Fatalf("GetHome: default city = %q, want 重庆", hd.City)
	}
	if len(hd.HotDemands) != 10 {
		t.Fatalf("GetHome: truncated len=%d, want 10", len(hd.HotDemands))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// insurance_finance.go — ListAllPolicies / ListAllLoans
// ─────────────────────────────────────────────────────────────────────────────

func TestInsuranceService_ListAllPolicies(t *testing.T) {
	svc := service.NewInsuranceService(memory.NewPolicyRepository(), memory.NewInspectionRepository())
	if _, err := svc.CreatePolicy(context.Background(), individualActor(), "M300", "SN", "liability", 100, 10000, time.Now(), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}
	list, total, err := svc.ListAllPolicies(context.Background(), 0, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("ListAllPolicies: total=%d len=%d err=%v", total, len(list), err)
	}
}

func TestFinanceService_ListAllLoans(t *testing.T) {
	svc := service.NewFinanceService(memory.NewLoanRepository())
	if _, err := svc.ApplyLoan(context.Background(), individualActor(), 100000, 12, "购机"); err != nil {
		t.Fatalf("ApplyLoan: %v", err)
	}
	list, total, err := svc.ListAllLoans(context.Background(), 0, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("ListAllLoans: total=%d len=%d err=%v", total, len(list), err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// jobs.go — ListAllJobs / DeleteJob / ListAllResumes / ListApplicantsForJob
// ─────────────────────────────────────────────────────────────────────────────

func TestJobService_ListAllAndDelete(t *testing.T) {
	svc := service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository())

	j, err := svc.CreateJob(context.Background(), entActor(), "飞手", "描述", "重庆", 100000)
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	list, total, err := svc.ListAllJobs(context.Background(), 0, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("ListAllJobs: total=%d len=%d err=%v", total, len(list), err)
	}
	if err := svc.DeleteJob(context.Background(), j.ID); err != nil {
		t.Fatalf("DeleteJob: %v", err)
	}
	if err := svc.DeleteJob(context.Background(), "nope"); err == nil {
		t.Fatal("DeleteJob: expected error for unknown id")
	}

	if _, err := svc.CreateResume(context.Background(), indActor(), "简历", "张三", "138", "a@b", "本科", "经验", nil, "", "内容", "public"); err != nil {
		t.Fatalf("CreateResume: %v", err)
	}
	res, total, err := svc.ListAllResumes(context.Background(), 0, 10)
	if err != nil || total != 1 || len(res) != 1 {
		t.Fatalf("ListAllResumes: total=%d len=%d err=%v", total, len(res), err)
	}
}

func TestJobService_ListApplicantsForJob(t *testing.T) {
	jobRepo := memory.NewJobRepository()
	resumeRepo := memory.NewResumeRepository()
	appRepo := memory.NewJobApplicationRepository()
	svc := service.NewJobService(jobRepo, resumeRepo, appRepo)

	j, _ := svc.CreateJob(context.Background(), entActor(), "飞手", "描述", "重庆", 100000)

	// 非职位所属企业 → 报错
	if _, err := svc.ListApplicantsForJob(context.Background(), indActor(), j.ID); err == nil {
		t.Fatal("ListApplicantsForJob: expected error for non-owner")
	}
	// 无投递 → 空
	apps, err := svc.ListApplicantsForJob(context.Background(), entActor(), j.ID)
	if err != nil || len(apps) != 0 {
		t.Fatalf("ListApplicantsForJob(empty): len=%d err=%v", len(apps), err)
	}

	// 发布 + 投递 + 简历快照
	if _, err := svc.PublishJob(context.Background(), entActor(), j.ID); err != nil {
		t.Fatalf("PublishJob: %v", err)
	}
	resume, _ := svc.CreateResume(context.Background(), indActor(), "简历", "张三", "138", "a@b", "本科", "经验", nil, "", "内容", "public")
	if _, err := svc.Apply(context.Background(), indActor(), j.ID, resume.ID); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	views, err := svc.ListApplicantsForJob(context.Background(), entActor(), j.ID)
	if err != nil || len(views) != 1 || views[0].Resume.ID != resume.ID {
		t.Fatalf("ListApplicantsForJob: len=%d err=%v", len(views), err)
	}

	// 简历已删 → 跳过该投递
	if _, err := appRepo.Create(context.Background(), domain.JobApplication{
		ID: "app-bogus", JobID: j.ID, ResumeID: "nope", ApplicantID: "user-9",
		Status: domain.AppSubmitted, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	views, err = svc.ListApplicantsForJob(context.Background(), entActor(), j.ID)
	if err != nil || len(views) != 1 {
		t.Fatalf("ListApplicantsForJob(skip missing resume): len=%d err=%v", len(views), err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// listings_labour.go — CreateAssignment / ListAssignments / ListMyAssignments
// ─────────────────────────────────────────────────────────────────────────────

func TestLabourService_Assignments(t *testing.T) {
	svc := service.NewLabourService(memory.NewLabourOrderRepository())
	employer := domain.Actor{ID: "emp-1", Role: domain.RoleEnterprise}
	worker := domain.Actor{ID: "worker-1", Role: domain.RoleIndividual}
	admin := domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}

	o, err := svc.CreateOrder(context.Background(), employer, "标题", "描述", 2, time.Now(), time.Now().Add(time.Hour), 100000)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	// CreateAssignment：订单不存在
	if _, err := svc.CreateAssignment(context.Background(), employer, "nope", worker.ID); err == nil {
		t.Fatal("CreateAssignment: expected error for unknown order")
	}
	// 归属校验：非雇主非管理员被拒
	if _, err := svc.CreateAssignment(context.Background(), worker, o.ID, worker.ID); err == nil {
		t.Fatal("CreateAssignment: expected permission denied for non-employer")
	}
	// 雇主成功
	a, err := svc.CreateAssignment(context.Background(), employer, o.ID, worker.ID)
	if err != nil || a.Status != "assigned" {
		t.Fatalf("CreateAssignment(employer): status=%q err=%v", a.Status, err)
	}
	// 管理员成功
	if _, err := svc.CreateAssignment(context.Background(), admin, o.ID, "worker-2"); err != nil {
		t.Fatalf("CreateAssignment(admin): %v", err)
	}

	// ListAssignments：订单不存在
	if _, err := svc.ListAssignments(context.Background(), employer, "nope"); err == nil {
		t.Fatal("ListAssignments: expected error for unknown order")
	}
	// 归属校验
	if _, err := svc.ListAssignments(context.Background(), worker, o.ID); err == nil {
		t.Fatal("ListAssignments: expected permission denied for non-employer")
	}
	// 成功
	list, err := svc.ListAssignments(context.Background(), employer, o.ID)
	if err != nil || len(list) != 2 {
		t.Fatalf("ListAssignments: len=%d err=%v", len(list), err)
	}
	// ListMyAssignments
	mine, err := svc.ListMyAssignments(context.Background(), worker)
	if err != nil || len(mine) != 1 {
		t.Fatalf("ListMyAssignments: len=%d err=%v", len(mine), err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// messages.go — Get / ListAll / Delete
// ─────────────────────────────────────────────────────────────────────────────

func TestMessageService_GetListAllDelete(t *testing.T) {
	svc := service.NewMessageService(memory.NewMessageRepository())

	m, err := svc.Send(context.Background(), "s", "r", "标题", "内容", "demand", "d1")
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	got, err := svc.Get(context.Background(), m.ID)
	if err != nil || got.ID != m.ID {
		t.Fatalf("Get: id=%q err=%v", got.ID, err)
	}
	if _, err := svc.Get(context.Background(), "nope"); err == nil {
		t.Fatal("Get: expected error for unknown id")
	}

	list, total, err := svc.ListAll(context.Background(), 0, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("ListAll: total=%d len=%d err=%v", total, len(list), err)
	}

	if err := svc.Delete(context.Background(), m.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := svc.Delete(context.Background(), "nope"); err == nil {
		t.Fatal("Delete: expected error for unknown id")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// phase3.go — TradeOrderService 售后 / 管理端 / 删除 / 列表 / 详情
// ─────────────────────────────────────────────────────────────────────────────

func TestTradeOrderService_AftersaleFlow(t *testing.T) {
	svc := service.NewTradeOrderService(memory.NewTradeOrderRepository())

	o, err := svc.Create(context.Background(), "buyer-1", "p1", "seller-1", 100000)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// 订单不存在
	if _, err := svc.ApplyAftersale(context.Background(), "buyer-1", "nope", "refund", "理由", "描述", 100000); err == nil {
		t.Fatal("ApplyAftersale: expected error for unknown order")
	}
	// 非买家 → 拒
	if _, err := svc.ApplyAftersale(context.Background(), "seller-1", o.ID, "refund", "", "", 100000); err == nil {
		t.Fatal("ApplyAftersale: expected permission denied for seller")
	}
	// pending → aftersale 非法流转
	if _, err := svc.ApplyAftersale(context.Background(), "buyer-1", o.ID, "refund", "", "", 100000); err == nil {
		t.Fatal("ApplyAftersale: expected invalid transition from pending")
	}
	// pending → paid
	if _, err := svc.UpdateStatus(context.Background(), o.ID, "buyer-1", "paid"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	// 售后申请成功
	ao, err := svc.ApplyAftersale(context.Background(), "buyer-1", o.ID, "refund", "理由", "描述", 100000)
	if err != nil || ao.Status != "aftersale" || ao.AftersaleStatus != "pending" {
		t.Fatalf("ApplyAftersale: status=%q aftersale=%q err=%v", ao.Status, ao.AftersaleStatus, err)
	}
	// 重复申请
	if _, err := svc.ApplyAftersale(context.Background(), "buyer-1", o.ID, "refund", "", "", 100000); err == nil {
		t.Fatal("ApplyAftersale: expected duplicate aftersale error")
	}
	// ReviewAftersale 订单不存在
	if _, err := svc.ReviewAftersale(context.Background(), "nope", true); err == nil {
		t.Fatal("ReviewAftersale: expected error for unknown order")
	}
	// 同意
	ro, err := svc.ReviewAftersale(context.Background(), o.ID, true)
	if err != nil || ro.AftersaleStatus != "approved" || ro.Status != "completed" {
		t.Fatalf("ReviewAftersale(approve): aftersale=%q status=%q err=%v", ro.AftersaleStatus, ro.Status, err)
	}
	// 已结案 → 再次审核报错
	if _, err := svc.ReviewAftersale(context.Background(), o.ID, true); err == nil {
		t.Fatal("ReviewAftersale: expected error when not pending")
	}

	// 驳回分支
	o2, _ := svc.Create(context.Background(), "buyer-2", "p2", "seller-2", 200000)
	svc.UpdateStatus(context.Background(), o2.ID, "buyer-2", "paid")
	if _, err := svc.ApplyAftersale(context.Background(), "buyer-2", o2.ID, "return", "理由", "描述", 200000); err != nil {
		t.Fatalf("ApplyAftersale(2nd): %v", err)
	}
	rj, err := svc.ReviewAftersale(context.Background(), o2.ID, false)
	if err != nil || rj.AftersaleStatus != "rejected" {
		t.Fatalf("ReviewAftersale(reject): aftersale=%q err=%v", rj.AftersaleStatus, err)
	}
}

func TestTradeOrderService_AdminDeleteListFind(t *testing.T) {
	svc := service.NewTradeOrderService(memory.NewTradeOrderRepository())

	o, _ := svc.Create(context.Background(), "buyer-1", "p1", "seller-1", 100000)
	// UpdateStatusAdmin：订单不存在
	if _, err := svc.UpdateStatusAdmin(context.Background(), "nope", "paid"); err == nil {
		t.Fatal("UpdateStatusAdmin: expected error for unknown order")
	}
	// 非法流转 pending → shipped
	if _, err := svc.UpdateStatusAdmin(context.Background(), o.ID, "shipped"); err == nil {
		t.Fatal("UpdateStatusAdmin: expected invalid transition")
	}
	// 合法流转
	if _, err := svc.UpdateStatusAdmin(context.Background(), o.ID, "paid"); err != nil {
		t.Fatalf("UpdateStatusAdmin: %v", err)
	}
	// FindByID
	if _, err := svc.FindByID(context.Background(), o.ID); err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if _, err := svc.FindByID(context.Background(), "nope"); err == nil {
		t.Fatal("FindByID: expected error for unknown id")
	}
	// ListAll
	list, total, err := svc.ListAll(context.Background(), 0, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("ListAll: total=%d len=%d err=%v", total, len(list), err)
	}
	// Delete
	if err := svc.Delete(context.Background(), o.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := svc.Delete(context.Background(), "nope"); err == nil {
		t.Fatal("Delete: expected error for unknown id")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// service_listing.go — 服务能力展示全分支
// ─────────────────────────────────────────────────────────────────────────────

func TestServiceListingService_CRUD(t *testing.T) {
	repo := memory.NewServiceListingRepository()
	svc := service.NewServiceListingService(repo)

	sl, err := svc.CreateListing(context.Background(), "prov-1", "服务商", "巡检服务", "巡检", "描述", "重庆", 10000, "次", "img")
	if err != nil || sl.Status != "published" {
		t.Fatalf("CreateListing: status=%q err=%v", sl.Status, err)
	}
	// 空状态（视为上架）+ 下架 用于 ListPublished 过滤
	if _, err := repo.Create(context.Background(), domain.ServiceListing{ID: "sl-empty", Title: "空状态", Status: ""}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Create(context.Background(), domain.ServiceListing{ID: "sl-offline", Title: "下架", Status: "offline"}); err != nil {
		t.Fatal(err)
	}

	lp, err := svc.ListPublished(context.Background())
	if err != nil || len(lp) != 2 {
		t.Fatalf("ListPublished: len=%d err=%v, want 2 (published + empty)", len(lp), err)
	}

	if _, err := svc.Get(context.Background(), sl.ID); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if _, err := svc.Get(context.Background(), "nope"); err == nil {
		t.Fatal("Get: expected error for unknown id")
	}

	// ListAdmin 关键词/分类过滤
	if la, _ := svc.ListAdmin(context.Background(), "巡检", ""); len(la) != 1 {
		t.Fatalf("ListAdmin(keyword): len=%d, want 1", len(la))
	}
	if la, _ := svc.ListAdmin(context.Background(), "", "巡检"); len(la) != 1 {
		t.Fatalf("ListAdmin(category): len=%d, want 1", len(la))
	}
	if la, _ := svc.ListAdmin(context.Background(), "不存在", ""); len(la) != 0 {
		t.Fatalf("ListAdmin(no keyword hit): len=%d, want 0", len(la))
	}
	if la, _ := svc.ListAdmin(context.Background(), "", "不匹配"); len(la) != 0 {
		t.Fatalf("ListAdmin(no category hit): len=%d, want 0", len(la))
	}

	sl.Title = "更新"
	up, err := svc.UpdateListing(context.Background(), sl)
	if err != nil || up.Title != "更新" {
		t.Fatalf("UpdateListing: title=%q err=%v", up.Title, err)
	}

	if err := svc.DeleteListing(context.Background(), sl.ID); err != nil {
		t.Fatalf("DeleteListing: %v", err)
	}
	if err := svc.DeleteListing(context.Background(), "nope"); err == nil {
		t.Fatal("DeleteListing: expected error for unknown id")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// services.go — Demand 管理端操作 / Enterprise 薄封装 / 合同模板
// ─────────────────────────────────────────────────────────────────────────────

func TestDemandService_AdminOps(t *testing.T) {
	svc := service.NewDemandService(memory.NewDemandRepository(nil))
	pub := enterpriseActor()
	admin := platformAdminActor()

	// ListAll 管理端全量
	all, err := svc.ListAll(context.Background(), repository.DemandFilter{})
	if err != nil || len(all) == 0 {
		t.Fatalf("ListAll: len=%d err=%v", len(all), err)
	}

	d, err := svc.Create(context.Background(), pub, service.CreateDemandInput{PublisherName: "企业", Contact: "138", Title: "管理端操作需求"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// ListByPublisher
	mine, err := svc.ListByPublisher(context.Background(), pub.ID)
	if err != nil || len(mine) != 1 {
		t.Fatalf("ListByPublisher: len=%d err=%v", len(mine), err)
	}

	// CloseByAdmin：非管理员被拒
	if _, err := svc.CloseByAdmin(context.Background(), pub, d.ID, "reason"); err == nil {
		t.Fatal("CloseByAdmin: expected error for non-admin")
	}
	// 未公开（pending）不能关闭
	if _, err := svc.CloseByAdmin(context.Background(), admin, d.ID, "reason"); err == nil {
		t.Fatal("CloseByAdmin: expected error for unpublished demand")
	}
	// 发布后关闭成功
	if _, err := svc.Approve(context.Background(), admin, d.ID); err != nil {
		t.Fatalf("Approve: %v", err)
	}
	closed, err := svc.CloseByAdmin(context.Background(), admin, d.ID, "reason")
	if err != nil || closed.Status != domain.DemandCancelled {
		t.Fatalf("CloseByAdmin: status=%q err=%v", closed.Status, err)
	}

	// SetOfflineAmount：非管理员 / 负数 / 非公开状态 / 成功
	if _, err := svc.SetOfflineAmount(context.Background(), pub, d.ID, 100); err == nil {
		t.Fatal("SetOfflineAmount: expected error for non-admin")
	}
	if _, err := svc.SetOfflineAmount(context.Background(), admin, d.ID, -1); err == nil {
		t.Fatal("SetOfflineAmount: expected error for negative amount")
	}
	d2, _ := svc.Create(context.Background(), pub, service.CreateDemandInput{PublisherName: "企业", Contact: "138", Title: "成交登记需求"})
	if _, err := svc.SetOfflineAmount(context.Background(), admin, d2.ID, 100); err == nil {
		t.Fatal("SetOfflineAmount: expected error for unpublished demand")
	}
	if _, err := svc.Approve(context.Background(), admin, d2.ID); err != nil {
		t.Fatalf("Approve d2: %v", err)
	}
	set, err := svc.SetOfflineAmount(context.Background(), admin, d2.ID, 50000)
	if err != nil || set.OfflineAmountFen != 50000 {
		t.Fatalf("SetOfflineAmount: amount=%d err=%v", set.OfflineAmountFen, err)
	}

	// Delete：非管理员 / 已公开不可删 / 不存在 / 已取消可删
	if err := svc.Delete(context.Background(), pub, d2.ID); err == nil {
		t.Fatal("Delete: expected error for non-admin")
	}
	if err := svc.Delete(context.Background(), admin, d2.ID); err == nil {
		t.Fatal("Delete: expected error for published demand")
	}
	if err := svc.Delete(context.Background(), admin, "nope"); err == nil {
		t.Fatal("Delete: expected error for unknown demand")
	}
	d3, _ := svc.Create(context.Background(), pub, service.CreateDemandInput{PublisherName: "企业", Contact: "138", Title: "可删除需求"})
	if _, err := svc.Cancel(context.Background(), pub, d3.ID); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if err := svc.Delete(context.Background(), admin, d3.ID); err != nil {
		t.Fatalf("Delete(cancelled): %v", err)
	}
}

func TestEnterpriseService_CRUD(t *testing.T) {
	svc := service.NewEnterpriseService(memory.NewEnterpriseRepository(nil))

	e, err := svc.Create(context.Background(), domain.Enterprise{ID: "ent-x", Name: "企业", Status: domain.EnterpriseDraft})
	if err != nil || e.ID != "ent-x" {
		t.Fatalf("EnterpriseService.Create: id=%q err=%v", e.ID, err)
	}
	list, total, err := svc.ListByStatus(context.Background(), "", 0, 10)
	if err != nil || total == 0 || len(list) == 0 {
		t.Fatalf("EnterpriseService.ListByStatus: total=%d len=%d err=%v", total, len(list), err)
	}
	e.Name = "更新"
	up, err := svc.Update(context.Background(), e.ID, e)
	if err != nil || up.Name != "更新" {
		t.Fatalf("EnterpriseService.Update: name=%q err=%v", up.Name, err)
	}
	if err := svc.Delete(context.Background(), e.ID); err != nil {
		t.Fatalf("EnterpriseService.Delete: %v", err)
	}
}

func TestContractTemplateService_List(t *testing.T) {
	svc := service.NewContractTemplateService(memory.NewContractTemplateRepository())
	list, err := svc.List(context.Background())
	if err != nil || len(list) == 0 {
		t.Fatalf("ContractTemplateService.List: len=%d err=%v", len(list), err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// trading.go — 商品详情/更新/删除 + 维修全量列表
// ─────────────────────────────────────────────────────────────────────────────

func TestTradingService_ProductCRUD(t *testing.T) {
	svc := service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository())

	p, err := svc.CreateProduct(context.Background(), individualActor(), domain.ProductDrone, "无人机", "描述", "品牌", "型号", "new", 100000, nil)
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	// GetProduct
	if _, err := svc.GetProduct(context.Background(), p.ID); err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if _, err := svc.GetProduct(context.Background(), "nope"); err == nil {
		t.Fatal("GetProduct: expected error for unknown id")
	}
	// GetProductAndCountView
	v, err := svc.GetProductAndCountView(context.Background(), p.ID)
	if err != nil || v.Views != 1 {
		t.Fatalf("GetProductAndCountView: views=%d err=%v", v.Views, err)
	}
	if _, err := svc.GetProductAndCountView(context.Background(), "nope"); err == nil {
		t.Fatal("GetProductAndCountView: expected error for unknown id")
	}
	// UpdateProduct
	p.Title = "更新"
	up, err := svc.UpdateProduct(context.Background(), p)
	if err != nil || up.Title != "更新" {
		t.Fatalf("UpdateProduct: title=%q err=%v", up.Title, err)
	}
	// DeleteProduct
	if err := svc.DeleteProduct(context.Background(), p.ID); err != nil {
		t.Fatalf("DeleteProduct: %v", err)
	}
	if err := svc.DeleteProduct(context.Background(), "nope"); err == nil {
		t.Fatal("DeleteProduct: expected error for unknown id")
	}
	// ListAllRepairs
	if _, err := svc.CreateRepair(context.Background(), individualActor(), "无人机", "故障"); err != nil {
		t.Fatalf("CreateRepair: %v", err)
	}
	list, total, err := svc.ListAllRepairs(context.Background(), 0, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("ListAllRepairs: total=%d len=%d err=%v", total, len(list), err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// training.go — 课程/证书 CRUD + 飞手认证全分支 + certTypeName
// ─────────────────────────────────────────────────────────────────────────────

func TestTrainingService_CourseCertCRUD(t *testing.T) {
	svc := service.NewTrainingService(
		memory.NewCertificateRepository(),
		memory.NewCourseRepository(),
		memory.NewInstructorRepository(),
		memory.NewPilotRepository(nil),
	)
	actor := individualActor()

	c, err := svc.AddCertificate(context.Background(), actor, domain.CertCAAC, "编号", "等级", "机构", time.Now(), time.Now().AddDate(1, 0, 0))
	if err != nil {
		t.Fatalf("AddCertificate: %v", err)
	}
	// GetCert
	if _, err := svc.GetCert(context.Background(), c.ID); err != nil {
		t.Fatalf("GetCert: %v", err)
	}
	if _, err := svc.GetCert(context.Background(), "nope"); err == nil {
		t.Fatal("GetCert: expected error for unknown id")
	}
	// UpdateCertificate
	up, err := svc.UpdateCertificate(context.Background(), c.ID, "caac", "新编号", "新等级", "新机构", "approved", time.Now(), time.Now().AddDate(1, 0, 0))
	if err != nil || up.Status != "approved" {
		t.Fatalf("UpdateCertificate: status=%q err=%v", up.Status, err)
	}
	if _, err := svc.UpdateCertificate(context.Background(), "nope", "", "", "", "", "", time.Now(), time.Now()); err == nil {
		t.Fatal("UpdateCertificate: expected error for unknown id")
	}
	// DeleteCertificate
	if err := svc.DeleteCertificate(context.Background(), c.ID); err != nil {
		t.Fatalf("DeleteCertificate: %v", err)
	}
	if err := svc.DeleteCertificate(context.Background(), "nope"); err == nil {
		t.Fatal("DeleteCertificate: expected error for unknown id")
	}

	// GetCourse
	course, err := svc.CreateCourse(context.Background(), actor, domain.TrainingCourse{Title: "课程"})
	if err != nil {
		t.Fatalf("CreateCourse: %v", err)
	}
	if _, err := svc.GetCourse(context.Background(), course.ID); err != nil {
		t.Fatalf("GetCourse: %v", err)
	}
	if _, err := svc.GetCourse(context.Background(), "nope"); err == nil {
		t.Fatal("GetCourse: expected error for unknown id")
	}
	// UpdateCourse
	course.Title = "更新课程"
	upc, err := svc.UpdateCourse(context.Background(), course)
	if err != nil || upc.Title != "更新课程" {
		t.Fatalf("UpdateCourse: title=%q err=%v", upc.Title, err)
	}
	if _, err := svc.UpdateCourse(context.Background(), domain.TrainingCourse{ID: "nope"}); err == nil {
		t.Fatal("UpdateCourse: expected error for unknown id")
	}
	// DeleteCourse
	if err := svc.DeleteCourse(context.Background(), course.ID); err != nil {
		t.Fatalf("DeleteCourse: %v", err)
	}
	if err := svc.DeleteCourse(context.Background(), "nope"); err == nil {
		t.Fatal("DeleteCourse: expected error for unknown id")
	}
}

func TestTrainingService_PilotLifecycle(t *testing.T) {
	certRepo := memory.NewCertificateRepository()
	pilotRepo := memory.NewPilotRepository(nil)
	svc := service.NewTrainingService(certRepo, memory.NewCourseRepository(), memory.NewInstructorRepository(), pilotRepo)

	actor := domain.Actor{ID: "pilot-user", Role: domain.RoleIndividual}
	admin := domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}

	// 4 种证书类型全部已审核通过（覆盖 certTypeName 全分支 + 自动关联）
	// 直接种子仓储，避免 AddCertificate 的 UnixNano ID 碰撞。
	types := []domain.CertType{domain.CertCAAC, domain.CertUTCDJI, domain.CertGovLevel, domain.CertType("custom")}
	for i, ct := range types {
		if _, err := certRepo.Create(context.Background(), domain.Certificate{
			ID: fmt.Sprintf("cert-%d", i), UserID: actor.ID, CertType: ct,
			CertNumber: "n", Level: "A", IssuerOrg: "机构", Status: "approved",
			IssueDate: time.Now(), ExpireDate: time.Now().AddDate(1, 0, 0),
		}); err != nil {
			t.Fatalf("seed cert %s: %v", ct, err)
		}
	}

	// 首次申请 → pending，自动关联 4 张已认证证书
	p, err := svc.RegisterPilot(context.Background(), actor, "张三", "id", 100, "bio", "avatar", "重庆")
	if err != nil || p.Status != "pending" || len(p.CertIDs) != 4 {
		t.Fatalf("RegisterPilot(fresh): status=%q certs=%d err=%v", p.Status, len(p.CertIDs), err)
	}
	// 审核中重复申请
	if _, err := svc.RegisterPilot(context.Background(), actor, "张三", "id", 100, "", "", ""); err == nil {
		t.Fatal("RegisterPilot: expected error when pending")
	}

	// RejectPilot：非管理员被拒
	if _, err := svc.RejectPilot(context.Background(), actor, p.ID); err == nil {
		t.Fatal("RejectPilot: expected error for non-admin")
	}
	// 通过认证
	if _, err := svc.ApprovePilot(context.Background(), admin, p.ID); err != nil {
		t.Fatalf("ApprovePilot: %v", err)
	}
	// 已通过重复申请
	if _, err := svc.RegisterPilot(context.Background(), actor, "张三", "id", 100, "", "", ""); err == nil {
		t.Fatal("RegisterPilot: expected error when approved")
	}
	// 驳回 → rejected
	rp, err := svc.RejectPilot(context.Background(), admin, p.ID)
	if err != nil || rp.Status != "rejected" {
		t.Fatalf("RejectPilot: status=%q err=%v", rp.Status, err)
	}

	// GetPilot
	if _, err := svc.GetPilot(context.Background(), p.ID); err != nil {
		t.Fatalf("GetPilot: %v", err)
	}

	// 被驳回后重新申请 → 覆盖重提为 pending（Update 路径）
	r2, err := svc.RegisterPilot(context.Background(), actor, "张三2", "id2", 200, "bio2", "avatar2", "北京")
	if err != nil || r2.Status != "pending" || r2.RealName != "张三2" {
		t.Fatalf("RegisterPilot(resubmit): status=%q name=%q err=%v", r2.Status, r2.RealName, err)
	}

	// GetPilotByOwner：命中 + 未申请返回零值
	gp, err := svc.GetPilotByOwner(context.Background(), actor.ID)
	if err != nil || gp.ID != p.ID {
		t.Fatalf("GetPilotByOwner: id=%q err=%v", gp.ID, err)
	}
	np, err := svc.GetPilotByOwner(context.Background(), "nope")
	if err != nil || np.ID != "" {
		t.Fatalf("GetPilotByOwner(nope): id=%q err=%v", np.ID, err)
	}

	// GetPilotDetail：证书明细（certTypeName 覆盖）
	detail, err := svc.GetPilotDetail(context.Background(), p.ID)
	if err != nil || len(detail.Certificates) != 4 {
		t.Fatalf("GetPilotDetail: certs=%d err=%v", len(detail.Certificates), err)
	}
	names := map[string]bool{}
	for _, cb := range detail.Certificates {
		names[cb.CertName] = true
	}
	if !names["CAAC无人机驾驶员执照"] || !names["DJI UTC 植保无人机驾驶证"] || !names["政府职业技能等级证书"] || !names["custom"] {
		t.Fatalf("GetPilotDetail: certTypeName coverage missing: %v", names)
	}
	if _, err := svc.GetPilotDetail(context.Background(), "nope"); err == nil {
		t.Fatal("GetPilotDetail: expected error for unknown id")
	}
}
