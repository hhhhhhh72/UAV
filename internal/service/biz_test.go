package service_test

import (
	"context"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

func TestExpertCRUD(t *testing.T) {
	svc := service.NewExpertService(memory.NewExpertRepository())
	e, err := svc.Create(context.Background(), "张三", "教授", "重庆大学", "无人机", "研究方向:无人机产业", "", "active", []string{"CAAC"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(context.Background(), e.ID); err != nil {
		t.Fatal(err)
	}
	if list, err := svc.List(context.Background(), ""); err != nil || len(list) != 1 {
		t.Fatalf("list: %v, %d", err, len(list))
	}
	if err := svc.Delete(context.Background(), e.ID); err != nil {
		t.Fatal(err)
	}
}

func TestCaseCRUD(t *testing.T) {
	svc := service.NewCaseService(memory.NewCaseRepository())
	c, err := svc.Create(context.Background(), "无人机物流案例", "logistics", "配送方案", []string{"img.jpg"}, "XX物流", "降本30%")
	if err != nil {
		t.Fatal(err)
	}
	if _, total, err := svc.List(context.Background(), "logistics", 1, 20); err != nil || total != 1 {
		t.Fatalf("list fail")
	}
	_ = c
}

func TestComplianceCRUD(t *testing.T) {
	svc := service.NewComplianceService(memory.NewComplianceRepository())
	svc.CreateDoc(context.Background(), "适航条例", "policy", "民航局", "2026-01-01", "draft", "内容", "", []string{"适航"})
	svc.CreateStandard(context.Background(), "团体标准", "group", "T/CDA-001", "协会", "2026-07-01", "draft", "标准内容", "")
	if _, total, err := svc.ListDocs(context.Background(), "", 1, 20); err != nil || total != 1 {
		t.Fatalf("list docs fail")
	}
}

func TestReportCRUD(t *testing.T) {
	svc := service.NewReportService(memory.NewIndustryReportRepository())
	r, err := svc.Create(context.Background(), "无人机产业报告", "2026H1", "行业", "摘要", "全文", "", "协会", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, total, err := svc.List(context.Background(), 1, 20); err != nil || total != 1 {
		t.Fatalf("list fail")
	}
	_ = r
}

func TestPortfolioCRUD(t *testing.T) {
	svc := service.NewPortfolioService(memory.NewPortfolioRepository())
	p, err := svc.Create(context.Background(), "ent-001", "巡航科技", "logo.png", "cover.png", "无人机方案商", "138", []string{"巡检"}, []string{"优秀"}, "")
	if err != nil {
		t.Fatal(err)
	}
	_ = p
}

func TestAchievementCRUD(t *testing.T) {
	svc := service.NewAchievementService(memory.NewAchievementRepository())
	a, err := svc.Create(context.Background(), "user-1", "AI避障算法", "patent", "自动避障", "无人机", "lab", "138", []string{"diagram.jpg"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, total, err := svc.List(context.Background(), "无人机", 1, 20); err != nil || total != 1 {
		t.Fatalf("list fail")
	}
	_ = a
}

func TestRDChallengeCRUD(t *testing.T) {
	svc := service.NewRDChallengeService(memory.NewRDChallengeRepository())
	c, err := svc.Create(context.Background(), "ent-001", "长续航电池", "电池", ">2小时轻量方案", 500000, time.Now().AddDate(0, 3, 0), "")
	if err != nil {
		t.Fatal(err)
	}
	if _, total, err := svc.List(context.Background(), "", 1, 20); err != nil || total != 1 {
		t.Fatalf("list fail")
	}
	_ = c
}

func TestResearchProjCRUD(t *testing.T) {
	svc := service.NewResearchProjectService(memory.NewResearchProjectRepository())
	p, err := svc.Create(context.Background(), "航路规划", "空域", "低空航路", "重庆大学", "Q1调研", []string{"巡航科技"}, 200000, time.Now(), time.Now().AddDate(1, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if _, total, err := svc.List(context.Background(), 1, 20); err != nil || total != 1 {
		t.Fatalf("list fail")
	}
	_ = p
}

func TestProjectAppCRUD(t *testing.T) {
	svc := service.NewProjectAppService(memory.NewProjectAppRepository())
	a, err := svc.Create(context.Background(), "user-1", "示范项目", "示范", "无人机产业示范", 1000000, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Review(context.Background(), a.ID, "同意立项", "approve"); err != nil {
		t.Fatal(err)
	}
}

func TestCompetitionCRUD(t *testing.T) {
	svc := service.NewCompetitionService(memory.NewCompetitionRepository(nil))
	c, err := svc.Create(context.Background(), domain.Competition{
		Title: "竞速赛", Category: "racing", Description: "FPV竞速", Location: "巴南", Sponsor: "协会",
		StartDate: time.Now().AddDate(0, 1, 0), EndDate: time.Now().AddDate(0, 1, 3), MaxTeams: 50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, total, err := svc.List(context.Background(), 1, 20); err != nil || total != 1 {
		t.Fatalf("list fail")
	}
	svc.Register(context.Background(), c.ID, "user-1", "闪电队", 3, "138", "张三", "13800000000", "500101199001011234", "/uploads/photo.jpg", "/uploads/idcard.jpg")
}

func TestEventCRUD(t *testing.T) {
	svc := service.NewEventService(memory.NewEventRepository())
	e, err := svc.Create(context.Background(), "产业论坛", "forum", "年度论坛", "博览中心", "", time.Now().AddDate(0, 2, 0), time.Now().AddDate(0, 2, 1), 500, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, total, err := svc.List(context.Background(), 1, 20); err != nil || total != 1 {
		t.Fatalf("list fail")
	}
	svc.Register(context.Background(), e.ID, "user-1", "张三", "138", "巡航科技")
}

func TestEmergencyCRUD(t *testing.T) {
	svc := service.NewEmergencyService(memory.NewEmergencyRepository())
	r, err := svc.CreateResource(context.Background(), "user-1", "应急无人机01", "drone", "M300RTK+热成像", "南岸", "138", 2, "")
	if err != nil {
		t.Fatal(err)
	}
	svc.CreateDispatch(context.Background(), r.ID, "山火应急", "北碚", "张指挥", "成功灭火", "", time.Now(), time.Now().Add(3*time.Hour))
}
