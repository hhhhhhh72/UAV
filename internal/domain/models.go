// Package domain defines the core business entities, status enums,
// and type constants used across all layers of the drone platform.
//
// This package is the single source of truth for data structures shared
// between handlers, services, and repositories. It contains no business
// logic, no HTTP concerns, and no database concerns.
//
// Status machines: every entity with a status field has a defined set of
// legal transitions enforced by the service layer. See the state diagrams
// in docs/business-flows.md.
package domain

import (
	"strings"
	"time"
)

// Role is a user role in the 4-level RBAC hierarchy.
// Roles are ordered: platform_admin > association_admin > enterprise > individual.
type Role string

const (
	// RolePlatformAdmin has full access to all resources and admin functions.
	RolePlatformAdmin Role = "platform_admin"
	// RoleAssociationAdmin can review enterprises, manage content, and view analytics.
	RoleAssociationAdmin Role = "association_admin"
	// RoleEnterprise can publish demands, post jobs, create contracts, and bid on work.
	RoleEnterprise Role = "enterprise"
	// RoleIndividual can bid on demands, apply for jobs, trade in the marketplace.
	RoleIndividual Role = "individual"
)

// Actor represents the authenticated identity of the caller.
// It is extracted from the Bearer token by auth middleware and stored
// in the request context for the duration of the request.
type Actor struct {
	ID   string `json:"id"`
	Role Role   `json:"role"`
	// TokenVersion 签发时的用户令牌版本：authenticate 与库中对比，
	// 不一致/用户被删/被禁 → 立即失效（0 兼容历史 token）。
	TokenVersion int64 `json:"tv,omitempty"`
}

// 用户账号状态（users.status）。鉴权中间件按"非 active 即失效"处理
// （见 httpapi.revalidateActor）：注销/封禁后已签发 token 立即 401。
const (
	UserActive  = "active"  // 正常
	UserBanned  = "banned"  // 封禁
	UserDeleted = "deleted" // 已注销（软删：账号不可登录，发布内容保留）
)

// User represents a registered platform user linked to a WeChat identity.
// Phone numbers are stored as AES-256-GCM ciphertext and excluded from JSON
// serialization (json:"-"). Use crypto.Decrypt or MaskPhone for display.
type User struct {
	ID           string `json:"id"`
	WechatOpenID string `json:"wechat_openid"`
	PhoneCipher  string `json:"-"` // AES-256-GCM encrypted, never serialized
	// PasswordHash is the bcrypt hash for password login. Never serialized.
	PasswordHash string `json:"-"`
	Name         string `json:"name"` // 昵称（users.name）
	AvatarURL    string `json:"avatar_url"`
	Gender       string `json:"gender"`   // 性别（男/女）
	Birthday     string `json:"birthday"` // 生日 YYYY-MM-DD
	Region       string `json:"region"`   // 所在地区
	Bio          string `json:"bio"`      // 个人简介
	Role         Role   `json:"role"`
	Status       string `json:"status"`
	// TokenVersion 令牌版本：删除/封禁/改角色时自增，使已签发 token 立即失效。
	TokenVersion int64      `json:"token_version"`
	Version      int        `json:"version"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// UserProfile carries editable profile fields for PATCH /api/v1/me.
// Phone is plaintext here; repositories encrypt it before persistence.
type UserProfile struct {
	Gender   string
	Birthday string
	Region   string
	Bio      string
	Phone    string
}

// EnterpriseStatus represents the approval lifecycle of a business registration.
type EnterpriseStatus string

// EnterpriseBrief 发布者企业摘要（需求/服务等详情页展示用，不含敏感字段）
type EnterpriseBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

const (
	// EnterpriseDraft is the initial state; the owner can edit freely.
	EnterpriseDraft EnterpriseStatus = "draft"
	// EnterpriseSubmitted means the registration is awaiting admin review.
	EnterpriseSubmitted EnterpriseStatus = "submitted"
	// EnterpriseSupplementRequired means the admin requested additional materials.
	EnterpriseSupplementRequired EnterpriseStatus = "supplement_required"
	// EnterpriseApproved means the business is fully onboarded.
	EnterpriseApproved EnterpriseStatus = "approved"
	// EnterpriseRejected means the registration was denied.
	EnterpriseRejected EnterpriseStatus = "rejected"
)

// Enterprise represents a registered business entity on the platform.
// Sensitive fields (LicenseURL, AccountName) are encrypted at rest.
type Enterprise struct {
	ID               string           `json:"id"`
	OwnerUserID      string           `json:"owner_user_id"`
	Name             string           `json:"name"`
	CreditCode       string           `json:"credit_code"`       // 统一社会信用代码
	LegalPerson      string           `json:"legal_person"`      // 法定代表人
	ContactPhone     string           `json:"contact_phone"`     // 联系电话
	IndustryCategory string           `json:"industry_category"` // 产业分类（整机/零部件/飞控/载荷/运营服务/实训院校…）
	Scale            string           `json:"scale"`             // 企业规模
	Address          string           `json:"address"`           // 地址
	Description      string           `json:"description"`       // 简介
	BusinessHours    string           `json:"business_hours"`    // 营业时间
	Logo             string           `json:"logo"`              // 机构 logo
	CoverImage       string           `json:"cover_image"`       // 机构封面图
	LicenseURL       string           `json:"license_url"`
	AccountName      string           `json:"account_name"`
	ContactPerson    string           `json:"contact_person"`  // 联系人（PRD FR-2.1）
	Email            string           `json:"email"`           // 邮箱
	FoundedAt        string           `json:"founded_at"`      // 成立时间（YYYY-MM）
	CapabilityTags   string           `json:"capability_tags"` // 能力标签，逗号分隔（预设标签库多选）
	Status           EnterpriseStatus `json:"status"`
	ReviewComment    string           `json:"review_comment"` // 审核意见：驳回/需补充原因，用户端展示
	IsMember         bool             `json:"is_member"`
	Version          int              `json:"version"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// EmploymentStatus represents the lifecycle of a labour dispatch request.
type EmploymentStatus string

const (
	EmploymentPending   EmploymentStatus = "pending"
	EmploymentQuoted    EmploymentStatus = "quoted"
	EmploymentConfirmed EmploymentStatus = "confirmed"
	EmploymentFulfilled EmploymentStatus = "fulfilled"
	EmploymentSettled   EmploymentStatus = "settled"
	EmploymentCancelled EmploymentStatus = "cancelled"
)

// EmploymentRequest is a labour dispatch order posted by an enterprise.
type EmploymentRequest struct {
	ID           string           `json:"id"`
	EnterpriseID string           `json:"enterprise_id"`
	Position     string           `json:"position"`
	Headcount    int              `json:"headcount"`
	StartDate    time.Time        `json:"start_date"`
	EndDate      time.Time        `json:"end_date"`
	Status       EmploymentStatus `json:"status"`
	Version      int              `json:"version"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// ContractStatus represents the signing lifecycle tracked via webhook callbacks.
type ContractStatus string

const (
	ContractDraft   ContractStatus = "draft"
	ContractSent    ContractStatus = "sent"
	ContractSigning ContractStatus = "signing"
	ContractSigned  ContractStatus = "signed"
	ContractVoided  ContractStatus = "voided"
	ContractExpired ContractStatus = "expired"
)

// Contract is a legal agreement between the platform and an enterprise,
// optionally tracked through an external e-signature service.
type Contract struct {
	ID           string         `json:"id"`
	EnterpriseID string         `json:"enterprise_id"`
	TemplateID   string         `json:"template_id"`
	Status       ContractStatus `json:"status"`
	SignURL      string         `json:"sign_url"`
	Version      int            `json:"version"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// ContractTemplate is a reusable contract text template (contract_templates table).
type ContractTemplate struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Version   int       `json:"version"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DefaultContractTemplates are the built-in templates seeded on first use
// (memory repo 初始化 / PG 种子迁移 000062)。
var DefaultContractTemplates = []ContractTemplate{
	{ID: "tpl-001", Name: "标准无人机服务合同", Version: 1, Status: "active"},
	{ID: "tpl-002", Name: "无人机买卖协议", Version: 1, Status: "active"},
}

// BizType categorises a demand by the type of drone operation needed.
type BizType string

const (
	// BizCableInspection covers power line inspection work.
	BizCableInspection BizType = "cable_inspection"
	// BizPlantTransport covers agricultural transport work.
	BizPlantTransport BizType = "plant_transport"
	// BizSprayPesticide covers crop spraying work.
	BizSprayPesticide BizType = "spray_pesticide"
	// BizCleanPaint covers cleaning and painting work.
	BizCleanPaint BizType = "clean_paint"
	// BizTradeLease covers buying, selling, and leasing.
	BizTradeLease BizType = "trade_lease"
	// BizOther covers all other types not classified above.
	BizOther BizType = "other"
)

// DemandStatus represents the lifecycle of a demand from draft to completion.
type DemandStatus string

const (
	DemandPending   DemandStatus = "pending"   // awaiting admin review
	DemandPublished DemandStatus = "published" // publicly visible with contact info
	DemandAssigned  DemandStatus = "assigned"  // 已接受意向并生成工单（锁新意向/接单）
	DemandCompleted DemandStatus = "completed" // marked done by publisher
	DemandCancelled DemandStatus = "cancelled" // withdrawn by publisher
	DemandRejected  DemandStatus = "rejected"  // declined by admin
)

// Demand is a job request posted by a publisher (enterprise or individual)
// that other users can bid on. It is the core entity of the platform.
type Demand struct {
	ID                  string           `json:"id"`
	PublisherID         string           `json:"publisher_id"`
	PublisherName       string           `json:"publisher_name"`
	PublisherEnterprise *EnterpriseBrief `json:"publisher_enterprise,omitempty"` // 发布者已认证企业摘要（无则省略）
	Contact             string           `json:"contact"`                        // encrypted at rest, masked in public responses
	BizType             BizType          `json:"biz_type"`
	District            string           `json:"district"`
	CityCode            string           `json:"city_code"`
	Title               string           `json:"title"`
	Description         string           `json:"description"`
	Images              []string         `json:"images"`
	Attachments         []string         `json:"attachments"` // 附件（图片/PDF，≤10MB，公开详情可下载）
	Aircraft            []string         `json:"aircraft"`    // 机型要求（多旋翼/固定翼/垂直起降/直升机）
	PilotCount          int              `json:"pilot_count"` // 需要飞手数量（0=不限）
	Latitude            float64          `json:"latitude"`
	Longitude           float64          `json:"longitude"`
	BudgetFen           int64            `json:"budget_fen"`         // 预算上限（分，0=面议）
	BudgetMinFen        int64            `json:"budget_min_fen"`     // 预算下限（分，0=不限）
	OfflineAmountFen    int64            `json:"offline_amount_fen"` // 线下成交金额（联系对接模式撮合价值度量）
	// Deadline 需求有效期截止日（YYYY-MM-DD，服务端校验非过去；空串=长期有效）。
	// 发布时效管理：过期需求前端不再展示可对接入口。
	Deadline  string         `json:"deadline,omitempty"`
	BizFields map[string]any `json:"biz_fields"`
	Status    DemandStatus   `json:"status"`
	IsMine    bool           `json:"is_mine,omitempty"` // 详情接口标记：当前请求者即发布者（前端据此禁用自登记）
	Version   int            `json:"version"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// DemandIntent records an intent to contact a demand publisher.
// It is the data basis of the "联系对接" (contact-deal) model: an interested
// party registers intent with contact info, the publisher sees the list,
// and deal outcome is tracked via status (pending → contacted / done / closed).
type DemandIntent struct {
	ID           string    `json:"id"`
	DemandID     string    `json:"demand_id"`
	IntentorID   string    `json:"intentor_id"`
	IntentorName string    `json:"intentor_name"`
	Contact      string    `json:"contact"`
	Remark       string    `json:"remark"`
	Status       string    `json:"status"` // pending / contacted / done / closed
	// DemandTitle 是跨需求聚合查询（ListByPublisher）带出的展示字段，
	// 让发布方的「接单申请」聚合页不必再逐条回查需求标题。
	// omitempty：单需求查询不填，其余接口的 JSON 契约保持不变。
	DemandTitle string `json:"demand_title,omitempty"`
	Version      int       `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// WorkOrderStatus represents the lifecycle of a work order (订单状态流转, PRD FR-6.3/6.4).
//
//	pending → ongoing → awaiting_accept → completed
//	   └─────────┴───────┴───────────────→ cancelled
type WorkOrderStatus string

const (
	WorkOrderPending        WorkOrderStatus = "pending"         // 待开始：企业确认接单后生成
	WorkOrderOngoing        WorkOrderStatus = "ongoing"         // 进行中：飞手确认开始
	WorkOrderAwaitingAccept WorkOrderStatus = "awaiting_accept" // 待确认完成：飞手确认完成
	WorkOrderCompleted      WorkOrderStatus = "completed"       // 已完成：企业验收通过
	WorkOrderCancelled      WorkOrderStatus = "cancelled"       // 已取消：任意一方发起
)

// WorkOrder is the order generated when a publisher confirms an intent (接单派单闭环).
type WorkOrder struct {
	ID            string          `json:"id"`
	OrderNo       string          `json:"order_no"`
	DemandID      string          `json:"demand_id"`
	DemandTitle   string          `json:"demand_title"` // 需求标题（服务层填充，便于列表/详情展示）
	IntentID      string          `json:"intent_id"`    // 来源意向（B 批：唯一约束防并发双建单）
	PublisherID   string          `json:"publisher_id"` // 需求方（企业）
	PublisherName string          `json:"publisher_name"`
	WorkerID      string          `json:"worker_id"` // 接单飞手
	WorkerName    string          `json:"worker_name"`
	AmountFen     int64           `json:"amount_fen"` // 订单金额（企业确认接单时填写，面议为 0）
	Status        WorkOrderStatus `json:"status"`
	ResultPhotos  []string        `json:"result_photos"` // 作业成果照片（飞手确认完成时上传）
	ReworkNote    string          `json:"rework_note"`   // 企业整改要求
	CancelReason  string          `json:"cancel_reason"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// ---- Training & Certification ----

// CertType represents the issuing authority for a drone operation certificate.
type CertType string

const (
	CertCAAC     CertType = "caac"      // CAAC civil aviation certificate
	CertUTCDJI   CertType = "utc_dji"   // DJI agricultural UTC certificate
	CertGovLevel CertType = "gov_level" // government vocational certificate
)

// Certificate is a drone operation qualification held by a user.
type Certificate struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	CertType   CertType  `json:"cert_type"`
	CertNumber string    `json:"cert_number"`
	Level      string    `json:"level"`
	IssueDate  time.Time `json:"issue_date"`
	ExpireDate time.Time `json:"expire_date"`
	IssuerOrg  string    `json:"issuer_org"`
	ImageURL   string    `json:"image_url"`
	Status     string    `json:"status"` // pending / approved / rejected
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CoursePrice is a price plan for a training course (enroll/register page).
type CoursePrice struct {
	Name  string `json:"name"`
	Price int    `json:"price"` // 元
}

// TrainingCourse is a training class offered by an organisation.
// 字段与小程序 pages/training/{courses,enroll,register}.vue 读取的名称对齐。
type TrainingCourse struct {
	ID            string        `json:"id"`
	OrgID         string        `json:"org_id"`
	OrgName       string        `json:"org_name"` // 机构名（页面 org_name || enterprise_name || name）
	Title         string        `json:"title"`
	CertType      CertType      `json:"cert_type"`
	Description   string        `json:"description"`
	StartDate     time.Time     `json:"start_date"`
	EndDate       time.Time     `json:"end_date"`
	MaxStudents   int           `json:"max_students"`
	EnrolledCount int           `json:"enrolled_count"`
	Location      string        `json:"location"`
	District      string        `json:"district"` // 区县（页面筛选 district || region）
	PriceFen      int64         `json:"price_fen"`
	Rating        string        `json:"rating"`
	ReviewCount   int           `json:"review_count"`
	DurationDays  int           `json:"duration_days"`
	Image         string        `json:"image"` // 封面（页面 image||cover_image||image_url）
	Tags          []string      `json:"tags"`
	Certificate   string        `json:"certificate"` // 证书/结业证书图（页面 certificate || certificate_url）
	Courses       []CoursePrice `json:"courses"`     // 课程方案 [{name,price}]
	Prices        []CoursePrice `json:"prices"`      // 价格方案 [{name,price}]
	BusinessHours string        `json:"business_hours"`
	Phone         string        `json:"phone"`        // 报名电话（页面 phone || contact_phone）
	Remain        int           `json:"remain"`       // 剩余名额（页面"仅剩N个"徽章）
	Environment   []string      `json:"environment"`  // 培训环境图集（页面 environment || env_images）
	CourseTypes   []string      `json:"course_types"` // 课程类型列表（页面 course_types）
	Status        string        `json:"status"`       // draft / published / recruiting / full / upcoming / urgent
	// PassRate 通过考试率（页面"通过考试"统计，如 "92"；空为"—"）
	PassRate string `json:"pass_rate"`
	// Years 机构年限（页面"机构年限"统计；0 为"—"）
	Years   int `json:"years"`
	Version int `json:"version"`

	// MyEnrollmentStatus 当前请求者的报名状态（""=未报名；enrolled/paid/approved/rejected/completed）。
	// 运行时按认证用户填充（详情页按钮态：已报名显示进度而非"立即报名"），不入库。
	MyEnrollmentStatus string    `json:"my_enrollment_status,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Instructor is a certified training instructor registered on the platform.
type Instructor struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Photo     string    `json:"photo"` // 教练照片
	CertTypes []string  `json:"cert_types"`
	Bio       string    `json:"bio"`
	OrgID     string    `json:"org_id"`
	Status    string    `json:"status"` // pending / approved
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ---- Jobs ----

// JobStatus represents the publishing lifecycle of a job posting.
type JobStatus string

const (
	JobDraft     JobStatus = "draft"
	JobPublished JobStatus = "published"
	JobClosed    JobStatus = "closed"
	JobArchived  JobStatus = "archived"
)

// Job is a recruitment posting by an enterprise.
type Job struct {
	ID           string    `json:"id"`
	EnterpriseID string    `json:"enterprise_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Location     string    `json:"location"`
	JobType      string    `json:"job_type"`
	SalaryFen    int64     `json:"salary_fen"`
	Status       JobStatus `json:"status"`
	Version      int       `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AppStatus represents the application review pipeline.
type AppStatus string

const (
	AppSubmitted    AppStatus = "submitted"
	AppViewed       AppStatus = "viewed"
	AppInterviewing AppStatus = "interviewing"
	AppOffered      AppStatus = "offered"
	AppRejected     AppStatus = "rejected"
	AppWithdrawn    AppStatus = "withdrawn"
)

// Resume is a job applicant's profile and work history.
type Resume struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Title          string    `json:"title"`
	Name           string    `json:"name"`            // 姓名
	Phone          string    `json:"phone"`           // 联系电话
	Email          string    `json:"email"`           // 邮箱
	Education      string    `json:"education"`       // 学历
	WorkExperience string    `json:"work_experience"` // 工作经历
	Skills         []string  `json:"skills"`          // 技能标签
	CertificateURL string    `json:"certificate_url"` // 证书图
	Content        string    `json:"content"`
	Visibility     string    `json:"visibility"` // private / public
	Version        int       `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// JobApplication records a user's application to a specific job.
type JobApplication struct {
	ID          string    `json:"id"`
	JobID       string    `json:"job_id"`
	ResumeID    string    `json:"resume_id"`
	ApplicantID string    `json:"applicant_id"`
	Status      AppStatus `json:"status"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FileRecord stores metadata about an uploaded file.
type FileRecord struct {
	ID          string    `json:"id"`
	StorageKey  string    `json:"storage_key"`
	SHA256      string    `json:"sha256"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	Visibility  string    `json:"visibility"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	// Width/Height 图片像素尺寸，上传时从文件头解出。0 表示未知（非图片，或迁移前的存量数据）。
	// 消费方：商品卡片要按每张图自己的比例预留空间——只知道 URL 无法预留，
	// 加载完再撑开会造成列表跳动（CLS），写死比例又会裁图。
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Message is an in-app notification sent between users or from the system.
type Message struct {
	ID           string    `json:"id"`
	SenderID     string    `json:"sender_id"`
	ReceiverID   string    `json:"receiver_id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	IsRead       bool      `json:"is_read"`
	CreatedAt    time.Time `json:"created_at"`
}

// Enrollment records a user's registration for a training course.
type Enrollment struct {
	ID            string    `json:"id"`
	CourseID      string    `json:"course_id"`
	UserID        string    `json:"user_id"`
	Name          string    `json:"name"`          // 报名人姓名
	Phone         string    `json:"phone"`         // 联系电话
	IDCard        string    `json:"id_card"`       // 身份证号
	Gender        string    `json:"gender"`        // 性别
	Birthday      time.Time `json:"birthday"`      // 生日（DATE）
	Email         string    `json:"email"`         // 邮箱
	Education     string    `json:"education"`     // 学历
	Experience    string    `json:"experience"`    // 从业经验
	PhotoURL      string    `json:"photo_url"`     // 证件照
	IDCardImage   string    `json:"id_card_image"` // 身份证正面照片
	IDCardBack    string    `json:"id_card_back"`  // 身份证反面照片（正反面 3 张材料闭环）
	NoCrime       string    `json:"no_crime"`      // 无犯罪证明
	Status        string    `json:"status"`
	ReviewNote    string    `json:"review_note"`     // 机构审核备注（通过可留空/拒绝需原因）
	PaidAmountFen int64     `json:"paid_amount_fen"` // 报名时冻结的学费（分），完成时按此金额释放——与课程实时价格解耦
	CreatedAt     time.Time `json:"created_at"`
}

// TradeOrder is a purchase order in the drone marketplace.
type TradeOrder struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	// ProductName 响应增强字段（不入库）：列表接口按 product_id 关联填充，
	// 供小程序订单展示商品名（避免前端仅显示"订单N"）。
	ProductName string `json:"product_name"`
	BuyerID     string `json:"buyer_id"`
	SellerID    string `json:"seller_id"`
	AmountFen   int64  `json:"amount_fen"`
	Status      string `json:"status"`
	// Reviewed 响应增强字段（**不入库**）：**当前请求者**是否已评价过本单。
	// 由 reviews 表推导（target_type='order' 且 target_id=本单 且 reviewer=本人 且
	// status<>'rejected'），**不新增冗余列** —— 同一个事实记两处迟早对不上
	//（2026-09-20 的 reg_count 漂移就是这一类）。
	//
	// 为什么必须有它：订单状态机的 completed 在小程序里被展示成「待评价」，
	// 而「已评价」此前**只写在手机本地存储**里（miniprogram/utils/orderAdapter.js 的
	// order_reviewed_prod），换设备或清缓存就回到「待评价」，服务端根本没有这个事实。
	Reviewed bool `json:"reviewed"`
	// 售后契约：aftersale_type=refund(仅退款)/return(退货退款)；
	// aftersale_status 为空串表示该订单从未申请过售后，其余取值：
	//   refund 单：pending(待审核) → approved(已退款) / rejected(已驳回)
	//   return 单：pending → returning(已同意退货，待买家寄回)
	//                     → returned(买家已寄回，待卖家确认收到)
	//                     → approved(确认收货后退款) / rejected(驳回)
	// 退款只在 approved 那一步发生：退货单在买家寄回、卖家确认之前不动钱。
	AftersaleType      string    `json:"aftersale_type"`
	AftersaleReason    string    `json:"aftersale_reason"`
	AftersaleDesc      string    `json:"aftersale_desc"`
	AftersaleAmountFen int64     `json:"aftersale_amount_fen"`
	AftersaleStatus    string    `json:"aftersale_status"`
	AftersaleTime      time.Time `json:"aftersale_time"`
	// 收货信息（下单时快照到订单）。此前订单表完全没有地址——发布表单里却有
	// 「物流发货」这个交付方式，卖家选它卖出去之后不知道寄给谁。
	ReceiverName    string `json:"receiver_name"`
	ReceiverPhone   string `json:"receiver_phone"`
	ReceiverRegion  string `json:"receiver_region"`  // 省市区
	ReceiverAddress string `json:"receiver_address"` // 详细地址
	// 发货信息（卖家发货时写入）。shipping_tracking 是**出库**单号，
	// 与 ReturnTracking（买家退货寄回的单号）方向相反，不要混用。
	ShippingCompany  string     `json:"shipping_company"`
	ShippingTracking string     `json:"shipping_tracking"`
	ShippedAt        *time.Time `json:"shipped_at,omitempty"`
	// AftersaleFrom 申请售后前的订单状态（paid/shipped/completed）；驳回时恢复原状态。
	AftersaleFrom string `json:"aftersale_from"`
	// 退货物流（一期）：仅 aftersale_type=return 且有值；买家在 returning 状态下提交。
	ReturnTracking string     `json:"return_tracking"`
	ReturnNote     string     `json:"return_note"`
	ReturnedAt     *time.Time `json:"returned_at,omitempty"`
	Version        int        `json:"version"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// EscrowAccount holds a user's balance and frozen funds in the escrow system.
type EscrowAccount struct {
	UserID     string    `json:"user_id"`
	BalanceFen int64     `json:"balance_fen"`
	FrozenFen  int64     `json:"frozen_fen"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// 托管金流水渠道（escrow_transactions.channel）：区分真实资金与平台内部记账，
// 对账/报表只有按渠道过滤才能答出「这个月真实收了多少、账上多出来的钱是哪来的」。
const (
	ChannelInternal = "internal" // 平台内部业务记账（默认值）
	// ChannelInternalAdmin 管理员手工入账：线下对公转账/现金收款后补记，无线上支付凭证。
	ChannelInternalAdmin = "internal_admin"
	// ChannelInternalSelf 用户自助充值（模拟通道：没有真实资金进平台，仅用于联调/演示）。
	ChannelInternalSelf = "internal_self"
	// ChannelWeChat 微信支付：真实资金，入账必须带微信支付单号（external_txn_id）。
	ChannelWeChat = "wechat"
)

// IsRealChannel 报告渠道是否为真实资金渠道。内部记账渠道（internal 与 internal_* 前缀）
// 返回 false——它们代表平台自己记出来的余额，不对应任何外部资金流。
func IsRealChannel(channel string) bool {
	return channel != "" && !strings.HasPrefix(channel, "internal")
}

// EscrowTransaction records a single escrow operation (deposit/freeze/release/refund).
type EscrowTransaction struct {
	ID            string `json:"id"`
	FromUser      string `json:"from_user"`
	ToUser        string `json:"to_user"`
	AmountFen     int64  `json:"amount_fen"`
	TxType        string `json:"tx_type"`
	ReferenceType string `json:"reference_type"`
	ReferenceID   string `json:"reference_id"`
	Status        string `json:"status"`
	// Channel 资金渠道（domain.Channel*）：internal* = 平台内部记账，wechat = 真实微信支付。
	Channel string `json:"channel"`
	// ExternalTxnID 外部支付单号（微信支付 transaction_id）。真实渠道入账的幂等键：
	// 同一 (channel, external_txn_id) 只能入账一次，微信回调重试不会重复加钱。
	ExternalTxnID string    `json:"external_txn_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// EscrowReconcile 托管金对账结果：某个渠道在 [From, To) 区间内的入金合计 + 全量流水明细。
// DepositFen 只统计 tx_type=deposit 的入金（真实资金对账关心的就是"钱进来多少"），
// Transactions 同时给出冻结/放款/退款等资金动作，便于逐笔核对。
type EscrowReconcile struct {
	Channel      string              `json:"channel"`
	RealFunds    bool                `json:"real_funds"`
	From         time.Time           `json:"from"`
	To           time.Time           `json:"to"`
	DepositCount int                 `json:"deposit_count"`
	DepositFen   int64               `json:"deposit_fen"`
	Transactions []EscrowTransaction `json:"transactions"`
}

// Article is a news article published by an admin on the platform.
type Article struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Summary   string    `json:"summary"`
	Category  string    `json:"category"`
	Source    string    `json:"source"`
	Author    string    `json:"author"`
	IsPinned  bool      `json:"is_pinned"`
	Status    string    `json:"status"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Review is a user rating and comment on a target (enterprise, demand, product, etc.).
type Review struct {
	ID         string    `json:"id"`
	ReviewerID string    `json:"reviewer_id"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Rating     int       `json:"rating"`
	Content    string    `json:"content"`
	Status     string    `json:"status"` // pending / approved / rejected
	CreatedAt  time.Time `json:"created_at"`
}

// Venue is a physical location available for booking (training ground, test field, etc.).
type Venue struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Name      string    `json:"name"`
	VenueType string    `json:"venue_type"`
	Location  string    `json:"location"`
	PriceFen  int64     `json:"price_fen"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// VenueBooking records a time-slot reservation for a venue.
type VenueBooking struct {
	ID        string    `json:"id"`
	VenueID   string    `json:"venue_id"`
	UserID    string    `json:"user_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Banner is a carousel image on the home page, configurable by admins.
type Banner struct {
	ID        string    `json:"id"`
	ImageURL  string    `json:"image_url"`
	LinkURL   string    `json:"link_url"`
	SortOrder int       `json:"sort_order"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// HomeQuickEntry is a shortcut icon on the home page grid.
type HomeQuickEntry struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	IconURL   string `json:"icon_url"`
	LinkURL   string `json:"link_url"`
	SortOrder int    `json:"sort_order"`
	Status    string `json:"status"`
}

// CertifiedPilot is a verified drone pilot with documented flight hours and ratings.
type CertifiedPilot struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	RealName      string    `json:"real_name"`
	IDCard        string    `json:"id_card"` // encrypted at rest
	Avatar        string    `json:"avatar"`  // 头像 URL（/uploads/...）
	Region        string    `json:"region"`  // 所在地区（如：重庆·渝北区）
	CertIDs       []string  `json:"cert_ids"`
	FlightHours   int       `json:"flight_hours"`
	Bio           string    `json:"bio"` // 擅长领域/简介
	Rating        float64   `json:"rating"`
	CompletedJobs int       `json:"completed_jobs"`
	Status        string    `json:"status"`        // pending / approved / rejected
	RejectReason  string    `json:"reject_reason"` // 驳回理由（管理端审核留痕）
	Version       int       `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CertifiedPilotDetail 名录/详情输出：cert_ids 扩展为证书对象数组（清单 4：关联证书详情）
type CertifiedPilotDetail struct {
	CertifiedPilot
	Certificates []CertificateBrief `json:"certificates"`
}

// PilotReviewDetail 管理端审核飞手申请所需的完整资料。
//
// 与 CertifiedPilotDetail（公开档案展示用）刻意分开：
//   - 公开档案只给"已通过且未过期"的证书摘要，身份证脱敏；
//   - 审核必须看到**全部**随附证书（含 pending/rejected），且要看得到证书编号与照片
//     才能核对真伪——CertificateBrief 既没有 cert_number 也没有 image_url。
type PilotReviewDetail struct {
	CertifiedPilot
	Certificates []Certificate `json:"certificates"`
}

// CertificateBrief 飞手展示用证书摘要（不含用户隐私字段）
type CertificateBrief struct {
	ID        string `json:"id"`
	CertType  string `json:"cert_type"`
	CertName  string `json:"cert_name"`
	IssuerOrg string `json:"issuer_org"`
	Level     string `json:"level"`
	Status    string `json:"status"`
}

// ReviewRecord is an immutable audit entry for an admin review action.
type ReviewRecord struct {
	ID           string    `json:"id"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	Action       string    `json:"action"`
	Reason       string    `json:"reason"`
	ReviewerID   string    `json:"reviewer_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// EnterpriseDocument is a supporting file (business license, ID card, etc.) attached to an enterprise registration.
type EnterpriseDocument struct {
	ID           string    `json:"id"`
	EnterpriseID string    `json:"enterprise_id"`
	FileID       string    `json:"file_id"`
	DocumentType string    `json:"document_type"`
	ReviewStatus string    `json:"review_status"`
	CreatedAt    time.Time `json:"created_at"`
}

// ---- Drone Trading ----

// ProductType categorises a product listing in the drone marketplace.
type ProductType string

const (
	ProductDrone       ProductType = "drone"       // complete drone unit
	ProductPart        ProductType = "part"        // spare part or accessory
	ProductRepair      ProductType = "repair"      // repair service
	ProductAerial      ProductType = "aerial"      // 航拍服务（需求②-2 供给能力展示）
	ProductTestFly     ProductType = "test_fly"    // 试飞测试（场地预约）
	ProductCalibration ProductType = "calibration" // 检测标定
	ProductAirspace    ProductType = "airspace"    // 空域协调
)

// 商品审核状态（drone_products.check_status），与上架状态 status 正交。
const (
	ProductCheckPending  = "pending"  // 待审核（用户刚发布）
	ProductCheckPassed   = "passed"   // 审核通过
	ProductCheckRejected = "rejected" // 审核驳回（必须带 check_reason）
)

// 交付方式（drone_products.delivery）。
const (
	DeliveryPickup     = "pickup"     // 自提
	DeliveryCity       = "city"       // 同城配送
	DeliveryLogistics  = "logistics"  // 物流发货
	DeliveryNegotiable = "negotiable" // 可协商
)

// 价格模式（drone_products.price_mode）。
const (
	PriceModeFixed      = "fixed"      // 明码标价：PriceFen 必须 > 0
	PriceModeNegotiable = "negotiable" // 面议：PriceFen 必须 = 0
)

// DroneProduct is a marketplace listing for a drone, part, or repair service.
type DroneProduct struct {
	ID          string      `json:"id"`
	SellerID    string      `json:"seller_id"`
	SellerName  string      `json:"seller_name"`
	ProdType    ProductType `json:"prod_type"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	PriceFen    int64       `json:"price_fen"`
	// Category/Region/Unit 三个服务类字段：从 service_listings 并入（migration 000110）。
	// Category 比 ProdType 更细——ProdType 只到"航拍服务"，而"测绘/巡检/应急"是更细的类目；
	// Unit 是报价单位（次/天/公里），没有它"¥800"不知道是每次还是每天。
	Category string `json:"category"`
	Region   string `json:"region"`
	Unit     string `json:"unit"`
	// Delivery 交付方式：pickup(自提) / city(同城配送) / logistics(物流发货) / negotiable(可协商)。
	// 空串表示卖家未选——下单时按商品类型兜底判断是否需要收货地址
	//（此前表单采集了这个字段但提交时丢弃，见 migration 000109）。
	Delivery string `json:"delivery"`
	// PriceMode 价格模式：fixed(明码标价) / negotiable(面议)。
	// 此前靠 price_fen==0 隐式表达面议，与"卖家填了 0 元"不可区分——见 migration 000107。
	PriceMode string `json:"price_mode"`
	Images      []string    `json:"images"`
	// DetailImages 详情图（详情区长图）。与 Images 分工不同：Images 是顶部图集（封面，
	// 买家第一眼看到的），DetailImages 是往下翻时铺在详情区的细节图（内部结构/铭牌/检测报告/实拍）。
	// 参考 Tigshop 的图文详情（descArr），先做轻量版：纯图数组，不含图文块混排。
	DetailImages []string `json:"detail_images"`
	// CoverWidth/CoverHeight 封面图（Images[0]）的像素宽高，**仅供响应**，不落商品表。
	// 来源是 uploads 台账（迁移 000115）。用途：供给大厅的卡片在渲染前按每张图自己的
	// 比例预留位置——只知道 URL 无法预留（加载完再撑开会跳动），写死比例又会裁图。
	// 0 表示未知（封面不是本站上传、或存量数据尚未回填），前端退化为 1:1。
	CoverWidth  int `json:"cover_width,omitempty"`
	CoverHeight int `json:"cover_height,omitempty"`
	// CoverURL 封面原图地址（与 Images[0] 同值）。前端请求裁剪尺寸时作为路径参数用。
	CoverURL string `json:"cover_url,omitempty"`
	Brand       string      `json:"brand"`
	Model       string      `json:"model"`
	Condition   string      `json:"condition"` // new / used
	Views       int         `json:"views"`     // detail view counter
	// Status 只表达**上架维度**：pending(未上架，刚发布) / listed(在售) / sold(已售) / removed(已下架)。
	// 审核维度见 CheckStatus——两者正交；此前混在一个字段里，"驳回"只能写成 removed，
	// 与"卖家主动下架"同值，卖家分不清自己是被驳回还是自己下的。
	Status string `json:"status"`
	// CheckStatus 审核维度：pending / passed / rejected。公开商城要求
	// CheckStatus=passed 且 Status=listed，缺一不可（service.ProductVisibleInHall）。
	CheckStatus string `json:"check_status"`
	// CheckReason 驳回原因（CheckStatus=rejected 时有意义，审核驳回必填）。
	CheckReason string `json:"check_reason"`
	// ReviewedAt/ReviewedBy 审核留痕。配套地，管理端审核动作会写 s.audit——
	// 此前商品路径完全没有审计，而订单路径有。
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy string     `json:"reviewed_by,omitempty"`
	Version    int        `json:"version"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	// DeletedAt 软删除时间（回收站）。非空表示已进回收站：公开列表/详情/我的商品
	// 一律不可见，订单仍能按 ID 取到商品名（ListByIDs 不过滤），避免"删商品丢订单上下文"。
	// 此前是物理 DELETE，订单里的 product_id 会变成指向不存在行的孤儿。
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ServiceListing is an enterprise service capability showcase (PRD ②-2 供给能力展示).
type ServiceListing struct {
	ID           string    `json:"id"`
	ProviderID   string    `json:"provider_id"`   // 企业用户 ID（管理端录入可为空）
	ProviderName string    `json:"provider_name"` // 企业名称（展示用）
	Title        string    `json:"title"`         // 服务标题
	Category     string    `json:"category"`      // 巡检/航拍/测绘/应急 等
	Description  string    `json:"description"`
	Region       string    `json:"region"`    // 服务区域
	PriceFen     int64     `json:"price_fen"` // 报价（分），0 为面议
	Unit         string    `json:"unit"`      // 单位：次/天/公里 等
	Image        string    `json:"image"`     // 封面图
	Status       string    `json:"status"`    // published / offline
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RepairOrder is a repair service request submitted by a customer.
type RepairOrder struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	ProductDesc string    `json:"product_desc"`
	FaultDesc   string    `json:"fault_desc"`
	QuoteFen    int64     `json:"quote_fen"`
	Status      string    `json:"status"` // submitted / quoted / repairing / completed
	Technician  string    `json:"technician"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ---- Community ----

// Post is a user-generated content item in the community forum.
// Posts are invisible until an admin publishes them.
type Post struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"author_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Images    []string  `json:"images"`
	CityCode  string    `json:"city_code"`
	Status    string    `json:"status"` // pending / published / removed
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Comment is a user reply on a post.
type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Report is a user-submitted complaint about abusive content.
type Report struct {
	ID           string    `json:"id"`
	ReporterID   string    `json:"reporter_id"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	Reason       string    `json:"reason"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// ---- Listings ----

// Listing is a second-hand product posted for sale.
type Listing struct {
	ID          string    `json:"id"`
	SellerID    string    `json:"seller_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	PriceFen    int64     `json:"price_fen"`
	Images      []string  `json:"images"`
	District    string    `json:"district"`
	Status      string    `json:"status"` // listed / sold / removed
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ---- Labour ----

// LabourOrder is a temporary labour dispatch request.
type LabourOrder struct {
	ID          string    `json:"id"`
	EmployerID  string    `json:"employer_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	WorkerCount int       `json:"worker_count"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	BudgetFen   int64     `json:"budget_fen"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LabourQuote is a price quotation submitted in response to a labour order.
type LabourQuote struct {
	ID         string    `json:"id"`
	OrderID    string    `json:"order_id"`
	QuoterID   string    `json:"quoter_id"`
	QuoterName string    `json:"quoter_name"`
	AmountFen  int64     `json:"amount_fen"`
	Proposal   string    `json:"proposal"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// Assignment assigns a specific worker to a labour order.
type Assignment struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	WorkerID  string    `json:"worker_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ---- Insurance & Finance ----

// InsurancePolicy is a drone insurance policy registered by a user.
type InsurancePolicy struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	DroneModel  string    `json:"drone_model"`
	DroneSN     string    `json:"drone_sn"`
	PolicyType  string    `json:"policy_type"` // liability / hull / third_party
	PremiumFen  int64     `json:"premium_fen"`
	CoverageFen int64     `json:"coverage_fen"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Insurer     string    `json:"insurer"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AnnualInspection records a mandatory yearly drone safety inspection.
type AnnualInspection struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	DroneModel  string    `json:"drone_model"`
	DroneSN     string    `json:"drone_sn"`
	InspectDate time.Time `json:"inspect_date"`
	ExpireDate  time.Time `json:"expire_date"`
	Result      string    `json:"result"`
	ReportURL   string    `json:"report_url"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LoanApplication is an installment loan request for drone purchases.
type LoanApplication struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	AmountFen     int64     `json:"amount_fen"`
	TermMonths    int       `json:"term_months"`
	Purpose       string    `json:"purpose"`
	Status        string    `json:"status"` // submitted / approved / rejected / active / paid_off
	ApprovedFen   int64     `json:"approved_fen"`
	MonthlyPayFen int64     `json:"monthly_pay_fen"`
	Version       int       `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
