package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"drone-platform/internal/domain"
)

func (s *Server) registerBatch2Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/admin/transformations", s.createTransformation)
	mux.HandleFunc("GET /api/v1/transformations", s.listTransformations)
	mux.HandleFunc("POST /api/v1/transformations/{id}/advance", s.advanceStage)
	mux.HandleFunc("POST /api/v1/transformations/{id}/milestones", s.addMilestone)

	mux.HandleFunc("POST /api/v1/admin/colleges", s.createCollege)
	mux.HandleFunc("GET /api/v1/colleges", s.listColleges)

	mux.HandleFunc("POST /api/v1/cooperation-programs", s.createCooperation)
	mux.HandleFunc("GET /api/v1/cooperation-programs", s.listCooperations)
	mux.HandleFunc("POST /api/v1/cooperation-programs/{id}/status", s.updateCooperationStatus)
}

// ── Transformation ──
func (s *Server) createTransformation(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	var in struct {
		Title         string `json:"title"`
		AchievementID string `json:"achievement_id"`
		PartnerID     string `json:"partner_id"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	t, err := s.transSvc.Create(r.Context(), in.Title, in.AchievementID, a.ID, in.PartnerID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, t)
}
func (s *Server) listTransformations(w http.ResponseWriter, r *http.Request) {
	// 公开查询：按成果(achievement_id)或归属(owner_id)过滤，无参返回全部
	var (
		list []domain.Transformation
		err  error
	)
	if aid := r.URL.Query().Get("achievement_id"); aid != "" {
		list, err = s.transSvc.ListByAchievement(r.Context(), aid)
	} else {
		list, err = s.transSvc.List(r.Context(), r.URL.Query().Get("owner_id"))
	}
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	if list == nil {
		list = []domain.Transformation{}
	}
	respond(w, r, http.StatusOK, list)
}
func (s *Server) advanceStage(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct{ Stage, Progress string }
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	t, err := s.transSvc.AdvanceStage(r.Context(), a, r.PathValue("id"), domain.TransformationStage(in.Stage), in.Progress)
	if err != nil {
		fail(w, r, mutationErrorCode(err), err)
		return
	}
	respond(w, r, http.StatusOK, t)
}
func (s *Server) addMilestone(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct{ Name, Evidence string }
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	t, err := s.transSvc.AddMilestone(r.Context(), a, r.PathValue("id"), in.Name, in.Evidence)
	if err != nil {
		fail(w, r, mutationErrorCode(err), err)
		return
	}
	respond(w, r, http.StatusOK, t)
}

// ── College ──
func (s *Server) createCollege(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name        string `json:"name"`
		Region      string `json:"region"`
		Description string `json:"description"`
		LogoURL     string `json:"logo_url"`
		CoopType    string `json:"coop_type"` // research/talent/both
		Majors      []string
		Facilities  []string
		// 小程序院校页扩展字段（colleges/list + detail）
		City         string                  `json:"city"`
		Tags         []string                `json:"tags"`
		ShortName    string                  `json:"short_name"`
		LevelTags    string                  `json:"level_tags"`
		Specialties  []string                `json:"specialties"`
		MajorCount   int                     `json:"major_count"`
		PartnerCount int                     `json:"partner_count"`
		TeacherCount int                     `json:"teacher_count"`
		StudentCount int                     `json:"student_count"`
		GraduateRate string                  `json:"graduate_rate"`
		Partners     []domain.CollegePartner `json:"partners"`
		Cover        string                  `json:"cover"`
		Photos       []string                `json:"photos"`
		Phone        string                  `json:"phone"`
		Website      string                  `json:"website"`
		Intro        string                  `json:"intro"`
		MajorsDetail []domain.CollegeMajor   `json:"majors_detail"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c, err := s.collegeSvc.Create(r.Context(), domain.College{
		Name: in.Name, Region: in.Region, City: in.City, Description: in.Description,
		LogoURL: in.LogoURL, CoopType: in.CoopType, Majors: in.Majors, Facilities: in.Facilities,
		Tags: in.Tags, ShortName: in.ShortName, LevelTags: in.LevelTags, Specialties: in.Specialties,
		MajorCount: in.MajorCount, PartnerCount: in.PartnerCount, TeacherCount: in.TeacherCount,
		StudentCount: in.StudentCount, GraduateRate: in.GraduateRate, Partners: in.Partners,
		CoverURL: in.Cover, Photos: in.Photos, Phone: in.Phone, Website: in.Website,
		Intro: in.Intro, MajorsDetail: in.MajorsDetail,
	})
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, c)
}

// listColleges 支持小程序 colleges/list.vue 的分页 + type/keyword 筛选。
// type: undergraduate(非专科，默认) / vocational(专科/高职)；基于 tags 判定（与页面 collegeLevel 一致）。
func (s *Server) listColleges(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")
	keyword := r.URL.Query().Get("keyword")
	collegeType := r.URL.Query().Get("type")
	list, err := s.collegeSvc.List(r.Context(), region)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	var out []domain.College
	for _, c := range list {
		if collegeType != "" && !matchCollegeType(collegeType, c) {
			continue
		}
		if keyword != "" && !strings.Contains(c.Name, keyword) && !strings.Contains(c.City, keyword) && !strings.Contains(c.Region, keyword) {
			continue
		}
		out = append(out, c)
	}
	paginatedRespond(w, r, out, len(out))
}

// matchCollegeType 按 tags 判定院校类型，与页面 collegeLevel() 语义一致。
func matchCollegeType(tp string, c domain.College) bool {
	isVocational := false
	for _, t := range c.Tags {
		if t == "专科" || t == "高职" {
			isVocational = true
			break
		}
	}
	switch tp {
	case "vocational":
		return isVocational
	default: // undergraduate / top 等非专科类型
		return !isVocational
	}
}

// ── Cooperation ──

// requireAdmin 复用于非 /admin/ 前缀的管理写操作（C2 修复）：
// 这些路由不受 adminGate 覆盖，必须显式校验管理员角色。
func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return false
	}
	return true
}

// mutationErrorCode 区分归属拒绝(403)与资源不存在(404)；未识别的错误
// 一律 500（服务端日志记详情），绝不把 DB/SQL 细节以 403/404 暴露给客户端
// （此前默认 404 会把 pgx 错误文本原样回传）。
func mutationErrorCode(err error) int {
	if strings.Contains(err.Error(), "only the owner") {
		return http.StatusForbidden
	}
	if strings.Contains(err.Error(), "not found") {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

func (s *Server) createCooperation(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var in struct {
		Title, CollegeID, EnterpriseID, CoopType, Description, StartDate, EndDate string
		StudentQuota                                                              int
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// P2 修复：严格解析日期——非法日期此前被 ParseTime 静默写成当前时间落库。
	sd, ok := strictDate(w, r, in.StartDate)
	if !ok {
		return
	}
	ed, ok := strictDate(w, r, in.EndDate)
	if !ok {
		return
	}
	cp, err := s.coopSvc.Create(r.Context(), in.Title, in.CollegeID, in.EnterpriseID, in.CoopType, in.Description, sd, ed, in.StudentQuota)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, cp)
}
func (s *Server) listCooperations(w http.ResponseWriter, r *http.Request) {
	list, err := s.coopSvc.List(r.Context(), r.URL.Query().Get("enterprise_id"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, list)
}
func (s *Server) updateCooperationStatus(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var in struct{ Status string }
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	cp, err := s.coopSvc.UpdateStatus(r.Context(), r.PathValue("id"), in.Status)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, cp)
}
