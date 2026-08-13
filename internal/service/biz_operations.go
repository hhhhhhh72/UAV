package service

import (
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
func (s *CompetitionService) Create(c domain.Competition) (domain.Competition, error) {
	now := time.Now()
	if c.ID == "" {
		c.ID = fmt.Sprintf("comp-%d", now.UnixNano())
	}
	if c.Status == "" {
		c.Status = "published"
	}
	c.CreatedAt = now
	c.UpdatedAt = now
	return s.repo.Create(c)
}

func (s *CompetitionService) List(page, pageSize int) ([]domain.Competition, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(offset, pageSize)
}

func (s *CompetitionService) Get(id string) (domain.Competition, error) {
	return s.repo.FindByID(id)
}

func (s *CompetitionService) Update(c domain.Competition) (domain.Competition, error) {
	old, err := s.repo.FindByID(c.ID)
	if err != nil {
		return domain.Competition{}, err
	}
	c.CreatedAt = old.CreatedAt // 保留原创建时间
	c.UpdatedAt = time.Now()
	return s.repo.Update(c)
}

func (s *CompetitionService) Delete(id string) error {
	return s.repo.Delete(id)
}

// Register 报名：name/phone/idCard 为参赛人实名信息，photoURL/idCardImage 为证件影像（C13 补字段）。
func (s *CompetitionService) Register(competitionID, userID, teamName string, memberCount int, contactInfo, name, phone, idCard, photoURL, idCardImage string) (domain.CompetitionReg, error) {
	now := time.Now()
	// Check competition exists
	if _, err := s.repo.FindByID(competitionID); err != nil {
		return domain.CompetitionReg{}, err
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
	return s.repo.CreateReg(cr)
}

func (s *CompetitionService) ListRegs(competitionID string) ([]domain.CompetitionReg, error) {
	return s.repo.ListRegs(competitionID)
}

// ---- EventService (协会活动) ----

type EventService struct {
	repo repository.EventRepository
}

func NewEventService(repo repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) Create(title, eventType, description, location, coverURL string, startTime, endTime time.Time, maxAttendees int) (domain.AssociationEvent, error) {
	now := time.Now()
	e := domain.AssociationEvent{
		ID:           fmt.Sprintf("event-%d", now.UnixNano()),
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
	return s.repo.Create(e)
}

func (s *EventService) List(page, pageSize int) ([]domain.AssociationEvent, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(offset, pageSize)
}

func (s *EventService) Get(id string) (domain.AssociationEvent, error) {
	return s.repo.FindByID(id)
}

func (s *EventService) Update(id, title, eventType, description, location, coverURL, status string, startTime, endTime time.Time, maxAttendees int) (domain.AssociationEvent, error) {
	ev, err := s.repo.FindByID(id)
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
	return s.repo.Update(ev)
}

func (s *EventService) Delete(id string) error { return s.repo.Delete(id) }

func (s *EventService) Register(eventID, userID, name, phone, org string) (domain.EventRegistration, error) {
	now := time.Now()
	if _, err := s.repo.FindByID(eventID); err != nil {
		return domain.EventRegistration{}, err
	}
	er := domain.EventRegistration{
		ID:        fmt.Sprintf("evtreg-%d", now.UnixNano()),
		EventID:   eventID,
		UserID:    userID,
		Name:      name,
		Phone:     phone,
		Org:       org,
		Status:    "registered",
		CreatedAt: now,
	}
	return s.repo.CreateReg(er)
}

func (s *EventService) ListRegs(eventID string) ([]domain.EventRegistration, error) {
	return s.repo.ListRegs(eventID)
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

func (s *ResourceService) Create(ownerID, name, resType, model, specs, location, bookingInfo string, priceFen int64, visibilityLevel string) (domain.IndustryResource, error) {
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
	return s.repo.Create(r)
}

func (s *ResourceService) List(resType string, page, pageSize int) ([]domain.IndustryResource, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(resType, offset, pageSize)
}

func (s *ResourceService) Get(id string) (domain.IndustryResource, error) {
	return s.repo.FindByID(id)
}

func (s *ResourceService) Update(id, name, resType, model, specs, location, bookingInfo string, priceFen int64, visibilityLevel, status string) (domain.IndustryResource, error) {
	r, err := s.repo.FindByID(id)
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
	return s.repo.Update(r)
}

func (s *ResourceService) Delete(id string) error { return s.repo.Delete(id) }

// Book 提交资源预约（C11：小程序资源详情页 → POST /api/v1/industry-resources/{id}/book）。
// date 为 YYYY-MM-DD（小程序日期选择器格式，格式校验在 Handler 层）。
func (s *ResourceService) Book(userID, resourceID, date, purpose, contactName, contactPhone string) (domain.IndustryResourceBooking, error) {
	if _, err := s.repo.FindByID(resourceID); err != nil {
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
	return s.repo.CreateBooking(b)
}

// ListBookingsByResource 某资源的全部预约（供测试与管理端查询）。
func (s *ResourceService) ListBookingsByResource(resourceID string) ([]domain.IndustryResourceBooking, error) {
	return s.repo.ListBookingsByResource(resourceID)
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

func (s *EmergencyService) CreateResource(ownerID, name, resType, specs, location, contactInfo string, quantity int) (domain.EmergencyResource, error) {
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
	return s.repo.CreateResource(r)
}

func (s *EmergencyService) ListResources(resType, q string, page, pageSize int) ([]domain.EmergencyResource, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListResources(normalizeEmergencyResType(resType), strings.TrimSpace(q), offset, pageSize)
}

func (s *EmergencyService) GetResource(id string) (domain.EmergencyResource, error) {
	return s.repo.FindResourceByID(id)
}

func (s *EmergencyService) FindDispatchByID(id string) (domain.EmergencyDispatch, error) {
	return s.repo.FindDispatchByID(id)
}

// Emergency Dispatches

func (s *EmergencyService) CreateDispatch(resourceID, eventDesc, location, commander, result string, startTime, endTime time.Time) (domain.EmergencyDispatch, error) {
	now := time.Now()
	if _, err := s.repo.FindResourceByID(resourceID); err != nil {
		return domain.EmergencyDispatch{}, err
	}
	d := domain.EmergencyDispatch{
		ID:         fmt.Sprintf("emdisp-%d", now.UnixNano()),
		ResourceID: resourceID,
		EventDesc:  eventDesc,
		Location:   location,
		StartTime:  startTime,
		EndTime:    endTime,
		Commander:  commander,
		Result:     result,
		Status:     "dispatched",
		CreatedAt:  now,
	}
	return s.repo.CreateDispatch(d)
}

func (s *EmergencyService) ListDispatches(page, pageSize int) ([]domain.EmergencyDispatch, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListDispatches(offset, pageSize)
}

func (s *EmergencyService) UpdateResource(id, name, resType, specs, location, contactInfo, status string, quantity int) (domain.EmergencyResource, error) {
	r, err := s.repo.FindResourceByID(id)
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
	return s.repo.UpdateResource(r)
}

func (s *EmergencyService) DeleteResource(id string) error { return s.repo.DeleteResource(id) }
func (s *EmergencyService) DeleteDispatch(id string) error { return s.repo.DeleteDispatch(id) }

func (s *EmergencyService) UpdateDispatch(id, resourceID, eventDesc, location, commander, result, status string, startTime, endTime time.Time) (domain.EmergencyDispatch, error) {
	d, err := s.repo.FindDispatchByID(id)
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
	return s.repo.UpdateDispatch(d)
}
