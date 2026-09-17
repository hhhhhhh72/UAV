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

// containsID 证书 ID 去重：本次随申请提交的证书若因查重命中已有记录，
// 会与"已备案的有效证书"列表重叠，直接用会重复关联。
func containsID(ids []string, id string) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

// ownerHasValidCert 核验飞手档案所关联的证书中，是否仍有"approved 且未过期"的。
// 审批时复核用：覆盖"申请时有效、审批时已过期/被撤销"的空档——认证飞手身份
// 必须始终有有效证书支撑，无证不批（产品口径：不允许先审核后补证）。
func (s *TrainingService) ownerHasValidCert(ctx context.Context, p domain.CertifiedPilot) bool {
	if len(p.CertIDs) == 0 {
		return false
	}
	certs, err := s.certRepo.ListByUser(ctx, p.UserID)
	if err != nil {
		return false // 取不到证书时不放行：门禁宁可保守
	}
	for _, c := range certs {
		if !certValid(c) {
			continue
		}
		for _, id := range p.CertIDs {
			if c.ID == id {
				return true
			}
		}
	}
	return false
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

var (
	// ErrTrainingNotFound 培训/认证类记录不存在（Handler → 404）。
	ErrTrainingNotFound = errors.New("记录不存在")
	// ErrTrainingStateConflict 当前状态不允许该操作（Handler → 409）。
	ErrTrainingStateConflict = errors.New("当前状态不允许该操作")
	// ErrAdminRequired 需要管理员权限（Handler → 403）。
	ErrAdminRequired = errors.New("admin permission required")
	// ErrInvalidInput 入参不合法（Handler → 400）。
	ErrInvalidInput = errors.New("参数不合法")
)

func (s *TrainingService) ApproveCertificate(ctx context.Context, a domain.Actor, id string) (domain.Certificate, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Certificate{}, ErrAdminRequired
	}
	// 状态机前置：approved 幂等；已驳回不得翻转为通过（纠错请重走发证/审核流程）。
	cur, err := s.certRepo.FindByID(ctx, id)
	if err != nil {
		return domain.Certificate{}, notFoundErr(ErrTrainingNotFound, "certificate", id, err)
	}
	if cur.Status == "approved" {
		return cur, nil
	}
	if cur.Status == "rejected" {
		return domain.Certificate{}, fmt.Errorf("%w：已驳回的证书不能改为通过", ErrTrainingStateConflict)
	}
	c, err := s.certRepo.UpdateStatus(ctx, id, "approved")
	if err != nil {
		return domain.Certificate{}, notFoundErr(ErrTrainingNotFound, "certificate", id, err)
	}
	return c, nil
}

func (s *TrainingService) ListMyCertificates(ctx context.Context, a domain.Actor) ([]domain.Certificate, error) {
	return s.certRepo.ListByUser(ctx, a.ID)
}

// RejectCertificate 管理端驳回证书（用户可重新提交覆盖为 pending）。
func (s *TrainingService) RejectCertificate(ctx context.Context, a domain.Actor, id string) (domain.Certificate, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Certificate{}, ErrAdminRequired
	}
	cur, err := s.certRepo.FindByID(ctx, id)
	if err != nil {
		return domain.Certificate{}, notFoundErr(ErrTrainingNotFound, "certificate", id, err)
	}
	if cur.Status == "approved" {
		return domain.Certificate{}, fmt.Errorf("%w：已通过的证书不能驳回，请走吊销流程", ErrTrainingStateConflict)
	}
	if cur.Status == "rejected" {
		return cur, nil
	}
	c, err := s.certRepo.UpdateStatus(ctx, id, "rejected")
	if err != nil {
		return domain.Certificate{}, notFoundErr(ErrTrainingNotFound, "certificate", id, err)
	}
	return c, nil
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

// remainOf 剩余名额 = max(0, 总名额 - 已报)。总名额为 0 表示不限额，剩余无意义，返回 0。
//
// remain 是**派生值**，不是独立字段：建课/改课时按此公式算，报名时由仓储的 BumpEnrolled 重算。
// 此前建课直接采信表单里手填的 remain——它与 max_students 是两个独立输入框，
// 于是 max=11 / remain=10 这种不一致从建课那一刻就存在；而 BumpEnrolled 只在
// **有人报名时**才纠正，没人报名就永远是错的（生产上已实际出现，列表页据此
// 显示「仅剩 10 个」，详情页却按容量算出「已报 0 / 11」，同一门课两个答案）。
func remainOf(maxStudents, enrolledCount int) int {
	if maxStudents <= 0 {
		return 0
	}
	if left := maxStudents - enrolledCount; left > 0 {
		return left
	}
	return 0
}

// CreateCourse 接收完整领域对象（含小程序页面字段 org_name/rating/district/courses 等）。
func (s *TrainingService) CreateCourse(ctx context.Context, a domain.Actor, c domain.TrainingCourse) (domain.TrainingCourse, error) {
	if c.PriceFen < 0 {
		return domain.TrainingCourse{}, errors.New("price cannot be negative")
	}
	now := time.Now()
	if c.ID == "" {
		c.ID = nextID("course")
	}
	// 课程归属决定学费结算方向（completeEnrollment 用 Release(学员, course.OrgID, 学费)）。
	//
	// 默认挂发布者本人——个人/企业发布必须如此：若能指定任意 OrgID，
	// 就能把别人的学费结算指向自己或第三方账户。
	// 平台/协会管理员例外：他们要代机构建课（协会的课挂协会、协会替商家建课），
	// 需要能显式指定归属；未指定时同样回落到本人，语义不变。
	//
	// 此前是无条件 c.OrgID = a.ID，谁都指定不了机构归属。
	if c.OrgID == "" || (a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
		c.OrgID = a.ID
	}
	if c.Status == "" {
		c.Status = "draft"
	}
	// 派生值，不采信表单：建课时已报数为 0，所以剩余名额等于总名额。
	c.Remain = remainOf(c.MaxStudents, c.EnrolledCount)
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
	// 已报数与剩余名额是累计/派生值，一律**不接受表单传入**：
	//   - 管理端表单不带 enrolled_count，采信它会把已报数清零（内存实现是整条替换，尤其明显）；
	//   - remain 由总名额与已报数推出，此前直接写表单里的 remain，与 max_students 无约束。
	c.EnrolledCount = old.EnrolledCount
	c.Remain = remainOf(c.MaxStudents, old.EnrolledCount)
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
		return domain.Instructor{}, ErrAdminRequired
	}
	// 状态机前置：approved 幂等；已驳回不得翻转为通过。
	cur, err := s.instructorRepo.FindByID(ctx, id)
	if err != nil {
		return domain.Instructor{}, notFoundErr(ErrTrainingNotFound, "instructor", id, err)
	}
	if cur.Status == "approved" {
		return cur, nil
	}
	if cur.Status == "rejected" {
		return domain.Instructor{}, fmt.Errorf("%w：已驳回的培训师不能改为通过", ErrTrainingStateConflict)
	}
	i, err := s.instructorRepo.UpdateStatus(ctx, id, "approved")
	if err != nil {
		return domain.Instructor{}, notFoundErr(ErrTrainingNotFound, "instructor", id, err)
	}
	return i, nil
}

func (s *TrainingService) ListInstructors(ctx context.Context) ([]domain.Instructor, error) {
	return s.instructorRepo.List(ctx)
}

// ---- Certified Pilots ----

// PilotCertInput 随飞手认证申请一并提交的证书。
// 字段与 AddCertificate 一致；ExpireDate 为零值表示长期有效。
type PilotCertInput struct {
	CertType   domain.CertType
	CertNumber string
	Level      string
	IssuerOrg  string
	ImageURL   string
	IssueDate  time.Time
	ExpireDate time.Time
}

// RegisterPilot 申请飞手认证（不随附证书）：证书已在平台备案的用户走这里。
// 需要随申请一并提交证书的用 RegisterPilotWithCerts。
func (s *TrainingService) RegisterPilot(ctx context.Context, a domain.Actor, realName, idCard string, flightHours int, bio, avatar, region string) (domain.CertifiedPilot, error) {
	return s.RegisterPilotWithCerts(ctx, a, realName, idCard, flightHours, bio, avatar, region, nil)
}

// RegisterPilotWithCerts 申请飞手认证，可随申请一并提交证书（合并审核）。
//
// 合并之前：用户必须先把证书**单独**提交给管理端、审核通过后才允许申请飞手认证——
// 同一批证据走两道人工审核，且申请页只能提示"请先提交证书"（没有填写位置）。
// 现在证书随申请一并落 pending 并关联进档案，管理端审一次飞手申请即同时裁定这批证书
// （ApprovePilot / RejectPilot 会联动它们的状态）。
//
// 门禁口径不变，仍是"无证不批"：有已备案的有效证书、或本次提交了证书，二者其一即可。
func (s *TrainingService) RegisterPilotWithCerts(ctx context.Context, a domain.Actor, realName, idCard string, flightHours int, bio, avatar, region string, certs []PilotCertInput) (domain.CertifiedPilot, error) {
	// 0) 无副作用的重复申请预检，必须放在建证书**之前**：
	//    否则已认证/审核中的用户再提交一次，证书会被建出来而申请被拒，
	//    留下永远没人审的孤儿 pending 证书。
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
			}
		}
	}

	// 1) 随申请提交的证书先落库（pending，等本次审核一并裁定）。
	//    AddCertificate 自带查重：撞号或他人已占用直接报错，不静默吞掉。
	submitted := make([]string, 0, len(certs))
	for _, in := range certs {
		c, err := s.AddCertificate(ctx, a, in.CertType, in.CertNumber, in.Level, in.IssuerOrg, in.ImageURL, in.IssueDate, in.ExpireDate)
		if err != nil {
			return domain.CertifiedPilot{}, err
		}
		submitted = append(submitted, c.ID)
	}

	// 2) 已备案的有效证书（approved 且未过期）+ 本次提交的证书，共同构成申请关联的证书集。
	//    本次提交的是 pending，certValid 不认，所以必须显式并入。
	certIDs := []string{}
	if existing, err := s.certRepo.ListByUser(ctx, a.ID); err == nil {
		for _, c := range existing {
			if certValid(c) {
				certIDs = append(certIDs, c.ID)
			}
		}
	}
	for _, id := range submitted {
		if !containsID(certIDs, id) {
			certIDs = append(certIDs, id)
		}
	}
	// 无证不批：既没有已备案的有效证书，本次也没提交任何证书 → 拒绝申请
	if len(certIDs) == 0 {
		return domain.CertifiedPilot{}, errors.New("请至少提交一张证书（如 CAAC / AOPA / 大疆 UTC）后再申请飞手认证")
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
				// 重提即清掉上一轮的驳回理由，否则待审核记录会一直挂着它
				//（管理端列表与用户端"审核中"卡片都会显示这条过期理由）。
				e.RejectReason = ""
				return s.pilotRepo.Update(ctx, e)
			}
		}
	}
	p := domain.CertifiedPilot{ID: nextID("pilot"), UserID: a.ID, RealName: realName,
		IDCard: idCard, Avatar: avatar, Region: region, CertIDs: certIDs, FlightHours: flightHours, Bio: bio,
		Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.pilotRepo.Create(ctx, p)
}

// syncAttachedCerts 把档案所关联的证书从 fromStatus 批量改为 toStatus（合并审核联动）。
//
// 只处理 p.CertIDs 名单内的证书：用户后来**单独**提交、尚未随申请进入档案的证书不受影响，
// 避免"审一次飞手申请"把无关的待审证书也一并放行/驳回。
// 某一类证书没有状态机前置校验（证书本身只有 pending/approved/rejected），
// 因此这里只做状态过滤，不引入新的业务判断。
func (s *TrainingService) syncAttachedCerts(ctx context.Context, p domain.CertifiedPilot, fromStatus, toStatus string) error {
	if len(p.CertIDs) == 0 {
		return nil
	}
	certs, err := s.certRepo.ListByUser(ctx, p.UserID)
	if err != nil {
		return fmt.Errorf("list certificates for pilot %s: %w", p.ID, err)
	}
	for _, c := range certs {
		if c.Status != fromStatus || !containsID(p.CertIDs, c.ID) {
			continue
		}
		if _, err := s.certRepo.UpdateStatus(ctx, c.ID, toStatus); err != nil {
			return fmt.Errorf("set certificate %s to %s: %w", c.ID, toStatus, err)
		}
	}
	return nil
}

func (s *TrainingService) ApprovePilot(ctx context.Context, a domain.Actor, id string) (domain.CertifiedPilot, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.CertifiedPilot{}, ErrAdminRequired
	}
	// 状态机前置：approved 幂等；已驳回不得翻转为通过（驳回后需重新申请产生新记录）。
	cur, err := s.pilotRepo.FindByID(ctx, id)
	if err != nil {
		return domain.CertifiedPilot{}, notFoundErr(ErrTrainingNotFound, "pilot", id, err)
	}
	if cur.Status == "approved" {
		return cur, nil
	}
	if cur.Status == "rejected" {
		return domain.CertifiedPilot{}, fmt.Errorf("%w：已驳回的飞手申请不能改为通过", ErrTrainingStateConflict)
	}
	// 合并审核：随申请提交的证书此刻还是 pending——先把**档案关联的 pending 证书**
	// 置为 approved（管理端这次审的就是这批证据），再做下面的"仍有有效证书"复核。
	// 若先复核再放行证书，pending 永远过不了 certValid，合并审核就死在这里了。
	if err := s.syncAttachedCerts(ctx, cur, "pending", "approved"); err != nil {
		return domain.CertifiedPilot{}, err
	}
	// 审批门禁（与申请同规则）：批准时复核申请人仍持有至少一张未过期的 approved 证书。
	// 产品决策：不允许"先审核后补证"——申请时无证已被拒，审批时证书若已过期/被撤销也不得放行。
	if !s.ownerHasValidCert(ctx, cur) {
		return domain.CertifiedPilot{}, fmt.Errorf("%w：申请人当前没有有效的证书，不能通过飞手认证", ErrTrainingStateConflict)
	}
	p, err := s.pilotRepo.UpdateStatus(ctx, id, "approved")
	if err != nil {
		return domain.CertifiedPilot{}, notFoundErr(ErrTrainingNotFound, "pilot", id, err)
	}
	return p, nil
}

// RejectPilot 驳回飞手认证申请（管理员），reason 为驳回理由（审核留痕）。
// 状态机前置：rejected 幂等；approved 可驳回（撤销已通过的认证——纠错需连带
// 降级身份，因此留痕 reason 必填）；pending/rejected 常规驳回。
func (s *TrainingService) RejectPilot(ctx context.Context, a domain.Actor, id, reason string) (domain.CertifiedPilot, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.CertifiedPilot{}, ErrAdminRequired
	}
	if strings.TrimSpace(reason) == "" {
		return domain.CertifiedPilot{}, fmt.Errorf("%w：驳回理由必填", ErrInvalidInput)
	}
	cur, err := s.pilotRepo.FindByID(ctx, id)
	if err != nil {
		return domain.CertifiedPilot{}, notFoundErr(ErrTrainingNotFound, "pilot", id, err)
	}
	if cur.Status == "rejected" {
		return cur, nil
	}
	// 联同驳回：档案关联的 pending 证书一并置为 rejected。
	// 否则会出现"飞手申请被驳回、随附证书还挂在管理端待审列表"的不一致。
	// 只动 pending 的：已 approved 的证书不因飞手身份被撤销而作废——
	// 撤销认证属于纠错，不等于证书本身造假。
	if err := s.syncAttachedCerts(ctx, cur, "pending", "rejected"); err != nil {
		return domain.CertifiedPilot{}, err
	}
	p, err := s.pilotRepo.UpdateReject(ctx, id, reason)
	if err != nil {
		return domain.CertifiedPilot{}, notFoundErr(ErrTrainingNotFound, "pilot", id, err)
	}
	return p, nil
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

// GetPilotReviewDetail 管理端审核用：返回档案 + 该用户的**全部**证书。
//
// 与 GetPilotDetail 的区别是刻意的：后者按 certValid 过滤（只给已通过且未过期的），
// 那是"展示已认证飞手"的口径；审核要看的是"申请人提交了什么"——
// 待审的、被驳回的、已过期的都必须能看到，否则审核人无从核对。
func (s *TrainingService) GetPilotReviewDetail(ctx context.Context, id string) (domain.PilotReviewDetail, error) {
	p, err := s.pilotRepo.FindByID(ctx, id)
	if err != nil || p.ID == "" {
		return domain.PilotReviewDetail{}, ErrResourceNotFound
	}
	certs, err := s.certRepo.ListByUser(ctx, p.UserID)
	if err != nil {
		return domain.PilotReviewDetail{}, fmt.Errorf("list certificates for pilot %s: %w", p.ID, err)
	}
	return domain.PilotReviewDetail{CertifiedPilot: p, Certificates: certs}, nil
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
