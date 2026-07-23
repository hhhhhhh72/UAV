package domain

import "time"

// ── 成果转化追踪 (per .doc ③-5) ──

// TransformationStage represents a stage in lab→pilot→industrialized pipeline.
type TransformationStage string

const (
	StageLab           TransformationStage = "lab"
	StagePilot         TransformationStage = "pilot"
	StageIndustrialized TransformationStage = "industrialized"
	StageListed        TransformationStage = "listed"
)

// Transformation tracks an achievement through commercialization stages.
type Transformation struct {
	ID            string              `json:"id"`
	AchievementID string              `json:"achievement_id"`
	OwnerID       string              `json:"owner_id"`
	Title         string              `json:"title"`
	Stage         TransformationStage `json:"stage"`
	Progress      string              `json:"progress"`   // 当前进度描述
	Milestones    []TransMilestone    `json:"milestones"`
	PartnerID     string              `json:"partner_id"` // 合作产业化伙伴
	Status        string              `json:"status"`     // active / completed
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// TransMilestone is a checkpoint in the transformation process.
type TransMilestone struct {
	Name       string    `json:"name"`
	Completed  bool      `json:"completed"`
	Date       time.Time `json:"date"`
	Evidence   string    `json:"evidence"` // URL or description
}

// ── 院校展示 (per .doc ⑤-2) ──

// College showcases a school's drone-related programs and resources.
type College struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Region       string   `json:"region"`
	Majors       []string `json:"majors"`       // 无人机相关专业
	Facilities   []string `json:"facilities"`   // 实训基地/实验室
	StudentCount int      `json:"student_count"`
	GraduateRate string   `json:"graduate_rate"` // 就业率
	Partners     []string `json:"partners"`      // 合作企业
	LogoURL      string   `json:"logo_url"`
	Description  string   `json:"description"`
	Status       string   `json:"status"` // active / inactive
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ── 校企共建 (per .doc ⑤-5) ──

// CooperationProgram represents a school-enterprise cooperation.
type CooperationProgram struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	CollegeID   string    `json:"college_id"`
	EnterpriseID string   `json:"enterprise_id"`
	CoopType    string    `json:"coop_type"` // directed_training / internship_base / joint_lab / curriculum
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	StudentQuota int      `json:"student_quota"` // 定向培养名额
	Status      string    `json:"status"` // proposed / active / completed / cancelled
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
