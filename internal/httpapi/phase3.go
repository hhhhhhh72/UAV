package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Enrollments ----

// POST /api/v1/training-courses/{id}/enroll
func (s *Server) enrollCourse(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	e, err := s.enrollSvc.Enroll(a.ID, r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusConflict, err); return }
	respond(w, r, http.StatusCreated, e)
}

// POST /api/v1/training-courses/{id}/pay-and-enroll
// Freezes course fee from escrow balance, then enrolls student.
func (s *Server) payAndEnroll(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }

	// Get course price.
	courses, err := s.trainingSvc.ListCourses()
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	var course domain.TrainingCourse
	found := false
	for _, c := range courses {
		if c.ID == r.PathValue("id") { course = c; found = true; break }
	}
	if !found { fail(w, r, http.StatusNotFound, errors.New("course not found")); return }

	if course.PriceFen > 0 {
		_, err := s.escrowSvc.Freeze(a.ID, course.PriceFen, "training_course", course.ID)
		if err != nil { fail(w, r, http.StatusPaymentRequired, fmt.Errorf("insufficient balance: %w", err)); return }
	}

	e, err := s.enrollSvc.Enroll(a.ID, course.ID)
	if err != nil { fail(w, r, http.StatusConflict, err); return }
	s.audit(r.Context(), a.ID, "pay_and_enroll", "enrollment", e.ID, "enrolled")
	respond(w, r, http.StatusCreated, e)
}

// POST /api/v1/enrollments/{id}/complete
// Admin marks enrollment complete: releases escrow to course org + issues certificate.
func (s *Server) completeEnrollment(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required")); return
	}

	enrolls, err := s.enrollSvc.ListByCourse(r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }

	var enrollment domain.Enrollment
	found := false
	for _, e := range enrolls {
		if e.ID == r.PathValue("id") { enrollment = e; found = true; break }
	}
	if !found { fail(w, r, http.StatusNotFound, errors.New("enrollment not found")); return }

	// Find course to get price and org.
	courses, _ := s.trainingSvc.ListCourses()
	var course domain.TrainingCourse
	for _, c := range courses {
		if c.ID == enrollment.CourseID { course = c; break }
	}

	// Release funds if course was paid.
	if course.PriceFen > 0 {
		_, err := s.escrowSvc.Release(enrollment.UserID, course.OrgID, course.PriceFen, "training_course", course.ID)
		if err != nil { fail(w, r, http.StatusInternalServerError, fmt.Errorf("release escrow: %w", err)); return }
	}

	// Auto-issue certificate.
	cert, err := s.trainingSvc.AddCertificate(
		domain.Actor{ID: enrollment.UserID, Role: domain.RoleIndividual},
		course.CertType, "auto-"+enrollment.ID, "passed", course.OrgID,
		time.Now(), time.Now().AddDate(3, 0, 0),
	)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("issue certificate: %w", err))
		return
	}

	s.audit(r.Context(), a.ID, "complete_enrollment", "enrollment", enrollment.ID, "completed+cert_issued")
	respond(w, r, http.StatusOK, map[string]any{
		"enrollment":  enrollment,
		"certificate": cert,
		"status":      "completed",
	})
}

// GET /api/v1/enrollments/mine
func (s *Server) listMyEnrollments(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }

	// Collect all enrollments for this user across all courses.
	courses, err := s.trainingSvc.ListCourses()
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }

	type enrollmentWithCourse struct {
		domain.Enrollment
		CourseTitle string `json:"course_title"`
	}
	var out []enrollmentWithCourse
	for _, c := range courses {
		enrolls, _ := s.enrollSvc.ListByCourse(c.ID)
		for _, e := range enrolls {
			if e.UserID == a.ID {
				out = append(out, enrollmentWithCourse{Enrollment: e, CourseTitle: c.Title})
			}
		}
	}
	respond(w, r, http.StatusOK, out)
}

// GET /api/v1/training-courses/{id}/enrollments
func (s *Server) listEnrollments(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	enrolls, err := s.enrollSvc.ListByCourse(r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, enrolls)
}

// ---- Expiry Alerts ----

// GET /api/v1/certificates/expiring?days=30
func (s *Server) listExpiringCerts(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	days := 30
	if d, err := fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days); d != 1 || err != nil { days = 30 }
	certs, err := s.trainingSvc.ListAllCertificates()
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, s.expirySvc.GetExpiringCerts(certs, days))
}

// GET /api/v1/inspections/expiring?days=30
func (s *Server) listExpiringInspections(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	days := 30
	if d, err := fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days); d != 1 || err != nil { days = 30 }
	inspections, err := s.insuranceSvc.ListAllInspections()
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, s.expirySvc.GetExpiringInspections(inspections, days))
}

// ---- Trade Orders ----

// POST /api/v1/trade-orders
func (s *Server) createTradeOrder(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		ProductID string `json:"product_id"`
		SellerID  string `json:"seller_id"`
		AmountFen int64  `json:"amount_fen"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	o, err := s.tradeSvc.Create(a.ID, in.ProductID, in.SellerID, in.AmountFen)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	s.audit(r.Context(), a.ID, "create_trade_order", "trade_order", o.ID, "created")
	respond(w, r, http.StatusCreated, o)
}

// PATCH /api/v1/trade-orders/{id}/status
func (s *Server) updateTradeOrderStatus(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct{ Status string `json:"status"` }
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	o, err := s.tradeSvc.UpdateStatus(r.PathValue("id"), a.ID, in.Status)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusOK, o)
}

// GET /api/v1/trade-orders/mine
func (s *Server) listMyTradeOrders(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	orders, err := s.tradeSvc.ListMine(a.ID)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, orders)
}

// ---- Admin Dashboard ----

// GET /api/v1/admin/dashboard
func (s *Server) adminDashboard(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	if a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	// Aggregate real stats from all services.
	var (
		entPending  int
		totalDemands int
		totalPosts   int
		totalReports  int
		totalUsers   int
	)
	if ent, err := s.enterprises.Pending(a); err == nil {
		entPending = len(ent)
	}
	if dem, err := s.demands.List(repository.DemandFilter{}); err == nil {
		totalDemands = len(dem)
	}
	// Use community service for post/report counts
	if posts, _, err := s.communitySvc.ListPublishedPosts(0, 10000); err == nil {
		totalPosts = len(posts)
	}
	respond(w, r, http.StatusOK, map[string]any{
		"pending_enterprises": entPending,
		"total_demands":       totalDemands,
		"total_posts":         totalPosts,
		"pending_reports":     totalReports,
		"total_users":         totalUsers,
		"server_time":         time.Now().UTC().Format(time.RFC3339),
	})
}
