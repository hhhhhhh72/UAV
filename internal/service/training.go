package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

// certValid 有效证书：状态 approved，且未过期（expire_date 未设置视为长期有效，
// 仅显式过期日期判断）——过期证书不参与飞手/导师认证关联与公开名录
// （此前只判断 status，过期证书被当作有效资质）。
func certValid(c domain.Certificate) bool {
	if c.Status != "approved" {
		return false
	}
	if c.ExpireDate.IsZero() {
		return true
	}
	return c.ExpireDate.After(time.Now())
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

func (s *TrainingService) AddCertificate(ctx context.Context, a domain.Actor, certType domain.CertType, certNumber, level, issuer, imageURL string, issueDate, expireDate time.Time) (domain.Certificate, error) {
	// 幂等/防撞号：cert_number 已存在时——本人持有则幂等返回已有证书（completeEnrollment 重试）；
	// 本人持有但已驳回 → 允许重新提交（覆盖回 pending，信息以本次为准）；
	// 他人持有则报错（防用户提交他人已占用的证书号静默返回错误结果）。
	if certNumber != "" {
		if existing, err := s.certRepo.FindByNumber(ctx, certNumber); err == nil {
			if existing.UserID == a.ID {
				if existing.Status == "rejected" {
					existing.CertType = certType
					existing.Level = level
					existing.IssuerOrg = issuer
					existing.ImageURL = imageURL
					existing.IssueDate = issueDate
					existing.ExpireDate = expireDate
					existing.Status = "pending"
					existing.Version++
					existing.UpdatedAt = time.Now()
					return s.certRepo.Update(ctx, existing)
				}
				return existing, nil
			}
			return domain.Certificate{}, fmt.Errorf("certificate number %q already exists", certNumber)
		}
	}
	now := time.Now()
	c := domain.Certificate{ID: nextID("cert"), UserID: a.ID, CertType: certType,
		CertNumber: certNumber, Level: level, IssueDate: issueDate, ExpireDate: expireDate,
		IssuerOrg: issuer, ImageURL: imageURL, Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	if _, err := s.certRepo.Create(ctx, c); err != nil {
		// 并发撞号：check-then-insert 竞态由唯一索引兜底，转为与原预检一致的友好错误。
		if errors.Is(err, repository.ErrCertNumberTaken) {
			return domain.Certificate{}, fmt.Errorf("certificate number %q already exists", certNumber)
		}
		return domain.Certificate{}, err
	}
	return c, nil
}

func (s *TrainingService) ApproveCertificate(ctx context.Context, a domain.Actor, id string) (domain.Certificate, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Certificate{}, errors.New("admin permission required")
	}
	// 状态机前置：approved 幂等；已驳回不得翻转为通过（纠错请重走发证/审核流程）。
	cur, err := s.certRepo.FindByID(ctx, id)
	if err != nil {
		return domain.Certificate{}, err
	}
	if cur.Status == "approved" {
		return cur, nil
	}
	if cur.Status == "rejected" {
		return domain.Certificate{}, errors.New("已驳回的证书不能改为通过")
	}
	return s.certRepo.UpdateStatus(ctx, id, "approved")
}

func (s *TrainingService) ListMyCertificates(ctx context.Context, a domain.Actor) ([]domain.Certificate, error) {
	return s.certRepo.ListByUser(ctx, a.ID)
}

// RejectCertificate 管理端驳回证书（用户可重新提交覆盖为 pending）。
func (s *TrainingService) RejectCertificate(ctx context.Context, a domain.Actor, id string) (domain.Certificate, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Certificate{}, errors.New("admin permission required")
	}
	cur, err := s.certRepo.FindByID(ctx, id)
	if err != nil {
		return domain.Certificate{}, err
	}
	if cur.Status == "approved" {
		return domain.Certificate{}, errors.New("已通过的证书不能驳回，请走吊销流程")
	}
	if cur.Status == "rejected" {
		return cur, nil
	}
	return s.certRepo.UpdateStatus(ctx, id, "rejected")
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
	// 幂等锚点保护：auto- 前缀证书号为系统签发（completeEnrollment 以
	// auto-<报名ID> 判定"是否已发证"），改号会破坏锚点导致重复发证。
	if strings.HasPrefix(c.CertNumber, "auto-") && certNumber != c.CertNumber {
		return domain.Certificate{}, errors.New("系统签发证书号不可修改")
	}
	c.CertType = domain.CertType(certType)
	c.CertNumber = certNumber
	c.Level = level
	c.IssuerOrg = issuer
	c.Status = status
	c.IssueDate = issueDate
	c.ExpireDate = expireDate
	if _, err := s.certRepo.Update(ctx, c); err != nil {
		if errors.Is(err, repository.ErrCertNumberTaken) {
			return domain.Certificate{}, fmt.Errorf("certificate number %q already exists", certNumber)
		}
		return domain.Certificate{}, err
	}
	return c, nil
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
	// 状态机前置：approved 幂等；已驳回不得翻转为通过。
	cur, err := s.instructorRepo.FindByID(ctx, id)
	if err != nil {
		return domain.Instructor{}, err
	}
	if cur.Status == "approved" {
		return cur, nil
	}
	if cur.Status == "rejected" {
		return domain.Instructor{}, errors.New("已驳回的培训师不能改为通过")
	}
	return s.instructorRepo.UpdateStatus(ctx, id, "approved")
}

func (s *TrainingService) ListInstructors(ctx context.Context) ([]domain.Instructor, error) {
	return s.instructorRepo.List(ctx)
}

// ---- Certified Pilots ----

func (s *TrainingService) RegisterPilot(ctx context.Context, a domain.Actor, realName, idCard string, flightHours int, bio, avatar, region string) (domain.CertifiedPilot, error) {
	// 自动关联有效证书：仅"approved 且未过期"计入（此前只看 approved，
	// 过期证书仍被当作有效资质，持过期证书者可获"已认证飞手"身份）。
	certIDs := []string{}
	if certs, err := s.certRepo.ListByUser(ctx, a.ID); err == nil {
		for _, c := range certs {
			if certValid(c) {
				certIDs = append(certIDs, c.ID)
			}
		}
	}
	// 审批门禁：至少持有一张未过期 approved 证书，否则不批准。
	// 实现为申请校验：无有效证书直接拒绝申请（与"无证不批"同效）。
	if len(certIDs) == 0 {
		return domain.CertifiedPilot{}, errors.New("需要至少一张未过期的有效证书才能申请飞手认证")
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
	// 状态机前置：approved 幂等；已驳回不得翻转为通过（驳回后需重新申请产生新记录）。
	cur, err := s.pilotRepo.FindByID(ctx, id)
	if err != nil {
		return domain.CertifiedPilot{}, err
	}
	if cur.Status == "approved" {
		return cur, nil
	}
	if cur.Status == "rejected" {
		return domain.CertifiedPilot{}, errors.New("已驳回的飞手申请不能改为通过")
	}
	return s.pilotRepo.UpdateStatus(ctx, id, "approved")
}

// RejectPilot 驳回飞手认证申请（管理员），reason 为驳回理由（审核留痕）。
// 状态机前置：rejected 幂等；approved 可驳回（撤销已通过的认证——纠错需连带
// 降级身份，因此留痕 reason 必填）；pending/rejected 常规驳回。
func (s *TrainingService) RejectPilot(ctx context.Context, a domain.Actor, id, reason string) (domain.CertifiedPilot, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.CertifiedPilot{}, errors.New("admin permission required")
	}
	cur, err := s.pilotRepo.FindByID(ctx, id)
	if err != nil {
		return domain.CertifiedPilot{}, err
	}
	if cur.Status == "rejected" {
		return cur, nil
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
// 未找到返回 ErrResourceNotFound（Handler 区分 404/500，不把 DB 故障伪装成 not found）。
func (s *TrainingService) GetPilotDetail(ctx context.Context, id string) (domain.CertifiedPilotDetail, error) {
	p, err := s.pilotRepo.FindByID(ctx, id)
	if err != nil || p.ID == "" {
		return domain.CertifiedPilotDetail{}, ErrResourceNotFound
	}
	certs, err := s.certRepo.ListAll(ctx)
	if err != nil {
		return domain.CertifiedPilotDetail{}, err
	}
	d := domain.CertifiedPilotDetail{CertifiedPilot: p}
	for _, c := range certs {
		if c.UserID != p.UserID || !certValid(c) {
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
	return s.attachCertificates(ctx, pilots)
}

// ListPilotsDetailedPaged 公开名录分页（SQL 端 COUNT + LIMIT/OFFSET，不再整表加载）：
// keyword 匹配姓名；证书关联仅对本页飞手做一次 ListAll 分组（无 N+1）。
// 返回 total 为过滤后的已认证飞手总数。
func (s *TrainingService) ListPilotsDetailedPaged(ctx context.Context, keyword string, offset, limit int) ([]domain.CertifiedPilotDetail, int, error) {
	pilots, total, err := s.pilotRepo.ListApproved(ctx, keyword, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	out, err := s.attachCertificates(ctx, pilots)
	if err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// attachCertificates 为一组飞手补充已认证证书明细（证书整表一次加载后按 UserID 分组）。
func (s *TrainingService) attachCertificates(ctx context.Context, pilots []domain.CertifiedPilot) ([]domain.CertifiedPilotDetail, error) {
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
			if !certValid(c) {
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
