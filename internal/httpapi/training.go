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
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		CertType   string    `json:"cert_type"`
		CertNumber string    `json:"cert_number"`
		Level      string    `json:"level"`
		IssuerOrg  string    `json:"issuer_org"`
		IssueDate  time.Time `json:"issue_date"`
		ExpireDate time.Time `json:"expire_date"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	c, err := s.trainingSvc.AddCertificate(a, domain.CertType(in.CertType), in.CertNumber, in.Level, in.IssuerOrg, in.IssueDate, in.ExpireDate)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	s.audit(r.Context(), a.ID, "add_cert", "certificate", c.ID, "created")
	respond(w, r, http.StatusCreated, c)
}

// POST /api/v1/admin/certificates/{id}/approve
func (s *Server) approveCertificate(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	c, err := s.trainingSvc.ApproveCertificate(a, r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusOK, c)
}

// GET /api/v1/certificates/mine
func (s *Server) listMyCertificates(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	certs, err := s.trainingSvc.ListMyCertificates(a)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, certs)
}

// ---- Courses ----

// POST /api/v1/training-courses
func (s *Server) createCourse(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
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
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	startDate := domain.ParseTime(in.StartDate)
	endDate := domain.ParseTime(in.EndDate)
	c, err := s.trainingSvc.CreateCourse(a, in.Title, domain.CertType(in.CertType), in.Description, in.Location, startDate, endDate, in.MaxStudents, in.PriceFen)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, c)
}

// GET /api/v1/training-courses
func (s *Server) listCourses(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	status := r.URL.Query().Get("status")
	courses, err := s.trainingSvc.ListCourses()
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	// filter
	var out []domain.TrainingCourse
	for _, c := range courses {
		if keyword != "" && !strings.Contains(c.Title, keyword) { continue }
		if status != "" && c.Status != status { continue }
		out = append(out, c)
	}
	paginatedRespond(w, r, out, len(out))
}

// ---- Instructors ----

// POST /api/v1/instructors
func (s *Server) registerInstructor(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		Name      string   `json:"name"`
		Bio       string   `json:"bio"`
		OrgID     string   `json:"org_id"`
		CertTypes []string `json:"cert_types"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	i, err := s.trainingSvc.RegisterInstructor(a, in.Name, in.Bio, in.OrgID, in.CertTypes)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, i)
}

// POST /api/v1/admin/instructors/{id}/approve
func (s *Server) approveInstructor(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	i, err := s.trainingSvc.ApproveInstructor(a, r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusOK, i)
}

// GET /api/v1/instructors
func (s *Server) listInstructors(w http.ResponseWriter, r *http.Request) {
	instructors, err := s.trainingSvc.ListInstructors()
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, instructors)
}

// ---- Certified Pilots ----

// POST /api/v1/certified-pilots
func (s *Server) registerPilot(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct{ RealName string `json:"real_name"` }
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	p, err := s.trainingSvc.RegisterPilot(a, in.RealName)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, p)
}

// POST /api/v1/admin/certified-pilots/{id}/approve
func (s *Server) approvePilot(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	p, err := s.trainingSvc.ApprovePilot(a, r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	if p.IDCard != "" {
		p.IDCard = crypto.MaskIDCard(p.IDCard)
	}
	respond(w, r, http.StatusOK, p)
}

// GET /api/v1/certified-pilots
func (s *Server) listPilots(w http.ResponseWriter, r *http.Request) {
	pilots, err := s.trainingSvc.ListPilots()
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	// 脱敏身份证号
	for i := range pilots {
		if pilots[i].IDCard != "" {
			pilots[i].IDCard = crypto.MaskIDCard(pilots[i].IDCard)
		}
	}
	respond(w, r, http.StatusOK, pilots)
}
