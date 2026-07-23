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
	var in struct{ Title, EventType, Location, DroneModel, TeamName, Summary, Result, Lessons, Source, Date string }
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	rc, err := s.rescueCaseSvc.Create(in.Title, in.EventType, in.Location, in.DroneModel, in.TeamName, in.Summary, in.Result, in.Lessons, in.Source, domain.ParseTime(in.Date))
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusCreated, rc)
}
func (s *Server) listRescueCases(w http.ResponseWriter, r *http.Request) {
	page, ps := paginationFromQuery(r)
	list, total, err := s.rescueCaseSvc.List(r.URL.Query().Get("event_type"), page, ps)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	paginatedRespond(w, r, list, total)
}

// ── EmergencyDept ──
func (s *Server) createEmergencyDept(w http.ResponseWriter, r *http.Request) {
	var in struct{ Name, DeptType, Region, ContactName, ContactPhone, ProtocolURL string }
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	d, err := s.emergDeptSvc.CreateDept(in.Name, in.DeptType, in.Region, in.ContactName, in.ContactPhone, in.ProtocolURL)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusCreated, d)
}
func (s *Server) listEmergencyDepts(w http.ResponseWriter, r *http.Request) {
	list, err := s.emergDeptSvc.ListDepts()
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, list)
}
func (s *Server) createEmergencyDrill(w http.ResponseWriter, r *http.Request) {
	var in struct{ DeptID, Title, Scenario, Date, Result string; Participants, DroneCount int }
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	d, err := s.emergDeptSvc.CreateDrill(in.DeptID, in.Title, in.Scenario, domain.ParseTime(in.Date), in.Participants, in.DroneCount, in.Result)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusCreated, d)
}
func (s *Server) listEmergencyDrills(w http.ResponseWriter, r *http.Request) {
	list, err := s.emergDeptSvc.ListDrills(r.URL.Query().Get("dept_id"))
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, list)
}

// ── AssociationMember ──
func (s *Server) addAssociationMember(w http.ResponseWriter, r *http.Request) {
	var in struct{ UserID, EnterpriseID, Role string }
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	m, err := s.assocMemberSvc.AddMember(in.UserID, in.EnterpriseID, domain.AssociationRole(in.Role))
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusCreated, m)
}
func (s *Server) listAssociationMembers(w http.ResponseWriter, r *http.Request) {
	page, ps := paginationFromQuery(r)
	list, total, err := s.assocMemberSvc.ListMembers(r.URL.Query().Get("role"), page, ps)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	paginatedRespond(w, r, list, total)
}
func (s *Server) getMyAssociationRole(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("auth required")); return }
	m, err := s.assocMemberSvc.GetByUserID(a.ID)
	if err != nil { fail(w, r, http.StatusNotFound, err); return }
	respond(w, r, http.StatusOK, m)
}
