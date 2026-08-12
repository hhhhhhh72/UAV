package service

import (
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ── ResourcePool Service ──

type ResourcePoolService struct {
	repo repository.ResourcePoolRepository
}

func NewResourcePoolService(r repository.ResourcePoolRepository) *ResourcePoolService {
	return &ResourcePoolService{repo: r}
}

func (s *ResourcePoolService) Create(name, poolType, description, ownerID string) (domain.ResourcePool, error) {
	p := domain.ResourcePool{ID: fmt.Sprintf("pool-%d", time.Now().UnixNano()),
		Name: name, PoolType: poolType, Description: description, OwnerID: ownerID,
		Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(p)
}
func (s *ResourcePoolService) List(poolType string) ([]domain.ResourcePool, error) {
	return s.repo.List(poolType)
}
func (s *ResourcePoolService) Get(id string) (domain.ResourcePool, error) {
	return s.repo.FindByID(id)
}
func (s *ResourcePoolService) AddMember(poolID, resID, resType string, quantity int) (domain.ResourcePoolMember, error) {
	m := domain.ResourcePoolMember{ID: fmt.Sprintf("rpm-%d", time.Now().UnixNano()),
		PoolID: poolID, ResID: resID, ResType: resType, Quantity: quantity,
		Status: "standby", JoinedAt: time.Now()}
	return s.repo.AddMember(m)
}
func (s *ResourcePoolService) ListMembers(poolID string) ([]domain.ResourcePoolMember, error) {
	return s.repo.ListMembers(poolID)
}

// ── TestSite Service ──

type TestSiteService struct{ repo repository.TestSiteRepository }

func NewTestSiteService(r repository.TestSiteRepository) *TestSiteService {
	return &TestSiteService{repo: r}
}

func (s *TestSiteService) Create(name, siteType, location, bookingRule, ownerID string, priceFen int64, facilities []string, status string) (domain.TestSite, error) {
	if status == "" {
		status = "available"
	}
	ts := domain.TestSite{ID: fmt.Sprintf("tst-%d", time.Now().UnixNano()),
		Name: name, SiteType: siteType, OwnerID: ownerID, Location: location,
		Facilities: facilities, PriceFen: priceFen, BookingRule: bookingRule,
		Status: status, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(ts)
}
func (s *TestSiteService) List(siteType string) ([]domain.TestSite, error) {
	return s.repo.List(siteType)
}
func (s *TestSiteService) Get(id string) (domain.TestSite, error) {
	return s.repo.FindByID(id)
}

func (s *TestSiteService) UpdateSite(id, name, siteType, location, bookingRule, status string, priceFen int64, facilities []string) (domain.TestSite, error) {
	site, err := s.repo.FindByID(id)
	if err != nil {
		return domain.TestSite{}, err
	}
	site.Name = name
	site.SiteType = siteType
	site.Location = location
	site.BookingRule = bookingRule
	site.Status = status
	site.PriceFen = priceFen
	site.Facilities = facilities
	site.UpdatedAt = time.Now()
	return s.repo.UpdateSite(site)
}

func (s *TestSiteService) Book(siteID, userID, purpose, contactName, contactPhone string, startTime, endTime time.Time) (domain.TestSiteBooking, error) {
	// Check conflicts
	bookings, err := s.repo.ListBookings(siteID)
	if err != nil {
		return domain.TestSiteBooking{}, err
	}
	for _, b := range bookings {
		if b.Status == "approved" && !(endTime.Before(b.StartTime) || startTime.After(b.EndTime)) {
			return domain.TestSiteBooking{}, fmt.Errorf("time slot conflicted")
		}
	}
	bk := domain.TestSiteBooking{ID: fmt.Sprintf("tsbk-%d", time.Now().UnixNano()),
		SiteID: siteID, UserID: userID, Purpose: purpose,
		ContactName: contactName, ContactPhone: contactPhone,
		StartTime: startTime, EndTime: endTime, Status: "pending", CreatedAt: time.Now()}
	return s.repo.CreateBooking(bk)
}
func (s *TestSiteService) ReviewBooking(bookingID, status, note string) (domain.TestSiteBooking, error) {
	return s.repo.UpdateBookingStatus(bookingID, status, note)
}

// ListAllBookings 管理端全量预约记录（分页）。
func (s *TestSiteService) ListAllBookings(offset, limit int) ([]domain.TestSiteBooking, int, error) {
	return s.repo.ListAllBookings(offset, limit)
}
func (s *TestSiteService) ListBookings(siteID string) ([]domain.TestSiteBooking, error) {
	return s.repo.ListBookings(siteID)
}

// ListMyBookings 我的预约：当前用户提交的全部场地预约（最新在前）。
func (s *TestSiteService) ListMyBookings(userID string) ([]domain.TestSiteBooking, error) {
	return s.repo.ListBookingsByUser(userID)
}

func (s *TestSiteService) DeleteSite(id string) error { return s.repo.DeleteSite(id) }

// ── Exhibition Service ──

type ExhibitionService struct {
	repo repository.ExhibitionRepository
}

func NewExhibitionService(r repository.ExhibitionRepository) *ExhibitionService {
	return &ExhibitionService{repo: r}
}

func (s *ExhibitionService) Create(title, category, description, location, organizer, coverURL string, startDate, endDate time.Time, boothCount int, boothPrice int64, status string) (domain.Exhibition, error) {
	if status == "" {
		status = "draft"
	}
	e := domain.Exhibition{ID: fmt.Sprintf("expo-%d", time.Now().UnixNano()),
		Title: title, Category: category, Description: description, Location: location,
		Organizer: organizer, CoverURL: coverURL, StartDate: startDate, EndDate: endDate,
		BoothCount: boothCount, BoothPrice: boothPrice, Status: status,
		CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(e)
}
func (s *ExhibitionService) List(page, pageSize int) ([]domain.Exhibition, int, error) {
	return s.repo.List((page-1)*pageSize, pageSize)
}
func (s *ExhibitionService) Get(id string) (domain.Exhibition, error) {
	return s.repo.FindByID(id)
}

func (s *ExhibitionService) Update(id, title, category, description, location, organizer string, startDate, endDate time.Time, boothCount int, boothPrice int64, status string) (domain.Exhibition, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Exhibition{}, err
	}
	e.Title = title
	e.Category = category
	e.Description = description
	e.Location = location
	e.Organizer = organizer
	e.StartDate = startDate
	e.EndDate = endDate
	e.BoothCount = boothCount
	e.BoothPrice = boothPrice
	e.Status = status
	e.UpdatedAt = time.Now()
	return s.repo.Update(e)
}

func (s *ExhibitionService) Delete(id string) error { return s.repo.Delete(id) }

func (s *ExhibitionService) ApplyBooth(exhibitionID, exhibitorID, boothNumber, exhibitName, exhibitDesc string) (domain.ExhibitionBooth, error) {
	b := domain.ExhibitionBooth{ID: fmt.Sprintf("exbk-%d", time.Now().UnixNano()),
		ExhibitionID: exhibitionID, ExhibitorID: exhibitorID, BoothNumber: boothNumber,
		ExhibitName: exhibitName, ExhibitDesc: exhibitDesc, Status: "applied", CreatedAt: time.Now()}
	return s.repo.CreateBooth(b)
}
func (s *ExhibitionService) ListBooths(exhibitionID string) ([]domain.ExhibitionBooth, error) {
	return s.repo.ListBooths(exhibitionID)
}
func (s *ExhibitionService) ReviewBooth(boothID, status string) (domain.ExhibitionBooth, error) {
	return s.repo.UpdateBoothStatus(boothID, status)
}
