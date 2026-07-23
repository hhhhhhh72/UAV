package service

import (
	"fmt"
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

func (s *CompetitionService) Create(title, category, description, location, sponsor string, startDate, endDate time.Time, maxTeams int) (domain.Competition, error) {
	now := time.Now()
	c := domain.Competition{
		ID:          fmt.Sprintf("comp-%d", now.UnixNano()),
		Title:       title,
		Category:    category,
		Description: description,
		Location:    location,
		Sponsor:     sponsor,
		StartDate:   startDate,
		EndDate:     endDate,
		MaxTeams:    maxTeams,
		Status:      "published",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.Create(c)
}

func (s *CompetitionService) List(page, pageSize int) ([]domain.Competition, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(offset, pageSize)
}

func (s *CompetitionService) Get(id string) (domain.Competition, error) {
	return s.repo.FindByID(id)
}

func (s *CompetitionService) Update(id, title, category, description, location, sponsor string, startDate, endDate time.Time, maxTeams int) (domain.Competition, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Competition{}, err
	}
	c.Title = title
	c.Category = category
	c.Description = description
	c.Location = location
	c.Sponsor = sponsor
	c.StartDate = startDate
	c.EndDate = endDate
	c.MaxTeams = maxTeams
	c.UpdatedAt = time.Now()
	return s.repo.Update(c)
}

func (s *CompetitionService) Register(competitionID, userID, teamName string, memberCount int, contactInfo string) (domain.CompetitionReg, error) {
	now := time.Now()
	// Check competition exists
	if _, err := s.repo.FindByID(competitionID); err != nil {
		return domain.CompetitionReg{}, err
	}
	cr := domain.CompetitionReg{
		ID:            fmt.Sprintf("compreg-%d", now.UnixNano()),
		CompetitionID: competitionID,
		UserID:        userID,
		TeamName:      teamName,
		MemberCount:   memberCount,
		ContactInfo:   contactInfo,
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

func (s *EventService) Update(id, title, eventType, description, location, coverURL string, startTime, endTime time.Time, maxAttendees int) (domain.AssociationEvent, error) {
	ev, err := s.repo.FindByID(id)
	if err != nil {
		return domain.AssociationEvent{}, err
	}
	ev.Title = title
	ev.EventType = eventType
	ev.Description = description
	ev.Location = location
	ev.CoverURL = coverURL
	ev.StartTime = startTime
	ev.EndTime = endTime
	ev.MaxAttendees = maxAttendees
	ev.UpdatedAt = time.Now()
	return s.repo.Update(ev)
}

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

type ResourceService struct {
	repo repository.ResourceRepository
}

func NewResourceService(repo repository.ResourceRepository) *ResourceService {
	return &ResourceService{repo: repo}
}

func (s *ResourceService) Create(ownerID, name, resType, model, specs, location, bookingInfo string, priceFen int64) (domain.IndustryResource, error) {
	now := time.Now()
	r := domain.IndustryResource{
		ID:          fmt.Sprintf("res-%d", now.UnixNano()),
		OwnerID:     ownerID,
		Name:        name,
		ResType:     resType,
		Model:       model,
		Specs:       specs,
		Location:    location,
		PriceFen:    priceFen,
		BookingInfo: bookingInfo,
		Status:      "available",
		CreatedAt:   now,
		UpdatedAt:   now,
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

func (s *ResourceService) Update(id, name, resType, model, specs, location, bookingInfo string, priceFen int64) (domain.IndustryResource, error) {
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
	r.UpdatedAt = time.Now()
	return s.repo.Update(r)
}

// ---- EmergencyService (应急管理) ----

type EmergencyService struct {
	repo repository.EmergencyRepository
}

func NewEmergencyService(repo repository.EmergencyRepository) *EmergencyService {
	return &EmergencyService{repo: repo}
}

// Emergency Resources

func (s *EmergencyService) CreateResource(ownerID, name, resType, specs, location, contactInfo string, quantity int) (domain.EmergencyResource, error) {
	now := time.Now()
	r := domain.EmergencyResource{
		ID:          fmt.Sprintf("emres-%d", now.UnixNano()),
		OwnerID:     ownerID,
		Name:        name,
		ResType:     resType,
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

func (s *EmergencyService) ListResources(page, pageSize int) ([]domain.EmergencyResource, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListResources(offset, pageSize)
}

func (s *EmergencyService) GetResource(id string) (domain.EmergencyResource, error) {
	return s.repo.FindResourceByID(id)
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
