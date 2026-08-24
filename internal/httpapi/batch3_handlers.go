package httpapi

import (
	"errors"
	"net/http"

	"drone-platform/internal/domain"
)

func (s *Server) registerBatch3Routes(mux *http.ServeMux) {
	// Rescue Cases
	mux.HandleFunc("POST /api/v1/admin/rescue-cases", s.createRescueCase)
	mux.HandleFunc("GET /api/v1/rescue-cases", s.listRescueCases)
	// Emergency Depts
	mux.HandleFunc("POST /api/v1/admin/emergency-depts", s.createEmergencyDept)
	mux.HandleFunc("GET /api/v1/emergency-depts", s.listEmergencyDepts)
	mux.HandleFunc("POST /api/v1/admin/emergency-drills", s.createEmergencyDrill)
	mux.HandleFunc("GET /api/v1/emergency-drills", s.listEmergencyDrills)
	// Association Members
	mux.HandleFunc("POST /api/v1/admin/association-members", s.addAssociationMember)
	mux.HandleFunc("GET /api/v1/association-members", s.listAssociationMembers)
	mux.HandleFunc("GET /api/v1/association-members/me", s.getMyAssociationRole)
}

// ── RescueCase ──
func (s *Server) createRescueCase(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title      string `json:"title"`
		EventType  string `json:"event_type"`
		Location   string `json:"location"`
		DroneModel string `json:"drone_model"`
		TeamName   string `json:"team_name"`
		Summary    string `json:"summary"`
		Result     string `json:"result"`
		Lessons    string `json:"lessons"`
		Source     string `json:"source"`
		Date       string `json:"date"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// P2 修复：严格解析日期——非法日期此前被 ParseTime 静默写成当前时间落库。
	date, ok := strictDate(w, r, in.Date)
	if !ok {
		return
	}
	rc, err := s.rescueCaseSvc.Create(r.Context(), in.Title, in.EventType, in.Location, in.DroneModel, in.TeamName, in.Summary, in.Result, in.Lessons, in.Source, date)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, rc)
}
func (s *Server) listRescueCases(w http.ResponseWriter, r *http.Request) {
	// 性能审查：repo 支持 event_type/q 过滤 → 分页下沉 SQL，respondPage 不再二次切片。
	// event_type 支持中文（山火/洪水/…）与英文别名（mountain_fire/…），服务层归一
	page, pageSize := paginationFromQuery(r)
	list, total, err := s.rescueCaseSvc.List(r.Context(), r.URL.Query().Get("event_type"), r.URL.Query().Get("q"), page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respondPage(w, r, list, total, page, pageSize)
}

// ── EmergencyDept ──
func (s *Server) createEmergencyDept(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name        string `json:"name"`
		DeptType    string `json:"dept_type"`
		Region      string `json:"region"`
		ContactName string `json:"contact_name"`
		ContactPhone string `json:"contact_phone"`
		ProtocolURL string `json:"protocol_url"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	d, err := s.emergDeptSvc.CreateDept(r.Context(), in.Name, in.DeptType, in.Region, in.ContactName, in.ContactPhone, in.ProtocolURL)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, d)
}
func (s *Server) listEmergencyDepts(w http.ResponseWriter, r *http.Request) {
	list, err := s.emergDeptSvc.ListDepts(r.Context())
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// 公开接口脱敏：联系方式不随公开列表外泄（仅管理端/协会成员可见原文）
	for i := range list {
		list[i].ContactPhone = maskPhone(list[i].ContactPhone)
	}
	respond(w, r, http.StatusOK, list)
}
func (s *Server) createEmergencyDrill(w http.ResponseWriter, r *http.Request) {
	var in struct {
		DeptID       string `json:"dept_id"`
		Title        string `json:"title"`
		Scenario     string `json:"scenario"`
		Date         string `json:"date"`
		Result       string `json:"result"`
		Participants int    `json:"participants"`
		DroneCount   int    `json:"drone_count"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// P2 修复：严格解析日期——非法日期此前被 ParseTime 静默写成当前时间落库。
	date, ok := strictDate(w, r, in.Date)
	if !ok {
		return
	}
	d, err := s.emergDeptSvc.CreateDrill(r.Context(), in.DeptID, in.Title, in.Scenario, date, in.Participants, in.DroneCount, in.Result)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, d)
}
func (s *Server) listEmergencyDrills(w http.ResponseWriter, r *http.Request) {
	list, err := s.emergDeptSvc.ListDrills(r.Context(), r.URL.Query().Get("dept_id"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, list)
}

// ── AssociationMember ──
func (s *Server) addAssociationMember(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID       string `json:"user_id"`
		EnterpriseID string `json:"enterprise_id"`
		Role         string `json:"role"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	m, err := s.assocMemberSvc.AddMember(r.Context(), in.UserID, in.EnterpriseID, domain.AssociationRole(in.Role))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, m)
}
func (s *Server) listAssociationMembers(w http.ResponseWriter, r *http.Request) {
	// 性能审查：repo 支持 role 过滤 → 分页下沉 SQL，respondPage 不再二次切片。
	page, pageSize := paginationFromQuery(r)
	list, total, err := s.assocMemberSvc.ListMembers(r.Context(), r.URL.Query().Get("role"), page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// P1 脱敏：公开响应返回前替换手机号注册用户的 user_id/enterprise_id，
	// 防止手机号被批量抓取（协会成员名录可匿名访问）。
	for i := range list {
		list[i].UserID = maskUserID(list[i].UserID)
		list[i].EnterpriseID = maskUserID(list[i].EnterpriseID)
	}
	respondPage(w, r, list, total, page, pageSize)
}
func (s *Server) getMyAssociationRole(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	m, err := s.assocMemberSvc.GetByUserID(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, m)
}
