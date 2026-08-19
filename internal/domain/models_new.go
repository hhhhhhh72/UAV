package domain

import "time"

// ---- New Business Modules (per 重庆市无人机产业协会 requirements) ----

// Expert is a think-tank expert in the drone / low-altitude economy domain.
type Expert struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Title     string    `json:"title"`
	Org       string    `json:"org"`
	Field     string    `json:"field"`
	Tags      []string  `json:"tags"`
	Bio       string    `json:"bio"`
	AvatarURL string    `json:"avatar_url"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Application is a service application submitted from the miniprogram
// (formerly JSON-file backed h5Application; now persisted in service_applications).
type Application struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	ServiceID   string         `json:"service_id"`
	ServiceName string         `json:"service_name"`
	OrderNo     string         `json:"order_no"`
	Status      string         `json:"status"`
	ApplyTime   string         `json:"apply_time"`
	FormData    map[string]any `json:"form_data"`
	CreatedAt   time.Time      `json:"created_at"`
}

// CaseEntry is a successful project case or industry best-practice.
type CaseEntry struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Images      []string  `json:"images"`
	ClientName  string    `json:"client_name"`
	Result      string    `json:"result"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ComplianceDoc is a regulatory guidance document.
type ComplianceDoc struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	Publisher   string    `json:"publisher"`
	PublishDate time.Time `json:"publish_date"`
	Summary     string    `json:"summary"`
	FileURL     string    `json:"file_url"`
	Tags        []string  `json:"tags"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StandardDoc is an industry group standard document.
type StandardDoc struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Category      string    `json:"category"` // 国家标准/行业标准/团体标准/企业标准（C11 补列）
	StandardNo    string    `json:"standard_no"`
	Publisher     string    `json:"publisher"`
	EffectiveDate time.Time `json:"effective_date"`
	Status        string    `json:"status"`
	Scope         string    `json:"scope"`
	Summary       string    `json:"summary"`
	FileURL       string    `json:"file_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ProjectApplication is a government/association project subsidy application.
type ProjectApplication struct {
	ID          string    `json:"id"`
	ApplicantID string    `json:"applicant_id"`
	ProjectName string    `json:"project_name"`
	Category    string    `json:"category"`
	BudgetFen   int64     `json:"budget_fen"`
	Description string    `json:"description"`
	Attachments []string  `json:"attachments"`
	Status      string    `json:"status"`
	ReviewNote  string    `json:"review_note"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Achievement is a technology achievement / patent / innovation.
type Achievement struct {
	ID          string       `json:"id"`
	OwnerID     string       `json:"owner_id"`
	Title       string       `json:"title"`
	AchieveType string       `json:"achieve_type"`
	Description string       `json:"description"`
	Field       string       `json:"field"`
	Stage       string       `json:"stage"`
	Images      []string     `json:"images"`
	Attachments []Attachment `json:"attachments"`
	ContactInfo string       `json:"contact_info"`
	Status      string       `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Attachment 成果附件（detail_pd.html 原型：名称/大小/URL）
type Attachment struct {
	Name string `json:"name"`
	Size string `json:"size"`
	URL  string `json:"url"`
}

// RDChallenge is an enterprise R&D challenge posted for collaboration.
type RDChallenge struct {
	ID          string    `json:"id"`
	PosterID    string    `json:"poster_id"`
	Title       string    `json:"title"`
	Field       string    `json:"field"`
	Description string    `json:"description"`
	BudgetFen   int64     `json:"budget_fen"`
	Deadline    time.Time `json:"deadline"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ResearchProject is a joint research project between industry and academia.
type ResearchProject struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Field       string    `json:"field"`
	Description string    `json:"description"`
	LeadOrg     string    `json:"lead_org"`
	Members     []string  `json:"members"`
	BudgetFen   int64     `json:"budget_fen"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Milestones  string    `json:"milestones"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CompetitionEvent is an event category inside a competition (register page selection).
type CompetitionEvent struct {
	Name   string `json:"name"`
	Type   string `json:"type"`   // 个人赛/团体赛
	Format string `json:"format"` // 计时排名/3人一队 等
	Fee    int    `json:"fee"`    // 元
}

// CompetitionRequirement is an entry requirement on the detail page.
type CompetitionRequirement struct {
	Icon  string `json:"icon"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Level string `json:"level"` // 初级/中级/高级
}

// CompetitionPrize is a prize tier on the detail page.
type CompetitionPrize struct {
	Level  string `json:"level"` // 冠军/亚军/季军
	Amount int    `json:"amount"`
	Metal  string `json:"metal"` // gold/silver/bronze
	Medal  string `json:"medal"`
}

// Competition is a drone competition / contest.
// 字段与小程序 pages/competitions/{list,detail,register}.vue 读取的名称对齐。
type Competition struct {
	ID                 string                   `json:"id"`
	Title              string                   `json:"title"`
	Category           string                   `json:"category"`
	Description        string                   `json:"description"`
	Location           string                   `json:"location"`
	StartDate          time.Time                `json:"start_date"`
	EndDate            time.Time                `json:"end_date"`
	Deadline           *time.Time               `json:"deadline"` // 报名截止（页面 deadline || enroll_deadline）
	MaxTeams           int                      `json:"max_teams"`
	RegCount           int                      `json:"reg_count"`
	Sponsor            string                   `json:"sponsor"`       // 主办方（页面 organizer || sponsor）
	OrganizerSub       string                   `json:"organizer_sub"` // 承办方
	Fee                int                      `json:"fee"`           // 报名费（元）
	MinFee             int                      `json:"min_fee"`       // 页面「报名费 ¥xx 起/人」（元）
	OriginalFee        int                      `json:"original_fee"`  // 划线原价（元，> Fee 时前端显示划线）
	Tags               []string                 `json:"tags"`
	Poster             string                   `json:"poster"` // 海报图（页面 poster||cover||image||banner）
	Requirements       []CompetitionRequirement `json:"requirements"`
	Events             []CompetitionEvent       `json:"events"`
	Prizes             []CompetitionPrize       `json:"prizes"`
	RegistrationStatus string                   `json:"registration_status"` // 页面报名状态（open/enrolling 等）
	Status             string                   `json:"status"`              // published / enrolling / ongoing / closed / full
	CreatedAt          time.Time                `json:"created_at"`
	UpdatedAt          time.Time                `json:"updated_at"`
}

// CompetitionReg is a team/individual registration for a competition.
type CompetitionReg struct {
	ID            string    `json:"id"`
	CompetitionID string    `json:"competition_id"`
	UserID        string    `json:"user_id"`
	TeamName      string    `json:"team_name"`
	MemberCount   int       `json:"member_count"`
	ContactInfo   string    `json:"contact_info"`
	Name          string    `json:"name"`          // 参赛人姓名（C13 补列）
	Phone         string    `json:"phone"`         // 参赛人手机号（C13 补列）
	IDCard        string    `json:"id_card"`       // 身份证号（C13 补列）
	PhotoURL      string    `json:"photo_url"`     // 白底免冠证件照（C13 补列）
	IDCardImage   string    `json:"id_card_image"` // 身份证正面影像（C13 补列）
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// AssociationEvent is an event organized by the drone association.
type AssociationEvent struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	EventType    string    `json:"event_type"`
	Description  string    `json:"description"`
	Location     string    `json:"location"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	MaxAttendees int       `json:"max_attendees"`
	RegCount     int       `json:"reg_count"`
	CoverURL     string    `json:"cover_url"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// EventRegistration is a user's registration for an association event.
type EventRegistration struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Org       string    `json:"org"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// MemberPortfolio is a member enterprise's brand showcase page.
type MemberPortfolio struct {
	ID           string    `json:"id"`
	EnterpriseID string    `json:"enterprise_id"`
	Name         string    `json:"name"`
	LogoURL      string    `json:"logo_url"`
	CoverURL     string    `json:"cover_url"`
	Description  string    `json:"description"`
	Products     []string  `json:"products"`
	Honors       []string  `json:"honors"`
	ContactInfo  string    `json:"contact_info"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// IndustryReport is a periodic industry analysis report.
type IndustryReport struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Period    string    `json:"period"`
	Category  string    `json:"category"`
	Summary   string    `json:"summary"`
	Content   string    `json:"content"`
	FileURL   string    `json:"file_url"`
	Author    string    `json:"author"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IndustryResource is a physical resource (drone, airfield, test site, etc.).
type IndustryResource struct {
	ID          string `json:"id"`
	OwnerID     string `json:"owner_id"`
	Name        string `json:"name"`
	ResType     string `json:"res_type"`
	Model       string `json:"model"`
	Specs       string `json:"specs"`
	Location    string `json:"location"`
	PriceFen    int64  `json:"price_fen"`
	BookingInfo string `json:"booking_info"`
	// VisibilityLevel: public(政府访客) < member(会员+) < partner(副会长单位+) < admin(仅协会管理员)
	VisibilityLevel string    `json:"visibility_level"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// IndustryResourceBooking is a booking request for an industry resource (C11).
type IndustryResourceBooking struct {
	ID           string    `json:"id"`
	ResourceID   string    `json:"resource_id"`
	UserID       string    `json:"user_id"`
	BookingDate  string    `json:"date"` // YYYY-MM-DD（小程序日期选择器格式）
	Purpose      string    `json:"purpose"`
	ContactName  string    `json:"contact_name"`
	ContactPhone string    `json:"contact_phone"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// EmergencyResource is a drone-related emergency response resource.
type EmergencyResource struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Name        string    `json:"name"`
	ResType     string    `json:"res_type"`
	Specs       string    `json:"specs"`
	Quantity    int       `json:"quantity"`
	Location    string    `json:"location"`
	ContactInfo string    `json:"contact_info"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EmergencyResourceBrief 调度关联资源摘要（列表展示用，避免全量字段）
type EmergencyResourceBrief struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	ResType string `json:"res_type"`
	Status string `json:"status"`
}

// EmergencyDispatch records an emergency resource dispatch event.
type EmergencyDispatch struct {
	ID         string                  `json:"id"`
	ResourceID string                  `json:"resource_id"`
	EventDesc  string                  `json:"event_desc"`
	Location   string                  `json:"location"`
	StartTime  time.Time               `json:"start_time"`
	EndTime    time.Time               `json:"end_time,omitempty"`
	Commander  string                  `json:"commander"`
	Result     string                  `json:"result"`
	Status     string                  `json:"status"`
	Related    *EmergencyResourceBrief `json:"related,omitempty"` // 关联资源摘要（列表查询内嵌）
	CreatedAt  time.Time               `json:"created_at"`
}
