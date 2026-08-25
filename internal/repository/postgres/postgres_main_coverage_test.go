package postgres_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/repository/postgres"
)

// 本文件补齐 internal/repository/postgres/postgres.go 主文件中 0% 覆盖的仓储实现。
// 复用 postgres_test.go 的 setupStore/ug/migrateOnce 基建，所有写入用 ug("cov-") 唯一
// 前缀隔离，不污染既有测试/开发数据。断言失败统一输出方法名 + err。

// containsByID 泛型辅助：判断 items 中是否存在 get(it)==id 的元素。
func containsByID[T any](items []T, id string, get func(T) string) bool {
	for _, it := range items {
		if get(it) == id {
			return true
		}
	}
	return false
}

// setupCipherStore 构造带 AES-256-GCM cipher 的 store，用于覆盖 userRepo 中
// PhoneCipher 加密/解密的 cipher 分支（UpdateProfile 加密落库，Find 解密回明文）。
func setupCipherStore(t *testing.T) *postgres.Store {
	t.Helper()
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	cipher, err := crypto.NewCipher(key)
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	store, err := postgres.NewStore(ctx, databaseURL(), cipher)
	if err != nil {
		t.Skipf("no PG: %v", err)
		return nil
	}
	migrateOnce.Do(func() {
		migrateErr = store.RunMigrationsFromDir(ctx, postgres.MigrationsDir())
	})
	if migrateErr != nil {
		t.Fatalf("migration: %v", migrateErr)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

// ── ContractTemplateRepository ──

func TestCoverage_ContractTemplateRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewContractTemplateRepository()
	list, err := r.List(context.Background())
	if err != nil {
		t.Fatalf("List contract templates: %v", err)
	}
	if !containsByID(list, "tpl-001", func(v domain.ContractTemplate) string { return v.ID }) {
		t.Fatalf("List: seed template tpl-001 missing (%d rows)", len(list))
	}
	id := ug("cov-tpl")
	created, err := r.Create(context.Background(), domain.ContractTemplate{ID: id, Name: "覆盖率模板", Version: 1, Content: "合同正文", Status: "active"})
	if err != nil {
		t.Fatalf("Create contract template: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: template id mismatch: got %q want %q", created.ID, id)
	}
	list2, err := r.List(context.Background())
	if err != nil {
		t.Fatalf("List after create: %v", err)
	}
	if !containsByID(list2, id, func(v domain.ContractTemplate) string { return v.ID }) {
		t.Fatalf("List: created template %s not found", id)
	}
}

// ── ContractRepository ──

func TestCoverage_ContractRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewContractRepository()
	eid := ug("cov-contract-ent")
	id := ug("cov-contract")
	created, err := r.Create(context.Background(), domain.Contract{ID: id, EnterpriseID: eid, TemplateID: "tpl-001", Status: domain.ContractDraft})
	if err != nil {
		t.Fatalf("Create contract: %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("Create: contract version = %d, want 1", created.Version)
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || got.Status != domain.ContractDraft {
		t.Fatalf("FindByID: contract mismatch: id=%q status=%q", got.ID, got.Status)
	}
	if l, total, err := r.ListByEnterprise(context.Background(), eid, 0, 20); err != nil {
		t.Fatalf("ListByEnterprise: %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.Contract) string { return v.ID }) {
		t.Fatalf("ListByEnterprise: contract %s not found (total=%d)", id, total)
	}
	if l, total, err := r.ListAll(context.Background(), 0, 20); err != nil {
		t.Fatalf("ListAll: %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.Contract) string { return v.ID }) {
		t.Fatalf("ListAll: contract %s not found (total=%d)", id, total)
	}
	upd, err := r.UpdateStatus(context.Background(), id, domain.ContractSent)
	if err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if upd.Status != domain.ContractSent {
		t.Fatalf("UpdateStatus: status = %q, want %q", upd.Status, domain.ContractSent)
	}
	if _, err := r.UpdateStatus(context.Background(), ug("cov-contract-missing"), domain.ContractSent); err == nil {
		t.Fatalf("UpdateStatus: expected not-found error on missing contract")
	}
	if _, err := r.FindByID(context.Background(), ug("cov-contract-missing")); err == nil {
		t.Fatalf("FindByID: expected not-found error on missing contract")
	}
}

// ── EmploymentRepository ──

func TestCoverage_EmploymentRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewEmploymentRepository()
	eid := ug("cov-emp-ent")
	id := ug("cov-emp")
	created, err := r.Create(context.Background(), domain.EmploymentRequest{ID: id, EnterpriseID: eid, Position: "飞手", Headcount: 3, Status: domain.EmploymentPending})
	if err != nil {
		t.Fatalf("Create employment: %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("Create: employment version = %d, want 1", created.Version)
	}
	if l, total, err := r.ListByEnterprise(context.Background(), eid, 0, 20); err != nil {
		t.Fatalf("ListByEnterprise: %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.EmploymentRequest) string { return v.ID }) {
		t.Fatalf("ListByEnterprise: employment %s not found (total=%d)", id, total)
	}
	if l, total, err := r.ListAll(context.Background(), 0, 20); err != nil {
		t.Fatalf("ListAll: %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.EmploymentRequest) string { return v.ID }) {
		t.Fatalf("ListAll: employment %s not found (total=%d)", id, total)
	}
}

// ── ResumeRepository ──

func TestCoverage_ResumeRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewResumeRepository()
	uid := ug("cov-resume-user")
	id := ug("cov-resume")
	in := domain.Resume{
		ID: id, UserID: uid, Title: "飞手简历", Name: "李四", Phone: "13800000000",
		Email: "l@test.cn", Education: "本科", WorkExperience: "2年",
		Skills: []string{"航拍"}, Visibility: "public",
	}
	if _, err := r.Create(context.Background(), in); err != nil {
		t.Fatalf("Create resume: %v", err)
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || len(got.Skills) != 1 || got.Skills[0] != "航拍" {
		t.Fatalf("FindByID: resume roundtrip mismatch: %+v", got)
	}
	got.Title = "更新简历"
	got.Skills = []string{"航拍", "测绘"}
	upd, err := r.Update(context.Background(), id, got)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if upd.Version != 2 || upd.Title != "更新简历" {
		t.Fatalf("Update: resume mismatch: %+v", upd)
	}
	if l, err := r.ListByUser(context.Background(), uid); err != nil {
		t.Fatalf("ListByUser: %v", err)
	} else if !containsByID(l, id, func(v domain.Resume) string { return v.ID }) {
		t.Fatalf("ListByUser: resume %s not found", id)
	}
	if l, total, err := r.ListAll(context.Background(), 0, 20); err != nil {
		t.Fatalf("ListAll: %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.Resume) string { return v.ID }) {
		t.Fatalf("ListAll: resume %s not found (total=%d)", id, total)
	}
	if _, err := r.Update(context.Background(), ug("cov-resume-missing"), got); err == nil {
		t.Fatalf("Update: expected not-found error on missing resume")
	}
}

// ── JobApplicationRepository ──

func TestCoverage_JobApplicationRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	userID := ug("cov-app-user")
	if _, err := store.NewUserRepository().Create(context.Background(), domain.User{
		ID: userID, WechatOpenID: ug("cov-app-openid"), Name: "求职者", Status: "active",
	}); err != nil {
		t.Fatalf("create parent user: %v", err)
	}
	jobID := ug("cov-app-job")
	if _, err := store.NewJobRepository().Create(context.Background(), domain.Job{
		ID: jobID, EnterpriseID: ug("cov-app-ent"), Title: "飞手", JobType: "full-time", Status: domain.JobPublished,
	}); err != nil {
		t.Fatalf("create parent job: %v", err)
	}
	resumeID := ug("cov-app-resume")
	if _, err := store.NewResumeRepository().Create(context.Background(), domain.Resume{
		ID: resumeID, UserID: userID, Title: "求职简历", Visibility: "public",
	}); err != nil {
		t.Fatalf("create parent resume: %v", err)
	}
	r := store.NewJobApplicationRepository()
	id := ug("cov-app")
	created, err := r.Create(context.Background(), domain.JobApplication{
		ID: id, JobID: jobID, ResumeID: resumeID, ApplicantID: userID, Status: domain.AppSubmitted,
	})
	if err != nil {
		t.Fatalf("Create application: %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("Create: application version = %d, want 1", created.Version)
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || got.JobID != jobID || got.ApplicantID != userID {
		t.Fatalf("FindByID: application mismatch: %+v", got)
	}
	upd, err := r.UpdateStatus(context.Background(), id, domain.AppViewed)
	if err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if upd.Status != domain.AppViewed {
		t.Fatalf("UpdateStatus: status = %q, want %q", upd.Status, domain.AppViewed)
	}
	if _, err := r.UpdateStatus(context.Background(), ug("cov-app-missing"), domain.AppViewed); err == nil {
		t.Fatalf("UpdateStatus: expected not-found error on missing application")
	}
	if l, err := r.ListByJob(context.Background(), jobID); err != nil {
		t.Fatalf("ListByJob: %v", err)
	} else if !containsByID(l, id, func(v domain.JobApplication) string { return v.ID }) {
		t.Fatalf("ListByJob: application %s not found", id)
	}
	if l, err := r.ListByApplicant(context.Background(), userID); err != nil {
		t.Fatalf("ListByApplicant: %v", err)
	} else if !containsByID(l, id, func(v domain.JobApplication) string { return v.ID }) {
		t.Fatalf("ListByApplicant: application %s not found", id)
	}
}

// ── PostRepository ──

func TestCoverage_PostRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewPostRepository()
	uid := ug("cov-post-user")
	id := ug("cov-post")
	created, err := r.Create(context.Background(), domain.Post{
		ID: id, AuthorID: uid, Title: "社区帖子", Content: "正文", Images: []string{"/p.jpg"}, Status: "published",
	})
	if err != nil {
		t.Fatalf("Create post: %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("Create: post version = %d, want 1", created.Version)
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || len(got.Images) != 1 {
		t.Fatalf("FindByID: post mismatch: %+v", got)
	}
	if l, total, err := r.ListPublished(context.Background(), 0, 20); err != nil {
		t.Fatalf("ListPublished: %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.Post) string { return v.ID }) {
		t.Fatalf("ListPublished: post %s not found (total=%d)", id, total)
	}
	if l, err := r.ListByAuthor(context.Background(), uid); err != nil {
		t.Fatalf("ListByAuthor: %v", err)
	} else if !containsByID(l, id, func(v domain.Post) string { return v.ID }) {
		t.Fatalf("ListByAuthor: post %s not found", id)
	}
	got.Title = "更新帖子"
	upd, err := r.Update(context.Background(), id, got)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if upd.Title != "更新帖子" {
		t.Fatalf("Update: post title mismatch: %q", upd.Title)
	}
}

// ── CommentRepository ──

func TestCoverage_CommentRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	postID := ug("cov-comment-post")
	if _, err := store.NewPostRepository().Create(context.Background(), domain.Post{
		ID: postID, AuthorID: "cov-u", Title: "帖子", Status: "published",
	}); err != nil {
		t.Fatalf("create parent post: %v", err)
	}
	r := store.NewCommentRepository()
	id := ug("cov-comment")
	created, err := r.Create(context.Background(), domain.Comment{ID: id, PostID: postID, AuthorID: "cov-u", Content: "评论内容"})
	if err != nil {
		t.Fatalf("Create comment: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: comment id mismatch: got %q want %q", created.ID, id)
	}
	if l, err := r.ListByPost(context.Background(), postID); err != nil {
		t.Fatalf("ListByPost: %v", err)
	} else if !containsByID(l, id, func(v domain.Comment) string { return v.ID }) {
		t.Fatalf("ListByPost: comment %s not found", id)
	}
}

// ── ReportRepository ──

func TestCoverage_ReportRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewReportRepository()
	id := ug("cov-report")
	created, err := r.Create(context.Background(), domain.Report{
		ID: id, ReporterID: "cov-u", ResourceType: "post", ResourceID: "cov-res", Reason: "违规内容",
	})
	if err != nil {
		t.Fatalf("Create report: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: report id mismatch: got %q want %q", created.ID, id)
	}
	if l, total, err := r.ListPending(context.Background(), 0, 20); err != nil {
		t.Fatalf("ListPending: %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.Report) string { return v.ID }) {
		t.Fatalf("ListPending: report %s not found (total=%d)", id, total)
	}
}

// ── ListingRepository ──

func TestCoverage_ListingRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewListingRepository()
	sid := ug("cov-listing-seller")
	id := ug("cov-listing")
	created, err := r.Create(context.Background(), domain.Listing{
		ID: id, SellerID: sid, Title: "二手无人机", Description: "9成新", Category: "整机",
		PriceFen: 100000, Images: []string{"/l.jpg"}, Status: "listed",
	})
	if err != nil {
		t.Fatalf("Create listing: %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("Create: listing version = %d, want 1", created.Version)
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || len(got.Images) != 1 {
		t.Fatalf("FindByID: listing mismatch: %+v", got)
	}
	if l, total, err := r.ListByStatus(context.Background(), "listed", 0, 20); err != nil {
		t.Fatalf("ListByStatus(listed): %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.Listing) string { return v.ID }) {
		t.Fatalf("ListByStatus(listed): listing %s not found (total=%d)", id, total)
	}
	if l, total, err := r.ListByStatus(context.Background(), "", 0, 20); err != nil {
		t.Fatalf("ListByStatus(all): %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.Listing) string { return v.ID }) {
		t.Fatalf("ListByStatus(all): listing %s not found (total=%d)", id, total)
	}
	if l, err := r.ListBySeller(context.Background(), sid); err != nil {
		t.Fatalf("ListBySeller: %v", err)
	} else if !containsByID(l, id, func(v domain.Listing) string { return v.ID }) {
		t.Fatalf("ListBySeller: listing %s not found", id)
	}
	got.Title = "更新二手"
	upd, err := r.Update(context.Background(), id, got)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if upd.Title != "更新二手" {
		t.Fatalf("Update: listing title mismatch: %q", upd.Title)
	}
	uid := ug("cov-listing-user")
	if err := r.AddFavorite(context.Background(), id, uid); err != nil {
		t.Fatalf("AddFavorite: %v", err)
	}
	if err := r.AddFavorite(context.Background(), id, uid); err != nil {
		t.Fatalf("AddFavorite (idempotent): %v", err)
	}
	if err := r.RemoveFavorite(context.Background(), id, uid); err != nil {
		t.Fatalf("RemoveFavorite: %v", err)
	}
}

// ── LabourOrderRepository ──

func TestCoverage_LabourOrderRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewLabourOrderRepository()
	eid := ug("cov-labour-emp")
	id := ug("cov-labour")
	created, err := r.Create(context.Background(), domain.LabourOrder{
		ID: id, EmployerID: eid, Title: "无人机巡检", Description: "电力巡检", WorkerCount: 2,
		StartDate: time.Now(), EndDate: time.Now().Add(24 * time.Hour), BudgetFen: 50000, Status: "draft",
	})
	if err != nil {
		t.Fatalf("Create labour order: %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("Create: labour version = %d, want 1", created.Version)
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id {
		t.Fatalf("FindByID: labour id mismatch: got %q want %q", got.ID, id)
	}
	if l, err := r.ListByEmployer(context.Background(), eid); err != nil {
		t.Fatalf("ListByEmployer: %v", err)
	} else if !containsByID(l, id, func(v domain.LabourOrder) string { return v.ID }) {
		t.Fatalf("ListByEmployer: labour %s not found", id)
	}
	if l, total, err := r.ListAll(context.Background(), 0, 20); err != nil {
		t.Fatalf("ListAll: %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.LabourOrder) string { return v.ID }) {
		t.Fatalf("ListAll: labour %s not found (total=%d)", id, total)
	}
	qid := ug("cov-quote")
	quote, err := r.CreateQuote(context.Background(), domain.LabourQuote{
		ID: qid, OrderID: id, QuoterID: ug("cov-quoter"), QuoterName: "报价方", AmountFen: 40000, Proposal: "方案",
	})
	if err != nil {
		t.Fatalf("CreateQuote: %v", err)
	}
	if quote.ID != qid {
		t.Fatalf("CreateQuote: quote id mismatch: got %q want %q", quote.ID, qid)
	}
	if l, err := r.ListQuotes(context.Background(), id); err != nil {
		t.Fatalf("ListQuotes: %v", err)
	} else if !containsByID(l, qid, func(v domain.LabourQuote) string { return v.ID }) {
		t.Fatalf("ListQuotes: quote %s not found", qid)
	}
	workerID := ug("cov-worker")
	aid := ug("cov-assign")
	if _, err := r.CreateAssignment(context.Background(), domain.Assignment{ID: aid, OrderID: id, WorkerID: workerID}); err != nil {
		t.Fatalf("CreateAssignment: %v", err)
	}
	if l, err := r.ListAssignmentsByOrder(context.Background(), id); err != nil {
		t.Fatalf("ListAssignmentsByOrder: %v", err)
	} else if !containsByID(l, aid, func(v domain.Assignment) string { return v.ID }) {
		t.Fatalf("ListAssignmentsByOrder: assignment %s not found", aid)
	}
	if l, err := r.ListAssignmentsByWorker(context.Background(), workerID); err != nil {
		t.Fatalf("ListAssignmentsByWorker: %v", err)
	} else if !containsByID(l, aid, func(v domain.Assignment) string { return v.ID }) {
		t.Fatalf("ListAssignmentsByWorker: assignment %s not found", aid)
	}
}

// ── UserRepository（带 cipher：加密落库 + 解密回明文） ──

func TestCoverage_UserRepoCipher(t *testing.T) {
	store := setupCipherStore(t)
	if store == nil {
		return
	}
	r := store.NewUserRepository()
	id := ug("cov-cipher-user")
	openid := ug("cov-cipher-openid")
	created, err := r.Create(context.Background(), domain.User{ID: id, WechatOpenID: openid, Name: "加密用户", Status: "active"})
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	if created.Version != 1 || created.Role != domain.RoleIndividual {
		t.Fatalf("Create: user mismatch: %+v", created)
	}
	phone := "13900001111"
	if err := r.UpdateProfile(context.Background(), id, domain.UserProfile{Gender: "男", Phone: phone}); err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	var stored string
	if err := store.Pool().QueryRow(context.Background(),
		`SELECT phone_ciphertext FROM users WHERE id=$1`, id).Scan(&stored); err != nil {
		t.Fatalf("read raw phone_ciphertext: %v", err)
	}
	if stored == "" || stored == phone {
		t.Fatalf("phone should be encrypted at rest, got %q", stored)
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.PhoneCipher != phone {
		t.Fatalf("FindByID: phone not decrypted: got %q want %q", got.PhoneCipher, phone)
	}
	byOpenID, err := r.FindByOpenID(context.Background(), openid)
	if err != nil {
		t.Fatalf("FindByOpenID: %v", err)
	}
	if byOpenID.PhoneCipher != phone {
		t.Fatalf("FindByOpenID: phone not decrypted: got %q want %q", byOpenID.PhoneCipher, phone)
	}
	all, err := r.All(context.Background())
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	found := false
	for _, u := range all {
		if u.ID == id {
			found = true
			if u.PhoneCipher != phone {
				t.Fatalf("All: phone not decrypted: got %q want %q", u.PhoneCipher, phone)
			}
		}
	}
	if !found {
		t.Fatalf("All: user %s not found", id)
	}
}

// ── PolicyRepository（保险） ──

func TestCoverage_PolicyRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewPolicyRepository()
	uid := ug("cov-policy-user")
	id := ug("cov-policy")
	created, err := r.Create(context.Background(), domain.InsurancePolicy{
		ID: id, UserID: uid, DroneModel: "M300", DroneSN: "SN-1", PolicyType: "liability",
		PremiumFen: 1000, CoverageFen: 100000, Insurer: "保险公司", Status: "active",
	})
	if err != nil {
		t.Fatalf("Create policy: %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("Create: policy version = %d, want 1", created.Version)
	}
	if l, err := r.ListByUser(context.Background(), uid); err != nil {
		t.Fatalf("ListByUser: %v", err)
	} else if !containsByID(l, id, func(v domain.InsurancePolicy) string { return v.ID }) {
		t.Fatalf("ListByUser: policy %s not found", id)
	}
	if l, total, err := r.ListAll(context.Background(), 0, 20); err != nil {
		t.Fatalf("ListAll: %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.InsurancePolicy) string { return v.ID }) {
		t.Fatalf("ListAll: policy %s not found (total=%d)", id, total)
	}
}

// ── InspectionRepository（年检） ──

func TestCoverage_InspectionRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewInspectionRepository()
	uid := ug("cov-inspect-user")
	id := ug("cov-inspect")
	created, err := r.Create(context.Background(), domain.AnnualInspection{
		ID: id, UserID: uid, DroneModel: "M300", DroneSN: "SN-1", Result: "合格",
		ReportURL: "/r.pdf", Status: "done",
	})
	if err != nil {
		t.Fatalf("Create inspection: %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("Create: inspection version = %d, want 1", created.Version)
	}
	if l, err := r.ListByUser(context.Background(), uid); err != nil {
		t.Fatalf("ListByUser: %v", err)
	} else if !containsByID(l, id, func(v domain.AnnualInspection) string { return v.ID }) {
		t.Fatalf("ListByUser: inspection %s not found", id)
	}
	if l, err := r.ListAll(context.Background()); err != nil {
		t.Fatalf("ListAll: %v", err)
	} else if !containsByID(l, id, func(v domain.AnnualInspection) string { return v.ID }) {
		t.Fatalf("ListAll: inspection %s not found", id)
	}
}

// ── LoanRepository ──

func TestCoverage_LoanRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewLoanRepository()
	uid := ug("cov-loan-user")
	id := ug("cov-loan")
	created, err := r.Create(context.Background(), domain.LoanApplication{
		ID: id, UserID: uid, AmountFen: 100000, TermMonths: 12, Purpose: "购机", Status: "submitted",
	})
	if err != nil {
		t.Fatalf("Create loan: %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("Create: loan version = %d, want 1", created.Version)
	}
	if l, err := r.ListByUser(context.Background(), uid); err != nil {
		t.Fatalf("ListByUser: %v", err)
	} else if !containsByID(l, id, func(v domain.LoanApplication) string { return v.ID }) {
		t.Fatalf("ListByUser: loan %s not found", id)
	}
	if l, total, err := r.ListAll(context.Background(), 0, 20); err != nil {
		t.Fatalf("ListAll: %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.LoanApplication) string { return v.ID }) {
		t.Fatalf("ListAll: loan %s not found (total=%d)", id, total)
	}
}

// ── RepairRepository ──

func TestCoverage_RepairRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewRepairRepository()
	uid := ug("cov-repair-user")
	id := ug("cov-repair")
	created, err := r.Create(context.Background(), domain.RepairOrder{
		ID: id, CustomerID: uid, ProductDesc: "M300", FaultDesc: "无法起飞", QuoteFen: 5000, Status: "submitted",
	})
	if err != nil {
		t.Fatalf("Create repair: %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("Create: repair version = %d, want 1", created.Version)
	}
	if l, err := r.ListByUser(context.Background(), uid); err != nil {
		t.Fatalf("ListByUser: %v", err)
	} else if !containsByID(l, id, func(v domain.RepairOrder) string { return v.ID }) {
		t.Fatalf("ListByUser: repair %s not found", id)
	}
	if l, total, err := r.ListAll(context.Background(), 0, 20); err != nil {
		t.Fatalf("ListAll: %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.RepairOrder) string { return v.ID }) {
		t.Fatalf("ListAll: repair %s not found (total=%d)", id, total)
	}
}

// ── MessageRepository ──

func TestCoverage_MessageRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewMessageRepository()
	sender := ug("cov-msg-sender")
	receiver := ug("cov-msg-receiver")
	id := ug("cov-msg")
	created, err := r.Create(context.Background(), domain.Message{
		ID: id, SenderID: sender, ReceiverID: receiver, Title: "通知", Content: "内容",
		ResourceType: "demand", ResourceID: "d-1",
	})
	if err != nil {
		t.Fatalf("Create message: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: message id mismatch: got %q want %q", created.ID, id)
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || got.ReceiverID != receiver {
		t.Fatalf("FindByID: message mismatch: %+v", got)
	}
	if l, err := r.ListByUser(context.Background(), receiver, false); err != nil {
		t.Fatalf("ListByUser(all): %v", err)
	} else if !containsByID(l, id, func(v domain.Message) string { return v.ID }) {
		t.Fatalf("ListByUser(all): message %s not found", id)
	}
	if l, err := r.ListByUser(context.Background(), receiver, true); err != nil {
		t.Fatalf("ListByUser(unread): %v", err)
	} else if !containsByID(l, id, func(v domain.Message) string { return v.ID }) {
		t.Fatalf("ListByUser(unread): message %s not found", id)
	}
	if n, err := r.UnreadCount(context.Background(), receiver); err != nil {
		t.Fatalf("UnreadCount: %v", err)
	} else if n < 1 {
		t.Fatalf("UnreadCount: expected >=1, got %d", n)
	}
	read, err := r.MarkRead(context.Background(), id)
	if err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
	if !read.IsRead {
		t.Fatalf("MarkRead: message should be read")
	}
	if l, err := r.ListByUser(context.Background(), receiver, true); err != nil {
		t.Fatalf("ListByUser(unread after read): %v", err)
	} else if containsByID(l, id, func(v domain.Message) string { return v.ID }) {
		t.Fatalf("ListByUser(unread after read): read message %s should be excluded", id)
	}
	if l, total, err := r.ListAll(context.Background(), 0, 20); err != nil {
		t.Fatalf("ListAll: %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.Message) string { return v.ID }) {
		t.Fatalf("ListAll: message %s not found (total=%d)", id, total)
	}
	if err := r.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

// ── IntentRepository（需求对接意向） ──

func TestCoverage_IntentRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewIntentRepository()
	demandID := ug("cov-intent-demand")
	if _, err := store.NewDemandRepository().Create(context.Background(), domain.Demand{
		ID: demandID, PublisherID: ug("cov-intent-pub"), Title: "需求", Contact: "13800000000", Status: domain.DemandPublished,
	}); err != nil {
		t.Fatalf("create parent demand: %v", err)
	}
	intentor := ug("cov-intentor")
	id := ug("cov-intent")
	created, err := r.Create(context.Background(), domain.DemandIntent{
		ID: id, DemandID: demandID, IntentorID: intentor, IntentorName: "对接方", Contact: "13800000000", Status: "pending",
	})
	if err != nil {
		t.Fatalf("Create intent: %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("Create: intent version = %d, want 1", created.Version)
	}
	if l, err := r.ListByDemand(context.Background(), demandID); err != nil {
		t.Fatalf("ListByDemand: %v", err)
	} else if !containsByID(l, id, func(v domain.DemandIntent) string { return v.ID }) {
		t.Fatalf("ListByDemand: intent %s not found", id)
	}
	if l, err := r.ListByIntentor(context.Background(), intentor); err != nil {
		t.Fatalf("ListByIntentor: %v", err)
	} else if !containsByID(l, id, func(v domain.DemandIntent) string { return v.ID }) {
		t.Fatalf("ListByIntentor: intent %s not found", id)
	}
	upd, err := r.UpdateStatus(context.Background(), id, "contacted")
	if err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if upd.Status != "contacted" {
		t.Fatalf("UpdateStatus: status = %q, want contacted", upd.Status)
	}
}

// ── WorkOrderRepository（接单派单闭环，含 FK 父行） ──

func TestCoverage_WorkOrderRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	// 父行：publisher user + worker user + demand（work_orders 有 FK）
	publisher := ug("cov-wo-pub")
	if _, err := store.NewUserRepository().Create(context.Background(), domain.User{
		ID: publisher, WechatOpenID: ug("cov-wo-pub-openid"), Status: "active",
	}); err != nil {
		t.Fatalf("create publisher user: %v", err)
	}
	worker := ug("cov-wo-worker")
	if _, err := store.NewUserRepository().Create(context.Background(), domain.User{
		ID: worker, WechatOpenID: ug("cov-wo-worker-openid"), Status: "active",
	}); err != nil {
		t.Fatalf("create worker user: %v", err)
	}
	demandID := ug("cov-wo-demand")
	if _, err := store.NewDemandRepository().Create(context.Background(), domain.Demand{
		ID: demandID, PublisherID: publisher, Title: "需求", Contact: "13800000000", Status: domain.DemandPublished,
	}); err != nil {
		t.Fatalf("create parent demand: %v", err)
	}
	r := store.NewWorkOrderRepository()
	id := ug("cov-wo")
	created, err := r.Create(context.Background(), domain.WorkOrder{
		ID: id, OrderNo: ug("cov-wo-no"), DemandID: demandID, PublisherID: publisher, PublisherName: "企业",
		WorkerID: worker, WorkerName: "飞手", AmountFen: 100000, Status: domain.WorkOrderPending,
		ResultPhotos: []string{"/w1.jpg"},
	})
	if err != nil {
		t.Fatalf("Create work order: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: work order id mismatch")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || len(got.ResultPhotos) != 1 {
		t.Fatalf("FindByID: work order mismatch: %+v", got)
	}
	if l, err := r.ListByPublisher(context.Background(), publisher); err != nil {
		t.Fatalf("ListByPublisher: %v", err)
	} else if !containsByID(l, id, func(v domain.WorkOrder) string { return v.ID }) {
		t.Fatalf("ListByPublisher: work order %s not found", id)
	}
	if l, err := r.ListByWorker(context.Background(), worker); err != nil {
		t.Fatalf("ListByWorker: %v", err)
	} else if !containsByID(l, id, func(v domain.WorkOrder) string { return v.ID }) {
		t.Fatalf("ListByWorker: work order %s not found", id)
	}
	upd, err := r.UpdateStatus(context.Background(), id, domain.WorkOrderPending, domain.WorkOrderOngoing)
	if err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if upd.Status != domain.WorkOrderOngoing {
		t.Fatalf("UpdateStatus: status = %q, want ongoing", upd.Status)
	}
	pupd, err := r.UpdatePhotos(context.Background(), id, []string{"/a.jpg", "/b.jpg"})
	if err != nil {
		t.Fatalf("UpdatePhotos: %v", err)
	}
	if len(pupd.ResultPhotos) != 2 {
		t.Fatalf("UpdatePhotos: photos = %v, want 2", pupd.ResultPhotos)
	}
	rupd, err := r.UpdateRework(context.Background(), id, "整改")
	if err != nil {
		t.Fatalf("UpdateRework: %v", err)
	}
	if rupd.ReworkNote != "整改" {
		t.Fatalf("UpdateRework: note = %q", rupd.ReworkNote)
	}
	cupd, err := r.UpdateCancel(context.Background(), id, "取消原因")
	if err != nil {
		t.Fatalf("UpdateCancel: %v", err)
	}
	if cupd.CancelReason != "取消原因" {
		t.Fatalf("UpdateCancel: reason = %q", cupd.CancelReason)
	}
}

// ── CollegeRepository（院校） ──

func TestCoverage_CollegeRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewCollegeRepository()
	id := ug("cov-college")
	created, err := r.Create(context.Background(), domain.College{
		ID: id, Name: "测试航空学院", Region: "重庆", City: "重庆", Description: "无人机专业",
		Status: "active", Majors: []string{"无人机工程"}, Facilities: []string{"实训基地"},
		Tags: []string{"本科"}, Specialties: []string{"无人机"},
	})
	if err != nil {
		t.Fatalf("Create college: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: college id mismatch")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || len(got.Majors) != 1 {
		t.Fatalf("FindByID: college mismatch: %+v", got)
	}
	if l, err := r.List(context.Background(), "重庆"); err != nil {
		t.Fatalf("List(region): %v", err)
	} else if !containsByID(l, id, func(v domain.College) string { return v.ID }) {
		t.Fatalf("List(region): college %s not found", id)
	}
	if l, err := r.List(context.Background(), ""); err != nil {
		t.Fatalf("List(all): %v", err)
	} else if !containsByID(l, id, func(v domain.College) string { return v.ID }) {
		t.Fatalf("List(all): college %s not found", id)
	}
	got.Name = "更新学院"
	upd, err := r.Update(context.Background(), got)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if upd.Name != "更新学院" {
		t.Fatalf("Update: name = %q", upd.Name)
	}
	if err := r.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

// ── StudyTourRepository（研学） ──

func TestCoverage_StudyTourRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewStudyTourRepository()
	id := ug("cov-tour")
	created, err := r.Create(context.Background(), domain.StudyTour{
		ID: id, Title: "低空研学", Destination: "重庆", Duration: "2天", Capacity: 30,
		Status: "draft", Description: "研学", Location: "重庆", OrganizerID: ug("cov-tour-org"),
		StartDate: time.Now(), EndDate: time.Now().Add(24 * time.Hour), CoverImage: "/t.jpg", PriceFen: 10000,
		Schedule: []domain.StudySchedule{{Day: 1, Title: "第一天", Items: []string{"参观"}}},
	})
	if err != nil {
		t.Fatalf("Create study tour: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: study tour id mismatch")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || len(got.Schedule) != 1 {
		t.Fatalf("FindByID: study tour mismatch: %+v", got)
	}
	if l, err := r.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	} else if !containsByID(l, id, func(v domain.StudyTour) string { return v.ID }) {
		t.Fatalf("List: study tour %s not found", id)
	}
	got.Title = "更新研学"
	upd, err := r.Update(context.Background(), got)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if upd.Title != "更新研学" {
		t.Fatalf("Update: title = %q", upd.Title)
	}
	if err := r.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

// ── ExhibitionRepository（展会 + 展位） ──

func TestCoverage_ExhibitionRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewExhibitionRepository()
	id := ug("cov-exh")
	created, err := r.Create(context.Background(), domain.Exhibition{
		ID: id, Title: "无人机展", Category: "drone_show", Location: "重庆",
		StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 3), BoothCount: 10, BoothPrice: 100000,
		Organizer: "组委会", Status: "recruiting",
	})
	if err != nil {
		t.Fatalf("Create exhibition: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: exhibition id mismatch")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id {
		t.Fatalf("FindByID: exhibition id mismatch")
	}
	if l, total, err := r.List(context.Background(), 0, 20); err != nil {
		t.Fatalf("List: %v", err)
	} else if total == 0 || !containsByID(l, id, func(v domain.Exhibition) string { return v.ID }) {
		t.Fatalf("List: exhibition %s not found (total=%d)", id, total)
	}
	got.Title = "更新展会"
	upd, err := r.Update(context.Background(), got)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if upd.Title != "更新展会" {
		t.Fatalf("Update: title = %q", upd.Title)
	}
	bid := ug("cov-booth")
	b, err := r.CreateBooth(context.Background(), domain.ExhibitionBooth{
		ID: bid, ExhibitionID: id, ExhibitorID: ug("cov-exhibitor"), BoothNumber: "A01",
		ExhibitName: "无人机", Status: "applied",
	})
	if err != nil {
		t.Fatalf("CreateBooth: %v", err)
	}
	if b.ID != bid {
		t.Fatalf("CreateBooth: booth id mismatch")
	}
	if l, err := r.ListBooths(context.Background(), id); err != nil {
		t.Fatalf("ListBooths: %v", err)
	} else if !containsByID(l, bid, func(v domain.ExhibitionBooth) string { return v.ID }) {
		t.Fatalf("ListBooths: booth %s not found", bid)
	}
	ub, err := r.UpdateBoothStatus(context.Background(), bid, "approved")
	if err != nil {
		t.Fatalf("UpdateBoothStatus: %v", err)
	}
	if ub.Status != "approved" {
		t.Fatalf("UpdateBoothStatus: status = %q", ub.Status)
	}
	if err := r.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

// ── TestSiteRepository（测试场地 + 预约） ──

func TestCoverage_TestSiteRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewTestSiteRepository()
	id := ug("cov-site")
	created, err := r.Create(context.Background(), domain.TestSite{
		ID: id, Name: "测试场地", SiteType: "flying_field", OwnerID: ug("cov-site-owner"),
		Location: "重庆", BookingRule: "工作日", Status: "available", PriceFen: 50000, Facilities: []string{"5G"},
	})
	if err != nil {
		t.Fatalf("Create test site: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: test site id mismatch")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || len(got.Facilities) != 1 {
		t.Fatalf("FindByID: test site mismatch: %+v", got)
	}
	if l, err := r.List(context.Background(), "flying_field"); err != nil {
		t.Fatalf("List(type): %v", err)
	} else if !containsByID(l, id, func(v domain.TestSite) string { return v.ID }) {
		t.Fatalf("List(type): test site %s not found", id)
	}
	if l, err := r.List(context.Background(), ""); err != nil {
		t.Fatalf("List(all): %v", err)
	} else if !containsByID(l, id, func(v domain.TestSite) string { return v.ID }) {
		t.Fatalf("List(all): test site %s not found", id)
	}
	got.Name = "更新场地"
	upd, err := r.UpdateSite(context.Background(), got)
	if err != nil {
		t.Fatalf("UpdateSite: %v", err)
	}
	if upd.Name != "更新场地" {
		t.Fatalf("UpdateSite: name = %q", upd.Name)
	}
	userID := ug("cov-site-user")
	bookID := ug("cov-booking")
	bk, err := r.CreateBooking(context.Background(), domain.TestSiteBooking{
		ID: bookID, SiteID: id, UserID: userID, Purpose: "certification",
		StartTime: time.Now(), EndTime: time.Now().Add(time.Hour), ContactName: "张三", ContactPhone: "13800000000",
	})
	if err != nil {
		t.Fatalf("CreateBooking: %v", err)
	}
	if bk.ID != bookID {
		t.Fatalf("CreateBooking: booking id mismatch")
	}
	ub, err := r.UpdateBookingStatus(context.Background(), bookID, "approved", "通过")
	if err != nil {
		t.Fatalf("UpdateBookingStatus: %v", err)
	}
	if ub.Status != "approved" {
		t.Fatalf("UpdateBookingStatus: status = %q", ub.Status)
	}
	if l, err := r.ListBookings(context.Background(), id); err != nil {
		t.Fatalf("ListBookings: %v", err)
	} else if !containsByID(l, bookID, func(v domain.TestSiteBooking) string { return v.ID }) {
		t.Fatalf("ListBookings: booking %s not found", bookID)
	}
	if l, err := r.ListBookingsByUser(context.Background(), userID); err != nil {
		t.Fatalf("ListBookingsByUser: %v", err)
	} else if !containsByID(l, bookID, func(v domain.TestSiteBooking) string { return v.ID }) {
		t.Fatalf("ListBookingsByUser: booking %s not found", bookID)
	}
	if l, total, err := r.ListAllBookings(context.Background(), 0, 20); err != nil {
		t.Fatalf("ListAllBookings: %v", err)
	} else if total == 0 || !containsByID(l, bookID, func(v domain.TestSiteBooking) string { return v.ID }) {
		t.Fatalf("ListAllBookings: booking %s not found (total=%d)", bookID, total)
	}
	if err := r.DeleteSite(context.Background(), id); err != nil {
		t.Fatalf("DeleteSite: %v", err)
	}
}

// ── TransformationRepository（成果转化） ──

func TestCoverage_TransformationRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewTransformationRepository()
	owner := ug("cov-trans-owner")
	id := ug("cov-trans")
	created, err := r.Create(context.Background(), domain.Transformation{
		ID: id, Title: "成果转化", AchievementID: ug("cov-ach"), OwnerID: owner,
		Progress: "50%", PartnerID: ug("cov-partner"), Status: "active", Stage: domain.StageLab,
	})
	if err != nil {
		t.Fatalf("Create transformation: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: transformation id mismatch")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id {
		t.Fatalf("FindByID: transformation id mismatch")
	}
	if l, err := r.List(context.Background(), owner); err != nil {
		t.Fatalf("List(owner): %v", err)
	} else if !containsByID(l, id, func(v domain.Transformation) string { return v.ID }) {
		t.Fatalf("List(owner): transformation %s not found", id)
	}
	if l, err := r.List(context.Background(), ""); err != nil {
		t.Fatalf("List(all): %v", err)
	} else if !containsByID(l, id, func(v domain.Transformation) string { return v.ID }) {
		t.Fatalf("List(all): transformation %s not found", id)
	}
	got.Title = "更新转化"
	upd, err := r.Update(context.Background(), got)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if upd.Title != "更新转化" {
		t.Fatalf("Update: title = %q", upd.Title)
	}
	if err := r.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

// TestCoverage_UploadRepo 直测 uploads 台账：Create 落库 + SumBytesSince 按 owner/时间聚合。
func TestCoverage_UploadRepo(t *testing.T) {
	store := setupCipherStore(t)
	r := store.NewUploadRepository()
	// 测试 DB 跨运行保留数据：owner/ID 一律用运行级唯一后缀，避免与残留行串扰。
	uniq := time.Now().UnixNano()
	owner := fmt.Sprintf("cov-owner-%d", uniq)
	other := fmt.Sprintf("cov-other-%d", uniq)
	now := time.Now()
	recs := []domain.FileRecord{
		{ID: fmt.Sprintf("cov-up-%d-1", uniq), OwnerID: owner, SizeBytes: 100, Visibility: "public", CreatedAt: now.Add(-2 * time.Hour)},
		{ID: fmt.Sprintf("cov-up-%d-2", uniq), OwnerID: owner, SizeBytes: 250, Visibility: "private", CreatedAt: now.Add(-time.Hour)},
		{ID: fmt.Sprintf("cov-up-%d-3", uniq), OwnerID: other, SizeBytes: 999, Visibility: "public", CreatedAt: now.Add(-time.Hour)},
	}
	for _, rec := range recs {
		if err := r.Create(context.Background(), rec); err != nil {
			t.Fatalf("Create(%s): %v", rec.ID, err)
		}
	}
	// 今日累计：owner 自己的两条（100+250），不含他人。
	total, err := r.SumBytesSince(context.Background(), owner, time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()))
	if err != nil {
		t.Fatalf("SumBytesSince: %v", err)
	}
	if total != 350 {
		t.Fatalf("SumBytesSince = %d, want 350", total)
	}
	// since 推进到 1.5 小时前：只剩 cov-up-2（250）。
	late, err := r.SumBytesSince(context.Background(), owner, now.Add(-90*time.Minute))
	if err != nil {
		t.Fatalf("SumBytesSince(late): %v", err)
	}
	if late != 250 {
		t.Fatalf("SumBytesSince(late) = %d, want 250", late)
	}
	// 他人不受影响。
	otherSum, err := r.SumBytesSince(context.Background(), other, time.Time{})
	if err != nil {
		t.Fatalf("SumBytesSince(other): %v", err)
	}
	if otherSum != 999 {
		t.Fatalf("SumBytesSince(other) = %d, want 999", otherSum)
	}
}
