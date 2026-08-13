package service

import (
	"fmt"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ── RescueCase Service ──

// rescueEventTypeAliases: 救援案例事件类型英文键 → 中文规范值。
// 存量数据与小程序 tab 均为中文（山火/洪水/地震/搜救/其他），
// 英文键仅作 API 兼容别名（domain 注释中的 mountain_fire 等历史值）。
var rescueEventTypeAliases = map[string]string{
	"fire":          "山火",
	"mountain_fire": "山火",
	"flood":         "洪水",
	"earthquake":    "地震",
	"search_rescue": "搜救",
	"rescue":        "搜救",
	"other":         "其他",
}

// normalizeRescueEventType 将英文别名归一到中文规范值；未知值原样返回。
func normalizeRescueEventType(s string) string {
	s = strings.TrimSpace(s)
	if v, ok := rescueEventTypeAliases[strings.ToLower(s)]; ok {
		return v
	}
	return s
}

type RescueCaseService struct{ repo repository.RescueCaseRepository }

func NewRescueCaseService(r repository.RescueCaseRepository) *RescueCaseService {
	return &RescueCaseService{repo: r}
}
func (s *RescueCaseService) Create(title, eventType, location, droneModel, teamName, summary, result, lessons, source string, date time.Time) (domain.RescueCase, error) {
	rc := domain.RescueCase{ID: fmt.Sprintf("rc-%d", time.Now().UnixNano()),
		Title: title, EventType: normalizeRescueEventType(eventType), Location: location, Date: date,
		DroneModel: droneModel, TeamName: teamName, Summary: summary,
		Result: result, Lessons: lessons, Source: source,
		Status: "draft", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(rc)
}
func (s *RescueCaseService) List(eventType, q string, page, pageSize int) ([]domain.RescueCase, int, error) {
	return s.repo.List(normalizeRescueEventType(eventType), strings.TrimSpace(q), (page-1)*pageSize, pageSize)
}
func (s *RescueCaseService) Get(id string) (domain.RescueCase, error) {
	return s.repo.FindByID(id)
}

// ── EmergencyDept Service ──

type EmergencyDeptService struct{ repo repository.EmergencyDeptRepository }

func NewEmergencyDeptService(r repository.EmergencyDeptRepository) *EmergencyDeptService {
	return &EmergencyDeptService{repo: r}
}
func (s *EmergencyDeptService) CreateDept(name, deptType, region, contactName, contactPhone, protocolURL string) (domain.EmergencyDept, error) {
	d := domain.EmergencyDept{ID: fmt.Sprintf("dept-%d", time.Now().UnixNano()),
		Name: name, DeptType: deptType, Region: region, ContactName: contactName,
		ContactPhone: contactPhone, ProtocolURL: protocolURL,
		Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.CreateDept(d)
}
func (s *EmergencyDeptService) ListDepts() ([]domain.EmergencyDept, error) {
	return s.repo.ListDepts()
}
func (s *EmergencyDeptService) CreateDrill(deptID, title, scenario string, date time.Time, participants, droneCount int, result string) (domain.EmergencyDrill, error) {
	d := domain.EmergencyDrill{ID: fmt.Sprintf("drill-%d", time.Now().UnixNano()),
		DeptID: deptID, Title: title, Scenario: scenario, Date: date,
		Participants: participants, DroneCount: droneCount, Result: result, CreatedAt: time.Now()}
	return s.repo.CreateDrill(d)
}
func (s *EmergencyDeptService) ListDrills(deptID string) ([]domain.EmergencyDrill, error) {
	return s.repo.ListDrills(deptID)
}

// ── AssociationMember Service ──

type AssociationMemberService struct{ repo repository.AssociationMemberRepository }

func NewAssociationMemberService(r repository.AssociationMemberRepository) *AssociationMemberService {
	return &AssociationMemberService{repo: r}
}
func (s *AssociationMemberService) AddMember(userID, enterpriseID string, role domain.AssociationRole) (domain.AssociationMember, error) {
	m := domain.AssociationMember{ID: fmt.Sprintf("am-%d", time.Now().UnixNano()),
		UserID: userID, EnterpriseID: enterpriseID, Role: role,
		JoinDate: time.Now(), ExpireDate: time.Now().AddDate(1, 0, 0),
		Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(m)
}
func (s *AssociationMemberService) ListMembers(role string, page, pageSize int) ([]domain.AssociationMember, int, error) {
	return s.repo.List(role, (page-1)*pageSize, pageSize)
}
func (s *AssociationMemberService) GetByUserID(userID string) (domain.AssociationMember, error) {
	return s.repo.FindByUserID(userID)
}
func (s *AssociationMemberService) UpdateRole(id string, role domain.AssociationRole) (domain.AssociationMember, error) {
	return s.repo.UpdateRole(id, role)
}
