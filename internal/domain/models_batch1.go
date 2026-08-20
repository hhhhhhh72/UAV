package domain

import (
	"fmt"
	"time"
)

// ── 产业资源池 (per .doc ②-5) ──

// ResourcePool is a categorized collection of industry resources.
type ResourcePool struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PoolType    string    `json:"pool_type"` // emergency / event / team / equipment
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	Resources   []string  `json:"resources"` // resource IDs in this pool
	Status      string    `json:"status"`    // active / inactive
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ResourcePoolMember links a resource to a pool with metadata.
type ResourcePoolMember struct {
	ID       string    `json:"id"`
	PoolID   string    `json:"pool_id"`
	ResID    string    `json:"res_id"`
	ResType  string    `json:"res_type"` // drone / equipment / team / vehicle
	Quantity int       `json:"quantity"`
	Status   string    `json:"status"` // standby / engaged
	JoinedAt time.Time `json:"joined_at"`
}

// ── 测试环境预约 (per .doc ③-3) ──

// TestSite is a specialized testing facility (flying field, lab, etc.).
type TestSite struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	SiteType    string    `json:"site_type"` // flying_field / lab / anechoic_chamber / wind_tunnel
	OwnerID     string    `json:"owner_id"`
	Location    string    `json:"location"`
	Facilities  []string  `json:"facilities"` // 5G / RTK / radar / spectrum_analyzer
	PriceFen    int64     `json:"price_fen"`
	BookingRule string    `json:"booking_rule"` // "工作日9-18点,需提前3天"
	Status      string    `json:"status"`       // available / maintenance / reserved
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TestSiteBooking is a reservation for a test site time slot.
type TestSiteBooking struct {
	ID           string    `json:"id"`
	SiteID       string    `json:"site_id"`
	UserID       string    `json:"user_id"`
	Purpose      string    `json:"purpose"` // R&D / certification / demonstration
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	ContactName  string    `json:"contact_name"`  // 预约联系人
	ContactPhone string    `json:"contact_phone"` // 联系电话
	Status       string    `json:"status"`        // pending / approved / rejected / completed
	ReviewNote   string    `json:"review_note"`
	CreatedAt    time.Time `json:"created_at"`
}

// ── 产业展会管理 (per .doc ⑥-3) ──

// Exhibition is a tradeshow/exhibition event.
type Exhibition struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Category    string    `json:"category"` // drone_show / equipment_expo / innovation_week
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	BoothCount  int       `json:"booth_count"`
	BoothPrice  int64     `json:"booth_price_fen"`
	Organizer   string    `json:"organizer"`
	CoverURL    string    `json:"cover_url"`
	Status      string    `json:"status"` // draft / recruiting / underway / ended
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ExhibitionBooth is a booth booking at an exhibition.
type ExhibitionBooth struct {
	ID           string    `json:"id"`
	ExhibitionID string    `json:"exhibition_id"`
	ExhibitorID  string    `json:"exhibitor_id"` // enterprise ID
	BoothNumber  string    `json:"booth_number"`
	ExhibitName  string    `json:"exhibit_name"` // what's being shown
	ExhibitDesc  string    `json:"exhibit_desc"`
	Status       string    `json:"status"` // applied / approved / paid / rejected
	CreatedAt    time.Time `json:"created_at"`
}

// ParseTime parses an ISO8601 time string, falling back to current time.
func ParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Now()
	}
	return t
}

// ParseTimeStrict 严格解析时间字符串，解析失败返回错误（不静默回退当前时间）。
// 支持 RFC3339、"2006-01-02 15:04"、"2006-01-02 15:04:05"、"2006-01-02" 四种格式；
// 空串视为"未设置"，返回零值时间（前端按零值渲染"未设置"）。
// 用户可触达的 handler 入口用其校验非法日期，避免非法日期被静默写成当前时间落库（P2）。
func ParseTimeStrict(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04", "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format %q", s)
}

// Shop represents a marketplace shop/store.
type Shop struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	OwnerID      string    `json:"owner_id"`
	Description  string    `json:"description"`
	LogoURL      string    `json:"logo_url"`
	LicenseURL   string    `json:"license_url"`
	AccountName  string    `json:"account_name"`
	ContactPhone string    `json:"contact_phone"`
	Address      string    `json:"address"`
	IsMember     bool      `json:"is_member"`
	Version      int       `json:"version"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// StudyTour is a study tour / research trip event.
type StudyTour struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Destination string    `json:"destination"`
	OrganizerID string    `json:"organizer_id"`
	CoverImage  string    `json:"cover_image"` // 封面图 URL（/uploads/...）
	PriceFen    int64     `json:"price_fen"`   // 价格（分），0 表示免费/面议
	Schedule    []StudySchedule `json:"schedule"` // 行程安排（JSONB）
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Duration    string    `json:"duration"`
	Capacity    int       `json:"capacity"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StudySchedule 研学行程中的一天安排
type StudySchedule struct {
	Day   int      `json:"day"`
	Title string   `json:"title"`
	Items []string `json:"items"`
}
