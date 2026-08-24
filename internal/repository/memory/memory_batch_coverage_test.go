package memory_test

import (
	"context"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
)

// 以下 helper 统一断言输出格式：方法名 + got/want。
func mbErr(t *testing.T, method string, err error, wantErr bool) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Fatalf("%s: got nil error, want error", method)
		}
		return
	}
	if err != nil {
		t.Fatalf("%s: got error %v, want nil", method, err)
	}
}

func mbStr(t *testing.T, method, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got=%q want=%q", method, got, want)
	}
}

func mbInt(t *testing.T, method string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got=%d want=%d", method, got, want)
	}
}

func mbBool(t *testing.T, method string, got, want bool) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got=%v want=%v", method, got, want)
	}
}

// ── memory_batch1.go ──

func TestResourcePoolRepoCoverage(t *testing.T) {
	r := memory.NewResourcePoolRepository()

	p1, err := r.Create(context.Background(), domain.ResourcePool{ID: "pool-1", Name: "应急池", PoolType: "emergency"})
	mbErr(t, "Create", err, false)
	mbStr(t, "Create.ID", p1.ID, "pool-1")
	_, err = r.Create(context.Background(), domain.ResourcePool{ID: "pool-2", Name: "设备池", PoolType: "equipment"})
	mbErr(t, "Create2", err, false)

	f, err := r.FindByID(context.Background(), "pool-1")
	mbErr(t, "FindByID", err, false)
	mbStr(t, "FindByID.Name", f.Name, "应急池")
	_, err = r.FindByID(context.Background(), "pool-missing")
	mbErr(t, "FindByID(missing)", err, true)

	all, err := r.List(context.Background(), "")
	mbErr(t, "List(all)", err, false)
	mbInt(t, "List(all).len", len(all), 2)

	emergency, err := r.List(context.Background(), "emergency")
	mbErr(t, "List(emergency)", err, false)
	mbInt(t, "List(emergency).len", len(emergency), 1)
	mbStr(t, "List(emergency)[0].PoolType", emergency[0].PoolType, "emergency")

	noMatch, err := r.List(context.Background(), "team")
	mbErr(t, "List(team)", err, false)
	mbInt(t, "List(team).len", len(noMatch), 0)

	m1, err := r.AddMember(context.Background(), domain.ResourcePoolMember{ID: "m-1", PoolID: "pool-1", ResID: "res-1", ResType: "drone"})
	mbErr(t, "AddMember", err, false)
	mbStr(t, "AddMember.ID", m1.ID, "m-1")
	_, err = r.AddMember(context.Background(), domain.ResourcePoolMember{ID: "m-2", PoolID: "pool-2", ResID: "res-2", ResType: "equipment"})
	mbErr(t, "AddMember2", err, false)

	members, err := r.ListMembers(context.Background(), "pool-1")
	mbErr(t, "ListMembers", err, false)
	mbInt(t, "ListMembers.len", len(members), 1)
	mbStr(t, "ListMembers[0].ResID", members[0].ResID, "res-1")
}

func TestTestSiteRepoCoverage(t *testing.T) {
	r := memory.NewTestSiteRepository()

	_, err := r.Create(context.Background(), domain.TestSite{ID: "site-1", Name: "飞行场", SiteType: "flying_field"})
	mbErr(t, "Create", err, false)
	_, err = r.Create(context.Background(), domain.TestSite{ID: "site-2", Name: "实验室", SiteType: "lab"})
	mbErr(t, "Create2", err, false)

	f, err := r.FindByID(context.Background(), "site-1")
	mbErr(t, "FindByID", err, false)
	mbStr(t, "FindByID.Name", f.Name, "飞行场")
	_, err = r.FindByID(context.Background(), "site-missing")
	mbErr(t, "FindByID(missing)", err, true)

	all, err := r.List(context.Background(), "")
	mbErr(t, "List(all)", err, false)
	mbInt(t, "List(all).len", len(all), 2)
	ff, err := r.List(context.Background(), "flying_field")
	mbErr(t, "List(flying_field)", err, false)
	mbInt(t, "List(flying_field).len", len(ff), 1)

	now := time.Now()
	_, err = r.CreateBooking(context.Background(), domain.TestSiteBooking{ID: "bk-1", SiteID: "site-1", UserID: "u-1", CreatedAt: now})
	mbErr(t, "CreateBooking", err, false)
	_, err = r.CreateBooking(context.Background(), domain.TestSiteBooking{ID: "bk-2", SiteID: "site-1", UserID: "u-2", CreatedAt: now.Add(time.Hour)})
	mbErr(t, "CreateBooking2", err, false)
	_, err = r.CreateBooking(context.Background(), domain.TestSiteBooking{ID: "bk-3", SiteID: "site-2", UserID: "u-1", CreatedAt: now.Add(2 * time.Hour)})
	mbErr(t, "CreateBooking3", err, false)

	upd, err := r.UpdateBookingStatus(context.Background(), "bk-1", "approved", "通过")
	mbErr(t, "UpdateBookingStatus", err, false)
	mbStr(t, "UpdateBookingStatus.Status", upd.Status, "approved")
	mbStr(t, "UpdateBookingStatus.ReviewNote", upd.ReviewNote, "通过")
	_, err = r.UpdateBookingStatus(context.Background(), "bk-missing", "approved", "")
	mbErr(t, "UpdateBookingStatus(missing)", err, true)

	bks, err := r.ListBookings(context.Background(), "site-1")
	mbErr(t, "ListBookings(site-1)", err, false)
	mbInt(t, "ListBookings(site-1).len", len(bks), 2)
	bks2, err := r.ListBookings(context.Background(), "site-2")
	mbErr(t, "ListBookings(site-2)", err, false)
	mbInt(t, "ListBookings(site-2).len", len(bks2), 1)

	byUser, err := r.ListBookingsByUser(context.Background(), "u-1")
	mbErr(t, "ListBookingsByUser", err, false)
	mbInt(t, "ListBookingsByUser.len", len(byUser), 2)
	mbStr(t, "ListBookingsByUser[0].ID", byUser[0].ID, "bk-3") // 最新在前

	allBk, total, err := r.ListAllBookings(context.Background(), 0, 10)
	mbErr(t, "ListAllBookings", err, false)
	mbInt(t, "ListAllBookings.total", total, 3)
	mbInt(t, "ListAllBookings.len", len(allBk), 3)
	page, total2, err := r.ListAllBookings(context.Background(), 1, 1)
	mbErr(t, "ListAllBookings(page)", err, false)
	mbInt(t, "ListAllBookings(page).total", total2, 3)
	mbInt(t, "ListAllBookings(page).len", len(page), 1)
	empty, total3, err := r.ListAllBookings(context.Background(), 100, 10)
	mbErr(t, "ListAllBookings(overflow)", err, false)
	mbInt(t, "ListAllBookings(overflow).total", total3, 3)
	mbInt(t, "ListAllBookings(overflow).len", len(empty), 0)

	us, err := r.UpdateSite(context.Background(), domain.TestSite{ID: "site-1", Name: "飞行场更新", SiteType: "flying_field"})
	mbErr(t, "UpdateSite", err, false)
	mbStr(t, "UpdateSite.Name", us.Name, "飞行场更新")
	_, err = r.UpdateSite(context.Background(), domain.TestSite{ID: "site-missing"})
	mbErr(t, "UpdateSite(missing)", err, true)

	err = r.DeleteSite(context.Background(), "site-2")
	mbErr(t, "DeleteSite", err, false)
	err = r.DeleteSite(context.Background(), "site-missing")
	mbErr(t, "DeleteSite(missing)", err, true)
	_, err = r.FindByID(context.Background(), "site-2")
	mbErr(t, "FindByID(after delete)", err, true)
}

func TestExhibitionRepoCoverage(t *testing.T) {
	r := memory.NewExhibitionRepository()

	e1, err := r.Create(context.Background(), domain.Exhibition{ID: "expo-1", Title: "无人机展", Category: "drone_show"})
	mbErr(t, "Create", err, false)
	mbStr(t, "Create.ID", e1.ID, "expo-1")
	mbBool(t, "Create.CreatedAt set", e1.CreatedAt.IsZero(), false)
	mbBool(t, "Create.UpdatedAt set", e1.UpdatedAt.IsZero(), false)

	f, err := r.FindByID(context.Background(), "expo-1")
	mbErr(t, "FindByID", err, false)
	mbStr(t, "FindByID.Title", f.Title, "无人机展")
	_, err = r.FindByID(context.Background(), "expo-missing")
	mbErr(t, "FindByID(missing)", err, true)

	list, total, err := r.List(context.Background(), 0, 10)
	mbErr(t, "List", err, false)
	mbInt(t, "List.total", total, 1)
	mbInt(t, "List.len", len(list), 1)

	ue, err := r.Update(context.Background(), domain.Exhibition{ID: "expo-1", Title: "无人机展更新"})
	mbErr(t, "Update", err, false)
	mbStr(t, "Update.Title", ue.Title, "无人机展更新")
	mbBool(t, "Update.UpdatedAt set", ue.UpdatedAt.IsZero(), false)
	_, err = r.Update(context.Background(), domain.Exhibition{ID: "expo-missing"})
	mbErr(t, "Update(missing)", err, true)

	_, err = r.CreateBooth(context.Background(), domain.ExhibitionBooth{ID: "booth-1", ExhibitionID: "expo-1", BoothNumber: "A1"})
	mbErr(t, "CreateBooth", err, false)
	_, err = r.CreateBooth(context.Background(), domain.ExhibitionBooth{ID: "booth-2", ExhibitionID: "expo-1", BoothNumber: "A2"})
	mbErr(t, "CreateBooth2", err, false)

	booths, err := r.ListBooths(context.Background(), "expo-1")
	mbErr(t, "ListBooths", err, false)
	mbInt(t, "ListBooths.len", len(booths), 2)

	ub, err := r.UpdateBoothStatus(context.Background(), "booth-1", "approved")
	mbErr(t, "UpdateBoothStatus", err, false)
	mbStr(t, "UpdateBoothStatus.Status", ub.Status, "approved")
	_, err = r.UpdateBoothStatus(context.Background(), "booth-missing", "approved")
	mbErr(t, "UpdateBoothStatus(missing)", err, true)

	err = r.Delete(context.Background(), "expo-1")
	mbErr(t, "Delete", err, false)
	err = r.Delete(context.Background(), "expo-missing")
	mbErr(t, "Delete(missing)", err, true)
	_, err = r.FindByID(context.Background(), "expo-1")
	mbErr(t, "FindByID(after delete)", err, true)
}

// ── memory_batch2.go ──

func TestTransformationRepoCoverage(t *testing.T) {
	r := memory.NewTransformationRepository()

	_, err := r.Create(context.Background(), domain.Transformation{ID: "trans-1", OwnerID: "u-1", Title: "成果转化", Stage: domain.StageLab})
	mbErr(t, "Create", err, false)
	_, err = r.Create(context.Background(), domain.Transformation{ID: "trans-2", OwnerID: "u-2", Title: "成果转化2", Stage: domain.StagePilot})
	mbErr(t, "Create2", err, false)

	f, err := r.FindByID(context.Background(), "trans-1")
	mbErr(t, "FindByID", err, false)
	mbStr(t, "FindByID.Title", f.Title, "成果转化")
	_, err = r.FindByID(context.Background(), "trans-missing")
	mbErr(t, "FindByID(missing)", err, true)

	all, err := r.List(context.Background(), "")
	mbErr(t, "List(all)", err, false)
	mbInt(t, "List(all).len", len(all), 2)
	byOwner, err := r.List(context.Background(), "u-1")
	mbErr(t, "List(u-1)", err, false)
	mbInt(t, "List(u-1).len", len(byOwner), 1)
	mbStr(t, "List(u-1)[0].OwnerID", byOwner[0].OwnerID, "u-1")

	ut, err := r.Update(context.Background(), domain.Transformation{ID: "trans-1", OwnerID: "u-1", Title: "updated", Stage: domain.StagePilot})
	mbErr(t, "Update", err, false)
	mbStr(t, "Update.Title", ut.Title, "updated")
	mbStr(t, "Update.Stage", string(ut.Stage), string(domain.StagePilot))
	mbBool(t, "Update.UpdatedAt set", ut.UpdatedAt.IsZero(), false)
	_, err = r.Update(context.Background(), domain.Transformation{ID: "trans-missing"})
	mbErr(t, "Update(missing)", err, true)

	err = r.Delete(context.Background(), "trans-2")
	mbErr(t, "Delete", err, false)
	err = r.Delete(context.Background(), "trans-missing")
	mbErr(t, "Delete(missing)", err, true)
	_, err = r.FindByID(context.Background(), "trans-2")
	mbErr(t, "FindByID(after delete)", err, true)
}

func TestCollegeRepoCoverage(t *testing.T) {
	r := memory.NewCollegeRepository()

	_, err := r.Create(context.Background(), domain.College{ID: "col-1", Name: "西工大", Region: "陕西", Status: "active"})
	mbErr(t, "Create", err, false)
	_, err = r.Create(context.Background(), domain.College{ID: "col-2", Name: "北航", Region: "北京", Status: "active"})
	mbErr(t, "Create2", err, false)

	f, err := r.FindByID(context.Background(), "col-1")
	mbErr(t, "FindByID", err, false)
	mbStr(t, "FindByID.Name", f.Name, "西工大")
	_, err = r.FindByID(context.Background(), "col-missing")
	mbErr(t, "FindByID(missing)", err, true)

	all, err := r.List(context.Background(), "")
	mbErr(t, "List(all)", err, false)
	mbInt(t, "List(all).len", len(all), 2)
	byRegion, err := r.List(context.Background(), "陕西")
	mbErr(t, "List(陕西)", err, false)
	mbInt(t, "List(陕西).len", len(byRegion), 1)
	mbStr(t, "List(陕西)[0].Region", byRegion[0].Region, "陕西")

	uc, err := r.Update(context.Background(), domain.College{ID: "col-1", Name: "西北工业大学", Region: "陕西"})
	mbErr(t, "Update", err, false)
	mbStr(t, "Update.Name", uc.Name, "西北工业大学")
	_, err = r.Update(context.Background(), domain.College{ID: "col-missing"})
	mbErr(t, "Update(missing)", err, true)

	err = r.Delete(context.Background(), "col-2")
	mbErr(t, "Delete", err, false)
	err = r.Delete(context.Background(), "col-missing")
	mbErr(t, "Delete(missing)", err, true)
	_, err = r.FindByID(context.Background(), "col-2")
	mbErr(t, "FindByID(after delete)", err, true)
}

func TestStudyTourRepoCoverage(t *testing.T) {
	r := memory.NewStudyTourRepository()

	_, err := r.Create(context.Background(), domain.StudyTour{ID: "tour-1", Title: "研学游"})
	mbErr(t, "Create", err, false)
	_, err = r.Create(context.Background(), domain.StudyTour{ID: "tour-2", Title: "研学游2"})
	mbErr(t, "Create2", err, false)

	f, err := r.FindByID(context.Background(), "tour-1")
	mbErr(t, "FindByID", err, false)
	mbStr(t, "FindByID.Title", f.Title, "研学游")
	_, err = r.FindByID(context.Background(), "tour-missing")
	mbErr(t, "FindByID(missing)", err, true)

	all, err := r.List(context.Background())
	mbErr(t, "List", err, false)
	mbInt(t, "List.len", len(all), 2)

	ut, err := r.Update(context.Background(), domain.StudyTour{ID: "tour-1", Title: "updated"})
	mbErr(t, "Update", err, false)
	mbStr(t, "Update.Title", ut.Title, "updated")
	_, err = r.Update(context.Background(), domain.StudyTour{ID: "tour-missing"})
	mbErr(t, "Update(missing)", err, true)

	err = r.Delete(context.Background(), "tour-2")
	mbErr(t, "Delete", err, false)
	err = r.Delete(context.Background(), "tour-missing")
	mbErr(t, "Delete(missing)", err, true)
	_, err = r.FindByID(context.Background(), "tour-2")
	mbErr(t, "FindByID(after delete)", err, true)
}

func TestCooperationRepoCoverage(t *testing.T) {
	r := memory.NewCooperationRepository()

	_, err := r.Create(context.Background(), domain.CooperationProgram{ID: "coop-1", EnterpriseID: "ent-1", Title: "校企共建"})
	mbErr(t, "Create", err, false)
	_, err = r.Create(context.Background(), domain.CooperationProgram{ID: "coop-2", EnterpriseID: "ent-2", Title: "校企共建2"})
	mbErr(t, "Create2", err, false)

	f, err := r.FindByID(context.Background(), "coop-1")
	mbErr(t, "FindByID", err, false)
	mbStr(t, "FindByID.Title", f.Title, "校企共建")
	_, err = r.FindByID(context.Background(), "coop-missing")
	mbErr(t, "FindByID(missing)", err, true)

	all, err := r.List(context.Background(), "")
	mbErr(t, "List(all)", err, false)
	mbInt(t, "List(all).len", len(all), 2)
	byEnt, err := r.List(context.Background(), "ent-1")
	mbErr(t, "List(ent-1)", err, false)
	mbInt(t, "List(ent-1).len", len(byEnt), 1)
	mbStr(t, "List(ent-1)[0].EnterpriseID", byEnt[0].EnterpriseID, "ent-1")

	us, err := r.UpdateStatus(context.Background(), "coop-1", "active")
	mbErr(t, "UpdateStatus", err, false)
	mbStr(t, "UpdateStatus.Status", us.Status, "active")
	mbBool(t, "UpdateStatus.UpdatedAt set", us.UpdatedAt.IsZero(), false)
	_, err = r.UpdateStatus(context.Background(), "coop-missing", "active")
	mbErr(t, "UpdateStatus(missing)", err, true)
}

// ── memory_batch3.go ──

func TestRescueCaseRepoCoverage(t *testing.T) {
	r := memory.NewRescueCaseRepository()

	_, err := r.Create(context.Background(), domain.RescueCase{ID: "rc-1", Title: "山火救援", EventType: "mountain_fire", Location: "重庆", DroneModel: "M300", TeamName: "蓝天救援队", Summary: "无人机热成像搜索", Status: "published"})
	mbErr(t, "Create", err, false)
	_, err = r.Create(context.Background(), domain.RescueCase{ID: "rc-2", Title: "洪水救援", EventType: "flood", Location: "武汉", DroneModel: "M30", TeamName: "红十字", Summary: "抛投救生圈", Status: "published"})
	mbErr(t, "Create2", err, false)

	f, err := r.FindByID(context.Background(), "rc-1")
	mbErr(t, "FindByID", err, false)
	mbStr(t, "FindByID.Title", f.Title, "山火救援")
	_, err = r.FindByID(context.Background(), "rc-missing")
	mbErr(t, "FindByID(missing)", err, true)

	all, total, err := r.List(context.Background(), "", "", 0, 10)
	mbErr(t, "List(all)", err, false)
	mbInt(t, "List(all).total", total, 2)
	mbInt(t, "List(all).len", len(all), 2)

	byType, total2, err := r.List(context.Background(), "mountain_fire", "", 0, 10)
	mbErr(t, "List(mountain_fire)", err, false)
	mbInt(t, "List(mountain_fire).total", total2, 1)
	mbInt(t, "List(mountain_fire).len", len(byType), 1)
	mbStr(t, "List(mountain_fire)[0].EventType", byType[0].EventType, "mountain_fire")

	byQuery, _, err := r.List(context.Background(), "", "热成像", 0, 10)
	mbErr(t, "List(q=热成像)", err, false)
	mbInt(t, "List(q=热成像).len", len(byQuery), 1)
	mbStr(t, "List(q=热成像)[0].ID", byQuery[0].ID, "rc-1")

	noMatch, _, err := r.List(context.Background(), "", "不存在关键词", 0, 10)
	mbErr(t, "List(q=none)", err, false)
	mbInt(t, "List(q=none).len", len(noMatch), 0)

	page, _, err := r.List(context.Background(), "", "", 1, 1)
	mbErr(t, "List(page)", err, false)
	mbInt(t, "List(page).len", len(page), 1)
}

func TestEmergencyDeptRepoCoverage(t *testing.T) {
	r := memory.NewEmergencyDeptRepository()

	_, err := r.CreateDept(context.Background(), domain.EmergencyDept{ID: "dept-1", Name: "消防队", DeptType: "fire"})
	mbErr(t, "CreateDept", err, false)
	_, err = r.CreateDept(context.Background(), domain.EmergencyDept{ID: "dept-2", Name: "应急局", DeptType: "emergency_bureau"})
	mbErr(t, "CreateDept2", err, false)

	depts, err := r.ListDepts(context.Background())
	mbErr(t, "ListDepts", err, false)
	mbInt(t, "ListDepts.len", len(depts), 2)

	_, err = r.CreateDrill(context.Background(), domain.EmergencyDrill{ID: "drill-1", DeptID: "dept-1", Title: "联合演练"})
	mbErr(t, "CreateDrill", err, false)
	_, err = r.CreateDrill(context.Background(), domain.EmergencyDrill{ID: "drill-2", DeptID: "dept-2", Title: "联合演练2"})
	mbErr(t, "CreateDrill2", err, false)

	all, err := r.ListDrills(context.Background(), "")
	mbErr(t, "ListDrills(all)", err, false)
	mbInt(t, "ListDrills(all).len", len(all), 2)
	byDept, err := r.ListDrills(context.Background(), "dept-1")
	mbErr(t, "ListDrills(dept-1)", err, false)
	mbInt(t, "ListDrills(dept-1).len", len(byDept), 1)
	mbStr(t, "ListDrills(dept-1)[0].DeptID", byDept[0].DeptID, "dept-1")
}

func TestAssociationMemberRepoCoverage(t *testing.T) {
	r := memory.NewAssociationMemberRepository()

	_, err := r.Create(context.Background(), domain.AssociationMember{ID: "am-1", UserID: "u-1", Role: domain.AssocMember})
	mbErr(t, "Create", err, false)
	_, err = r.Create(context.Background(), domain.AssociationMember{ID: "am-2", UserID: "u-2", Role: domain.AssocPresident})
	mbErr(t, "Create2", err, false)

	f, err := r.FindByUserID(context.Background(), "u-1")
	mbErr(t, "FindByUserID", err, false)
	mbStr(t, "FindByUserID.ID", f.ID, "am-1")
	_, err = r.FindByUserID(context.Background(), "u-missing")
	mbErr(t, "FindByUserID(missing)", err, true)

	all, total, err := r.List(context.Background(), "", 0, 10)
	mbErr(t, "List(all)", err, false)
	mbInt(t, "List(all).total", total, 2)
	mbInt(t, "List(all).len", len(all), 2)

	byRole, total2, err := r.List(context.Background(), string(domain.AssocPresident), 0, 10)
	mbErr(t, "List(president)", err, false)
	mbInt(t, "List(president).total", total2, 1)
	mbInt(t, "List(president).len", len(byRole), 1)
	mbStr(t, "List(president)[0].Role", string(byRole[0].Role), string(domain.AssocPresident))

	ur, err := r.UpdateRole(context.Background(), "am-1", domain.AssocVicePresident)
	mbErr(t, "UpdateRole", err, false)
	mbStr(t, "UpdateRole.Role", string(ur.Role), string(domain.AssocVicePresident))
	mbBool(t, "UpdateRole.UpdatedAt set", ur.UpdatedAt.IsZero(), false)
	_, err = r.UpdateRole(context.Background(), "am-missing", domain.AssocPresident)
	mbErr(t, "UpdateRole(missing)", err, true)
}
