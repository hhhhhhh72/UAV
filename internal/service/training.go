package service

import (
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

func (s *TrainingService) AddCertificate(a domain.Actor, certType domain.CertType, certNumber, level, issuer string, issueDate, expireDate time.Time) (domain.Certificate, error) {
	now := time.Now()
	c := domain.Certificate{ID: fmt.Sprintf("cert-%d", now.UnixNano()), UserID: a.ID, CertType: certType,
		CertNumber: certNumber, Level: level, IssueDate: issueDate, ExpireDate: expireDate,
		IssuerOrg: issuer, Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.certRepo.Create(c)
}

func (s *TrainingService) ApproveCertificate(a domain.Actor, id string) (domain.Certificate, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Certificate{}, errors.New("admin permission required")
	}
	return s.certRepo.UpdateStatus(id, "approved")
}

func (s *TrainingService) ListMyCertificates(a domain.Actor) ([]domain.Certificate, error) {
	return s.certRepo.ListByUser(a.ID)
}

func (s *TrainingService) ListAllCertificates() ([]domain.Certificate, error) {
	return s.certRepo.ListAll()
}

func (s *TrainingService) GetCourse(id string) (domain.TrainingCourse, error) {
	return s.courseRepo.FindByID(id)
}

func (s *TrainingService) GetCert(id string) (domain.Certificate, error) {
	return s.certRepo.FindByID(id)
}

func (s *TrainingService) UpdateCertificate(id, certType, certNumber, level, issuer, status string, issueDate, expireDate time.Time) (domain.Certificate, error) {
	c, err := s.certRepo.FindByID(id)
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
	return s.certRepo.Update(c)
}

func (s *TrainingService) DeleteCertificate(id string) error {
	return s.certRepo.Delete(id)
}

// ---- Courses ----

// CreateCourse 接收完整领域对象（含小程序页面字段 org_name/rating/district/courses 等）。
func (s *TrainingService) CreateCourse(a domain.Actor, c domain.TrainingCourse) (domain.TrainingCourse, error) {
	now := time.Now()
	if c.ID == "" {
		c.ID = fmt.Sprintf("course-%d", now.UnixNano())
	}
	c.OrgID = a.ID
	if c.Status == "" {
		c.Status = "draft"
	}
	c.Version = 1
	c.CreatedAt = now
	c.UpdatedAt = now
	return s.courseRepo.Create(c)
}

func (s *TrainingService) ListCourses() ([]domain.TrainingCourse, error) {
	return s.courseRepo.List()
}

func (s *TrainingService) UpdateCourse(c domain.TrainingCourse) (domain.TrainingCourse, error) {
	old, err := s.courseRepo.FindByID(c.ID)
	if err != nil {
		return domain.TrainingCourse{}, err
	}
	c.Version = old.Version
	c.CreatedAt = old.CreatedAt // 保留原创建时间
	c.UpdatedAt = time.Now()
	return s.courseRepo.Update(c)
}

func (s *TrainingService) DeleteCourse(id string) error {
	return s.courseRepo.Delete(id)
}

// ---- Instructors ----

func (s *TrainingService) RegisterInstructor(a domain.Actor, name, photo, bio, orgID string, certTypes []string) (domain.Instructor, error) {
	now := time.Now()
	i := domain.Instructor{ID: fmt.Sprintf("instructor-%d", now.UnixNano()), UserID: a.ID, Name: name,
		Photo: photo, CertTypes: certTypes, Bio: bio, OrgID: orgID, Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.instructorRepo.Create(i)
}

func (s *TrainingService) ApproveInstructor(a domain.Actor, id string) (domain.Instructor, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Instructor{}, errors.New("admin permission required")
	}
	return s.instructorRepo.UpdateStatus(id, "approved")
}

func (s *TrainingService) ListInstructors() ([]domain.Instructor, error) {
	return s.instructorRepo.List()
}

// ---- Certified Pilots ----

func (s *TrainingService) RegisterPilot(a domain.Actor, realName, idCard string, flightHours int, bio string) (domain.CertifiedPilot, error) {
	// 自动关联已认证证书（审核管线 approved 状态，无手动勾选/造假空间）
	certIDs := []string{}
	if certs, err := s.certRepo.ListByUser(a.ID); err == nil {
		for _, c := range certs {
			if c.Status == "approved" {
				certIDs = append(certIDs, c.ID)
			}
		}
	}
	now := time.Now()
	// 已有记录：approved/pending 拒绝重复申请；rejected 覆盖重提（重置为 pending）
	if existing, err := s.pilotRepo.List(); err == nil {
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
				e.CertIDs = certIDs
				e.FlightHours = flightHours
				e.Bio = bio
				e.Status = "pending"
				return s.pilotRepo.Update(e)
			}
		}
	}
	p := domain.CertifiedPilot{ID: fmt.Sprintf("pilot-%d", now.UnixNano()), UserID: a.ID, RealName: realName,
		IDCard: idCard, CertIDs: certIDs, FlightHours: flightHours, Bio: bio,
		Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.pilotRepo.Create(p)
}

func (s *TrainingService) ApprovePilot(a domain.Actor, id string) (domain.CertifiedPilot, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.CertifiedPilot{}, errors.New("admin permission required")
	}
	return s.pilotRepo.UpdateStatus(id, "approved")
}

// RejectPilot 驳回飞手认证申请（管理员）。
func (s *TrainingService) RejectPilot(a domain.Actor, id string) (domain.CertifiedPilot, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.CertifiedPilot{}, errors.New("admin permission required")
	}
	return s.pilotRepo.UpdateStatus(id, "rejected")
}

func (s *TrainingService) ListPilots() ([]domain.CertifiedPilot, error) {
	return s.pilotRepo.List()
}

// GetPilotByOwner 查询我的飞手认证记录（未申请返回零值）。
func (s *TrainingService) GetPilotByOwner(userID string) (domain.CertifiedPilot, error) {
	pilots, err := s.pilotRepo.List()
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
