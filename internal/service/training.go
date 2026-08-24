package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type TrainingService struct {
	certRepo       repository.CertificateRepository
	courseRepo     repository.CourseRepository
	instructorRepo repository.InstructorRepository
	pilotRepo      repository.PilotRepository
}

func NewTrainingService(cr repository.CertificateRepository, cor repository.CourseRepository, ir repository.InstructorRepository, pr repository.PilotRepository) *TrainingService {
	return &TrainingService{
		certRepo:       cr,
		courseRepo:     cor,
		instructorRepo: ir,
		pilotRepo:      pr,
	}
}

// ---- Certificates ----

func (s *TrainingService) AddCertificate(ctx context.Context, a domain.Actor, certType domain.CertType, certNumber, level, issuer string, issueDate, expireDate time.Time) (domain.Certificate, error) {
	// 幂等/防撞号：cert_number 已存在时——本人持有则幂等返回已有证书（completeEnrollment 重试），
	// 他人持有则报错（防用户提交他人已占用的证书号静默返回错误结果）。
	if certNumber != "" {
		if existing, err := s.certRepo.FindByNumber(ctx, certNumber); err == nil {
			if existing.UserID == a.ID {
				return existing, nil
			}
			return domain.Certificate{}, fmt.Errorf("certificate number %q already exists", certNumber)
		}
	}
	now := time.Now()
	c := domain.Certificate{ID: nextID("cert"), UserID: a.ID, CertType: certType,
		CertNumber: certNumber, Level: level, IssueDate: issueDate, ExpireDate: expireDate,
		IssuerOrg: issuer, Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.certRepo.Create(ctx, c)
}

func (s *TrainingService) ApproveCertificate(ctx context.Context, a domain.Actor, id string) (domain.Certificate, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Certificate{}, errors.New("admin permission required")
	}
	return s.certRepo.UpdateStatus(ctx, id, "approved")
}

func (s *TrainingService) ListMyCertificates(ctx context.Context, a domain.Actor) ([]domain.Certificate, error) {
	return s.certRepo.ListByUser(ctx, a.ID)
}

func (s *TrainingService) ListAllCertificates(ctx context.Context) ([]domain.Certificate, error) {
	return s.certRepo.ListAll(ctx)
}

func (s *TrainingService) GetCourse(ctx context.Context, id string) (domain.TrainingCourse, error) {
	return s.courseRepo.FindByID(ctx, id)
}

// ToggleCourseFavorite 收藏/取消收藏培训课程（登录用户可收藏任意存在课程）。
func (s *TrainingService) ToggleCourseFavorite(ctx context.Context, userID, courseID string, favorite bool) error {
	if _, err := s.courseRepo.FindByID(ctx, courseID); err != nil {
		return err
	}
	if favorite {
		return s.courseRepo.FavoriteCourse(ctx, userID, courseID)
	}
	return s.courseRepo.UnfavoriteCourse(ctx, userID, courseID)
}

// ListFavoriteCourses 当前用户收藏的课程列表（按收藏时间倒序）。
func (s *TrainingService) ListFavoriteCourses(ctx context.Context, userID string) ([]domain.TrainingCourse, error) {
	return s.courseRepo.ListFavoriteCourses(ctx, userID)
}

func (s *TrainingService) GetCert(ctx context.Context, id string) (domain.Certificate, error) {
	return s.certRepo.FindByID(ctx, id)
}

// FindByNumber 按证书编号查证书（不存在返回错误）。
// completeEnrollment 幂等重试用：cert_number='auto-'+enrollment.ID 查证判断"是否已发证"。
func (s *TrainingService) FindByNumber(ctx context.Context, certNumber string) (domain.Certificate, error) {
	return s.certRepo.FindByNumber(ctx, certNumber)
}

func (s *TrainingService) UpdateCertificate(ctx context.Context, id, certType, certNumber, level, issuer, status string, issueDate, expireDate time.Time) (domain.Certificate, error) {
	c, err := s.certRepo.FindByID(ctx, id)
	if err != nil {
		return domain.Certificate{}, err
	}
	c.CertType = domain.CertType(certType)
	c.CertNumber = certNumber
	c.Level = level
	c.IssuerOrg = issuer
	c.Status = status
	c.IssueDate = issueDate
	c.ExpireDate = expireDate
	return s.certRepo.Update(ctx, c)
}

func (s *TrainingService) DeleteCertificate(ctx context.Context, id string) error {
	return s.certRepo.Delete(ctx, id)
}

// ---- Courses ----

// CreateCourse 接收完整领域对象（含小程序页面字段 org_name/rating/district/courses 等）。
func (s *TrainingService) CreateCourse(ctx context.Context, a domain.Actor, c domain.TrainingCourse) (domain.TrainingCourse, error) {
	if c.PriceFen < 0 {
		return domain.TrainingCourse{}, errors.New("price cannot be negative")
	}
	now := time.Now()
	if c.ID == "" {
		c.ID = nextID("course")
	}
	c.OrgID = a.ID
	if c.Status == "" {
		c.Status = "draft"
	}
	c.Version = 1
	c.CreatedAt = now
	c.UpdatedAt = now
	return s.courseRepo.Create(ctx, c)
}

func (s *TrainingService) ListCourses(ctx context.Context) ([]domain.TrainingCourse, error) {
	return s.courseRepo.List(ctx)
}

// validCourseStatus 课程状态白名单（与 domain.TrainingCourse.Status 注释一致，
// 另含 pending/closed 两个实际使用状态：用户发布待审核、管理端下架）。
func validCourseStatus(s string) bool {
	switch s {
	case "draft", "pending", "published", "recruiting", "full", "upcoming", "urgent", "closed":
		return true
	}
	return false
}

func (s *TrainingService) UpdateCourse(ctx context.Context, c domain.TrainingCourse) (domain.TrainingCourse, error) {
	old, err := s.courseRepo.FindByID(ctx, c.ID)
	if err != nil {
		return domain.TrainingCourse{}, err
	}
	if c.PriceFen < 0 {
		return domain.TrainingCourse{}, errors.New("price cannot be negative")
	}
	if !validCourseStatus(c.Status) {
		return domain.TrainingCourse{}, fmt.Errorf("invalid course status %q", c.Status)
	}
	// 机构归属不可经更新接口篡改/丢失：PG 模式 UPDATE 语句不含 org_id（天然保留），
	// 内存模式是整条替换——不显式保留会导致管理端改课程后 OrgID 清空，
	// completeEnrollment 学费释放目标变成空用户（资金路径断裂）。
	c.OrgID = old.OrgID
	c.Version = old.Version
	c.CreatedAt = old.CreatedAt // 保留原创建时间
	c.UpdatedAt = time.Now()
	return s.courseRepo.Update(ctx, c)
}

func (s *TrainingService) DeleteCourse(ctx context.Context, id string) error {
	return s.courseRepo.Delete(ctx, id)
}

// ---- Instructors ----

func (s *TrainingService) RegisterInstructor(ctx context.Context, a domain.Actor, name, photo, bio, orgID string, certTypes []string) (domain.Instructor, error) {
	now := time.Now()
	// 查重（与 RegisterPilot 对齐）：approved/pending 拒绝重复申请；rejected 覆盖重提
	if existing, err := s.instructorRepo.List(ctx); err == nil {
		for _, e := range existing {
			if e.UserID != a.ID {
				continue
			}
			switch e.Status {
			case "approved":
				return domain.Instructor{}, errors.New("你已经通过导师认证，无需重复申请")
			case "pending":
				return domain.Instructor{}, errors.New("导师认证审核中，请耐心等待")
			default: // rejected：覆盖重提
				e.Name = name
				e.Photo = photo
				e.CertTypes = certTypes
				e.Bio = bio
				e.OrgID = orgID
				e.Status = "pending"
				return s.instructorRepo.Update(ctx, e)
			}
		}
	}
	i := domain.Instructor{ID: nextID("instructor"), UserID: a.ID, Name: name,
		Photo: photo, CertTypes: certTypes, Bio: bio, OrgID: orgID, Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.instructorRepo.Create(ctx, i)
}

func (s *TrainingService) ApproveInstructor(ctx context.Context, a domain.Actor, id string) (domain.Instructor, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Instructor{}, errors.New("admin permission required")
	}
	return s.instructorRepo.UpdateStatus(ctx, id, "approved")
}

func (s *TrainingService) ListInstructors(ctx context.Context) ([]domain.Instructor, error) {
	return s.instructorRepo.List(ctx)
}

// ---- Certified Pilots ----

func (s *TrainingService) RegisterPilot(ctx context.Context, a domain.Actor, realName, idCard string, flightHours int, bio, avatar, region string) (domain.CertifiedPilot, error) {
	// 自动关联已认证证书（审核管线 approved 状态，无手动勾选/造假空间）
	certIDs := []string{}
	if certs, err := s.certRepo.ListByUser(ctx, a.ID); err == nil {
		for _, c := range certs {
			if c.Status == "approved" {
				certIDs = append(certIDs, c.ID)
			}
		}
	}
	now := time.Now()
	// 已有记录：approved/pending 拒绝重复申请；rejected 覆盖重提（重置为 pending）
	if existing, err := s.pilotRepo.List(ctx); err == nil {
		for _, e := range existing {
			if e.UserID != a.ID {
				continue
			}
			switch e.Status {
			case "approved":
				return domain.CertifiedPilot{}, errors.New("你已经通过飞手认证，无需重复申请")
			case "pending":
				return domain.CertifiedPilot{}, errors.New("飞手认证审核中，请耐心等待")
			default: // rejected：覆盖重提
				e.RealName = realName
				e.IDCard = idCard
				e.Avatar = avatar
				e.Region = region
				e.CertIDs = certIDs
				e.FlightHours = flightHours
				e.Bio = bio
				e.Status = "pending"
				return s.pilotRepo.Update(ctx, e)
			}
		}
	}
	p := domain.CertifiedPilot{ID: nextID("pilot"), UserID: a.ID, RealName: realName,
		IDCard: idCard, Avatar: avatar, Region: region, CertIDs: certIDs, FlightHours: flightHours, Bio: bio,
		Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.pilotRepo.Create(ctx, p)
}

func (s *TrainingService) ApprovePilot(ctx context.Context, a domain.Actor, id string) (domain.CertifiedPilot, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.CertifiedPilot{}, errors.New("admin permission required")
	}
	return s.pilotRepo.UpdateStatus(ctx, id, "approved")
}

// RejectPilot 驳回飞手认证申请（管理员），reason 为驳回理由（审核留痕）。
func (s *TrainingService) RejectPilot(ctx context.Context, a domain.Actor, id, reason string) (domain.CertifiedPilot, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.CertifiedPilot{}, errors.New("admin permission required")
	}
	return s.pilotRepo.UpdateReject(ctx, id, reason)
}

func (s *TrainingService) ListPilots(ctx context.Context) ([]domain.CertifiedPilot, error) {
	return s.pilotRepo.List(ctx)
}

// GetPilot 按 ID 单查飞手（详情页）。
func (s *TrainingService) GetPilot(ctx context.Context, id string) (domain.CertifiedPilot, error) {
	return s.pilotRepo.FindByID(ctx, id)
}

// GetPilotDetail 按 ID 单查飞手详情（含 certificates 证书明细，一次性 ListAll 关联防 N+1）。
func (s *TrainingService) GetPilotDetail(ctx context.Context, id string) (domain.CertifiedPilotDetail, error) {
	p, err := s.pilotRepo.FindByID(ctx, id)
	if err != nil || p.ID == "" {
		return domain.CertifiedPilotDetail{}, err
	}
	certs, err := s.certRepo.ListAll(ctx)
	if err != nil {
		return domain.CertifiedPilotDetail{}, err
	}
	d := domain.CertifiedPilotDetail{CertifiedPilot: p}
	for _, c := range certs {
		if c.UserID != p.UserID || c.Status != "approved" {
			continue
		}
		d.Certificates = append(d.Certificates, domain.CertificateBrief{
			ID: c.ID, CertType: string(c.CertType), CertName: certTypeName(c.CertType),
			IssuerOrg: c.IssuerOrg, Level: c.Level, Status: c.Status,
		})
	}
	return d, nil
}

// ListPilotsDetailed 名录输出：把 cert_ids 扩展为证书对象数组（certificates）。
// 一次性 ListAll 后按 UserID 分组，避免 N+1 查询。
func (s *TrainingService) ListPilotsDetailed(ctx context.Context) ([]domain.CertifiedPilotDetail, error) {
	pilots, err := s.pilotRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	certs, err := s.certRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	byUser := make(map[string][]domain.Certificate)
	for _, c := range certs {
		byUser[c.UserID] = append(byUser[c.UserID], c)
	}
	out := make([]domain.CertifiedPilotDetail, 0, len(pilots))
	for _, p := range pilots {
		d := domain.CertifiedPilotDetail{CertifiedPilot: p}
		for _, c := range byUser[p.UserID] {
			if c.Status != "approved" {
				continue
			}
			d.Certificates = append(d.Certificates, domain.CertificateBrief{
				ID: c.ID, CertType: string(c.CertType), CertName: certTypeName(c.CertType),
				IssuerOrg: c.IssuerOrg, Level: c.Level, Status: c.Status,
			})
		}
		out = append(out, d)
	}
	return out, nil
}

// certTypeName 证书类型 → 展示名称（前端详情页证书卡用）。
func certTypeName(t domain.CertType) string {
	switch t {
	case domain.CertCAAC:
		return "CAAC无人机驾驶员执照"
	case domain.CertUTCDJI:
		return "DJI UTC 植保无人机驾驶证"
	case domain.CertGovLevel:
		return "政府职业技能等级证书"
	default:
		return string(t)
	}
}

// GetPilotByOwner 查询我的飞手认证记录（未申请返回零值）。
func (s *TrainingService) GetPilotByOwner(ctx context.Context, userID string) (domain.CertifiedPilot, error) {
	pilots, err := s.pilotRepo.List(ctx)
	if err != nil {
		return domain.CertifiedPilot{}, err
	}
	for _, p := range pilots {
		if p.UserID == userID {
			return p, nil
		}
	}
	return domain.CertifiedPilot{}, nil
}
