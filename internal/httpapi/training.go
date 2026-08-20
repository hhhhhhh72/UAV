package httpapi

import (
	"errors"
	"io"
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
	c, err := s.trainingSvc.AddCertificate(r.Context(), a, domain.CertType(in.CertType), in.CertNumber, in.Level, in.IssuerOrg, in.IssueDate, in.ExpireDate)
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
	c, err := s.trainingSvc.ApproveCertificate(r.Context(), a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "approve_certificate", "certificate", c.ID, "approved")
	respond(w, r, http.StatusOK, c)
}

// GET /api/v1/certificates/mine
func (s *Server) listMyCertificates(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	certs, err := s.trainingSvc.ListMyCertificates(r.Context(), a)
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
		// 小程序培训页扩展字段（training/courses + enroll + register）
		OrgName       string               `json:"org_name"`
		Rating        string               `json:"rating"`
		ReviewCount   int                  `json:"review_count"`
		District      string               `json:"district"`
		DurationDays  int                  `json:"duration_days"`
		Image         string               `json:"image"`
		Tags          []string             `json:"tags"`
		Certificate   string               `json:"certificate"`
		Courses       []domain.CoursePrice `json:"courses"`
		Prices        []domain.CoursePrice `json:"prices"`
		BusinessHours string               `json:"business_hours"`
		Phone         string               `json:"phone"`
		Remain        int                  `json:"remain"`
		Environment   []string             `json:"environment"`
		CourseTypes   []string             `json:"course_types"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// P2 修复：严格解析日期——非法日期此前被 ParseTime 静默写成当前时间落库。
	startDate, ok := strictDate(w, r, in.StartDate)
	if !ok {
		return
	}
	endDate, ok := strictDate(w, r, in.EndDate)
	if !ok {
		return
	}
	c, err := s.trainingSvc.CreateCourse(r.Context(), a, domain.TrainingCourse{
		Title: in.Title, CertType: domain.CertType(in.CertType), Description: in.Description,
		Location: in.Location, StartDate: startDate, EndDate: endDate,
		MaxStudents: in.MaxStudents, PriceFen: in.PriceFen,
		OrgName: in.OrgName, Rating: in.Rating, ReviewCount: in.ReviewCount,
		District: in.District, DurationDays: in.DurationDays, Image: in.Image,
		Tags: in.Tags, Certificate: in.Certificate, Courses: in.Courses,
		Prices: in.Prices, BusinessHours: in.BusinessHours, Phone: in.Phone,
		Remain: in.Remain, Environment: in.Environment, CourseTypes: in.CourseTypes,
		// 用户发布课程待审核（pending），管理端审核（status 改 published）后进入公开列表——
		// 与需求/服务/商品发布一致；此前课程即时上架，无需审核。
		Status: "pending",
	})
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, c)
}

// GET /api/v1/training-courses
// 支持小程序 training/courses.vue 筛选：cert_type / region(district) / keyword / status / page / page_size
func (s *Server) listCourses(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	status := r.URL.Query().Get("status")
	certType := r.URL.Query().Get("cert_type")
	region := r.URL.Query().Get("region")
	courses, err := s.trainingSvc.ListCourses(r.Context())
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// mine=1：只看当前用户（机构）发布的课程（含草稿/未上架，未登录返回空）
	if r.URL.Query().Get("mine") == "1" {
		a, ok := authenticatedActor(r)
		if !ok {
			paginatedRespond(w, r, []domain.TrainingCourse{}, 0)
			return
		}
		mine := make([]domain.TrainingCourse, 0, len(courses))
		for _, c := range courses {
			if c.OrgID == a.ID {
				mine = append(mine, c)
			}
		}
		paginatedRespond(w, r, mine, len(mine))
		return
	}
	// filter（公开列表仅已上架：待审核/草稿/已关闭不公开；管理端请求可见全部）
	adminReq := false
	if a, ok := authenticatedActor(r); ok &&
		(a.Role == domain.RolePlatformAdmin || a.Role == domain.RoleAssociationAdmin) {
		adminReq = true
	}
	var out []domain.TrainingCourse
	for _, c := range courses {
		if !adminReq && isNonPublicCourseStatus(c.Status) {
			continue
		}
		if keyword != "" && !strings.Contains(c.Title, keyword) && !strings.Contains(c.OrgName, keyword) {
			continue
		}
		if status != "" && c.Status != status {
			continue
		}
		if certType != "" && string(c.CertType) != certType {
			continue
		}
		if region != "" && c.District != region && c.Location != region {
			continue
		}
		out = append(out, c)
	}
	paginatedRespond(w, r, out, len(out))
}

// isNonPublicCourseStatus 报告课程状态是否不对外公开（待审核/草稿/已关闭）。
// 公开接口（publicListCourses）与认证列表（listCourses）共用同一过滤逻辑，
// 保证匿名可见状态集与登录可见集一致。
func isNonPublicCourseStatus(s string) bool {
	return s == "pending" || s == "draft" || s == "closed"
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
		Photo     string   `json:"photo"`
		Bio       string   `json:"bio"`
		OrgID     string   `json:"org_id"`
		CertTypes []string `json:"cert_types"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	i, err := s.trainingSvc.RegisterInstructor(r.Context(), a, in.Name, in.Photo, in.Bio, in.OrgID, in.CertTypes)
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
	i, err := s.trainingSvc.ApproveInstructor(r.Context(), a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, i)
}

// GET /api/v1/instructors
// P2 修复：公开列表此前直出全量（含未审核教练与其 user_id）。
// 现仅返回已审核（approved）教练，并脱敏 user_id（不暴露账号 ID）。
func (s *Server) listInstructors(w http.ResponseWriter, r *http.Request) {
	instructors, err := s.trainingSvc.ListInstructors(r.Context())
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	out := make([]domain.Instructor, 0, len(instructors))
	for _, i := range instructors {
		if i.Status != "approved" {
			continue
		}
		i.UserID = ""
		out = append(out, i)
	}
	respond(w, r, http.StatusOK, out)
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
		Avatar      string `json:"avatar"`
		Region      string `json:"region"`
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
	p, err := s.trainingSvc.RegisterPilot(r.Context(), a, in.RealName, in.IDCard, in.FlightHours, in.Bio, in.Avatar, in.Region)
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
	p, err := s.trainingSvc.ApprovePilot(r.Context(), a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	if p.IDCard != "" {
		p.IDCard = crypto.MaskIDCard(p.IDCard)
	}
	s.audit(r.Context(), a.ID, "approve_pilot", "certified_pilot", p.ID, "approved")
	respond(w, r, http.StatusOK, p)
}

// GET /api/v1/certified-pilots — 公开名录：仅已认证（approved）飞手，身份证脱敏，支持 page/page_size 分页
func (s *Server) listPilots(w http.ResponseWriter, r *http.Request) {
	pilots, err := s.trainingSvc.ListPilotsDetailed(r.Context())
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// 公开名录只显示已认证（待审/驳回不入名录）
	approved := make([]domain.CertifiedPilotDetail, 0, len(pilots))
	for _, p := range pilots {
		if p.Status == "approved" {
			approved = append(approved, p)
		}
	}
	pilots = approved
	// 关键词过滤（姓名）
	if kw := strings.TrimSpace(r.URL.Query().Get("keyword")); kw != "" {
		filtered := make([]domain.CertifiedPilotDetail, 0, len(pilots))
		for _, p := range pilots {
			if strings.Contains(p.RealName, kw) {
				filtered = append(filtered, p)
			}
		}
		pilots = filtered
	}
	// 脱敏身份证号 + P1 脱敏用户 ID（手机号注册用户 user_id=user-<手机号>，公开名录不得泄露）
	for i := range pilots {
		if pilots[i].IDCard != "" {
			pilots[i].IDCard = crypto.MaskIDCard(pilots[i].IDCard)
		}
		pilots[i].UserID = maskUserID(pilots[i].UserID)
	}
	// 分页（兼容：显式传 page_size 才分页，否则保持全量返回）
	// 双重分页修复：不再手工切片——paginatedRespond 内部按 page/page_size 唯一一次分页。
	if r.URL.Query().Get("page_size") != "" {
		paginatedRespond(w, r, pilots, len(pilots))
		return
	}
	respond(w, r, http.StatusOK, pilots)
}

// GET /api/v1/certified-pilots/{id} — 飞手详情单查：仅 approved 可公开查看，身份证脱敏，返回 certificates 明细
func (s *Server) getPilot(w http.ResponseWriter, r *http.Request) {
	p, err := s.trainingSvc.GetPilotDetail(r.Context(), r.PathValue("id"))
	if err != nil || p.ID == "" {
		fail(w, r, http.StatusNotFound, errors.New("pilot not found"))
		return
	}
	if p.Status != "approved" {
		fail(w, r, http.StatusNotFound, errors.New("pilot not found"))
		return
	}
	if p.IDCard != "" {
		p.IDCard = crypto.MaskIDCard(p.IDCard)
	}
	// P1 脱敏：公开详情返回前替换手机号注册用户的 user_id，防止手机号泄露。
	p.UserID = maskUserID(p.UserID)
	respond(w, r, http.StatusOK, p)
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
	pilots, err := s.trainingSvc.ListPilots(r.Context())
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	status := r.URL.Query().Get("status")
	if status == "all" {
		status = ""
	}
	filtered, total := adminListFilter(pilots, r.URL.Query().Get("keyword"), status,
		func(p domain.CertifiedPilot) string { return p.RealName },
		func(p domain.CertifiedPilot) string { return p.Status })
	paginatedRespond(w, r, filtered, total)
}

// POST /api/v1/admin/certified-pilots/{id}/reject — 驳回飞手认证（可选携带驳回理由）
func (s *Server) rejectPilot(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Reason string `json:"reason"`
	}
	// 无 body 视为空理由（兼容旧客户端）
	if err := decode(r, &in); err != nil && !errors.Is(err, io.EOF) {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	p, err := s.trainingSvc.RejectPilot(r.Context(), a, r.PathValue("id"), in.Reason)
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
	p, err := s.trainingSvc.GetPilotByOwner(r.Context(), a.ID)
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
