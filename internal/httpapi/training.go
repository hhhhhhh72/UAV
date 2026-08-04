package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
)

// ---- Certificates ----

// POST /api/v1/certificates
func (s *Server) addCertificate(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		CertType   string    `json:"cert_type"`
		CertNumber string    `json:"cert_number"`
		Level      string    `json:"level"`
		IssuerOrg  string    `json:"issuer_org"`
		IssueDate  time.Time `json:"issue_date"`
		ExpireDate time.Time `json:"expire_date"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c, err := s.trainingSvc.AddCertificate(a, domain.CertType(in.CertType), in.CertNumber, in.Level, in.IssuerOrg, in.IssueDate, in.ExpireDate)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "add_cert", "certificate", c.ID, "created")
	respond(w, r, http.StatusCreated, c)
}

// POST /api/v1/admin/certificates/{id}/approve
func (s *Server) approveCertificate(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	c, err := s.trainingSvc.ApproveCertificate(a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, c)
}

// GET /api/v1/certificates/mine
func (s *Server) listMyCertificates(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	certs, err := s.trainingSvc.ListMyCertificates(a)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, certs)
}

// ---- Courses ----

// POST /api/v1/training-courses
func (s *Server) createCourse(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title       string `json:"title"`
		CertType    string `json:"cert_type"`
		Description string `json:"description"`
		Location    string `json:"location"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		MaxStudents int    `json:"max_students"`
		PriceFen    int64  `json:"price_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	startDate := domain.ParseTime(in.StartDate)
	endDate := domain.ParseTime(in.EndDate)
	c, err := s.trainingSvc.CreateCourse(a, in.Title, domain.CertType(in.CertType), in.Description, in.Location, startDate, endDate, in.MaxStudents, in.PriceFen)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, c)
}

// GET /api/v1/training-courses
func (s *Server) listCourses(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	status := r.URL.Query().Get("status")
	courses, err := s.trainingSvc.ListCourses()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// filter
	var out []domain.TrainingCourse
	for _, c := range courses {
		if keyword != "" && !strings.Contains(c.Title, keyword) {
			continue
		}
		if status != "" && c.Status != status {
			continue
		}
		out = append(out, c)
	}
	paginatedRespond(w, r, out, len(out))
}

// ---- Instructors ----

// POST /api/v1/instructors
func (s *Server) registerInstructor(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Name      string   `json:"name"`
		Bio       string   `json:"bio"`
		OrgID     string   `json:"org_id"`
		CertTypes []string `json:"cert_types"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	i, err := s.trainingSvc.RegisterInstructor(a, in.Name, in.Bio, in.OrgID, in.CertTypes)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, i)
}

// POST /api/v1/admin/instructors/{id}/approve
func (s *Server) approveInstructor(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	i, err := s.trainingSvc.ApproveInstructor(a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, i)
}

// GET /api/v1/instructors
func (s *Server) listInstructors(w http.ResponseWriter, r *http.Request) {
	instructors, err := s.trainingSvc.ListInstructors()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, instructors)
}

// ---- Certified Pilots ----

// POST /api/v1/certified-pilots
func (s *Server) registerPilot(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		RealName    string `json:"real_name"`
		IDCard      string `json:"id_card"`
		FlightHours int    `json:"flight_hours"`
		Bio         string `json:"bio"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if in.RealName == "" || in.IDCard == "" {
		fail(w, r, http.StatusBadRequest, errors.New("real_name and id_card are required"))
		return
	}
	p, err := s.trainingSvc.RegisterPilot(a, in.RealName, in.IDCard, in.FlightHours, in.Bio)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, p)
}

// POST /api/v1/admin/certified-pilots/{id}/approve
func (s *Server) approvePilot(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	p, err := s.trainingSvc.ApprovePilot(a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	if p.IDCard != "" {
		p.IDCard = crypto.MaskIDCard(p.IDCard)
	}
	respond(w, r, http.StatusOK, p)
}

// GET /api/v1/certified-pilots — 公开名录：仅已认证（approved）飞手，身份证脱敏
func (s *Server) listPilots(w http.ResponseWriter, r *http.Request) {
	pilots, err := s.trainingSvc.ListPilots()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// 公开名录只显示已认证（待审/驳回不入名录）
	approved := make([]domain.CertifiedPilot, 0, len(pilots))
	for _, p := range pilots {
		if p.Status == "approved" {
			approved = append(approved, p)
		}
	}
	pilots = approved
	// 关键词过滤（姓名）
	if kw := strings.TrimSpace(r.URL.Query().Get("keyword")); kw != "" {
		filtered := make([]domain.CertifiedPilot, 0, len(pilots))
		for _, p := range pilots {
			if strings.Contains(p.RealName, kw) {
				filtered = append(filtered, p)
			}
		}
		pilots = filtered
	}
	// 脱敏身份证号
	for i := range pilots {
		if pilots[i].IDCard != "" {
			pilots[i].IDCard = crypto.MaskIDCard(pilots[i].IDCard)
		}
	}
	respond(w, r, http.StatusOK, pilots)
}

// GET /api/v1/admin/certified-pilots — 管理端全量（含待审，身份证完整可见供审核核对）
func (s *Server) listAdminPilots(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	pilots, err := s.trainingSvc.ListPilots()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	page, pageSize := paginationFromQuery(r)
	status := r.URL.Query().Get("status")
	if status == "all" {
		status = ""
	}
	filtered, total := adminListFilter(pilots, r.URL.Query().Get("keyword"), status,
		func(p domain.CertifiedPilot) string { return p.RealName },
		func(p domain.CertifiedPilot) string { return p.Status })
	paginatedRespond(w, r, adminSlicePage(filtered, page, pageSize), total)
}

// POST /api/v1/admin/certified-pilots/{id}/reject — 驳回飞手认证
func (s *Server) rejectPilot(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	p, err := s.trainingSvc.RejectPilot(a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, p)
}

// GET /api/v1/certified-pilots/mine — 我的飞手认证状态
func (s *Server) getMyPilot(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	p, err := s.trainingSvc.GetPilotByOwner(a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	if p.ID == "" {
		respond(w, r, http.StatusOK, nil)
		return
	}
	if p.IDCard != "" {
		p.IDCard = crypto.MaskIDCard(p.IDCard)
	}
	respond(w, r, http.StatusOK, p)
}
