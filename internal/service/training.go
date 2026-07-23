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

func (s *TrainingService) RegisterPilot(a domain.Actor, realName string) (domain.CertifiedPilot, error) {
	now := time.Now()
	p := domain.CertifiedPilot{ID: fmt.Sprintf("pilot-%d", now.UnixNano()), UserID: a.ID, RealName: realName,
		Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.pilotRepo.Create(p)
}

func (s *TrainingService) ApprovePilot(a domain.Actor, id string) (domain.CertifiedPilot, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.CertifiedPilot{}, errors.New("admin permission required")
	}
	return s.pilotRepo.UpdateStatus(id, "approved")
}

func (s *TrainingService) ListPilots() ([]domain.CertifiedPilot, error) {
	return s.pilotRepo.List()
}
