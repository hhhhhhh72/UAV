package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/postgres"
)

// 本文件补齐 internal/repository/postgres 中 0% / 低覆盖的仓储实现。
// 复用 postgres_test.go 的 setupStore（迁移 + Skip 兜底），所有写入用
// cov-<nano> 唯一 id 前缀隔离，不污染既有测试/开发数据。

// ── batch3: RescueCase ──

func TestCoverage_RescueCaseRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewRescueCaseRepository()
	id := ug("cov-rescue")
	created, err := r.Create(context.Background(), domain.RescueCase{
		ID: id, Title: "山火救援", EventType: "mountain_fire", Location: "缙云山",
		Date: time.Now(), DroneModel: "M300", TeamName: "救援队",
		Summary: "火情侦查", Result: "成功", Lessons: "提前规划航线",
		MediaURLs: []string{"/m1.jpg", "/m2.mp4"}, Source: "内部", Status: "published",
	})
	if err != nil {
		t.Fatalf("create rescue case: %v", err)
	}
	if created.ID != id {
		t.Fatalf("id mismatch")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("find rescue case: %v", err)
	}
	if len(got.MediaURLs) != 2 || got.MediaURLs[0] != "/m1.jpg" {
		t.Fatalf("media_urls roundtrip mismatch: %v", got.MediaURLs)
	}
	// List: 事件类型 + 关键词 ILIKE + 分页
	list, total, err := r.List(context.Background(), "mountain_fire", "山火", 0, 20)
	if err != nil {
		t.Fatalf("list rescue cases: %v", err)
	}
	if total < 1 {
		t.Fatalf("expected at least 1 rescue case, total=%d", total)
	}
	found := false
	for _, c := range list {
		if c.ID == id {
			found = true
		}
	}
	if !found {
		t.Fatalf("rescue case %s not found in filtered list", id)
	}
	// 关键词过滤：不匹配的查询不应命中该条
	if others, _, _ := r.List(context.Background(), "mountain_fire", "不存在关键词xyz", 0, 20); containsRescueID(others, id) {
		t.Fatalf("unexpected hit with non-matching keyword")
	}
}

func containsRescueID(items []domain.RescueCase, id string) bool {
	for _, it := range items {
		if it.ID == id {
			return true
		}
	}
	return false
}

// ── batch3: EmergencyDept ──

func TestCoverage_EmergencyDeptRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewEmergencyDeptRepository()
	deptID := ug("cov-dept")
	d, err := r.CreateDept(context.Background(), domain.EmergencyDept{
		ID: deptID, Name: "渝北应急局", DeptType: "emergency_bureau", Region: "渝北区",
		ContactName: "张队", ContactPhone: "13800000000", ProtocolURL: "/p.pdf", Status: "active",
	})
	if err != nil {
		t.Fatalf("create dept: %v", err)
	}
	if d.ID != deptID {
		t.Fatalf("dept id mismatch")
	}
	depts, err := r.ListDepts(context.Background())
	if err != nil {
		t.Fatalf("list depts: %v", err)
	}
	if !containsDeptID(depts, deptID) {
		t.Fatalf("dept %s not found in list", deptID)
	}
	drillID := ug("cov-drill")
	if _, err := r.CreateDrill(context.Background(), domain.EmergencyDrill{
		ID: drillID, DeptID: deptID, Title: "森林火情联合演练", Scenario: "山火",
		Date: time.Now(), Participants: 30, DroneCount: 5, Result: "圆满",
	}); err != nil {
		t.Fatalf("create drill: %v", err)
	}
	// 带 dept_id 过滤
	drills, err := r.ListDrills(context.Background(), deptID)
	if err != nil {
		t.Fatalf("list drills: %v", err)
	}
	if !containsDrillID(drills, drillID) {
		t.Fatalf("drill %s not found in dept list", drillID)
	}
	// 全量（空 deptID）
	all, err := r.ListDrills(context.Background(), "")
	if err != nil {
		t.Fatalf("list all drills: %v", err)
	}
	if !containsDrillID(all, drillID) {
		t.Fatalf("drill %s not found in all list", drillID)
	}
}

func containsDeptID(items []domain.EmergencyDept, id string) bool {
	for _, it := range items {
		if it.ID == id {
			return true
		}
	}
	return false
}

func containsDrillID(items []domain.EmergencyDrill, id string) bool {
	for _, it := range items {
		if it.ID == id {
			return true
		}
	}
	return false
}

// ── batch3: AssociationMember ──

func TestCoverage_AssociationMemberRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewAssociationMemberRepository()
	userID := ug("cov-assoc-user")
	id := ug("cov-assoc")
	m, err := r.Create(context.Background(), domain.AssociationMember{
		ID: id, UserID: userID, EnterpriseID: "ent-1", Role: domain.AssocMember,
		Status: "active",
	})
	if err != nil {
		t.Fatalf("create member: %v", err)
	}
	if m.ID != id || m.Role != domain.AssocMember {
		t.Fatalf("member mismatch: %+v", m)
	}
	// JoinDate 零值应回填 CreatedAt
	if m.JoinDate.IsZero() {
		t.Fatalf("join_date should be backfilled")
	}
	got, err := r.FindByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("find member by user: %v", err)
	}
	if got.ID != id {
		t.Fatalf("find member id mismatch")
	}
	// List 带 role 过滤
	list, total, err := r.List(context.Background(), string(domain.AssocMember), 0, 20)
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if total < 1 || !containsMemberID(list, id) {
		t.Fatalf("member %s not found in role-filtered list (total=%d)", id, total)
	}
	// UpdateRole
	upd, err := r.UpdateRole(context.Background(), id, domain.AssocSecretary)
	if err != nil {
		t.Fatalf("update role: %v", err)
	}
	if upd.Role != domain.AssocSecretary {
		t.Fatalf("role not updated: %v", upd.Role)
	}
}

func containsMemberID(items []domain.AssociationMember, id string) bool {
	for _, it := range items {
		if it.ID == id {
			return true
		}
	}
	return false
}

// ── ServiceListing 全 CRUD ──

func TestCoverage_ServiceListingRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewServiceListingRepository()
	id := ug("cov-sl")
	created, err := r.Create(context.Background(), domain.ServiceListing{
		ID: id, ProviderID: "prov-1", ProviderName: "测试企业", Title: "电力巡检服务",
		Category: "巡检", Description: "高压线巡检", Region: "重庆", PriceFen: 50000,
		Unit: "公里", Image: "/sl.png", Status: "published",
	})
	if err != nil {
		t.Fatalf("create service listing: %v", err)
	}
	if created.ID != id {
		t.Fatalf("sl id mismatch")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("find service listing: %v", err)
	}
	if got.Title != "电力巡检服务" {
		t.Fatalf("sl title mismatch: %v", got.Title)
	}
	list, err := r.List(context.Background())
	if err != nil {
		t.Fatalf("list service listings: %v", err)
	}
	if !containsServiceListingID(list, id) {
		t.Fatalf("sl %s not found in list", id)
	}
	got.Title = "更新后的巡检服务"
	got.Status = "offline"
	upd, err := r.Update(context.Background(), got)
	if err != nil {
		t.Fatalf("update service listing: %v", err)
	}
	if upd.Title != "更新后的巡检服务" || upd.Status != "offline" {
		t.Fatalf("sl update mismatch: %+v", upd)
	}
	if err := r.Delete(context.Background(), id); err != nil {
		t.Fatalf("delete service listing: %v", err)
	}
	if _, err := r.FindByID(context.Background(), id); err == nil {
		t.Fatalf("expected find error after delete")
	}
}

func containsServiceListingID(items []domain.ServiceListing, id string) bool {
	for _, it := range items {
		if it.ID == id {
			return true
		}
	}
	return false
}

// ── Audit + AuditAdapter ──

func TestCoverage_Audit(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	ctx := context.Background()
	rid := ug("cov-audit")
	if err := store.WriteAudit(ctx, postgres.AuditEntry{
		ActorID: "u-1", Action: "create", ResourceType: "demand",
		ResourceID: rid, Result: "ok", RequestID: "req-1",
		Metadata: map[string]any{"key": "value"},
	}); err != nil {
		t.Fatalf("WriteAudit: %v", err)
	}
	// 写入后能查回
	var actor, action, rtype, meta string
	if err := store.Pool().QueryRow(ctx,
		`SELECT actor_id, action, resource_type, metadata::text FROM audit_logs WHERE resource_id=$1`, rid).
		Scan(&actor, &action, &rtype, &meta); err != nil {
		t.Fatalf("read audit back: %v", err)
	}
	if actor != "u-1" || action != "create" || rtype != "demand" {
		t.Fatalf("audit row mismatch: %s/%s/%s", actor, action, rtype)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(meta), &m); err != nil {
		t.Fatalf("unmarshal audit metadata: %v", err)
	}
	if m["key"] != "value" {
		t.Fatalf("audit metadata mismatch: %v", m)
	}

	// 适配器路径
	adapter := postgres.NewAuditAdapter(store)
	rid2 := ug("cov-audit2")
	if err := adapter.WriteAudit(ctx, repository.AuditEntry{
		ActorID: "u-2", Action: "approve", ResourceType: "enterprise",
		ResourceID: rid2, Result: "ok", Metadata: map[string]any{"n": float64(1)},
	}); err != nil {
		t.Fatalf("adapter WriteAudit: %v", err)
	}
	var n int
	if err := store.Pool().QueryRow(ctx,
		`SELECT count(*) FROM audit_logs WHERE resource_id=$1`, rid2).Scan(&n); err != nil {
		t.Fatalf("count adapter audit: %v", err)
	}
	if n != 1 {
		t.Fatalf("adapter audit row count=%d, want 1", n)
	}
}

// ── DemandRepository ──

func TestCoverage_DemandRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	repo := store.NewDemandRepository()
	pub := ug("cov-demand-pub")
	// 公开需求用于 List/Search/ListByPublisher
	if _, err := repo.Create(context.Background(), domain.Demand{
		ID: pub, PublisherID: pub, PublisherName: "覆盖测试企业", Contact: "13800000000",
		BizType: domain.BizCableInspection, District: "渝北区", CityCode: "500112",
		Title: "电力巡检需求", Description: "需要飞手巡检高压线路",
		Images: []string{"/a.jpg", "/b.jpg"}, Latitude: 29.5, Longitude: 106.5,
		BudgetFen: 100000, OfflineAmountFen: 0, BizFields: map[string]any{"lines": float64(10)},
		Status: domain.DemandPublished,
	}); err != nil {
		t.Fatalf("create demand: %v", err)
	}
	got, err := repo.FindByID(context.Background(), pub)
	if err != nil {
		t.Fatalf("find demand: %v", err)
	}
	if len(got.Images) != 2 || len(got.BizFields) != 1 || got.BizType != domain.BizCableInspection {
		t.Fatalf("demand roundtrip mismatch: images=%v bizfields=%v biztype=%v", got.Images, got.BizFields, got.BizType)
	}

	// List：仅已发布 + 地区/类型过滤
	if l, _ := repo.List(context.Background(), repository.DemandFilter{District: "渝北区"}); !containsDemandID(l, pub) {
		t.Fatalf("demand %s not in district-filtered list", pub)
	}
	if l, _ := repo.List(context.Background(), repository.DemandFilter{BizType: "cable_inspection"}); !containsDemandID(l, pub) {
		t.Fatalf("demand %s not in biztype-filtered list", pub)
	}

	// ListAll：status 过滤
	if l, _ := repo.ListAll(context.Background(), repository.DemandFilter{Status: "published"}); !containsDemandID(l, pub) {
		t.Fatalf("demand %s not in ListAll published", pub)
	}

	// Search
	if l, _ := repo.Search(context.Background(), "巡检"); !containsDemandID(l, pub) {
		t.Fatalf("demand %s not found by Search", pub)
	}

	// ListByPublisher
	if l, _ := repo.ListByPublisher(context.Background(), pub); !containsDemandID(l, pub) {
		t.Fatalf("demand %s not found by ListByPublisher", pub)
	}

	// Update：成功路径 + 乐观锁冲突
	got.Title = "第一次更新"
	if _, err := repo.Update(context.Background(), got); err != nil {
		t.Fatalf("update demand: %v", err)
	}
	// got 仍是 version=1 的旧值 → 第二次 Update 命中 0 行
	got.Title = "第二次更新（应冲突）"
	if _, err := repo.Update(context.Background(), got); err == nil {
		t.Fatalf("expected optimistic lock conflict, got nil error")
	}

	// CompareAndSetStatus：成功 / 错误旧状态 / 不存在
	ok, d, err := repo.CompareAndSetStatus(context.Background(), pub, domain.DemandPublished, domain.DemandCompleted)
	if err != nil || !ok || d.Status != domain.DemandCompleted {
		t.Fatalf("CAS success failed: ok=%v err=%v", ok, err)
	}
	if ok2, _, _ := repo.CompareAndSetStatus(context.Background(), pub, domain.DemandCancelled, domain.DemandPublished); ok2 {
		t.Fatalf("CAS should fail with wrong old status")
	}
	if _, _, err := repo.CompareAndSetStatus(context.Background(), ug("cov-demand-missing"), domain.DemandPending, domain.DemandPublished); err == nil {
		t.Fatalf("CAS on missing demand should return error")
	}

	// Delete：成功 / 再次删除报 not found
	if err := repo.Delete(context.Background(), pub); err != nil {
		t.Fatalf("delete demand: %v", err)
	}
	if err := repo.Delete(context.Background(), pub); err == nil {
		t.Fatalf("expected not-found error on second delete")
	}
}

func containsDemandID(items []domain.Demand, id string) bool {
	for _, it := range items {
		if it.ID == id {
			return true
		}
	}
	return false
}

// ── EnterpriseRepository ──

func TestCoverage_EnterpriseRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	repo := store.NewEnterpriseRepository()
	owner := ug("cov-ent-owner")
	id := ug("cov-ent")
	now := time.Now()
	if _, err := repo.Create(context.Background(), domain.Enterprise{
		ID: id, OwnerUserID: owner, Name: "覆盖率测试企业", CreditCode: "91150000TEST",
		LegalPerson: "张三", ContactPhone: "13800000000", IndustryCategory: "整机",
		Scale: "中型", Address: "重庆市渝北区", Description: "测试描述",
		BusinessHours: "9:00-18:00", Logo: "/logo.png", CoverImage: "/cover.png",
		LicenseURL: "license-url", AccountName: "account-name", ContactPerson: "李四",
		Email: "li@test.com", FoundedAt: "2020-01", CapabilityTags: "整机,飞控",
		Status: domain.EnterpriseSubmitted, IsMember: false,
		Version: 1, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create enterprise: %v", err)
	}
	got, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("find enterprise: %v", err)
	}
	if got.Name != "覆盖率测试企业" || got.Status != domain.EnterpriseSubmitted {
		t.Fatalf("enterprise mismatch: %+v", got)
	}
	// FindByOwner
	if owned, _ := repo.FindByOwner(context.Background(), owner); !containsEnterpriseID(owned, id) {
		t.Fatalf("enterprise %s not found by owner", id)
	}
	// ListByStatus：空 = 全部；submitted = 过滤
	if all, _, _ := repo.ListByStatus(context.Background(), "", 0, 20); !containsEnterpriseID(all, id) {
		t.Fatalf("enterprise %s not in ListByStatus all", id)
	}
	if sub, _, _ := repo.ListByStatus(context.Background(), "submitted", 0, 20); !containsEnterpriseID(sub, id) {
		t.Fatalf("enterprise %s not in ListByStatus submitted", id)
	}
	// Pending（status=submitted）
	if p, _ := repo.Pending(context.Background()); !containsEnterpriseID(p, id) {
		t.Fatalf("enterprise %s not in Pending", id)
	}
	// Search
	if s, _ := repo.Search(context.Background(), "覆盖率"); !containsEnterpriseID(s, id) {
		t.Fatalf("enterprise %s not found by Search", id)
	}
	// Update
	got.Name = "更新后的企业"
	got.Status = domain.EnterpriseApproved
	upd, err := repo.Update(context.Background(), id, got)
	if err != nil {
		t.Fatalf("update enterprise: %v", err)
	}
	if upd.Version != 2 || upd.Name != "更新后的企业" {
		t.Fatalf("enterprise update mismatch: %+v", upd)
	}
	// AddDocument / ListDocuments
	docID := ug("cov-doc")
	if _, err := repo.AddDocument(context.Background(), domain.EnterpriseDocument{
		ID: docID, EnterpriseID: id, FileID: "f-1", DocumentType: "license",
		ReviewStatus: "pending", CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("add document: %v", err)
	}
	docs, err := repo.ListDocuments(context.Background(), id)
	if err != nil {
		t.Fatalf("list documents: %v", err)
	}
	if !containsDocID(docs, docID) {
		t.Fatalf("document %s not found", docID)
	}
	// Delete
	if err := repo.Delete(context.Background(), id); err != nil {
		t.Fatalf("delete enterprise: %v", err)
	}
}

func containsEnterpriseID(items []domain.Enterprise, id string) bool {
	for _, it := range items {
		if it.ID == id {
			return true
		}
	}
	return false
}

func containsDocID(items []domain.EnterpriseDocument, id string) bool {
	for _, it := range items {
		if it.ID == id {
			return true
		}
	}
	return false
}

// ── JobRepository ──

func TestCoverage_JobRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	repo := store.NewJobRepository()
	eid := ug("cov-job-ent")
	id := ug("cov-job")
	if _, err := repo.Create(context.Background(), domain.Job{
		ID: id, EnterpriseID: eid, Title: "飞手", Description: "巡检飞手",
		Location: "重庆", SalaryFen: 50000, JobType: "full-time", Status: domain.JobPublished,
	}); err != nil {
		t.Fatalf("create job: %v", err)
	}
	got, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("find job: %v", err)
	}
	if got.Status != domain.JobPublished {
		t.Fatalf("job status mismatch: %v", got.Status)
	}
	// ListByEnterprise
	if l, _ := repo.ListByEnterprise(context.Background(), eid); !containsJobID(l, id) {
		t.Fatalf("job %s not in ListByEnterprise", id)
	}
	// ListPublished / ListAll
	if l, _, _ := repo.ListPublished(context.Background(), 0, 20); !containsJobID(l, id) {
		t.Fatalf("job %s not in ListPublished", id)
	}
	if l, _, _ := repo.ListAll(context.Background(), 0, 20); !containsJobID(l, id) {
		t.Fatalf("job %s not in ListAll", id)
	}
	// Update
	got.Title = "更新飞手"
	got.Status = domain.JobClosed
	upd, err := repo.Update(context.Background(), id, got)
	if err != nil {
		t.Fatalf("update job: %v", err)
	}
	if upd.Version != 2 || upd.Title != "更新飞手" {
		t.Fatalf("job update mismatch: %+v", upd)
	}
	// Delete
	if err := repo.Delete(context.Background(), id); err != nil {
		t.Fatalf("delete job: %v", err)
	}
}

func containsJobID(items []domain.Job, id string) bool {
	for _, it := range items {
		if it.ID == id {
			return true
		}
	}
	return false
}

// ── UserRepository ──

func TestCoverage_UserRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	repo := store.NewUserRepository()
	id := ug("cov-user")
	openid := ug("cov-openid")
	created, err := repo.Create(context.Background(), domain.User{
		ID: id, WechatOpenID: openid, Name: "覆盖率用户", AvatarURL: "/av.png",
		Role: domain.RoleEnterprise, Status: "active",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if created.Role != domain.RoleEnterprise || created.Version != 1 {
		t.Fatalf("user create mismatch: %+v", created)
	}
	// FindByOpenID
	u, err := repo.FindByOpenID(context.Background(), openid)
	if err != nil {
		t.Fatalf("find by openid: %v", err)
	}
	if u.ID != id {
		t.Fatalf("find by openid id mismatch")
	}
	// FindByID / All
	if _, err := repo.FindByID(context.Background(), id); err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if _, err := repo.All(context.Background()); err != nil {
		t.Fatalf("all users: %v", err)
	}
	// UpdateRole / UpdateAvatar / UpdateName
	if err := repo.UpdateRole(context.Background(), id, domain.RoleAssociationAdmin); err != nil {
		t.Fatalf("update role: %v", err)
	}
	if err := repo.UpdateAvatar(context.Background(), id, "/av2.png"); err != nil {
		t.Fatalf("update avatar: %v", err)
	}
	if err := repo.UpdateName(context.Background(), id, "改名用户"); err != nil {
		t.Fatalf("update name: %v", err)
	}
	// UpdateProfile（含手机号）
	if err := repo.UpdateProfile(context.Background(), id, domain.UserProfile{
		Gender: "男", Birthday: "1995-05-20", Region: "重庆", Bio: "飞手", Phone: "13900000000",
	}); err != nil {
		t.Fatalf("update profile: %v", err)
	}
	after, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("find after profile update: %v", err)
	}
	if after.Name != "改名用户" || after.Gender != "男" || after.Region != "重庆" {
		t.Fatalf("profile update mismatch: %+v", after)
	}
	// Delete 级联：先落 refresh_token + user_role，再删用户
	rt := store.NewRefreshTokenRepository()
	if err := rt.Store(context.Background(), id, "hash-"+id, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("store refresh token: %v", err)
	}
	if _, err := store.Pool().Exec(context.Background(),
		`INSERT INTO user_roles (user_id, role_code) VALUES ($1, $2)`, id, "enterprise"); err != nil {
		t.Fatalf("insert user_role: %v", err)
	}
	if err := repo.Delete(context.Background(), id); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	// refresh_token 与 user_role 应被级联删除
	if _, _, _, err := rt.Find(context.Background(), "hash-"+id); err == nil {
		t.Fatalf("refresh token should be cascade-deleted")
	}
	var roleCount int
	if err := store.Pool().QueryRow(context.Background(),
		`SELECT count(*) FROM user_roles WHERE user_id=$1`, id).Scan(&roleCount); err != nil {
		t.Fatalf("count user_roles: %v", err)
	}
	if roleCount != 0 {
		t.Fatalf("user_roles not cascade-deleted: %d", roleCount)
	}
}

// ── RefreshTokenRepository ──

func TestCoverage_RefreshTokenRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	// refresh_tokens.user_id 有外键，先建用户
	userID := ug("cov-rt-user")
	if _, err := store.NewUserRepository().Create(context.Background(), domain.User{
		ID: userID, WechatOpenID: ug("cov-rt-openid"), Name: "rt用户", Status: "active",
	}); err != nil {
		t.Fatalf("create user for refresh token: %v", err)
	}
	repo := store.NewRefreshTokenRepository()
	hash := "cov-rt-hash-" + userID
	exp := time.Now().Add(24 * time.Hour)
	if err := repo.Store(context.Background(), userID, hash, exp); err != nil {
		t.Fatalf("store refresh token: %v", err)
	}
	uid, gotExp, revoked, err := repo.Find(context.Background(), hash)
	if err != nil {
		t.Fatalf("find refresh token: %v", err)
	}
	if uid != userID || revoked {
		t.Fatalf("refresh token mismatch: uid=%s revoked=%v", uid, revoked)
	}
	if gotExp.Unix() != exp.Unix() {
		t.Fatalf("expires_at mismatch: got=%v want=%v", gotExp, exp)
	}
	// Revoke 后再查 revoked=true
	if err := repo.Revoke(context.Background(), hash); err != nil {
		t.Fatalf("revoke refresh token: %v", err)
	}
	if _, _, revoked, err := repo.Find(context.Background(), hash); err != nil || !revoked {
		t.Fatalf("token should be revoked: revoked=%v err=%v", revoked, err)
	}
}

// ── CertificateRepository：Update / Delete（其余已有覆盖） ──

func TestCoverage_CertificateRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	repo := store.NewCertificateRepository()
	id := ug("cov-cert")
	if _, err := repo.Create(context.Background(), domain.Certificate{
		ID: id, UserID: "u-1", CertType: domain.CertCAAC, CertNumber: "CN-001",
		Level: "Ⅱ", IssueDate: time.Now(), ExpireDate: time.Now().AddDate(2, 0, 0),
		IssuerOrg: "CAAC", ImageURL: "/c.png", Status: "approved",
	}); err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	got, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("find certificate: %v", err)
	}
	got.Level = "Ⅲ"
	got.Status = "rejected"
	upd, err := repo.Update(context.Background(), got)
	if err != nil {
		t.Fatalf("update certificate: %v", err)
	}
	if upd.Version != 2 || upd.Level != "Ⅲ" || upd.Status != "rejected" {
		t.Fatalf("certificate update mismatch: %+v", upd)
	}
	if err := repo.Delete(context.Background(), id); err != nil {
		t.Fatalf("delete certificate: %v", err)
	}
}

// ── EnrollmentRepository：去重 + FindByUserAndCourse ──

func TestCoverage_EnrollmentRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	repo := store.NewEnrollmentRepository()
	userID := ug("cov-enr-user")
	courseID := ug("cov-enr-course")
	id := ug("cov-enr")
	birthday := time.Date(1995, 5, 20, 0, 0, 0, 0, time.UTC)
	if _, err := repo.Create(context.Background(), domain.Enrollment{
		ID: id, CourseID: courseID, UserID: userID, Name: "张三", Phone: "13800000000",
		IDCard: "500101199505201234", Gender: "男", Birthday: birthday, Email: "z@t.com",
		Education: "本科", Experience: "3年", PhotoURL: "/p.jpg", IDCardImage: "/i.jpg",
		NoCrime: "有", Status: "enrolled",
	}); err != nil {
		t.Fatalf("create enrollment: %v", err)
	}
	// 唯一索引 (user_id, course_id) 重复报名 → 友好错误
	if _, err := repo.Create(context.Background(), domain.Enrollment{
		ID: ug("cov-enr-dup"), CourseID: courseID, UserID: userID, Status: "enrolled",
	}); err == nil {
		t.Fatalf("expected duplicate enrollment error")
	} else if !strings.Contains(err.Error(), "请勿重复报名") {
		t.Fatalf("unexpected duplicate error: %v", err)
	}
	// FindByUserAndCourse：命中 / 未命中
	found, ok, err := repo.FindByUserAndCourse(context.Background(), userID, courseID)
	if err != nil {
		t.Fatalf("find by user+course: %v", err)
	}
	if !ok || found.ID != id {
		t.Fatalf("expected found enrollment, ok=%v", ok)
	}
	if _, ok2, _ := repo.FindByUserAndCourse(context.Background(), ug("cov-enr-missing"), courseID); ok2 {
		t.Fatalf("expected not-found for missing user")
	}
	// FindByID / Update / ListAll / ListByCourse
	if _, err := repo.FindByID(context.Background(), id); err != nil {
		t.Fatalf("find enrollment by id: %v", err)
	}
	upd, err := repo.Update(context.Background(), domain.Enrollment{
		ID: id, Name: "张三改", Phone: "13900000000", IDCard: "500101199505201234",
		Gender: "男", Birthday: birthday, Email: "z2@t.com", Education: "硕士",
		Experience: "5年", PhotoURL: "/p2.jpg", IDCardImage: "/i2.jpg",
		NoCrime: "有", Status: "completed",
	})
	if err != nil {
		t.Fatalf("update enrollment: %v", err)
	}
	if upd.Name != "张三改" || upd.Status != "completed" {
		t.Fatalf("enrollment update mismatch: %+v", upd)
	}
	if _, _, err := repo.ListAll(context.Background(), 0, 20); err != nil {
		t.Fatalf("list all enrollments: %v", err)
	}
	if l, _ := repo.ListByCourse(context.Background(), courseID); !containsEnrollmentID(l, id) {
		t.Fatalf("enrollment %s not in ListByCourse", id)
	}
}

func containsEnrollmentID(items []domain.Enrollment, id string) bool {
	for _, it := range items {
		if it.ID == id {
			return true
		}
	}
	return false
}

// ── EscrowRepository：事务 + 余额不足哨兵错误 ──

func TestCoverage_EscrowRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewEscrowRepository()
	uid := ug("cov-esc")
	now := time.Now()
	// GetAccount 不存在 → 零值账户
	zero, err := r.GetAccount(context.Background(), uid)
	if err != nil {
		t.Fatalf("get account (absent): %v", err)
	}
	if zero.BalanceFen != 0 || zero.UserID != uid {
		t.Fatalf("absent account mismatch: %+v", zero)
	}
	// Deposit 即开户
	if _, err := r.Deposit(context.Background(), uid, 100000, domain.EscrowTransaction{
		ID: "cov-tx-dep-" + uid, FromUser: "sys", ToUser: uid, AmountFen: 100000,
		TxType: "deposit", Status: "completed", CreatedAt: now,
	}); err != nil {
		t.Fatalf("deposit: %v", err)
	}
	acct, err := r.GetAccount(context.Background(), uid)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if acct.BalanceFen != 100000 {
		t.Fatalf("balance after deposit = %d, want 100000", acct.BalanceFen)
	}
	// Freeze
	if _, err := r.Freeze(context.Background(), uid, 30000, domain.EscrowTransaction{
		ID: "cov-tx-frz-" + uid, FromUser: uid, ToUser: "escrow", AmountFen: 30000,
		TxType: "freeze", Status: "completed", CreatedAt: now,
	}); err != nil {
		t.Fatalf("freeze: %v", err)
	}
	// Release 给收款方（收款方无账户，upsert 开户）
	worker := ug("cov-esc-worker")
	if _, err := r.Release(context.Background(), uid, worker, 30000, domain.EscrowTransaction{
		ID: "cov-tx-rel-" + uid, FromUser: uid, ToUser: worker, AmountFen: 30000,
		TxType: "release", Status: "completed", CreatedAt: now,
	}); err != nil {
		t.Fatalf("release: %v", err)
	}
	workerAcct, err := r.GetAccount(context.Background(), worker)
	if err != nil {
		t.Fatalf("get worker account: %v", err)
	}
	if workerAcct.BalanceFen != 30000 {
		t.Fatalf("worker balance = %d, want 30000", workerAcct.BalanceFen)
	}
	// Refund
	if _, err := r.Freeze(context.Background(), uid, 20000, domain.EscrowTransaction{
		ID: "cov-tx-frz2-" + uid, FromUser: uid, ToUser: "escrow", AmountFen: 20000,
		TxType: "freeze", Status: "completed", CreatedAt: now,
	}); err != nil {
		t.Fatalf("freeze2: %v", err)
	}
	if _, err := r.Refund(context.Background(), uid, 20000, domain.EscrowTransaction{
		ID: "cov-tx-ref-" + uid, FromUser: "escrow", ToUser: uid, AmountFen: 20000,
		TxType: "refund", Status: "completed", CreatedAt: now,
	}); err != nil {
		t.Fatalf("refund: %v", err)
	}
	// 余额不足哨兵：Freeze 超过余额
	if _, err := r.Freeze(context.Background(), uid, 999999, domain.EscrowTransaction{
		ID: "cov-tx-fail-" + uid, FromUser: uid, ToUser: "escrow", AmountFen: 999999,
		TxType: "freeze", Status: "completed", CreatedAt: now,
	}); !errors.Is(err, repository.ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
	// 冻结不足哨兵：Refund / Release 超过冻结
	if _, err := r.Refund(context.Background(), uid, 999999, domain.EscrowTransaction{
		ID: "cov-tx-ref-fail-" + uid, FromUser: "escrow", ToUser: uid, AmountFen: 999999,
		TxType: "refund", Status: "completed", CreatedAt: now,
	}); !errors.Is(err, repository.ErrInsufficientFrozenBalance) {
		t.Fatalf("expected ErrInsufficientFrozenBalance, got %v", err)
	}
	if _, err := r.Release(context.Background(), uid, worker, 999999, domain.EscrowTransaction{
		ID: "cov-tx-rel-fail-" + uid, FromUser: uid, ToUser: worker, AmountFen: 999999,
		TxType: "release", Status: "completed", CreatedAt: now,
	}); !errors.Is(err, repository.ErrInsufficientFrozenBalance) {
		t.Fatalf("expected ErrInsufficientFrozenBalance on release, got %v", err)
	}
	// ListTransactions
	txs, err := r.ListTransactions(context.Background(), uid)
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if !containsTxID(txs, "cov-tx-dep-"+uid) {
		t.Fatalf("deposit tx not found in ListTransactions (%d rows)", len(txs))
	}
}

func containsTxID(items []domain.EscrowTransaction, id string) bool {
	for _, it := range items {
		if it.ID == id {
			return true
		}
	}
	return false
}
