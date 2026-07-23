package domain

import "time"

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
	ResType  string    `json:"res_type"`  // drone / equipment / team / vehicle
	Quantity int       `json:"quantity"`
	Status   string    `json:"status"`    // standby / engaged
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
	ID         string    `json:"id"`
	SiteID     string    `json:"site_id"`
	UserID     string    `json:"user_id"`
	Purpose    string    `json:"purpose"` // R&D / certification / demonstration
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     string    `json:"status"` // pending / approved / rejected / completed
	ReviewNote string    `json:"review_note"`
	CreatedAt  time.Time `json:"created_at"`
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
