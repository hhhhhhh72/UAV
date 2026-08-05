// Package repository defines the data access contracts (interfaces) for all
// domain entities. Implementations live in sub-packages:
//   - memory/   — in-memory storage with sync.RWMutex, suitable for development
//   - postgres/ — PostgreSQL storage via pgxpool, suitable for production
//
// The golden rule: repository methods accept and return domain types only.
// No role checks, no business logic, no HTTP knowledge belongs here.
package repository

import (
	"context"
	"time"

	"drone-platform/internal/domain"
)

// AuditEntry records a single write operation for compliance and traceability.
// Audit entries are written by the HTTP server's audit() method after every
// significant action (create, approve, reject, void, etc.).
type AuditEntry struct {
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	Result       string
	RequestID    string
	Metadata     map[string]any
}

// AuditWriter is implemented by storage backends that persist audit logs.
type AuditWriter interface {
	WriteAudit(ctx context.Context, entry AuditEntry) error
}

// UserRepository manages platform user accounts and their role assignments.
type UserRepository interface {
	FindByOpenID(openid string) (domain.User, error)
	Create(u domain.User) (domain.User, error)
	FindByID(id string) (domain.User, error)
	All() ([]domain.User, error)
	UpdateRole(id string, role domain.Role) error
	UpdateAvatar(userID, avatarURL string) error
	Delete(id string) error
}

// RefreshTokenRepository manages JWT refresh tokens for the rotating-token auth scheme.
type RefreshTokenRepository interface {
	Store(userID, tokenHash string, expiresAt time.Time) error
	Find(tokenHash string) (userID string, expiresAt time.Time, revoked bool, err error)
	Revoke(tokenHash string) error
}

// DemandFilter carries optional query parameters for listing demands.
type DemandFilter struct{ District, BizType, Sort, Status string }

// DemandRepository manages demand lifecycle: create, list, search, and status transitions.
// The CompareAndSetStatus method provides atomic CAS for concurrent bid selection safety.
type DemandRepository interface {
	Create(domain.Demand) (domain.Demand, error)
	Update(d domain.Demand) (domain.Demand, error)
	FindByID(id string) (domain.Demand, error)
	List(DemandFilter) ([]domain.Demand, error)
	Search(string) ([]domain.Demand, error)
	SetStatus(id string, status domain.DemandStatus) (domain.Demand, error)
	CompareAndSetStatus(id string, oldStatus, newStatus domain.DemandStatus) (bool, domain.Demand, error)
}

// EnterpriseRepository manages enterprise registrations and the admin review workflow.
type EnterpriseRepository interface {
	Create(domain.Enterprise) (domain.Enterprise, error)
	Update(id string, e domain.Enterprise) (domain.Enterprise, error)
	FindByID(id string) (domain.Enterprise, error)
	FindByOwner(userID string) ([]domain.Enterprise, error)
	ListByStatus(status string, offset, limit int) ([]domain.Enterprise, int, error)
	Pending() ([]domain.Enterprise, error)
	Search(string) ([]domain.Enterprise, error)
	Delete(id string) error
}

type ShopRepository interface {
	Create(domain.Shop) (domain.Shop, error)
	Update(domain.Shop) (domain.Shop, error)
	FindByID(id string) (domain.Shop, error)
	List(offset, limit int) ([]domain.Shop, int, error)
	Delete(id string) error
}

type EmploymentRepository interface {
	Create(domain.EmploymentRequest) (domain.EmploymentRequest, error)
	ListByEnterprise(enterpriseID string, offset, limit int) ([]domain.EmploymentRequest, int, error)
	ListAll(offset, limit int) ([]domain.EmploymentRequest, int, error)
}

// ContractRepository manages contracts and their signing lifecycle via webhook callbacks.
type ContractRepository interface {
	Create(domain.Contract) (domain.Contract, error)
	ListByEnterprise(enterpriseID string, offset, limit int) ([]domain.Contract, int, error)
	ListAll(offset, limit int) ([]domain.Contract, int, error)
	FindByID(id string) (domain.Contract, error)
	UpdateStatus(id string, status domain.ContractStatus) (domain.Contract, error)
}

type JobRepository interface {
	Create(domain.Job) (domain.Job, error)
	Update(id string, j domain.Job) (domain.Job, error)
	FindByID(id string) (domain.Job, error)
	ListByEnterprise(eid string) ([]domain.Job, error)
	ListAll(offset, limit int) ([]domain.Job, int, error) // 管理端全量（含草稿）
	ListPublished(offset, limit int) ([]domain.Job, int, error)
	Delete(id string) error
}

type ResumeRepository interface {
	Create(domain.Resume) (domain.Resume, error)
	Update(id string, r domain.Resume) (domain.Resume, error)
	FindByID(id string) (domain.Resume, error)
	ListByUser(userID string) ([]domain.Resume, error)
}

type JobApplicationRepository interface {
	Create(domain.JobApplication) (domain.JobApplication, error)
	UpdateStatus(id string, status domain.AppStatus) (domain.JobApplication, error)
	ListByJob(jobID string) ([]domain.JobApplication, error)
	ListByApplicant(userID string) ([]domain.JobApplication, error)
}

type PostRepository interface {
	Create(domain.Post) (domain.Post, error)
	Update(id string, p domain.Post) (domain.Post, error)
	FindByID(id string) (domain.Post, error)
	ListPublished(offset, limit int) ([]domain.Post, int, error)
	ListByAuthor(userID string) ([]domain.Post, error)
}

type CommentRepository interface {
	Create(domain.Comment) (domain.Comment, error)
	ListByPost(postID string) ([]domain.Comment, error)
}

type ReportRepository interface {
	Create(domain.Report) (domain.Report, error)
	ListPending(offset, limit int) ([]domain.Report, int, error)
}

type ListingRepository interface {
	Create(domain.Listing) (domain.Listing, error)
	Update(id string, l domain.Listing) (domain.Listing, error)
	FindByID(id string) (domain.Listing, error)
	ListByStatus(status string, offset, limit int) ([]domain.Listing, int, error)
	ListBySeller(userID string) ([]domain.Listing, error)
	AddFavorite(listingID, userID string) error
	RemoveFavorite(listingID, userID string) error
}

type LabourOrderRepository interface {
	Create(domain.LabourOrder) (domain.LabourOrder, error)
	FindByID(id string) (domain.LabourOrder, error)
	ListByEmployer(userID string) ([]domain.LabourOrder, error)
	ListAll(offset, limit int) ([]domain.LabourOrder, int, error)
	CreateQuote(domain.LabourQuote) (domain.LabourQuote, error)
	ListQuotes(orderID string) ([]domain.LabourQuote, error)
	CreateAssignment(domain.Assignment) (domain.Assignment, error)
}

// BidRepository manages demand bids (quotations). Bids are created by bidders,
// listed per demand, and atomically accepted via UpdateStatus.
type BidRepository interface {
	Create(domain.DemandBid) (domain.DemandBid, error)
	FindByID(id string) (domain.DemandBid, error)
	ListByDemand(demandID string) ([]domain.DemandBid, error)
	ListByBidder(bidderID string) ([]domain.DemandBid, error)
	UpdateStatus(id string, status string) (domain.DemandBid, error)
}

// ---- Phase 3+ Repositories (migrated from in-memory services) ----

// CertificateRepository manages drone operation certificates.
type CertificateRepository interface {
	Create(domain.Certificate) (domain.Certificate, error)
	FindByID(id string) (domain.Certificate, error)
	ListByUser(userID string) ([]domain.Certificate, error)
	UpdateStatus(id string, status string) (domain.Certificate, error)
	ListAll() ([]domain.Certificate, error)
	Update(domain.Certificate) (domain.Certificate, error)
	Delete(id string) error
}

// CourseRepository manages training courses.
type CourseRepository interface {
	Create(domain.TrainingCourse) (domain.TrainingCourse, error)
	List() ([]domain.TrainingCourse, error)
	FindByID(id string) (domain.TrainingCourse, error)
	Update(domain.TrainingCourse) (domain.TrainingCourse, error)
	Delete(id string) error
}

// InstructorRepository manages certified instructors.
type InstructorRepository interface {
	Create(domain.Instructor) (domain.Instructor, error)
	FindByID(id string) (domain.Instructor, error)
	List() ([]domain.Instructor, error)
	UpdateStatus(id string, status string) (domain.Instructor, error)
}

// PilotRepository manages certified pilots.
type PilotRepository interface {
	Create(domain.CertifiedPilot) (domain.CertifiedPilot, error)
	FindByID(id string) (domain.CertifiedPilot, error)
	List() ([]domain.CertifiedPilot, error)
	Update(domain.CertifiedPilot) (domain.CertifiedPilot, error) // 被驳回后重新申请（覆盖重提）
	UpdateStatus(id string, status string) (domain.CertifiedPilot, error)
}

// ProductRepository manages drone product listings.
type ProductRepository interface {
	Create(domain.DroneProduct) (domain.DroneProduct, error)
	FindByID(id string) (domain.DroneProduct, error)
	List(prodType string) ([]domain.DroneProduct, error)
	Update(p domain.DroneProduct) (domain.DroneProduct, error)
	Delete(id string) error
	IncrementViews(id string) error
}

// RepairRepository manages repair orders.
type RepairRepository interface {
	Create(domain.RepairOrder) (domain.RepairOrder, error)
	ListByUser(userID string) ([]domain.RepairOrder, error)
}

// PolicyRepository manages insurance policies.
type PolicyRepository interface {
	Create(domain.InsurancePolicy) (domain.InsurancePolicy, error)
	ListByUser(userID string) ([]domain.InsurancePolicy, error)
}

// InspectionRepository manages annual drone inspections.
type InspectionRepository interface {
	Create(domain.AnnualInspection) (domain.AnnualInspection, error)
	ListByUser(userID string) ([]domain.AnnualInspection, error)
	ListAll() ([]domain.AnnualInspection, error)
}

// LoanRepository manages loan applications.
type LoanRepository interface {
	Create(domain.LoanApplication) (domain.LoanApplication, error)
	ListByUser(userID string) ([]domain.LoanApplication, error)
}

// MessageRepository manages in-app messages.
type MessageRepository interface {
	Create(domain.Message) (domain.Message, error)
	ListByUser(userID string, unreadOnly bool) ([]domain.Message, error)
	MarkRead(id string) (domain.Message, error)
	UnreadCount(userID string) (int, error)
	ListAll(offset, limit int) ([]domain.Message, int, error)
	Delete(id string) error
}

// ArticleRepository manages news articles.
type ArticleRepository interface {
	Create(domain.Article) (domain.Article, error)
	FindByID(id string) (domain.Article, error)
	Update(domain.Article) (domain.Article, error)
	ListByCategory(category string, offset, limit int) ([]domain.Article, int, error)
}

// ReviewRepository manages user reviews.
type ReviewRepository interface {
	Create(domain.Review) (domain.Review, error)
	ListByTarget(targetType, targetID string) ([]domain.Review, error)
	ListAll(status string, offset, limit int) ([]domain.Review, int, error)
	UpdateStatus(id string, status string) (domain.Review, error)
	Delete(id string) error
}

// VenueRepository manages venues and bookings.
type VenueRepository interface {
	Create(domain.Venue) (domain.Venue, error)
	List(venueType string) ([]domain.Venue, error)
	FindByID(id string) (domain.Venue, error)
	CreateBooking(domain.VenueBooking) (domain.VenueBooking, error)
	ListBookings(venueID string) ([]domain.VenueBooking, error)
}

// EnrollmentRepository manages training course enrollments.
type EnrollmentRepository interface {
	Create(domain.Enrollment) (domain.Enrollment, error)
	ListByCourse(courseID string) ([]domain.Enrollment, error)
	ListAll(offset, limit int) ([]domain.Enrollment, int, error) // 管理端全量
	FindByUserAndCourse(userID, courseID string) (domain.Enrollment, bool, error)
}

// TradeOrderRepository manages marketplace trade orders.
type TradeOrderRepository interface {
	Create(domain.TradeOrder) (domain.TradeOrder, error)
	FindByID(id string) (domain.TradeOrder, error)
	UpdateStatus(id string, status string) (domain.TradeOrder, error)
	ListByUser(userID string) ([]domain.TradeOrder, error)
	ListAll(offset, limit int) ([]domain.TradeOrder, int, error)
	Delete(id string) error
}

// EscrowRepository manages escrow accounts and transactions.
type EscrowRepository interface {
	GetAccount(userID string) (domain.EscrowAccount, error)
	UpsertAccount(domain.EscrowAccount) error
	CreateTransaction(domain.EscrowTransaction) (domain.EscrowTransaction, error)
	ListTransactions(userID string) ([]domain.EscrowTransaction, error)
}

// ---- New Business Module Repositories ----

// ExpertRepository manages think-tank experts.
type ExpertRepository interface {
	Create(domain.Expert) (domain.Expert, error)
	FindByID(id string) (domain.Expert, error)
	List(field string) ([]domain.Expert, error)
	Update(domain.Expert) (domain.Expert, error)
	Delete(id string) error
}

// CaseRepository manages project cases.
type CaseRepository interface {
	Create(domain.CaseEntry) (domain.CaseEntry, error)
	FindByID(id string) (domain.CaseEntry, error)
	List(category string, offset, limit int) ([]domain.CaseEntry, int, error)
	Update(domain.CaseEntry) (domain.CaseEntry, error)
	Delete(id string) error
}

// ComplianceRepository manages compliance docs and standards.
type ComplianceRepository interface {
	CreateDoc(domain.ComplianceDoc) (domain.ComplianceDoc, error)
	FindDocByID(id string) (domain.ComplianceDoc, error)
	ListDocs(category string, offset, limit int) ([]domain.ComplianceDoc, int, error)
	UpdateDoc(domain.ComplianceDoc) (domain.ComplianceDoc, error)
	DeleteDoc(id string) error
	DeleteStandard(id string) error
	FindStandardByID(id string) (domain.StandardDoc, error)
	UpdateStandard(domain.StandardDoc) (domain.StandardDoc, error)
	CreateStandard(domain.StandardDoc) (domain.StandardDoc, error)
	ListStandards(category string, offset, limit int) ([]domain.StandardDoc, int, error)
}

// AchievementRepository manages technology achievements.
type AchievementRepository interface {
	Create(domain.Achievement) (domain.Achievement, error)
	FindByID(id string) (domain.Achievement, error)
	List(field string, offset, limit int) ([]domain.Achievement, int, error)
	Update(domain.Achievement) (domain.Achievement, error)
	Delete(id string) error
}

// RDChallengeRepository manages enterprise R&D challenges.
type RDChallengeRepository interface {
	Create(domain.RDChallenge) (domain.RDChallenge, error)
	FindByID(id string) (domain.RDChallenge, error)
	List(field string, offset, limit int) ([]domain.RDChallenge, int, error)
	Update(domain.RDChallenge) (domain.RDChallenge, error)
	Delete(id string) error
}

// ResearchProjectRepository manages joint research projects.
type ResearchProjectRepository interface {
	Create(domain.ResearchProject) (domain.ResearchProject, error)
	FindByID(id string) (domain.ResearchProject, error)
	List(offset, limit int) ([]domain.ResearchProject, int, error)
	Update(domain.ResearchProject) (domain.ResearchProject, error)
	Delete(id string) error
}

// ProjectAppRepository manages project subsidy applications.
type ProjectAppRepository interface {
	Create(domain.ProjectApplication) (domain.ProjectApplication, error)
	FindByID(id string) (domain.ProjectApplication, error)
	ListByUser(userID string) ([]domain.ProjectApplication, error)
	ListAll(status string, offset, limit int) ([]domain.ProjectApplication, int, error)
	Update(domain.ProjectApplication) (domain.ProjectApplication, error)
}

// CompetitionRepository manages competitions and registrations.
type CompetitionRepository interface {
	Create(domain.Competition) (domain.Competition, error)
	FindByID(id string) (domain.Competition, error)
	List(offset, limit int) ([]domain.Competition, int, error)
	Update(domain.Competition) (domain.Competition, error)
	Delete(id string) error
	CreateReg(domain.CompetitionReg) (domain.CompetitionReg, error)
	ListRegs(competitionID string) ([]domain.CompetitionReg, error)
}

// EventRepository manages association events and registrations.
type EventRepository interface {
	Create(domain.AssociationEvent) (domain.AssociationEvent, error)
	FindByID(id string) (domain.AssociationEvent, error)
	List(offset, limit int) ([]domain.AssociationEvent, int, error)
	Update(domain.AssociationEvent) (domain.AssociationEvent, error)
	Delete(id string) error
	CreateReg(domain.EventRegistration) (domain.EventRegistration, error)
	ListRegs(eventID string) ([]domain.EventRegistration, error)
}

// PortfolioRepository manages member brand portfolios.
type PortfolioRepository interface {
	Create(domain.MemberPortfolio) (domain.MemberPortfolio, error)
	FindByID(id string) (domain.MemberPortfolio, error)
	ListByEnterprise(eid string) ([]domain.MemberPortfolio, error)
	List(offset, limit int) ([]domain.MemberPortfolio, int, error) // 管理端全量（含草稿/待审）
	ListPublished(offset, limit int) ([]domain.MemberPortfolio, int, error)
	Update(domain.MemberPortfolio) (domain.MemberPortfolio, error)
	Delete(id string) error
}

// IndustryReportRepository manages industry reports.
type IndustryReportRepository interface {
	Create(domain.IndustryReport) (domain.IndustryReport, error)
	FindByID(id string) (domain.IndustryReport, error)
	List(offset, limit int) ([]domain.IndustryReport, int, error)
	Update(domain.IndustryReport) (domain.IndustryReport, error)
	Delete(id string) error
}

// ResourceRepository manages industry resources (drones, airfields, test sites).
type ResourceRepository interface {
	Create(domain.IndustryResource) (domain.IndustryResource, error)
	FindByID(id string) (domain.IndustryResource, error)
	List(resType string, offset, limit int) ([]domain.IndustryResource, int, error)
	Update(domain.IndustryResource) (domain.IndustryResource, error)
	Delete(id string) error
}

// EmergencyRepository manages emergency resources and dispatches.
type EmergencyRepository interface {
	CreateResource(domain.EmergencyResource) (domain.EmergencyResource, error)
	FindResourceByID(id string) (domain.EmergencyResource, error)
	ListResources(offset, limit int) ([]domain.EmergencyResource, int, error)
	UpdateResource(domain.EmergencyResource) (domain.EmergencyResource, error)
	DeleteResource(id string) error
	FindDispatchByID(id string) (domain.EmergencyDispatch, error)
	UpdateDispatch(domain.EmergencyDispatch) (domain.EmergencyDispatch, error)
	DeleteDispatch(id string) error
	CreateDispatch(domain.EmergencyDispatch) (domain.EmergencyDispatch, error)
	ListDispatches(offset, limit int) ([]domain.EmergencyDispatch, int, error)
}

// ── Batch1: 产业资源池 + 测试预约 + 展会 (per .doc) ──

type ResourcePoolRepository interface {
	Create(domain.ResourcePool) (domain.ResourcePool, error)
	FindByID(id string) (domain.ResourcePool, error)
	List(poolType string) ([]domain.ResourcePool, error)
	AddMember(domain.ResourcePoolMember) (domain.ResourcePoolMember, error)
	ListMembers(poolID string) ([]domain.ResourcePoolMember, error)
}

type TestSiteRepository interface {
	Create(domain.TestSite) (domain.TestSite, error)
	FindByID(id string) (domain.TestSite, error)
	List(siteType string) ([]domain.TestSite, error)
	CreateBooking(domain.TestSiteBooking) (domain.TestSiteBooking, error)
	UpdateBookingStatus(id, status, note string) (domain.TestSiteBooking, error)
	UpdateSite(domain.TestSite) (domain.TestSite, error)
	DeleteSite(id string) error
	ListBookings(siteID string) ([]domain.TestSiteBooking, error)
	ListAllBookings(offset, limit int) ([]domain.TestSiteBooking, int, error) // 管理端全量
}

type ExhibitionRepository interface {
	Create(domain.Exhibition) (domain.Exhibition, error)
	FindByID(id string) (domain.Exhibition, error)
	List(offset, limit int) ([]domain.Exhibition, int, error)
	Update(domain.Exhibition) (domain.Exhibition, error)
	Delete(id string) error
	CreateBooth(domain.ExhibitionBooth) (domain.ExhibitionBooth, error)
	ListBooths(exhibitionID string) ([]domain.ExhibitionBooth, error)
	UpdateBoothStatus(id, status string) (domain.ExhibitionBooth, error)
}

// ── Batch2: 成果转化 + 院校 + 校企 (per .doc) ──

type TransformationRepository interface {
	Create(domain.Transformation) (domain.Transformation, error)
	FindByID(id string) (domain.Transformation, error)
	List(ownerID string) ([]domain.Transformation, error)
	Update(domain.Transformation) (domain.Transformation, error)
	Delete(id string) error
}

type CollegeRepository interface {
	Create(domain.College) (domain.College, error)
	FindByID(id string) (domain.College, error)
	List(region string) ([]domain.College, error)
	Update(domain.College) (domain.College, error)
	Delete(id string) error
}

// StudyTourRepository manages research-study tours.
type StudyTourRepository interface {
	Create(domain.StudyTour) (domain.StudyTour, error)
	FindByID(id string) (domain.StudyTour, error)
	List() ([]domain.StudyTour, error)
	Update(domain.StudyTour) (domain.StudyTour, error)
	Delete(id string) error
}

type CooperationRepository interface {
	Create(domain.CooperationProgram) (domain.CooperationProgram, error)
	FindByID(id string) (domain.CooperationProgram, error)
	List(enterpriseID string) ([]domain.CooperationProgram, error)
	UpdateStatus(id, status string) (domain.CooperationProgram, error)
}

// ── Batch3: 救援案例 + 应急对接 + 协会权限 (per .doc) ──

type RescueCaseRepository interface {
	Create(domain.RescueCase) (domain.RescueCase, error)
	FindByID(id string) (domain.RescueCase, error)
	List(eventType string, offset, limit int) ([]domain.RescueCase, int, error)
}

type EmergencyDeptRepository interface {
	CreateDept(domain.EmergencyDept) (domain.EmergencyDept, error)
	ListDepts() ([]domain.EmergencyDept, error)
	CreateDrill(domain.EmergencyDrill) (domain.EmergencyDrill, error)
	ListDrills(deptID string) ([]domain.EmergencyDrill, error)
}

type AssociationMemberRepository interface {
	Create(domain.AssociationMember) (domain.AssociationMember, error)
	FindByUserID(userID string) (domain.AssociationMember, error)
	List(role string, offset, limit int) ([]domain.AssociationMember, int, error)
	UpdateRole(id string, role domain.AssociationRole) (domain.AssociationMember, error)
}
