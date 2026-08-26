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
	// ContactInfo 对接联系方式（track 页「联系发布方」复制门的数据源）
	ContactInfo string    `json:"contact_info"`
	Status      string    `json:"status"` // active / completed
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TransMilestone is a checkpoint in the transformation process.
type TransMilestone struct {
	Name       string    `json:"name"`
	Completed  bool      `json:"completed"`
	Date       time.Time `json:"date"`
	Evidence   string    `json:"evidence"` // URL or description
}

// ── 院校展示 (per .doc ⑤-2) ──

// CollegeMajor is a drone-related major shown on the miniapp college detail page.
type CollegeMajor struct {
	Name     string `json:"name"`
	Degree   string `json:"degree"`   // 本科/专科/硕士
	Duration int    `json:"duration"` // 学制（年）
	Key      string `json:"key"`      // 特色标签（国家级特色专业 等）
	Flagship bool   `json:"flagship"` // 王牌专业
}

// CollegePartner is a partner enterprise shown on the college detail page.
type CollegePartner struct {
	Icon string `json:"icon"`
	Name string `json:"name"`
	Type string `json:"type"` // 联合实验室/实习基地 等
}

// College showcases a school's drone-related programs and resources.
// 字段与小程序 pages/colleges/list.vue + detail.vue 读取的 snake_case 名称对齐。
type College struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Region       string   `json:"region"`
	City         string   `json:"city"`          // 所在城市（页面 city || region）
	Tags         []string `json:"tags"`          // 985/211/专科/高职 等（页面 collegeLevel/类型筛选依据）
	ShortName    string   `json:"short_name"`    // 简称
	LevelTags    string   `json:"level_tags"`    // 层次标签（页面展示）
	Majors       []string `json:"majors"`        // 无人机相关专业（字符串列表，管理端兼容）
	Specialties  []string `json:"specialties"`   // 特色专业（页面 specialties || majors || tags）
	Facilities   []string `json:"facilities"`    // 实训基地/实验室
	MajorCount   int      `json:"major_count"`   // 无人机专业数
	PartnerCount int      `json:"partner_count"` // 合作企业数
	TeacherCount int      `json:"teacher_count"` // 硕博导师数
	StudentCount int      `json:"student_count"` // 在读学生
	GraduateRate string   `json:"graduate_rate"` // 就业率
	Partners     []CollegePartner `json:"partners"` // 合作企业 [{icon,name,type}]
	LogoURL      string   `json:"logo_url"`
	CoverURL     string   `json:"cover"`         // 封面图（页面 cover||image||campus_image||cover_image）
	Photos       []string `json:"photos"`        // 校园环境图
	Phone        string   `json:"phone"`
	Website      string   `json:"website"`
	Intro        string   `json:"intro"`         // 院校介绍（页面 intro || description）
	MajorsDetail []CollegeMajor `json:"majors_detail"` // 专业对象数组（detail 页 major-item 渲染）
	Description  string   `json:"description"`
	Status       string   `json:"status"`    // active / inactive
	CoopType     string   `json:"coop_type"` // research(科研合作) / talent(人才培养) / both(综合) — 功能方案修订版 三·五 分域
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
