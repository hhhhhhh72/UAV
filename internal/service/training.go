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

func (s *TrainingService) CreateCourse(a domain.Actor, title string, certType domain.CertType, desc, location string, start, end time.Time, maxStudents int, priceFen int64) (domain.TrainingCourse, error) {
	now := time.Now()
	c := domain.TrainingCourse{ID: fmt.Sprintf("course-%d", now.UnixNano()), OrgID: a.ID, Title: title,
		CertType: certType, Description: desc, StartDate: start, EndDate: end,
		MaxStudents: maxStudents, Location: location, PriceFen: priceFen, Status: "draft",
		Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.courseRepo.Create(c)
}

func (s *TrainingService) ListCourses() ([]domain.TrainingCourse, error) {
	return s.courseRepo.List()
}

func (s *TrainingService) UpdateCourse(id string, title, certType, desc, location string, start, end time.Time, maxStudents int, priceFen int64, status string) (domain.TrainingCourse, error) {
	c, err := s.courseRepo.FindByID(id)
	if err != nil {
		return domain.TrainingCourse{}, err
	}
	c.Title = title
	c.CertType = domain.CertType(certType)
	c.Description = desc
	c.Location = location
	c.StartDate = start
	c.EndDate = end
	c.MaxStudents = maxStudents
	c.PriceFen = priceFen
	c.Status = status
	return s.courseRepo.Update(c)
}

func (s *TrainingService) DeleteCourse(id string) error {
	return s.courseRepo.Delete(id)
}

// ---- Instructors ----

func (s *TrainingService) RegisterInstructor(a domain.Actor, name, bio, orgID string, certTypes []string) (domain.Instructor, error) {
	now := time.Now()
	i := domain.Instructor{ID: fmt.Sprintf("instructor-%d", now.UnixNano()), UserID: a.ID, Name: name,
		CertTypes: certTypes, Bio: bio, OrgID: orgID, Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
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
