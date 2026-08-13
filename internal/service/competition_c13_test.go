package service_test

import (
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// TestCompetitionRegExtendedFields: 回归 C13——赛事报名的参赛人姓名/手机号/身份证/
// 证件照/身份证影像必须落库。旧实现 Register 只收 team_name/member_count/contact_info，
// 小程序 register.vue 提交的实名与证件字段被静默丢弃。
func TestCompetitionRegExtendedFields(t *testing.T) {
	svc := service.NewCompetitionService(memory.NewCompetitionRepository(nil))
	c, err := svc.Create(domain.Competition{
		Title: "竞速赛", Category: "racing", Description: "FPV竞速", Location: "巴南", Sponsor: "协会",
		StartDate: time.Now().AddDate(0, 1, 0), EndDate: time.Now().AddDate(0, 1, 3), MaxTeams: 50,
	})
	if err != nil {
		t.Fatal(err)
	}
	reg, err := svc.Register(c.ID, "user-1", "闪电队", 3, "13800000000", "张三", "13800000000",
		"500101199001011234", "/uploads/photo-a.jpg", "/uploads/idcard-a.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if reg.Name != "张三" || reg.Phone != "13800000000" || reg.IDCard != "500101199001011234" ||
		reg.PhotoURL != "/uploads/photo-a.jpg" || reg.IDCardImage != "/uploads/idcard-a.jpg" {
		t.Fatalf("reg=%+v, want extended fields persisted", reg)
	}
	// ListRegs 返回同样字段
	regs, err := svc.ListRegs(c.ID)
	if err != nil || len(regs) != 1 {
		t.Fatalf("list regs: %d, err=%v, want 1", len(regs), err)
	}
	if regs[0].Name != "张三" || regs[0].IDCard != "500101199001011234" || regs[0].IDCardImage != "/uploads/idcard-a.jpg" {
		t.Fatalf("ListRegs lost extended fields: %+v", regs[0])
	}
	// 赛事不存在 → 报错
	if _, err := svc.Register("comp-nope", "user-1", "闪电队", 3, "138", "", "", "", "", ""); err == nil {
		t.Fatal("register to missing competition should fail")
	}
}
