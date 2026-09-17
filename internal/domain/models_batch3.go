package domain

import "time"

// ── 救援案例库 (per .doc ⑦-3) ──

// RescueCase documents a real drone-assisted rescue operation.
type RescueCase struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	EventType  string    `json:"event_type"` // mountain_fire / flood / earthquake / search_rescue
	Location   string    `json:"location"`
	Date       time.Time `json:"date"`
	DroneModel string    `json:"drone_model"`
	TeamName   string    `json:"team_name"`
	Summary    string    `json:"summary"`    // 救援过程简述
	Result     string    `json:"result"`     // 救援成果
	Lessons    string    `json:"lessons"`    // 经验教训
	MediaURLs  []string  `json:"media_urls"` // 图片/视频
	Source     string    `json:"source"`     // 信息来源
	Status     string    `json:"status"`     // published / draft
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ── 应急管理部门对接 (per .doc ⑦-4) ──

// EmergencyDept represents a linked emergency management department.
type EmergencyDept struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	DeptType     string    `json:"dept_type"` // fire / police / civil_affairs / emergency_bureau
	Region       string    `json:"region"`
	ContactName  string    `json:"contact_name"`
	ContactPhone string    `json:"contact_phone"`
	ProtocolURL  string    `json:"protocol_url"` // 联动协议文件
	Status       string    `json:"status"`       // active / inactive
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// EmergencyDrill records a joint emergency drill with a department.
type EmergencyDrill struct {
	ID           string    `json:"id"`
	DeptID       string    `json:"dept_id"`
	Title        string    `json:"title"`
	Scenario     string    `json:"scenario"` // 演练场景
	Date         time.Time `json:"date"`
	Participants int       `json:"participants"`
	DroneCount   int       `json:"drone_count"`
	Result       string    `json:"result"` // 演练评估
	CreatedAt    time.Time `json:"created_at"`
}

// ── 协会会员分级：已移除 ──
//
// 这里曾有 AssociationRole（会长/副会长/秘书长/部门负责人/普通会员/副会长单位/
// 合作院校/访客 共 8 个）与 AssociationMember 实体，源自一份 .doc 需求。
// 实际情况：8 个角色里只有 partner 一个真正参与过判定（visitorResourceLevel），
// 会长与访客在权限上完全等价；association_members 表 0 行、前端零调用、
// 单条新增不校验角色、UpdateRole 没有接口。整块按"未启用"删除。
