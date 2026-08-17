package postgres_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
)

// 本文件第二轮补齐 internal/repository/postgres 中剩余 <70% 的仓储实现。
// 复用 postgres_test.go 的 setupStore/ug/migrateOnce、integration_coverage_test.go 与
// postgres_main_coverage_test.go 的 setupCipherStore/containsByID 基建；数据 id 用 ug("cov-") 前缀。
// 断言失败统一输出方法名 + err。

// round2Cipher 构造一个独立 AES-256-GCM cipher，用于覆盖 pilotRepo.enc/dec 的 cipher 分支。
func round2Cipher(t *testing.T) *crypto.Cipher {
	t.Helper()
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	cipher, err := crypto.NewCipher(key)
	if err != nil {
		t.Fatalf("round2Cipher: new cipher: %v", err)
	}
	return cipher
}

// ── batch3: ResourcePool + CooperationProgram（0% → 全 CRUD） ──

func TestRound2_ResourcePoolRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewResourcePoolRepository()
	id := ug("cov-pool")
	created, err := r.Create(context.Background(), domain.ResourcePool{
		ID: id, Name: "应急无人机池", PoolType: "emergency", Description: "低空应急",
		OwnerID: ug("cov-pool-owner"), Resources: []string{"r-1", "r-2"}, Status: "active",
	})
	if err != nil {
		t.Fatalf("Create resource pool: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: pool id mismatch: got %q want %q", created.ID, id)
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || len(got.Resources) != 2 || got.Resources[0] != "r-1" {
		t.Fatalf("FindByID: pool roundtrip mismatch: %+v", got)
	}
	// List：类型过滤 + 全量
	if l, err := r.List(context.Background(), "emergency"); err != nil || !containsByID(l, id, func(v domain.ResourcePool) string { return v.ID }) {
		t.Fatalf("List(emergency): pool %s not found err=%v", id, err)
	}
	if l, err := r.List(context.Background(), ""); err != nil || !containsByID(l, id, func(v domain.ResourcePool) string { return v.ID }) {
		t.Fatalf("List(all): pool %s not found err=%v", id, err)
	}
	// AddMember + ListMembers
	mid := ug("cov-pool-member")
	if _, err := r.AddMember(context.Background(), domain.ResourcePoolMember{
		ID: mid, PoolID: id, ResID: "r-1", ResType: "drone", Quantity: 2, Status: "standby",
	}); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	if l, err := r.ListMembers(context.Background(), id); err != nil || !containsByID(l, mid, func(v domain.ResourcePoolMember) string { return v.ID }) {
		t.Fatalf("ListMembers: member %s not found err=%v", mid, err)
	}
}

func TestRound2_CooperationRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewCooperationRepository()
	id := ug("cov-coop")
	eid := ug("cov-coop-ent")
	created, err := r.Create(context.Background(), domain.CooperationProgram{
		ID: id, Title: "定向培养", CollegeID: ug("cov-coop-college"), EnterpriseID: eid,
		CoopType: "directed_training", Description: "校企共建", StartDate: time.Now(),
		EndDate: time.Now().AddDate(1, 0, 0), StudentQuota: 30, Status: "proposed",
	})
	if err != nil {
		t.Fatalf("Create cooperation: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: cooperation id mismatch")
	}
	if got, err := r.FindByID(context.Background(), id); err != nil || got.ID != id || got.StudentQuota != 30 {
		t.Fatalf("FindByID: cooperation mismatch: %+v err=%v", got, err)
	}
	if l, err := r.List(context.Background(), eid); err != nil || !containsByID(l, id, func(v domain.CooperationProgram) string { return v.ID }) {
		t.Fatalf("List(enterprise): cooperation %s not found err=%v", id, err)
	}
	if l, err := r.List(context.Background(), ""); err != nil || !containsByID(l, id, func(v domain.CooperationProgram) string { return v.ID }) {
		t.Fatalf("List(all): cooperation %s not found err=%v", id, err)
	}
	if upd, err := r.UpdateStatus(context.Background(), id, "active"); err != nil || upd.Status != "active" {
		t.Fatalf("UpdateStatus: status=%q err=%v", upd.Status, err)
	}
}

// ── biz_repos: Case.Update / Portfolio.List+Delete / Application 全 CRUD ──

func TestRound2_CaseUpdate(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewCaseRepository()
	id := ug("cov-case")
	if _, err := r.Create(context.Background(), domain.CaseEntry{ID: id, Title: "巡检案例", Category: "inspection", Status: "draft"}); err != nil {
		t.Fatalf("Create case: %v", err)
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	got.Title = "更新案例"
	got.Images = []string{"/c1.jpg", "/c2.jpg"}
	upd, err := r.Update(context.Background(), got)
	if err != nil {
		t.Fatalf("Update case: %v", err)
	}
	if upd.Title != "更新案例" || len(upd.Images) != 2 {
		t.Fatalf("Update: case mismatch: %+v", upd)
	}
}

func TestRound2_PortfolioListDelete(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewPortfolioRepository()
	id := ug("cov-portfolio")
	if _, err := r.Create(context.Background(), domain.MemberPortfolio{
		ID: id, EnterpriseID: ug("cov-port-ent"), Name: "品牌展示", Status: "draft",
		Products: []string{"p-1"}, Honors: []string{"h-1"},
	}); err != nil {
		t.Fatalf("Create portfolio: %v", err)
	}
	if l, total, err := r.List(context.Background(), 0, 20); err != nil || total == 0 || !containsByID(l, id, func(v domain.MemberPortfolio) string { return v.ID }) {
		t.Fatalf("List: portfolio %s not found total=%d err=%v", id, total, err)
	}
	if err := r.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete portfolio: %v", err)
	}
	if _, err := r.FindByID(context.Background(), id); err == nil {
		t.Fatalf("Delete: expected not-found after delete")
	}
}

func TestRound2_ApplicationRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewApplicationRepository()
	userID := ug("cov-app-user")
	id := ug("cov-app")
	created, err := r.Create(context.Background(), domain.Application{
		ID: id, UserID: userID, ServiceID: "svc-1", ServiceName: "检测标定",
		OrderNo: ug("cov-app-no"), Status: "submitted", ApplyTime: "2026-08-20",
		FormData: map[string]any{"model": "M300"},
	})
	if err != nil {
		t.Fatalf("Create application: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: application id mismatch")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || got.FormData == nil || got.FormData["model"] != "M300" {
		t.Fatalf("FindByID: application roundtrip mismatch: %+v", got)
	}
	if l, total, err := r.ListByUser(context.Background(), userID, 0, 20); err != nil || total == 0 || !containsByID(l, id, func(v domain.Application) string { return v.ID }) {
		t.Fatalf("ListByUser: application %s not found total=%d err=%v", id, total, err)
	}
	if l, total, err := r.ListAll(context.Background(), 0, 20); err != nil || total == 0 || !containsByID(l, id, func(v domain.Application) string { return v.ID }) {
		t.Fatalf("ListAll: application %s not found total=%d err=%v", id, total, err)
	}
}

// ── biz_repos2: RDChallenge.Delete / ResearchProject.Delete ──

func TestRound2_RDChallengeDelete(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewRDChallengeRepository()
	id := ug("cov-rd")
	if _, err := r.Create(context.Background(), domain.RDChallenge{ID: id, PosterID: "u-1", Title: "难题", Status: "open"}); err != nil {
		t.Fatalf("Create challenge: %v", err)
	}
	if err := r.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete challenge: %v", err)
	}
	if _, err := r.FindByID(context.Background(), id); err == nil {
		t.Fatalf("Delete: expected not-found after delete")
	}
}

func TestRound2_ResearchProjectDelete(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewResearchProjectRepository()
	id := ug("cov-rp")
	if _, err := r.Create(context.Background(), domain.ResearchProject{ID: id, Title: "课题", Status: "planning", Members: []string{"m-1"}}); err != nil {
		t.Fatalf("Create project: %v", err)
	}
	if err := r.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete project: %v", err)
	}
	if _, err := r.FindByID(context.Background(), id); err == nil {
		t.Fatalf("Delete: expected not-found after delete")
	}
}

// ── biz_repos3: Competition.Delete / Event.Delete / Emergency 过滤+删除 ──

func TestRound2_CompetitionEventDelete(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	c := store.NewCompetitionRepository(nil)
	cid := ug("cov-comp")
	if _, err := c.Create(context.Background(), domain.Competition{ID: cid, Title: "比赛", Status: "draft"}); err != nil {
		t.Fatalf("Create competition: %v", err)
	}
	if err := c.Delete(context.Background(), cid); err != nil {
		t.Fatalf("Delete competition: %v", err)
	}

	e := store.NewEventRepository()
	eid := ug("cov-event")
	if _, err := e.Create(context.Background(), domain.AssociationEvent{ID: eid, Title: "活动", Status: "draft"}); err != nil {
		t.Fatalf("Create event: %v", err)
	}
	if err := e.Delete(context.Background(), eid); err != nil {
		t.Fatalf("Delete event: %v", err)
	}
	if _, err := e.FindByID(context.Background(), eid); err == nil {
		t.Fatalf("Delete event: expected not-found after delete")
	}
}

func TestRound2_EmergencyFilterAndDelete(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewEmergencyRepository()
	resID := ug("cov-er")
	if _, err := r.CreateResource(context.Background(), domain.EmergencyResource{
		ID: resID, OwnerID: "u-1", Name: "应急无人机", ResType: "drone", Specs: "M300",
		Quantity: 3, Location: "渝北区", ContactInfo: "张队", Status: "standby",
	}); err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	// ListResources：resType 过滤 + 关键词（WHERE + AND 组合分支）
	if l, total, err := r.ListResources(context.Background(), "drone", "无人机", 0, 20); err != nil || total == 0 || !containsByID(l, resID, func(v domain.EmergencyResource) string { return v.ID }) {
		t.Fatalf("ListResources(type+q): resource %s not found total=%d err=%v", resID, total, err)
	}
	if l, _, err := r.ListResources(context.Background(), "", "渝北", 0, 20); err != nil || !containsByID(l, resID, func(v domain.EmergencyResource) string { return v.ID }) {
		t.Fatalf("ListResources(q only): resource %s not found err=%v", resID, err)
	}
	// 非零 EndTime 的调度（覆盖 nullableEndTime 的 return t 分支）
	dID := ug("cov-ed")
	if _, err := r.CreateDispatch(context.Background(), domain.EmergencyDispatch{
		ID: dID, ResourceID: resID, EventDesc: "演练", Location: "渝北区",
		StartTime: time.Now(), EndTime: time.Now().Add(time.Hour), Commander: "张队",
		Result: "完成", Status: "completed",
	}); err != nil {
		t.Fatalf("CreateDispatch: %v", err)
	}
	if one, err := r.FindDispatchByID(context.Background(), dID); err != nil || one.EndTime.IsZero() {
		t.Fatalf("FindDispatchByID: dispatch mismatch err=%v end_time=%v", err, one.EndTime)
	}
	// 删除
	if err := r.DeleteDispatch(context.Background(), dID); err != nil {
		t.Fatalf("DeleteDispatch: %v", err)
	}
	if err := r.DeleteResource(context.Background(), resID); err != nil {
		t.Fatalf("DeleteResource: %v", err)
	}
	if _, err := r.FindResourceByID(context.Background(), resID); err == nil {
		t.Fatalf("DeleteResource: expected not-found after delete")
	}
}

// ── phase3_repos: Course.FindByID/Update/Delete + Instructor 全量 ──

func TestRound2_CourseRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewCourseRepository()
	id := ug("cov-course")
	created, err := r.Create(context.Background(), domain.TrainingCourse{
		ID: id, OrgID: ug("cov-course-org"), OrgName: "培训学校", Title: "无人机培训",
		CertType: domain.CertCAAC, Description: "考证", Location: "重庆", District: "渝北区",
		PriceFen: 100000, Rating: "4.8", ReviewCount: 10, DurationDays: 7, Image: "/c.jpg",
		Tags: []string{"无人机"}, Certificate: "/cert.png", BusinessHours: "9:00-18:00",
		Phone: "13800000000", Remain: 5, Environment: []string{"/e1.jpg"},
		CourseTypes: []string{"多旋翼"}, Status: "published",
	})
	if err != nil {
		t.Fatalf("Create course: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: course id mismatch")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || got.Title != "无人机培训" || len(got.Tags) != 1 || len(got.CourseTypes) != 1 {
		t.Fatalf("FindByID: course roundtrip mismatch: %+v", got)
	}
	got.Title = "更新课程"
	got.Status = "recruiting"
	upd, err := r.Update(context.Background(), got)
	if err != nil {
		t.Fatalf("Update course: %v", err)
	}
	if upd.Title != "更新课程" || upd.Version != 2 {
		t.Fatalf("Update: course mismatch: %+v", upd)
	}
	if err := r.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete course: %v", err)
	}
	if _, err := r.FindByID(context.Background(), id); err == nil {
		t.Fatalf("Delete: expected not-found after delete")
	}
}

func TestRound2_InstructorRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewInstructorRepository()
	id := ug("cov-instructor")
	created, err := r.Create(context.Background(), domain.Instructor{
		ID: id, UserID: ug("cov-instr-user"), Name: "王教练", Photo: "/p.jpg",
		CertTypes: []string{"CAAC", "AOPA"}, Bio: "10年飞龄", OrgID: ug("cov-instr-org"), Status: "pending",
	})
	if err != nil {
		t.Fatalf("Create instructor: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: instructor id mismatch")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || len(got.CertTypes) != 2 || got.CertTypes[0] != "CAAC" {
		t.Fatalf("FindByID: instructor roundtrip mismatch: %+v", got)
	}
	if l, err := r.List(context.Background()); err != nil || !containsByID(l, id, func(v domain.Instructor) string { return v.ID }) {
		t.Fatalf("List: instructor %s not found err=%v", id, err)
	}
	if upd, err := r.UpdateStatus(context.Background(), id, "approved"); err != nil || upd.Status != "approved" {
		t.Fatalf("UpdateStatus: status=%q err=%v", upd.Status, err)
	}
}

// ── phase3_repos: Pilot（含 cipher 分支 + Update） ──

func TestRound2_PilotRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	cipher := round2Cipher(t)
	r := store.NewPilotRepository(cipher)
	id := ug("cov-pilot")
	idCard := "500101199001011234"
	created, err := r.Create(context.Background(), domain.CertifiedPilot{
		ID: id, UserID: ug("cov-pilot-user"), RealName: "飞手", IDCard: idCard,
		Avatar: "/a.jpg", Region: "重庆", CertIDs: []string{"c-1", "c-2"}, FlightHours: 100,
		Bio: "巡检", Rating: 4.5, CompletedJobs: 20, Status: "pending",
	})
	if err != nil {
		t.Fatalf("Create pilot: %v", err)
	}
	if created.IDCard != idCard {
		t.Fatalf("Create: id_card not roundtripped to plaintext: %q", created.IDCard)
	}
	// 库内应为密文
	var stored string
	if err := store.Pool().QueryRow(context.Background(),
		`SELECT id_card FROM certified_pilots WHERE id=$1`, id).Scan(&stored); err != nil {
		t.Fatalf("read raw id_card: %v", err)
	}
	if stored == idCard {
		t.Fatalf("Create: id_card stored in plaintext")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.IDCard != idCard || len(got.CertIDs) != 2 {
		t.Fatalf("FindByID: pilot roundtrip mismatch: %+v", got)
	}
	// Update（覆盖 pilotRepo.Update）
	got.RealName = "更新飞手"
	got.Status = "approved"
	upd, err := r.Update(context.Background(), got)
	if err != nil {
		t.Fatalf("Update pilot: %v", err)
	}
	if upd.RealName != "更新飞手" || upd.Status != "approved" || upd.IDCard != idCard {
		t.Fatalf("Update: pilot mismatch: %+v", upd)
	}
	if l, err := r.List(context.Background()); err != nil || !containsByID(l, id, func(v domain.CertifiedPilot) string { return v.ID }) {
		t.Fatalf("List: pilot %s not found err=%v", id, err)
	}
	if us, err := r.UpdateStatus(context.Background(), id, "rejected"); err != nil || us.Status != "rejected" {
		t.Fatalf("UpdateStatus: status=%q err=%v", us.Status, err)
	}
}

// ── phase3_repos: Product.FindByID/IncrementViews/Update/Delete ──

func TestRound2_ProductRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewProductRepository()
	id := ug("cov-prod")
	created, err := r.Create(context.Background(), domain.DroneProduct{
		ID: id, SellerID: ug("cov-prod-seller"), SellerName: "卖家", ProdType: domain.ProductDrone,
		Title: "M300 RTK", Description: "行业机", PriceFen: 200000, Images: []string{"/d1.jpg", "/d2.jpg"},
		Brand: "DJI", Model: "M300", Condition: "new", Views: 0, Status: "listed",
	})
	if err != nil {
		t.Fatalf("Create product: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: product id mismatch")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || len(got.Images) != 2 || got.ProdType != domain.ProductDrone {
		t.Fatalf("FindByID: product roundtrip mismatch: %+v", got)
	}
	if err := r.IncrementViews(context.Background(), id); err != nil {
		t.Fatalf("IncrementViews: %v", err)
	}
	after, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID after IncrementViews: %v", err)
	}
	if after.Views != 1 {
		t.Fatalf("IncrementViews: views=%d, want 1", after.Views)
	}
	got.Title = "更新商品"
	got.Status = "sold"
	upd, err := r.Update(context.Background(), got)
	if err != nil {
		t.Fatalf("Update product: %v", err)
	}
	if upd.Title != "更新商品" || upd.Status != "sold" {
		t.Fatalf("Update: product mismatch: %+v", upd)
	}
	if err := r.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete product: %v", err)
	}
	if _, err := r.FindByID(context.Background(), id); err == nil {
		t.Fatalf("Delete: expected not-found after delete")
	}
}

// ── phase3_repos2: Review.ListByTarget 扫描循环 / Venue 全量 / TradeOrder 全量 ──

func TestRound2_ReviewListByTarget(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewReviewRepository()
	id := ug("cov-review")
	if _, err := r.Create(context.Background(), domain.Review{
		ID: id, ReviewerID: "u-1", TargetType: "enterprise", TargetID: "ent-1",
		Rating: 5, Content: "好评", Status: "approved",
	}); err != nil {
		t.Fatalf("Create review: %v", err)
	}
	if l, err := r.ListByTarget(context.Background(), "enterprise", "ent-1"); err != nil || !containsByID(l, id, func(v domain.Review) string { return v.ID }) {
		t.Fatalf("ListByTarget: review %s not found err=%v", id, err)
	}
}

func TestRound2_VenueRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewVenueRepository()
	id := ug("cov-venue")
	created, err := r.Create(context.Background(), domain.Venue{
		ID: id, OwnerID: ug("cov-venue-owner"), Name: "飞场", VenueType: "training_ground",
		Location: "重庆", PriceFen: 50000, Status: "available",
	})
	if err != nil {
		t.Fatalf("Create venue: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: venue id mismatch")
	}
	if got, err := r.FindByID(context.Background(), id); err != nil || got.ID != id || got.VenueType != "training_ground" {
		t.Fatalf("FindByID: venue mismatch: %+v err=%v", got, err)
	}
	if l, err := r.List(context.Background(), "training_ground"); err != nil || !containsByID(l, id, func(v domain.Venue) string { return v.ID }) {
		t.Fatalf("List(type): venue %s not found err=%v", id, err)
	}
	if l, err := r.List(context.Background(), ""); err != nil || !containsByID(l, id, func(v domain.Venue) string { return v.ID }) {
		t.Fatalf("List(all): venue %s not found err=%v", id, err)
	}
	bid := ug("cov-venue-booking")
	if _, err := r.CreateBooking(context.Background(), domain.VenueBooking{
		ID: bid, VenueID: id, UserID: ug("cov-venue-user"), StartTime: time.Now(),
		EndTime: time.Now().Add(time.Hour), Status: "confirmed",
	}); err != nil {
		t.Fatalf("CreateBooking: %v", err)
	}
	if l, err := r.ListBookings(context.Background(), id); err != nil || !containsByID(l, bid, func(v domain.VenueBooking) string { return v.ID }) {
		t.Fatalf("ListBookings: booking %s not found err=%v", bid, err)
	}
}

func TestRound2_TradeOrderRepo(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	r := store.NewTradeOrderRepository()
	buyer := ug("cov-to-buyer")
	id := ug("cov-to")
	created, err := r.Create(context.Background(), domain.TradeOrder{
		ID: id, ProductID: ug("cov-to-prod"), BuyerID: buyer, SellerID: ug("cov-to-seller"),
		AmountFen: 200000, Status: "pending",
	})
	if err != nil {
		t.Fatalf("Create trade order: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create: trade order id mismatch")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || got.Status != "pending" || got.AmountFen != 200000 {
		t.Fatalf("FindByID: trade order mismatch: %+v", got)
	}
	if upd, err := r.UpdateStatus(context.Background(), id, "paid"); err != nil || upd.Status != "paid" {
		t.Fatalf("UpdateStatus: status=%q err=%v", upd.Status, err)
	}
	// UpdateAftersale（一次性写售后字段）
	got.Status = "aftersale"
	got.AftersaleType = "refund"
	got.AftersaleReason = "质量问题"
	got.AftersaleDesc = "无法起飞"
	got.AftersaleAmountFen = 200000
	got.AftersaleStatus = "pending"
	got.AftersaleTime = time.Now()
	aft, err := r.UpdateAftersale(context.Background(), got)
	if err != nil {
		t.Fatalf("UpdateAftersale: %v", err)
	}
	if aft.AftersaleType != "refund" || aft.AftersaleStatus != "pending" || aft.Status != "aftersale" {
		t.Fatalf("UpdateAftersale: order mismatch: %+v", aft)
	}
	if l, err := r.ListByUser(context.Background(), buyer); err != nil || !containsByID(l, id, func(v domain.TradeOrder) string { return v.ID }) {
		t.Fatalf("ListByUser: order %s not found err=%v", id, err)
	}
	if l, total, err := r.ListAll(context.Background(), 0, 20); err != nil || total == 0 || !containsByID(l, id, func(v domain.TradeOrder) string { return v.ID }) {
		t.Fatalf("ListAll: order %s not found total=%d err=%v", id, total, err)
	}
	if err := r.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete trade order: %v", err)
	}
	if _, err := r.FindByID(context.Background(), id); err == nil {
		t.Fatalf("Delete: expected not-found after delete")
	}
}

// ── migration.go：错误分支（读目录失败 + 必错迁移回滚） ──

func TestRound2_MigrationErrors(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	ctx := context.Background()
	// 1) 不存在的目录 → read migrations dir 错误
	if err := store.RunMigrationsFromDir(ctx, filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatalf("RunMigrationsFromDir: expected read dir error, got nil")
	}
	// 2) 临时目录放一个未登记的必错迁移 → applyMigration 回滚并报错
	badDir := t.TempDir()
	version := fmt.Sprintf("998%d", time.Now().UnixNano())
	if err := os.WriteFile(filepath.Join(badDir, version+"_bad.up.sql"), []byte("INSERT INTO nonexistent_table_probe VALUES (1);"), 0o644); err != nil {
		t.Fatalf("write bad migration: %v", err)
	}
	if err := store.RunMigrationsFromDir(ctx, badDir); err == nil {
		t.Fatalf("RunMigrationsFromDir: expected run migration error, got nil")
	}
}

// ── postgres.go：cipher 分支（Demand 联系方式 + Enterprise 证照/账户加密落库解密回明文） ──

func TestRound2_DemandCipher(t *testing.T) {
	store := setupCipherStore(t)
	if store == nil {
		return
	}
	r := store.NewDemandRepository()
	id := ug("cov-demand-cipher")
	contact := "13800001234"
	if _, err := r.Create(context.Background(), domain.Demand{
		ID: id, PublisherID: ug("cov-dc-pub"), Title: "加密需求", Contact: contact,
		BizType: domain.BizCableInspection, Status: domain.DemandPublished,
	}); err != nil {
		t.Fatalf("Create demand: %v", err)
	}
	var stored string
	if err := store.Pool().QueryRow(context.Background(),
		`SELECT contact FROM demands WHERE id=$1`, id).Scan(&stored); err != nil {
		t.Fatalf("read raw contact: %v", err)
	}
	if stored == contact {
		t.Fatalf("Create: contact stored in plaintext")
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Contact != contact {
		t.Fatalf("FindByID: contact not decrypted: got %q want %q", got.Contact, contact)
	}
	// Update（覆盖 demandRepo.Update 的 cipher 加密分支）
	upd, err := r.Update(context.Background(), got)
	if err != nil {
		t.Fatalf("Update demand: %v", err)
	}
	if upd.Contact != contact {
		t.Fatalf("Update: contact not roundtripped: got %q want %q", upd.Contact, contact)
	}
	if us, err := r.SetStatus(context.Background(), id, domain.DemandCompleted); err != nil || us.Contact != contact {
		t.Fatalf("SetStatus: contact=%q err=%v", us.Contact, err)
	}
}

func TestRound2_EnterpriseCipher(t *testing.T) {
	store := setupCipherStore(t)
	if store == nil {
		return
	}
	r := store.NewEnterpriseRepository()
	id := ug("cov-ent-cipher")
	owner := ug("cov-entc-owner")
	license := "license-url-123"
	account := "account-name-123"
	if _, err := r.Create(context.Background(), domain.Enterprise{
		ID: id, OwnerUserID: owner, Name: "加密企业", LicenseURL: license, AccountName: account,
		Status: domain.EnterpriseSubmitted,
	}); err != nil {
		t.Fatalf("Create enterprise: %v", err)
	}
	var storedLicense, storedAccount string
	if err := store.Pool().QueryRow(context.Background(),
		`SELECT license_url, account_name FROM enterprises WHERE id=$1`, id).Scan(&storedLicense, &storedAccount); err != nil {
		t.Fatalf("read raw enterprise: %v", err)
	}
	if storedLicense == license || storedAccount == account {
		t.Fatalf("Create: enterprise PII stored in plaintext: license=%q account=%q", storedLicense, storedAccount)
	}
	got, err := r.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.LicenseURL != license || got.AccountName != account {
		t.Fatalf("FindByID: PII not decrypted: license=%q account=%q", got.LicenseURL, got.AccountName)
	}
	// Search / Pending / FindByOwner（scanEnterprises 解密分支）
	if l, err := r.Search(context.Background(), "加密企业"); err != nil || !containsByID(l, id, func(v domain.Enterprise) string { return v.ID }) {
		t.Fatalf("Search: enterprise %s not found err=%v", id, err)
	}
	if l, err := r.Pending(context.Background()); err != nil || !containsByID(l, id, func(v domain.Enterprise) string { return v.ID }) {
		t.Fatalf("Pending: enterprise %s not found err=%v", id, err)
	}
	if l, err := r.FindByOwner(context.Background(), owner); err != nil || !containsByID(l, id, func(v domain.Enterprise) string { return v.ID }) {
		t.Fatalf("FindByOwner: enterprise %s not found err=%v", id, err)
	}
	// Update（加密分支 + 更新路径）
	got.Name = "更新加密企业"
	got.LicenseURL = "license-new"
	got.AccountName = "account-new"
	upd, err := r.Update(context.Background(), id, got)
	if err != nil {
		t.Fatalf("Update enterprise: %v", err)
	}
	if upd.Name != "更新加密企业" {
		t.Fatalf("Update: name mismatch: %q", upd.Name)
	}
	if after, err := r.FindByID(context.Background(), id); err != nil || after.LicenseURL != "license-new" || after.AccountName != "account-new" {
		t.Fatalf("Update: PII roundtrip mismatch: %+v err=%v", after, err)
	}
}
