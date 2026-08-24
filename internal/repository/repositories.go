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
	"errors"
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

// UploadRepository 记录每次文件上传（台账 + 按用户配额统计）。
type UploadRepository interface {
	Create(ctx context.Context, rec domain.FileRecord) error
	// FindByID 按 ID 查上传台账（私有文件归属校验用）。
	FindByID(ctx context.Context, id string) (domain.FileRecord, error)
	// SumBytesSince 统计 owner 自 since 起的累计上传字节数。
	SumBytesSince(ctx context.Context, ownerID string, since time.Time) (int64, error)
}

// UserRepository manages platform user accounts and their role assignments.
type UserRepository interface {
	FindByOpenID(ctx context.Context, openid string) (domain.User, error)
	Create(ctx context.Context, u domain.User) (domain.User, error)
	FindByID(ctx context.Context, id string) (domain.User, error)
	All(ctx context.Context) ([]domain.User, error)
	// Count 统计用户总数（首页 stats 计数，聚合查询不物化行）。
	Count(ctx context.Context) (int, error)
	UpdateRole(ctx context.Context, id string, role domain.Role) error
	UpdateAvatar(ctx context.Context, userID, avatarURL string) error
	UpdateName(ctx context.Context, userID, name string) error
	// UpdateProfile 更新个人资料扩展字段（性别/生日/地区/简介/手机号）。
	// Phone 为明文，由实现加密后落库；空 Phone 表示不修改手机号。
	UpdateProfile(ctx context.Context, id string, p domain.UserProfile) error
	Delete(ctx context.Context, id string) error
}

// RefreshTokenRepository manages JWT refresh tokens for the rotating-token auth scheme.
type RefreshTokenRepository interface {
	Store(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	Find(ctx context.Context, tokenHash string) (userID string, expiresAt time.Time, revoked bool, err error)
	Revoke(ctx context.Context, tokenHash string) error
	// Consume 原子消费一枚有效的刷新令牌：仅当令牌未撤销且未过期时删除该行并
	// 返回 found=true——两个并发 refresh 用同一令牌时只有一个成功（防 TOCTOU 双签发）。
	Consume(ctx context.Context, tokenHash string) (found bool, userID string, expiresAt time.Time, err error)
}

// DemandFilter carries optional query parameters for listing demands.
type DemandFilter struct{ District, BizType, Sort, Status string }

// DemandRepository manages demand lifecycle: create, list, search, and status transitions.
// The CompareAndSetStatus method provides atomic CAS for concurrent bid selection safety.
type DemandRepository interface {
	Create(ctx context.Context, v domain.Demand) (domain.Demand, error)
	Update(ctx context.Context, d domain.Demand) (domain.Demand, error)
	FindByID(ctx context.Context, id string) (domain.Demand, error)
	List(ctx context.Context, v DemandFilter) ([]domain.Demand, error)    // 公开语义：仅已发布
	ListAll(ctx context.Context, v DemandFilter) ([]domain.Demand, error) // 管理端全量（含待审核等）
	// ListTop 公开语义（仅已发布）按 created_at 倒序取前 limit 条（首页 Top-N，
	// SQL 端 LIMIT，避免整表拉取）。
	ListTop(ctx context.Context, v DemandFilter, limit int) ([]domain.Demand, error)
	// Count 按 filter 统计条数（ListAll 语义：status 为空统计全部；
	// 首页公开计数传 Status=published）。
	Count(ctx context.Context, v DemandFilter) (int, error)
	Search(ctx context.Context, v string) ([]domain.Demand, error)
	// ListByPublisher 返回某发布者的全部需求（全状态），供"我的"页统计/查询。
	ListByPublisher(ctx context.Context, publisherID string) ([]domain.Demand, error)
	SetStatus(ctx context.Context, id string, status domain.DemandStatus) (domain.Demand, error)
	CompareAndSetStatus(ctx context.Context, id string, oldStatus, newStatus domain.DemandStatus) (bool, domain.Demand, error)
	Delete(ctx context.Context, id string) error
	// 需求收藏（按 (user_id, demand_id) 唯一）
	FavoriteDemand(ctx context.Context, userID, demandID string) error
	UnfavoriteDemand(ctx context.Context, userID, demandID string) error
	ListFavoriteDemandIDs(ctx context.Context, userID string) ([]string, error)
	// ListFavoriteDemands 按收藏时间倒序返回完整需求（我的收藏列表）。
	ListFavoriteDemands(ctx context.Context, userID string) ([]domain.Demand, error)
}

// EnterpriseRepository manages enterprise registrations and the admin review workflow.
type EnterpriseRepository interface {
	Create(ctx context.Context, v domain.Enterprise) (domain.Enterprise, error)
	Update(ctx context.Context, id string, e domain.Enterprise) (domain.Enterprise, error)
	FindByID(ctx context.Context, id string) (domain.Enterprise, error)
	FindByOwner(ctx context.Context, userID string) ([]domain.Enterprise, error)
	ListByStatus(ctx context.Context, status string, offset, limit int) ([]domain.Enterprise, int, error)
	Pending(ctx context.Context) ([]domain.Enterprise, error)
	Search(ctx context.Context, v string) ([]domain.Enterprise, error)
	Delete(ctx context.Context, id string) error
	AddDocument(ctx context.Context, v domain.EnterpriseDocument) (domain.EnterpriseDocument, error)
	ListDocuments(ctx context.Context, enterpriseID string) ([]domain.EnterpriseDocument, error)
}

type EmploymentRepository interface {
	Create(ctx context.Context, v domain.EmploymentRequest) (domain.EmploymentRequest, error)
	ListByEnterprise(ctx context.Context, enterpriseID string, offset, limit int) ([]domain.EmploymentRequest, int, error)
	ListAll(ctx context.Context, offset, limit int) ([]domain.EmploymentRequest, int, error)
}

// ContractRepository manages contracts and their signing lifecycle via webhook callbacks.
// ContractTemplateRepository manages reusable contract templates.
type ContractTemplateRepository interface {
	List(ctx context.Context) ([]domain.ContractTemplate, error)
	Create(ctx context.Context, v domain.ContractTemplate) (domain.ContractTemplate, error)
}

type ContractRepository interface {
	Create(ctx context.Context, v domain.Contract) (domain.Contract, error)
	ListByEnterprise(ctx context.Context, enterpriseID string, offset, limit int) ([]domain.Contract, int, error)
	ListAll(ctx context.Context, offset, limit int) ([]domain.Contract, int, error)
	FindByID(ctx context.Context, id string) (domain.Contract, error)
	UpdateStatus(ctx context.Context, id string, status domain.ContractStatus) (domain.Contract, error)
}

type JobRepository interface {
	Create(ctx context.Context, v domain.Job) (domain.Job, error)
	Update(ctx context.Context, id string, j domain.Job) (domain.Job, error)
	FindByID(ctx context.Context, id string) (domain.Job, error)
	ListByEnterprise(ctx context.Context, eid string) ([]domain.Job, error)
	ListAll(ctx context.Context, offset, limit int) ([]domain.Job, int, error) // 管理端全量（含草稿）
	ListPublished(ctx context.Context, offset, limit int) ([]domain.Job, int, error)
	Delete(ctx context.Context, id string) error
}

type ResumeRepository interface {
	Create(ctx context.Context, v domain.Resume) (domain.Resume, error)
	Update(ctx context.Context, id string, r domain.Resume) (domain.Resume, error)
	FindByID(ctx context.Context, id string) (domain.Resume, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Resume, error)
	ListAll(ctx context.Context, offset, limit int) ([]domain.Resume, int, error)
	// ListByIDs 批量按 ID 取简历（ListApplicantsForJob 防 N+1）。
	ListByIDs(ctx context.Context, ids []string) ([]domain.Resume, error)
}

type JobApplicationRepository interface {
	Create(ctx context.Context, v domain.JobApplication) (domain.JobApplication, error)
	FindByID(ctx context.Context, id string) (domain.JobApplication, error)
	UpdateStatus(ctx context.Context, id string, status domain.AppStatus) (domain.JobApplication, error)
	ListByJob(ctx context.Context, jobID string) ([]domain.JobApplication, error)
	ListByApplicant(ctx context.Context, userID string) ([]domain.JobApplication, error)
}

type PostRepository interface {
	Create(ctx context.Context, v domain.Post) (domain.Post, error)
	Update(ctx context.Context, id string, p domain.Post) (domain.Post, error)
	FindByID(ctx context.Context, id string) (domain.Post, error)
	ListPublished(ctx context.Context, offset, limit int) ([]domain.Post, int, error)
	ListAll(ctx context.Context, offset, limit int) ([]domain.Post, int, error) // 管理端全量（含 pending），审核上架入口
	ListByAuthor(ctx context.Context, userID string) ([]domain.Post, error)
}

type CommentRepository interface {
	Create(ctx context.Context, v domain.Comment) (domain.Comment, error)
	ListByPost(ctx context.Context, postID string) ([]domain.Comment, error)
}

type ReportRepository interface {
	Create(ctx context.Context, v domain.Report) (domain.Report, error)
	ListPending(ctx context.Context, offset, limit int) ([]domain.Report, int, error)
}

type ListingRepository interface {
	Create(ctx context.Context, v domain.Listing) (domain.Listing, error)
	Update(ctx context.Context, id string, l domain.Listing) (domain.Listing, error)
	FindByID(ctx context.Context, id string) (domain.Listing, error)
	ListByStatus(ctx context.Context, status string, offset, limit int) ([]domain.Listing, int, error)
	ListBySeller(ctx context.Context, userID string) ([]domain.Listing, error)
	AddFavorite(ctx context.Context, listingID, userID string) error
	RemoveFavorite(ctx context.Context, listingID, userID string) error
}

type LabourOrderRepository interface {
	Create(ctx context.Context, v domain.LabourOrder) (domain.LabourOrder, error)
	FindByID(ctx context.Context, id string) (domain.LabourOrder, error)
	ListByEmployer(ctx context.Context, userID string) ([]domain.LabourOrder, error)
	ListAll(ctx context.Context, offset, limit int) ([]domain.LabourOrder, int, error)
	CreateQuote(ctx context.Context, v domain.LabourQuote) (domain.LabourQuote, error)
	ListQuotes(ctx context.Context, orderID string) ([]domain.LabourQuote, error)
	CreateAssignment(ctx context.Context, v domain.Assignment) (domain.Assignment, error)
	ListAssignmentsByOrder(ctx context.Context, orderID string) ([]domain.Assignment, error)
	ListAssignmentsByWorker(ctx context.Context, workerID string) ([]domain.Assignment, error)
}

// BidRepository manages demand bids (quotations). Bids are created by bidders,
// listed per demand, and atomically accepted via UpdateStatus.
type BidRepository interface {
	Create(ctx context.Context, v domain.DemandBid) (domain.DemandBid, error)
	FindByID(ctx context.Context, id string) (domain.DemandBid, error)
	ListByDemand(ctx context.Context, demandID string) ([]domain.DemandBid, error)
	ListByBidder(ctx context.Context, bidderID string) ([]domain.DemandBid, error)
	UpdateStatus(ctx context.Context, id string, status string) (domain.DemandBid, error)
}

// IntentRepository manages demand contact intents (联系对接模式).
type IntentRepository interface {
	Create(ctx context.Context, v domain.DemandIntent) (domain.DemandIntent, error)
	ListByDemand(ctx context.Context, demandID string) ([]domain.DemandIntent, error)
	ListByIntentor(ctx context.Context, intentorID string) ([]domain.DemandIntent, error)
	UpdateStatus(ctx context.Context, id string, status string) (domain.DemandIntent, error)
}

// WorkOrderRepository manages work orders generated from confirmed intents (接单派单闭环).
type WorkOrderRepository interface {
	Create(ctx context.Context, v domain.WorkOrder) (domain.WorkOrder, error)
	FindByID(ctx context.Context, id string) (domain.WorkOrder, error)
	ListByPublisher(ctx context.Context, publisherID string) ([]domain.WorkOrder, error)
	ListByWorker(ctx context.Context, workerID string) ([]domain.WorkOrder, error)
	// UpdateStatus CAS 语义：仅当当前状态 == oldStatus 时更新为新状态，
	// 否则返回错误（并发取消/开始作业防已取消订单复活）。
	UpdateStatus(ctx context.Context, id string, oldStatus, status domain.WorkOrderStatus) (domain.WorkOrder, error)
	UpdatePhotos(ctx context.Context, id string, photos []string) (domain.WorkOrder, error)
	UpdateRework(ctx context.Context, id string, note string) (domain.WorkOrder, error)
	UpdateCancel(ctx context.Context, id string, reason string) (domain.WorkOrder, error)
}

// ---- Phase 3+ Repositories (migrated from in-memory services) ----

// CertificateRepository manages drone operation certificates.
type CertificateRepository interface {
	Create(ctx context.Context, v domain.Certificate) (domain.Certificate, error)
	FindByID(ctx context.Context, id string) (domain.Certificate, error)
	FindByNumber(ctx context.Context, certNumber string) (domain.Certificate, error) // 幂等发证查重
	ListByUser(ctx context.Context, userID string) ([]domain.Certificate, error)
	UpdateStatus(ctx context.Context, id string, status string) (domain.Certificate, error)
	ListAll(ctx context.Context) ([]domain.Certificate, error)
	Update(ctx context.Context, v domain.Certificate) (domain.Certificate, error)
	Delete(ctx context.Context, id string) error
}

// CourseRepository manages training courses.
type CourseRepository interface {
	Create(ctx context.Context, v domain.TrainingCourse) (domain.TrainingCourse, error)
	List(ctx context.Context) ([]domain.TrainingCourse, error)
	FindByID(ctx context.Context, id string) (domain.TrainingCourse, error)
	Update(ctx context.Context, v domain.TrainingCourse) (domain.TrainingCourse, error)
	Delete(ctx context.Context, id string) error
	// FavoriteCourse/UnfavoriteCourse/ListFavoriteCourses 培训课程收藏（我的收藏列表）。
	FavoriteCourse(ctx context.Context, userID, courseID string) error
	UnfavoriteCourse(ctx context.Context, userID, courseID string) error
	ListFavoriteCourses(ctx context.Context, userID string) ([]domain.TrainingCourse, error)
}

// InstructorRepository manages certified instructors.
type InstructorRepository interface {
	Create(ctx context.Context, v domain.Instructor) (domain.Instructor, error)
	FindByID(ctx context.Context, id string) (domain.Instructor, error)
	List(ctx context.Context) ([]domain.Instructor, error)
	UpdateStatus(ctx context.Context, id string, status string) (domain.Instructor, error)
	// Update 覆盖更新（驳回后重新提交用，与 PilotRepository 对齐）
	Update(ctx context.Context, v domain.Instructor) (domain.Instructor, error)
}

// PilotRepository manages certified pilots.
type PilotRepository interface {
	Create(ctx context.Context, v domain.CertifiedPilot) (domain.CertifiedPilot, error)
	FindByID(ctx context.Context, id string) (domain.CertifiedPilot, error)
	List(ctx context.Context) ([]domain.CertifiedPilot, error)
	Update(ctx context.Context, v domain.CertifiedPilot) (domain.CertifiedPilot, error) // 被驳回后重新申请（覆盖重提）
	UpdateStatus(ctx context.Context, id string, status string) (domain.CertifiedPilot, error)
	UpdateReject(ctx context.Context, id string, reason string) (domain.CertifiedPilot, error) // 驳回并记录理由
}

// ProductRepository manages drone product listings.
type ProductRepository interface {
	Create(ctx context.Context, v domain.DroneProduct) (domain.DroneProduct, error)
	FindByID(ctx context.Context, id string) (domain.DroneProduct, error)
	List(ctx context.Context, prodType string) ([]domain.DroneProduct, error)
	// ListTop 按创建时间倒序取前 limit 条（首页 Top-N，SQL 端 LIMIT 不整表）。
	ListTop(ctx context.Context, prodType string, limit int) ([]domain.DroneProduct, error)
	// ListByIDs 批量按 ID 取商品（订单列表补商品名防 N+1）。
	ListByIDs(ctx context.Context, ids []string) ([]domain.DroneProduct, error)
	// SumViews 商品浏览量总和（可选按类型；首页 stats.views 聚合查询）。
	SumViews(ctx context.Context, prodType string) (int, error)
	Update(ctx context.Context, p domain.DroneProduct) (domain.DroneProduct, error)
	Delete(ctx context.Context, id string) error
	IncrementViews(ctx context.Context, id string) error
	// MarkSold 条件更新：仅 listed/空状态可标记 sold（下单抢占，防一物多卖/超卖）；
	// 状态非可售返回错误。
	MarkSold(ctx context.Context, id string) error
	// Restore 条件更新：sold → listed（订单创建失败回滚用）。
	Restore(ctx context.Context, id string) error
	// FavoriteProduct/UnfavoriteProduct/ListFavoriteProducts 商品收藏（我的收藏列表）。
	FavoriteProduct(ctx context.Context, userID, productID string) error
	UnfavoriteProduct(ctx context.Context, userID, productID string) error
	ListFavoriteProducts(ctx context.Context, userID string) ([]domain.DroneProduct, error)
}

// ServiceListingRepository manages enterprise service capability listings (PRD ②-2).
type ServiceListingRepository interface {
	Create(ctx context.Context, v domain.ServiceListing) (domain.ServiceListing, error)
	FindByID(ctx context.Context, id string) (domain.ServiceListing, error)
	List(ctx context.Context) ([]domain.ServiceListing, error)
	Update(ctx context.Context, sl domain.ServiceListing) (domain.ServiceListing, error)
	Delete(ctx context.Context, id string) error
	// FavoriteListing/UnfavoriteListing/ListFavoriteListings 服务能力收藏（我的收藏列表）。
	FavoriteListing(ctx context.Context, userID, listingID string) error
	UnfavoriteListing(ctx context.Context, userID, listingID string) error
	ListFavoriteListings(ctx context.Context, userID string) ([]domain.ServiceListing, error)
}

// RepairRepository manages repair orders.
type RepairRepository interface {
	Create(ctx context.Context, v domain.RepairOrder) (domain.RepairOrder, error)
	ListByUser(ctx context.Context, userID string) ([]domain.RepairOrder, error)
	ListAll(ctx context.Context, offset, limit int) ([]domain.RepairOrder, int, error)
}

// PolicyRepository manages insurance policies.
type PolicyRepository interface {
	Create(ctx context.Context, v domain.InsurancePolicy) (domain.InsurancePolicy, error)
	ListByUser(ctx context.Context, userID string) ([]domain.InsurancePolicy, error)
	ListAll(ctx context.Context, offset, limit int) ([]domain.InsurancePolicy, int, error)
}

// InspectionRepository manages annual drone inspections.
type InspectionRepository interface {
	Create(ctx context.Context, v domain.AnnualInspection) (domain.AnnualInspection, error)
	ListByUser(ctx context.Context, userID string) ([]domain.AnnualInspection, error)
	ListAll(ctx context.Context) ([]domain.AnnualInspection, error)
}

// LoanRepository manages loan applications.
type LoanRepository interface {
	Create(ctx context.Context, v domain.LoanApplication) (domain.LoanApplication, error)
	ListByUser(ctx context.Context, userID string) ([]domain.LoanApplication, error)
	ListAll(ctx context.Context, offset, limit int) ([]domain.LoanApplication, int, error)
}

// MessageRepository manages in-app messages.
type MessageRepository interface {
	Create(ctx context.Context, v domain.Message) (domain.Message, error)
	FindByID(ctx context.Context, id string) (domain.Message, error)
	ListByUser(ctx context.Context, userID string, unreadOnly bool) ([]domain.Message, error)
	MarkRead(ctx context.Context, id string) (domain.Message, error)
	UnreadCount(ctx context.Context, userID string) (int, error)
	ListAll(ctx context.Context, offset, limit int) ([]domain.Message, int, error)
	Delete(ctx context.Context, id string) error
}

// ApplicationRepository manages service applications (miniprogram /api/submit).
type ApplicationRepository interface {
	Create(ctx context.Context, v domain.Application) (domain.Application, error)
	FindByID(ctx context.Context, id string) (domain.Application, error)
	ListByUser(ctx context.Context, userID string, offset, limit int) ([]domain.Application, int, error)
	ListAll(ctx context.Context, offset, limit int) ([]domain.Application, int, error)
}

// ArticleRepository manages news articles.
type ArticleRepository interface {
	Create(ctx context.Context, v domain.Article) (domain.Article, error)
	FindByID(ctx context.Context, id string) (domain.Article, error)
	Update(ctx context.Context, v domain.Article) (domain.Article, error)
	ListByCategory(ctx context.Context, category string, offset, limit int) ([]domain.Article, int, error)
}

// ReviewRepository manages user reviews.
type ReviewRepository interface {
	Create(ctx context.Context, v domain.Review) (domain.Review, error)
	ListByTarget(ctx context.Context, targetType, targetID string) ([]domain.Review, error)
	ListAll(ctx context.Context, status string, offset, limit int) ([]domain.Review, int, error)
	UpdateStatus(ctx context.Context, id string, status string) (domain.Review, error)
	Delete(ctx context.Context, id string) error
}

// VenueRepository manages venues and bookings.
type VenueRepository interface {
	Create(ctx context.Context, v domain.Venue) (domain.Venue, error)
	List(ctx context.Context, venueType string) ([]domain.Venue, error)
	FindByID(ctx context.Context, id string) (domain.Venue, error)
	CreateBooking(ctx context.Context, v domain.VenueBooking) (domain.VenueBooking, error)
	ListBookings(ctx context.Context, venueID string) ([]domain.VenueBooking, error)
}

// EnrollmentRepository manages training course enrollments.
type EnrollmentRepository interface {
	Create(ctx context.Context, v domain.Enrollment) (domain.Enrollment, error)
	Update(ctx context.Context, e domain.Enrollment) (domain.Enrollment, error)
	FindByID(ctx context.Context, id string) (domain.Enrollment, error) // 管理端编辑时取旧状态做防回退校验
	ListByCourse(ctx context.Context, courseID string) ([]domain.Enrollment, error)
	// ListByUser 某用户全部报名（"我的报名"一次查询，避免按课程 N+1）。
	ListByUser(ctx context.Context, userID string) ([]domain.Enrollment, error)
	ListAll(ctx context.Context, offset, limit int) ([]domain.Enrollment, int, error) // 管理端全量
	FindByUserAndCourse(ctx context.Context, userID, courseID string) (domain.Enrollment, bool, error)
}

// TradeOrderRepository manages marketplace trade orders.
type TradeOrderRepository interface {
	Create(ctx context.Context, v domain.TradeOrder) (domain.TradeOrder, error)
	FindByID(ctx context.Context, id string) (domain.TradeOrder, error)
	UpdateStatus(ctx context.Context, id string, status string) (domain.TradeOrder, error)
	// CompareAndSetStatus 原子状态迁移：仅当前状态等于 oldStatus 时更新为 newStatus，
	// 返回 bool 表示是否迁移成功（false = 状态已并发变更），防止后写覆盖前写。
	CompareAndSetStatus(ctx context.Context, id, oldStatus, newStatus string) (bool, domain.TradeOrder, error)
	// UpdateAftersale 一次性写订单状态 + 售后字段（申请/审核结案共用），只写这两类列
	UpdateAftersale(ctx context.Context, o domain.TradeOrder) (domain.TradeOrder, error)
	ListByUser(ctx context.Context, userID string) ([]domain.TradeOrder, error)
	ListAll(ctx context.Context, offset, limit int) ([]domain.TradeOrder, int, error)
	Delete(ctx context.Context, id string) error
}

// EscrowRepository manages escrow accounts and transactions.
// 资金操作哨兵错误（C6 修复）：Service 层用 errors.Is 判断，
// 避免依赖字符串匹配。
var (
	ErrInsufficientBalance       = errors.New("insufficient balance")
	ErrInsufficientFrozenBalance = errors.New("insufficient frozen balance")
)

type EscrowRepository interface {
	GetAccount(ctx context.Context, userID string) (domain.EscrowAccount, error)
	// 以下四个资金方法必须原子：调整余额并写入流水，全成或全败。
	// 余额/冻结不足时返回 ErrInsufficientBalance / ErrInsufficientFrozenBalance。
	// 旧接口的读-改-写（GetAccount→UpsertAccount）存在并发丢更新，已移除。
	Deposit(ctx context.Context, userID string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error)
	Freeze(ctx context.Context, userID string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error)
	Release(ctx context.Context, fromUser, toUser string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error)
	Refund(ctx context.Context, userID string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error)
	ListTransactions(ctx context.Context, userID string) ([]domain.EscrowTransaction, error)
	// HasReleased 报告 fromUser 对 (reference_type, reference_id) 是否已有完成（status='completed'）的
	// release 流水——completeEnrollment 幂等重试判断"学费是否已释放"用（避免重复释放）。
	HasReleased(ctx context.Context, fromUser, refType, refID string) (bool, error)
	// ListOrphanFreezes 列出"冻结但无对应业务记录"的孤儿冻结流水
	// （ref_type/ref_id 指定的业务记录不存在，且冻结时间早于 olderThan），
	// 供自动补偿解冻（如培训报名冻结后进程崩溃，报名未落库）。
	ListOrphanFreezes(ctx context.Context, refType string, olderThan time.Time, limit int) ([]domain.EscrowTransaction, error)
}

// ---- New Business Module Repositories ----

// ExpertRepository manages think-tank experts.
type ExpertRepository interface {
	Create(ctx context.Context, v domain.Expert) (domain.Expert, error)
	FindByID(ctx context.Context, id string) (domain.Expert, error)
	List(ctx context.Context, field string) ([]domain.Expert, error)
	Update(ctx context.Context, v domain.Expert) (domain.Expert, error)
	Delete(ctx context.Context, id string) error
}

// CaseRepository manages project cases.
type CaseRepository interface {
	Create(ctx context.Context, v domain.CaseEntry) (domain.CaseEntry, error)
	FindByID(ctx context.Context, id string) (domain.CaseEntry, error)
	List(ctx context.Context, category string, offset, limit int) ([]domain.CaseEntry, int, error)
	Update(ctx context.Context, v domain.CaseEntry) (domain.CaseEntry, error)
	Delete(ctx context.Context, id string) error
}

// ComplianceRepository manages compliance docs and standards.
type ComplianceRepository interface {
	CreateDoc(ctx context.Context, v domain.ComplianceDoc) (domain.ComplianceDoc, error)
	FindDocByID(ctx context.Context, id string) (domain.ComplianceDoc, error)
	ListDocs(ctx context.Context, category string, offset, limit int) ([]domain.ComplianceDoc, int, error)
	UpdateDoc(ctx context.Context, v domain.ComplianceDoc) (domain.ComplianceDoc, error)
	DeleteDoc(ctx context.Context, id string) error
	DeleteStandard(ctx context.Context, id string) error
	FindStandardByID(ctx context.Context, id string) (domain.StandardDoc, error)
	UpdateStandard(ctx context.Context, v domain.StandardDoc) (domain.StandardDoc, error)
	CreateStandard(ctx context.Context, v domain.StandardDoc) (domain.StandardDoc, error)
	ListStandards(ctx context.Context, category string, offset, limit int) ([]domain.StandardDoc, int, error)
}

// AchievementRepository manages technology achievements.
type AchievementRepository interface {
	Create(ctx context.Context, v domain.Achievement) (domain.Achievement, error)
	FindByID(ctx context.Context, id string) (domain.Achievement, error)
	List(ctx context.Context, field string, offset, limit int) ([]domain.Achievement, int, error)
	Update(ctx context.Context, v domain.Achievement) (domain.Achievement, error)
	Delete(ctx context.Context, id string) error
}

// RDChallengeRepository manages enterprise R&D challenges.
type RDChallengeRepository interface {
	Create(ctx context.Context, v domain.RDChallenge) (domain.RDChallenge, error)
	FindByID(ctx context.Context, id string) (domain.RDChallenge, error)
	List(ctx context.Context, field string, offset, limit int) ([]domain.RDChallenge, int, error)
	Update(ctx context.Context, v domain.RDChallenge) (domain.RDChallenge, error)
	Delete(ctx context.Context, id string) error
}

// ResearchProjectRepository manages joint research projects.
type ResearchProjectRepository interface {
	Create(ctx context.Context, v domain.ResearchProject) (domain.ResearchProject, error)
	FindByID(ctx context.Context, id string) (domain.ResearchProject, error)
	List(ctx context.Context, offset, limit int) ([]domain.ResearchProject, int, error)
	Update(ctx context.Context, v domain.ResearchProject) (domain.ResearchProject, error)
	Delete(ctx context.Context, id string) error
}

// ProjectAppRepository manages project subsidy applications.
type ProjectAppRepository interface {
	Create(ctx context.Context, v domain.ProjectApplication) (domain.ProjectApplication, error)
	FindByID(ctx context.Context, id string) (domain.ProjectApplication, error)
	ListByUser(ctx context.Context, userID string) ([]domain.ProjectApplication, error)
	ListAll(ctx context.Context, status string, offset, limit int) ([]domain.ProjectApplication, int, error)
	Update(ctx context.Context, v domain.ProjectApplication) (domain.ProjectApplication, error)
}

// CompetitionRepository manages competitions and registrations.
type CompetitionRepository interface {
	Create(ctx context.Context, v domain.Competition) (domain.Competition, error)
	FindByID(ctx context.Context, id string) (domain.Competition, error)
	List(ctx context.Context, offset, limit int) ([]domain.Competition, int, error)
	Update(ctx context.Context, v domain.Competition) (domain.Competition, error)
	Delete(ctx context.Context, id string) error
	CreateReg(ctx context.Context, v domain.CompetitionReg) (domain.CompetitionReg, error)
	ListRegs(ctx context.Context, competitionID string) ([]domain.CompetitionReg, error)
}

// EventRepository manages association events and registrations.
type EventRepository interface {
	Create(ctx context.Context, v domain.AssociationEvent) (domain.AssociationEvent, error)
	FindByID(ctx context.Context, id string) (domain.AssociationEvent, error)
	List(ctx context.Context, offset, limit int) ([]domain.AssociationEvent, int, error)
	Update(ctx context.Context, v domain.AssociationEvent) (domain.AssociationEvent, error)
	Delete(ctx context.Context, id string) error
	CreateReg(ctx context.Context, v domain.EventRegistration) (domain.EventRegistration, error)
	ListRegs(ctx context.Context, eventID string) ([]domain.EventRegistration, error)
}

// PortfolioRepository manages member brand portfolios.
type PortfolioRepository interface {
	Create(ctx context.Context, v domain.MemberPortfolio) (domain.MemberPortfolio, error)
	FindByID(ctx context.Context, id string) (domain.MemberPortfolio, error)
	ListByEnterprise(ctx context.Context, eid string) ([]domain.MemberPortfolio, error)
	List(ctx context.Context, offset, limit int) ([]domain.MemberPortfolio, int, error) // 管理端全量（含草稿/待审）
	ListPublished(ctx context.Context, offset, limit int) ([]domain.MemberPortfolio, int, error)
	Update(ctx context.Context, v domain.MemberPortfolio) (domain.MemberPortfolio, error)
	Delete(ctx context.Context, id string) error
}

// IndustryReportRepository manages industry reports.
type IndustryReportRepository interface {
	Create(ctx context.Context, v domain.IndustryReport) (domain.IndustryReport, error)
	FindByID(ctx context.Context, id string) (domain.IndustryReport, error)
	List(ctx context.Context, offset, limit int) ([]domain.IndustryReport, int, error)
	Update(ctx context.Context, v domain.IndustryReport) (domain.IndustryReport, error)
	Delete(ctx context.Context, id string) error
}

// ResourceRepository manages industry resources (drones, airfields, test sites).
type ResourceRepository interface {
	Create(ctx context.Context, v domain.IndustryResource) (domain.IndustryResource, error)
	FindByID(ctx context.Context, id string) (domain.IndustryResource, error)
	List(ctx context.Context, resType string, offset, limit int) ([]domain.IndustryResource, int, error)
	Update(ctx context.Context, v domain.IndustryResource) (domain.IndustryResource, error)
	Delete(ctx context.Context, id string) error
	// Bookings (C11: 小程序资源预约 → POST /api/v1/industry-resources/{id}/book)
	CreateBooking(ctx context.Context, v domain.IndustryResourceBooking) (domain.IndustryResourceBooking, error)
	ListBookingsByResource(ctx context.Context, resourceID string) ([]domain.IndustryResourceBooking, error)
	ListBookingsByUser(ctx context.Context, userID string) ([]domain.IndustryResourceBooking, error)
}

// EmergencyRepository manages emergency resources and dispatches.
type EmergencyRepository interface {
	CreateResource(ctx context.Context, v domain.EmergencyResource) (domain.EmergencyResource, error)
	FindResourceByID(ctx context.Context, id string) (domain.EmergencyResource, error)
	ListResources(ctx context.Context, resType, q string, offset, limit int) ([]domain.EmergencyResource, int, error)
	UpdateResource(ctx context.Context, v domain.EmergencyResource) (domain.EmergencyResource, error)
	DeleteResource(ctx context.Context, id string) error
	FindDispatchByID(ctx context.Context, id string) (domain.EmergencyDispatch, error)
	UpdateDispatch(ctx context.Context, v domain.EmergencyDispatch) (domain.EmergencyDispatch, error)
	DeleteDispatch(ctx context.Context, id string) error
	CreateDispatch(ctx context.Context, v domain.EmergencyDispatch) (domain.EmergencyDispatch, error)
	// ListDispatches 分页列出调度记录；resourceID 非空时仅返回该资源的调度；
	// 返回项内嵌 related 资源摘要（名称/类型/状态）。
	ListDispatches(ctx context.Context, resourceID string, offset, limit int) ([]domain.EmergencyDispatch, int, error)
}

// ── Batch1: 产业资源池 + 测试预约 + 展会 (per .doc) ──

type ResourcePoolRepository interface {
	Create(ctx context.Context, v domain.ResourcePool) (domain.ResourcePool, error)
	FindByID(ctx context.Context, id string) (domain.ResourcePool, error)
	List(ctx context.Context, poolType string) ([]domain.ResourcePool, error)
	AddMember(ctx context.Context, v domain.ResourcePoolMember) (domain.ResourcePoolMember, error)
	ListMembers(ctx context.Context, poolID string) ([]domain.ResourcePoolMember, error)
}

type TestSiteRepository interface {
	Create(ctx context.Context, v domain.TestSite) (domain.TestSite, error)
	FindByID(ctx context.Context, id string) (domain.TestSite, error)
	List(ctx context.Context, siteType string) ([]domain.TestSite, error)
	CreateBooking(ctx context.Context, v domain.TestSiteBooking) (domain.TestSiteBooking, error)
	UpdateBookingStatus(ctx context.Context, id, status, note string) (domain.TestSiteBooking, error)
	UpdateSite(ctx context.Context, v domain.TestSite) (domain.TestSite, error)
	DeleteSite(ctx context.Context, id string) error
	ListBookings(ctx context.Context, siteID string) ([]domain.TestSiteBooking, error)
	ListBookingsByUser(ctx context.Context, userID string) ([]domain.TestSiteBooking, error)       // 我的预约（按用户）
	ListAllBookings(ctx context.Context, offset, limit int) ([]domain.TestSiteBooking, int, error) // 管理端全量
}

type ExhibitionRepository interface {
	Create(ctx context.Context, v domain.Exhibition) (domain.Exhibition, error)
	FindByID(ctx context.Context, id string) (domain.Exhibition, error)
	List(ctx context.Context, offset, limit int) ([]domain.Exhibition, int, error)
	Update(ctx context.Context, v domain.Exhibition) (domain.Exhibition, error)
	Delete(ctx context.Context, id string) error
	CreateBooth(ctx context.Context, v domain.ExhibitionBooth) (domain.ExhibitionBooth, error)
	ListBooths(ctx context.Context, exhibitionID string) ([]domain.ExhibitionBooth, error)
	UpdateBoothStatus(ctx context.Context, id, status string) (domain.ExhibitionBooth, error)
}

// ── Batch2: 成果转化 + 院校 + 校企 (per .doc) ──

type TransformationRepository interface {
	Create(ctx context.Context, v domain.Transformation) (domain.Transformation, error)
	FindByID(ctx context.Context, id string) (domain.Transformation, error)
	List(ctx context.Context, ownerID string) ([]domain.Transformation, error)
	Update(ctx context.Context, v domain.Transformation) (domain.Transformation, error)
	Delete(ctx context.Context, id string) error
}

type CollegeRepository interface {
	Create(ctx context.Context, v domain.College) (domain.College, error)
	FindByID(ctx context.Context, id string) (domain.College, error)
	List(ctx context.Context, region string) ([]domain.College, error)
	Update(ctx context.Context, v domain.College) (domain.College, error)
	Delete(ctx context.Context, id string) error
}

// StudyTourRepository manages research-study tours.
type StudyTourRepository interface {
	Create(ctx context.Context, v domain.StudyTour) (domain.StudyTour, error)
	FindByID(ctx context.Context, id string) (domain.StudyTour, error)
	List(ctx context.Context) ([]domain.StudyTour, error)
	Update(ctx context.Context, v domain.StudyTour) (domain.StudyTour, error)
	Delete(ctx context.Context, id string) error
}

type CooperationRepository interface {
	Create(ctx context.Context, v domain.CooperationProgram) (domain.CooperationProgram, error)
	FindByID(ctx context.Context, id string) (domain.CooperationProgram, error)
	List(ctx context.Context, enterpriseID string) ([]domain.CooperationProgram, error)
	UpdateStatus(ctx context.Context, id, status string) (domain.CooperationProgram, error)
}

// ── Batch3: 救援案例 + 应急对接 + 协会权限 (per .doc) ──

type RescueCaseRepository interface {
	Create(ctx context.Context, v domain.RescueCase) (domain.RescueCase, error)
	FindByID(ctx context.Context, id string) (domain.RescueCase, error)
	List(ctx context.Context, eventType, q string, offset, limit int) ([]domain.RescueCase, int, error)
}

type EmergencyDeptRepository interface {
	CreateDept(ctx context.Context, v domain.EmergencyDept) (domain.EmergencyDept, error)
	ListDepts(ctx context.Context) ([]domain.EmergencyDept, error)
	CreateDrill(ctx context.Context, v domain.EmergencyDrill) (domain.EmergencyDrill, error)
	ListDrills(ctx context.Context, deptID string) ([]domain.EmergencyDrill, error)
}

type AssociationMemberRepository interface {
	Create(ctx context.Context, v domain.AssociationMember) (domain.AssociationMember, error)
	FindByUserID(ctx context.Context, userID string) (domain.AssociationMember, error)
	List(ctx context.Context, role string, offset, limit int) ([]domain.AssociationMember, int, error)
	UpdateRole(ctx context.Context, id string, role domain.AssociationRole) (domain.AssociationMember, error)
}
