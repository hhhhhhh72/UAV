package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- CompetitionService (竞赛管理) ----

type CompetitionService struct {
	repo repository.CompetitionRepository
}

func NewCompetitionService(repo repository.CompetitionRepository) *CompetitionService {
	return &CompetitionService{repo: repo}
}

// Create 接收完整领域对象（含小程序页面字段 fee/tags/poster/deadline/events 等）。
func (s *CompetitionService) Create(ctx context.Context, c domain.Competition) (domain.Competition, error) {
	now := time.Now()
	if c.ID == "" {
		c.ID = nextID("comp")
	}
	if c.Status == "" {
		c.Status = "published"
	}
	c.CreatedAt = now
	c.UpdatedAt = now
	return s.repo.Create(ctx, c)
}

func (s *CompetitionService) List(ctx context.Context, page, pageSize int) ([]domain.Competition, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

func (s *CompetitionService) Get(ctx context.Context, id string) (domain.Competition, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *CompetitionService) Update(ctx context.Context, c domain.Competition) (domain.Competition, error) {
	old, err := s.repo.FindByID(ctx, c.ID)
	if err != nil {
		return domain.Competition{}, err
	}
	c.CreatedAt = old.CreatedAt // 保留原创建时间
	c.UpdatedAt = time.Now()
	return s.repo.Update(ctx, c)
}

func (s *CompetitionService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Register 报名：name/phone/idCard 为参赛人实名信息，photoURL/idCardImage 为证件影像（C13 补字段）。
// 幂等：同一用户不可重复报名同一赛事；容量：已报名队数达到 max_teams 拒绝；deadline 过后拒绝。
func (s *CompetitionService) Register(ctx context.Context, competitionID, userID, teamName string, memberCount int, contactInfo, name, phone, idCard, photoURL, idCardImage string) (domain.CompetitionReg, error) {
	now := time.Now()
	// Check competition exists
	c, err := s.repo.FindByID(ctx, competitionID)
	if err != nil {
		return domain.CompetitionReg{}, err
	}
	regs, err := s.repo.ListRegs(ctx, competitionID)
	if err != nil {
		return domain.CompetitionReg{}, err
	}
	// 幂等：同一用户重复报名拒绝
	for _, r := range regs {
		if r.UserID == userID {
			return domain.CompetitionReg{}, fmt.Errorf("already registered")
		}
	}
	// 容量：max_teams > 0 时已报名队数达到上限拒绝
	if c.MaxTeams > 0 && len(regs) >= c.MaxTeams {
		return domain.CompetitionReg{}, fmt.Errorf("competition is full")
	}
	// 截止：deadline 已过拒绝
	if c.Deadline != nil && !c.Deadline.IsZero() && now.After(*c.Deadline) {
		return domain.CompetitionReg{}, fmt.Errorf("registration deadline passed")
	}
	cr := domain.CompetitionReg{
		ID:            fmt.Sprintf("compreg-%d-%d", now.UnixNano(), nextSeq()),
		CompetitionID: competitionID,
		UserID:        userID,
		TeamName:      teamName,
		MemberCount:   memberCount,
		ContactInfo:   contactInfo,
		Name:          name,
		Phone:         phone,
		IDCard:        idCard,
		PhotoURL:      photoURL,
		IDCardImage:   idCardImage,
		Status:        "registered",
		CreatedAt:     now,
	}
	return s.repo.CreateReg(ctx, cr)
}

func (s *CompetitionService) ListRegs(ctx context.Context, competitionID string) ([]domain.CompetitionReg, error) {
	return s.repo.ListRegs(ctx, competitionID)
}

// ---- EventService (协会活动) ----

type EventService struct {
	repo repository.EventRepository
}

func NewEventService(repo repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) Create(ctx context.Context, title, eventType, description, location, coverURL string, startTime, endTime time.Time, maxAttendees int) (domain.AssociationEvent, error) {
	if !endTime.IsZero() && endTime.Before(startTime) {
		return domain.AssociationEvent{}, errors.New("end time must not be earlier than start time")
	}
	now := time.Now()
	e := domain.AssociationEvent{
		ID:           nextID("event"),
		Title:        title,
		EventType:    eventType,
		Description:  description,
		Location:     location,
		CoverURL:     coverURL,
		StartTime:    startTime,
		EndTime:      endTime,
		MaxAttendees: maxAttendees,
		Status:       "published",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return s.repo.Create(ctx, e)
}

func (s *EventService) List(ctx context.Context, page, pageSize int) ([]domain.AssociationEvent, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

func (s *EventService) Get(ctx context.Context, id string) (domain.AssociationEvent, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *EventService) Update(ctx context.Context, id, title, eventType, description, location, coverURL, status string, startTime, endTime time.Time, maxAttendees int) (domain.AssociationEvent, error) {
	if !endTime.IsZero() && endTime.Before(startTime) {
		return domain.AssociationEvent{}, errors.New("end time must not be earlier than start time")
	}
	ev, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.AssociationEvent{}, err
	}
	ev.Title = title
	ev.EventType = eventType
	ev.Description = description
	ev.Location = location
	ev.CoverURL = coverURL
	ev.Status = status
	ev.StartTime = startTime
	ev.EndTime = endTime
	ev.MaxAttendees = maxAttendees
	ev.UpdatedAt = time.Now()
	return s.repo.Update(ctx, ev)
}

func (s *EventService) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }

func (s *EventService) Register(ctx context.Context, eventID, userID, name, phone, org string) (domain.EventRegistration, error) {
	now := time.Now()
	ev, err := s.repo.FindByID(ctx, eventID)
	if err != nil {
		return domain.EventRegistration{}, err
	}
	regs, err := s.repo.ListRegs(ctx, eventID)
	if err != nil {
		return domain.EventRegistration{}, err
	}
	// 幂等：同一用户重复报名拒绝
	for _, r := range regs {
		if r.UserID == userID {
			return domain.EventRegistration{}, fmt.Errorf("already registered")
		}
	}
	// 容量：max_attendees > 0 时已报名人数达到上限拒绝
	if ev.MaxAttendees > 0 && len(regs) >= ev.MaxAttendees {
		return domain.EventRegistration{}, fmt.Errorf("event is full")
	}
	er := domain.EventRegistration{
		ID:        nextID("evtreg"),
		EventID:   eventID,
		UserID:    userID,
		Name:      name,
		Phone:     phone,
		Org:       org,
		Status:    "registered",
		CreatedAt: now,
	}
	return s.repo.CreateReg(ctx, er)
}

func (s *EventService) ListRegs(ctx context.Context, eventID string) ([]domain.EventRegistration, error) {
	return s.repo.ListRegs(ctx, eventID)
}

// ---- ResourceService (产业资源共享) ----

// ErrResourceNotFound 预约/查询的资源不存在（Handler 映射为 404）。
var ErrResourceNotFound = errors.New("resource not found")

type ResourceService struct {
	repo repository.ResourceRepository
}

func NewResourceService(repo repository.ResourceRepository) *ResourceService {
	return &ResourceService{repo: repo}
}

func (s *ResourceService) Create(ctx context.Context, ownerID, name, resType, model, specs, location, bookingInfo string, priceFen int64, visibilityLevel string) (domain.IndustryResource, error) {
	now := time.Now()
	if visibilityLevel == "" {
		visibilityLevel = "public"
	}
	r := domain.IndustryResource{
		ID:              fmt.Sprintf("res-%d-%d", now.UnixNano(), nextSeq()),
		OwnerID:         ownerID,
		Name:            name,
		ResType:         resType,
		Model:           model,
		Specs:           specs,
		Location:        location,
		PriceFen:        priceFen,
		BookingInfo:     bookingInfo,
		VisibilityLevel: visibilityLevel,
		Status:          "available",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	return s.repo.Create(ctx, r)
}

func (s *ResourceService) List(ctx context.Context, resType string, page, pageSize int) ([]domain.IndustryResource, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, resType, offset, pageSize)
}

func (s *ResourceService) Get(ctx context.Context, id string) (domain.IndustryResource, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ResourceService) Update(ctx context.Context, id, name, resType, model, specs, location, bookingInfo string, priceFen int64, visibilityLevel, status string) (domain.IndustryResource, error) {
	r, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.IndustryResource{}, err
	}
	r.Name = name
	r.ResType = resType
	r.Model = model
	r.Specs = specs
	r.Location = location
	r.PriceFen = priceFen
	r.BookingInfo = bookingInfo
	if visibilityLevel != "" {
		r.VisibilityLevel = visibilityLevel
	}
	if status != "" {
		r.Status = status
	}
	r.UpdatedAt = time.Now()
	return s.repo.Update(ctx, r)
}

func (s *ResourceService) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }

// Book 提交资源预约（C11：小程序资源详情页 → POST /api/v1/industry-resources/{id}/book）。
// date 为 YYYY-MM-DD（小程序日期选择器格式，格式校验在 Handler 层）。
func (s *ResourceService) Book(ctx context.Context, userID, resourceID, date, purpose, contactName, contactPhone string) (domain.IndustryResourceBooking, error) {
	if _, err := s.repo.FindByID(ctx, resourceID); err != nil {
		return domain.IndustryResourceBooking{}, fmt.Errorf("%w: %s", ErrResourceNotFound, resourceID)
	}
	now := time.Now()
	b := domain.IndustryResourceBooking{
		ID:           fmt.Sprintf("resbook-%d-%d", now.UnixNano(), nextSeq()),
		ResourceID:   resourceID,
		UserID:       userID,
		BookingDate:  date,
		Purpose:      purpose,
		ContactName:  contactName,
		ContactPhone: contactPhone,
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return s.repo.CreateBooking(ctx, b)
}

// ListBookingsByResource 某资源的全部预约（供测试与管理端查询）。
func (s *ResourceService) ListBookingsByResource(ctx context.Context, resourceID string) ([]domain.IndustryResourceBooking, error) {
	return s.repo.ListBookingsByResource(ctx, resourceID)
}

// ---- EmergencyService (应急管理) ----

type EmergencyService struct {
	repo repository.EmergencyRepository
}

func NewEmergencyService(repo repository.EmergencyRepository) *EmergencyService {
	return &EmergencyService{repo: repo}
}

// Emergency Resources

// emergencyResTypeAliases: 应急资源类型中文 → 英文规范值。
// 小程序与 domain 约定英文键（drone/comm/vehicle/medical），管理端可能传中文。
var emergencyResTypeAliases = map[string]string{
	"无人机": "drone",
	"通讯":  "comm",
	"通信":  "comm",
	"车辆":  "vehicle",
	"医疗":  "medical",
}

// normalizeEmergencyResType 将中文别名归一到英文规范值；未知值原样返回。
func normalizeEmergencyResType(s string) string {
	if v, ok := emergencyResTypeAliases[strings.TrimSpace(s)]; ok {
		return v
	}
	return strings.TrimSpace(s)
}

func (s *EmergencyService) CreateResource(ctx context.Context, ownerID, name, resType, specs, location, contactInfo string, quantity int) (domain.EmergencyResource, error) {
	now := time.Now()
	r := domain.EmergencyResource{
		ID:          fmt.Sprintf("emres-%d-%d", now.UnixNano(), nextSeq()),
		OwnerID:     ownerID,
		Name:        name,
		ResType:     normalizeEmergencyResType(resType),
		Specs:       specs,
		Quantity:    quantity,
		Location:    location,
		ContactInfo: contactInfo,
		Status:      "available",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.CreateResource(ctx, r)
}

func (s *EmergencyService) ListResources(ctx context.Context, resType, q string, page, pageSize int) ([]domain.EmergencyResource, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListResources(ctx, normalizeEmergencyResType(resType), strings.TrimSpace(q), offset, pageSize)
}

func (s *EmergencyService) GetResource(ctx context.Context, id string) (domain.EmergencyResource, error) {
	return s.repo.FindResourceByID(ctx, id)
}

func (s *EmergencyService) FindDispatchByID(ctx context.Context, id string) (domain.EmergencyDispatch, error) {
	return s.repo.FindDispatchByID(ctx, id)
}

// Emergency Dispatches

// CreateDispatch 创建调度记录；status 可指定（dispatched/ongoing/completed），
// 空值或非法值回退默认 "dispatched"（补录历史调度时可传实际状态）。
func (s *EmergencyService) CreateDispatch(ctx context.Context, resourceID, eventDesc, location, commander, result, status string, startTime, endTime time.Time) (domain.EmergencyDispatch, error) {
	if !endTime.IsZero() && endTime.Before(startTime) {
		return domain.EmergencyDispatch{}, errors.New("end time must not be earlier than start time")
	}
	if status != "ongoing" && status != "completed" {
		status = "dispatched"
	}
	now := time.Now()
	if _, err := s.repo.FindResourceByID(ctx, resourceID); err != nil {
		return domain.EmergencyDispatch{}, err
	}
	d := domain.EmergencyDispatch{
		ID:         nextID("emdisp"),
		ResourceID: resourceID,
		EventDesc:  eventDesc,
		Location:   location,
		StartTime:  startTime,
		EndTime:    endTime,
		Commander:  commander,
		Result:     result,
		Status:     status,
		CreatedAt:  now,
	}
	return s.repo.CreateDispatch(ctx, d)
}

func (s *EmergencyService) ListDispatches(ctx context.Context, resourceID string, page, pageSize int) ([]domain.EmergencyDispatch, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListDispatches(ctx, resourceID, offset, pageSize)
}

func (s *EmergencyService) UpdateResource(ctx context.Context, id, name, resType, specs, location, contactInfo, status string, quantity int) (domain.EmergencyResource, error) {
	r, err := s.repo.FindResourceByID(ctx, id)
	if err != nil {
		return domain.EmergencyResource{}, err
	}
	r.Name = name
	r.ResType = normalizeEmergencyResType(resType)
	r.Specs = specs
	r.Location = location
	r.ContactInfo = contactInfo
	r.Status = status
	r.Quantity = quantity
	r.UpdatedAt = time.Now()
	return s.repo.UpdateResource(ctx, r)
}

func (s *EmergencyService) DeleteResource(ctx context.Context, id string) error {
	return s.repo.DeleteResource(ctx, id)
}
func (s *EmergencyService) DeleteDispatch(ctx context.Context, id string) error {
	return s.repo.DeleteDispatch(ctx, id)
}

func (s *EmergencyService) UpdateDispatch(ctx context.Context, id, resourceID, eventDesc, location, commander, result, status string, startTime, endTime time.Time) (domain.EmergencyDispatch, error) {
	if !endTime.IsZero() && endTime.Before(startTime) {
		return domain.EmergencyDispatch{}, errors.New("end time must not be earlier than start time")
	}
	d, err := s.repo.FindDispatchByID(ctx, id)
	if err != nil {
		return domain.EmergencyDispatch{}, err
	}
	d.ResourceID = resourceID
	d.EventDesc = eventDesc
	d.Location = location
	d.Commander = commander
	d.Result = result
	d.Status = status
	d.StartTime = startTime
	d.EndTime = endTime
	return s.repo.UpdateDispatch(ctx, d)
}
