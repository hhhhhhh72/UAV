package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/service"
)

// ---- Enrollments ----

// POST /api/v1/training-courses/{id}/enroll
func (s *Server) enrollCourse(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var form service.EnrollmentForm
	if err := decode(r, &form); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	e, err := s.enrollSvc.Enroll(a.ID, r.PathValue("id"), form)
	if err != nil {
		fail(w, r, http.StatusConflict, err)
		return
	}
	respond(w, r, http.StatusCreated, e)
}

// POST /api/v1/training-courses/{id}/pay-and-enroll
// Freezes course fee from escrow balance, then enrolls student.
func (s *Server) payAndEnroll(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	course, err := s.trainingSvc.GetCourse(r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, errors.New("course not found"))
		return
	}

	var form service.EnrollmentForm
	if err := decode(r, &form); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}

	if course.PriceFen > 0 {
		_, err := s.escrowSvc.Freeze(a.ID, course.PriceFen, "training_course", course.ID)
		if err != nil {
			fail(w, r, http.StatusPaymentRequired, fmt.Errorf("insufficient balance: %w", err))
			return
		}
	}

	e, err := s.enrollSvc.Enroll(a.ID, course.ID, form)
	if err != nil {
		fail(w, r, http.StatusConflict, err)
		return
	}
	s.audit(r.Context(), a.ID, "pay_and_enroll", "enrollment", e.ID, "enrolled")
	respond(w, r, http.StatusCreated, e)
}

// POST /api/v1/enrollments/{id}/complete
// Admin marks enrollment complete: releases escrow to course org + issues certificate.
func (s *Server) completeEnrollment(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}

	enrolls, err := s.enrollSvc.ListByCourse(r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

	var enrollment domain.Enrollment
	found := false
	for _, e := range enrolls {
		if e.ID == r.PathValue("id") {
			enrollment = e
			found = true
			break
		}
	}
	if !found {
		fail(w, r, http.StatusNotFound, errors.New("enrollment not found"))
		return
	}

	// Find course to get price and org.
	course, err := s.trainingSvc.GetCourse(enrollment.CourseID)
	if err != nil {
		fail(w, r, http.StatusNotFound, fmt.Errorf("course not found: %w", err))
		return
	}

	// Release funds if course was paid.
	if course.PriceFen > 0 {
		_, err := s.escrowSvc.Release(enrollment.UserID, course.OrgID, course.PriceFen, "training_course", course.ID)
		if err != nil {
			fail(w, r, http.StatusInternalServerError, fmt.Errorf("release escrow: %w", err))
			return
		}
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

// PUT /api/v1/admin/enrollments/{id} — 管理端编辑报名记录
func (s *Server) updateEnrollment(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Name        string `json:"name"`
		Phone       string `json:"phone"`
		IDCard      string `json:"id_card"`
		Gender      string `json:"gender"`
		Birthday    string `json:"birthday"`
		Email       string `json:"email"`
		Education   string `json:"education"`
		Experience  string `json:"experience"`
		PhotoURL    string `json:"photo_url"`
		IDCardImage string `json:"id_card_image"`
		NoCrime     string `json:"no_crime"`
		Status      string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// 查原记录保留 course_id/user_id/created_at
	all, _, err := s.enrollSvc.All(0, 10000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	var found *domain.Enrollment
	for i := range all {
		if all[i].ID == r.PathValue("id") {
			found = &all[i]
			break
		}
	}
	if found == nil {
		fail(w, r, http.StatusNotFound, errors.New("enrollment not found"))
		return
	}
	found.Name = in.Name
	found.Phone = in.Phone
	found.IDCard = in.IDCard
	found.Gender = in.Gender
	if in.Birthday == "" {
		found.Birthday = time.Time{} // 清空生日
	} else if t, err := time.Parse("2006-01-02", in.Birthday); err == nil {
		found.Birthday = t
	}
	found.Email = in.Email
	found.Education = in.Education
	found.Experience = in.Experience
	found.PhotoURL = in.PhotoURL
	found.IDCardImage = in.IDCardImage
	found.NoCrime = in.NoCrime
	found.Status = in.Status
	updated, err := s.enrollSvc.Update(a, *found)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			fail(w, r, http.StatusNotFound, err)
		case strings.Contains(err.Error(), "permission"):
			fail(w, r, http.StatusForbidden, err)
		default:
			fail(w, r, http.StatusBadRequest, err) // 非法状态 / 防回退
		}
		return
	}
	respond(w, r, http.StatusOK, updated)
}

// GET /api/v1/enrollments/mine
func (s *Server) listMyEnrollments(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	// Collect all enrollments for this user across all courses.
	courses, err := s.trainingSvc.ListCourses()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

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
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	enrolls, err := s.enrollSvc.ListByCourse(r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, enrolls)
}

// ---- Expiry Alerts ----

// GET /api/v1/certificates/expiring?days=30
func (s *Server) listExpiringCerts(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	days := 30
	if d, err := fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days); d != 1 || err != nil {
		days = 30
	}
	certs, err := s.trainingSvc.ListAllCertificates()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, s.expirySvc.GetExpiringCerts(certs, days))
}

// GET /api/v1/inspections/expiring?days=30
func (s *Server) listExpiringInspections(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	days := 30
	if d, err := fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days); d != 1 || err != nil {
		days = 30
	}
	inspections, err := s.insuranceSvc.ListAllInspections()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, s.expirySvc.GetExpiringInspections(inspections, days))
}

// ---- Trade Orders ----

// POST /api/v1/trade-orders
func (s *Server) createTradeOrder(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		ProductID string `json:"product_id"`
		SellerID  string `json:"seller_id"`
		AmountFen int64  `json:"amount_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	o, err := s.tradeSvc.Create(a.ID, in.ProductID, in.SellerID, in.AmountFen)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_trade_order", "trade_order", o.ID, "created")
	respond(w, r, http.StatusCreated, o)
}

// PATCH /api/v1/trade-orders/{id}/status
func (s *Server) updateTradeOrderStatus(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	o, err := s.tradeSvc.UpdateStatus(r.PathValue("id"), a.ID, in.Status)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, o)
}

// GET /api/v1/trade-orders/mine
func (s *Server) listMyTradeOrders(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	orders, err := s.tradeSvc.ListMine(a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, orders)
}

// ---- Admin Dashboard ----

// GET /api/v1/admin/dashboard
func (s *Server) adminDashboard(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}

	ent, err := s.enterprises.Pending(a)
	if err != nil {
		ent = nil
	}
	entPending := len(ent)

	dem, err := s.demands.ListAll(repository.DemandFilter{})
	if err != nil {
		dem = nil
	}
	totalDemands := len(dem)

	posts, _, err := s.communitySvc.ListPublishedPosts(0, 10000)
	if err != nil {
		posts = nil
	}
	totalPosts := len(posts)

	pendingList, _, _ := s.reviewSvc.ListAll("", 0, 10000)
	totalReports := len(pendingList)

	users, _ := s.userRepo.All()
	totalUsers := len(users)

	msgs, _, _ := s.msgSvc.ListAll(0, 10000)
	totalMessages := len(msgs)

	// Trends: monthly counts (last 12 months) — 4 维趋势（需求/帖子/用户/消息）
	trends := buildDemandTrends(dem)
	trendsDetail := map[string][]map[string]any{
		"demand":  trends,
		"post":    buildMonthlyTrends(posts, func(p domain.Post) time.Time { return p.CreatedAt }),
		"user":    buildMonthlyTrends(users, func(u domain.User) time.Time { return u.CreatedAt }),
		"message": buildMonthlyTrends(msgs, func(m domain.Message) time.Time { return m.CreatedAt }),
	}

	// 线下成交金额汇总（联系对接模式撮合价值）
	var offlineAmountTotal int64
	for _, dd := range dem {
		offlineAmountTotal += dd.OfflineAmountFen
	}

	// Category distribution
	categoryDist := buildCategoryDist(dem)

	// Module stats
	modules := map[string]map[string]int{"talent": {}, "events": {}, "industry": {}}
	// Talent
	certs, _ := s.trainingSvc.ListAllCertificates()
	modules["talent"]["certificates"] = len(certs)
	cols, _ := s.collegeSvc.List("")
	modules["talent"]["colleges"] = len(cols)
	jobs, _, _ := s.jobSvc.ListPublishedJobs(0, 10000)
	modules["talent"]["jobs"] = len(jobs)
	tours, _ := s.studyTourRepo.List()
	modules["talent"]["study_tours"] = len(tours)
	courses, _ := s.trainingSvc.ListCourses()
	modules["talent"]["training_courses"] = len(courses)

	// Events
	competitions, _, _ := s.competitionSvc.List(1, 10000)
	modules["events"]["competitions"] = competitionsIfNil(competitions)
	evs, _, _ := s.eventSvc.List(1, 10000)
	modules["events"]["events"] = evsIfNil(evs)
	exhs, _, _ := s.exhibitionSvc.List(1, 10000)
	modules["events"]["exhibitions"] = exhsIfNil(exhs)
	emergRes, _, _ := s.emergencySvc.ListResources(1, 10000)
	modules["events"]["emergency_resources"] = emgResIfNil(emergRes)
	disps, _, _ := s.emergencySvc.ListDispatches(1, 10000)
	modules["events"]["emergency_dispatches"] = dispIfNil(disps)

	// Industry
	achs, _, _ := s.achievementSvc.List("", 1, 10000)
	modules["industry"]["achievements"] = achsIfNil(achs)
	cases, _, _ := s.caseSvc.List("", 1, 10000)
	modules["industry"]["cases"] = casesIfNil(cases)
	exps, _ := s.expertSvc.List("")
	modules["industry"]["experts"] = len(exps)
	rpts, _, _ := s.reportSvc.List(1, 10000)
	modules["industry"]["industry_reports"] = rptsIfNil(rpts)
	res, _, _ := s.resourceSvc.List("", 1, 10000)
	modules["industry"]["industry_resources"] = resIfNil(res)
	ports, _, _ := s.portfolioSvc.ListPublished(1, 10000)
	modules["industry"]["portfolios"] = len(ports)
	rds, _, _ := s.rdService.List("", 1, 10000)
	modules["industry"]["rd_challenges"] = rdsIfNil(rds)
	projs, _, _ := s.researchSvc.List(1, 10000)
	modules["industry"]["research_projects"] = projsIfNil(projs)
	sites, _ := s.testSiteSvc.List("")
	modules["industry"]["test_sites"] = len(sites)

	// Status distribution
	statusDist := buildStatusDist(dem)

	respond(w, r, http.StatusOK, map[string]any{
		"pending_enterprises":  entPending,
		"total_demands":        totalDemands,
		"total_posts":          totalPosts,
		"pending_reports":      totalReports,
		"total_users":          totalUsers,
		"total_messages":       totalMessages,
		"offline_amount_total": offlineAmountTotal,
		"trends":               trends,
		"trends_detail":        trendsDetail,
		"category_dist":        categoryDist,
		"status_dist":          statusDist,
		"modules":              modules,
		"server_time":          time.Now().UTC().Format(time.RFC3339),
	})
}

func buildDemandTrends(dem []domain.Demand) []map[string]any {
	return buildMonthlyTrends(dem, func(d domain.Demand) time.Time { return d.CreatedAt })
}

// buildMonthlyTrends 近 12 个月月度计数（泛型：需求/帖子/用户/消息等任意带 CreatedAt 的实体）
func buildMonthlyTrends[T any](items []T, getTime func(T) time.Time) []map[string]any {
	now := time.Now()
	counts := map[string]int{}
	for i := 0; i < 12; i++ {
		m := now.AddDate(0, -i, 0).Format("2006-01")
		counts[m] = 0
	}
	for _, it := range items {
		m := getTime(it).Format("2006-01")
		if _, ok := counts[m]; ok {
			counts[m]++
		}
	}
	out := make([]map[string]any, 12)
	for i := 0; i < 12; i++ {
		m := now.AddDate(0, -11+i, 0).Format("2006-01")
		out[i] = map[string]any{"date": m, "count": counts[m]}
	}
	return out
}

func buildCategoryDist(dem []domain.Demand) map[string]int {
	dist := map[string]int{}
	for _, d := range dem {
		bt := string(d.BizType)
		if bt != "" {
			dist[bt]++
		}
	}
	return dist
}

func buildStatusDist(dem []domain.Demand) map[string]int {
	statusLabel := map[string]string{
		"published": "已发布", "pending": "待审核", "completed": "已完成",
		"cancelled": "已取消", "rejected": "已驳回", "draft": "草稿", "processing": "进行中",
	}
	dist := map[string]int{}
	for _, d := range dem {
		st := string(d.Status)
		label := statusLabel[st]
		if label == "" {
			label = st
		}
		dist[label] = dist[label] + 1
	}
	return dist
}

// Nil-safe helpers for service List results
func competitionsIfNil(v []domain.Competition) int { return len(v) }
func evsIfNil(v []domain.AssociationEvent) int     { return len(v) }
func exhsIfNil(v []domain.Exhibition) int          { return len(v) }
func emgResIfNil(v []domain.EmergencyResource) int { return len(v) }
func dispIfNil(v []domain.EmergencyDispatch) int   { return len(v) }
func achsIfNil(v []domain.Achievement) int         { return len(v) }
func casesIfNil(v []domain.CaseEntry) int          { return len(v) }
func rptsIfNil(v []domain.IndustryReport) int      { return len(v) }
func resIfNil(v []domain.IndustryResource) int     { return len(v) }
func rdsIfNil(v []domain.RDChallenge) int          { return len(v) }
func projsIfNil(v []domain.ResearchProject) int    { return len(v) }
