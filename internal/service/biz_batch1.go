package service

import (
	"context"
	"errors"
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

func (s *ResourcePoolService) Create(ctx context.Context, name, poolType, description, ownerID string) (domain.ResourcePool, error) {
	p := domain.ResourcePool{ID: nextID("pool"),
		Name: name, PoolType: poolType, Description: description, OwnerID: ownerID,
		Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(ctx, p)
}
func (s *ResourcePoolService) List(ctx context.Context, poolType string) ([]domain.ResourcePool, error) {
	return s.repo.List(ctx, poolType)
}
func (s *ResourcePoolService) Get(ctx context.Context, id string) (domain.ResourcePool, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *ResourcePoolService) AddMember(ctx context.Context, poolID, resID, resType string, quantity int) (domain.ResourcePoolMember, error) {
	if quantity <= 0 {
		return domain.ResourcePoolMember{}, errors.New("quantity must be positive")
	}
	m := domain.ResourcePoolMember{ID: nextID("rpm"),
		PoolID: poolID, ResID: resID, ResType: resType, Quantity: quantity,
		Status: "standby", JoinedAt: time.Now()}
	return s.repo.AddMember(ctx, m)
}
func (s *ResourcePoolService) ListMembers(ctx context.Context, poolID string) ([]domain.ResourcePoolMember, error) {
	return s.repo.ListMembers(ctx, poolID)
}

// ── TestSite Service ──

type TestSiteService struct{ repo repository.TestSiteRepository }

func NewTestSiteService(r repository.TestSiteRepository) *TestSiteService {
	return &TestSiteService{repo: r}
}

func (s *TestSiteService) Create(ctx context.Context, name, siteType, location, bookingRule, ownerID string, priceFen int64, facilities []string, status, airspaceRange, maxTakeoffWeight, runwayLength, maxFlightHeight, compatibleModels, imageURL string) (domain.TestSite, error) {
	if priceFen < 0 {
		return domain.TestSite{}, errors.New("price cannot be negative")
	}
	if status == "" {
		status = "available"
	}
	ts := domain.TestSite{ID: nextID("tst"),
		Name: name, SiteType: siteType, OwnerID: ownerID, Location: location,
		Facilities: facilities, PriceFen: priceFen, BookingRule: bookingRule,
		Status: status, AirspaceRange: airspaceRange, MaxTakeoffWeight: maxTakeoffWeight,
		RunwayLength: runwayLength, MaxFlightHeight: maxFlightHeight,
		CompatibleModels: compatibleModels, ImageURL: imageURL,
		CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(ctx, ts)
}
func (s *TestSiteService) List(ctx context.Context, siteType string) ([]domain.TestSite, error) {
	return s.repo.List(ctx, siteType)
}
func (s *TestSiteService) Get(ctx context.Context, id string) (domain.TestSite, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *TestSiteService) UpdateSite(ctx context.Context, id, name, siteType, location, bookingRule, status string, priceFen int64, facilities []string, airspaceRange, maxTakeoffWeight, runwayLength, maxFlightHeight, compatibleModels, imageURL string) (domain.TestSite, error) {
	if priceFen < 0 {
		return domain.TestSite{}, errors.New("price cannot be negative")
	}
	site, err := s.repo.FindByID(ctx, id)
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
	site.AirspaceRange = airspaceRange
	site.MaxTakeoffWeight = maxTakeoffWeight
	site.RunwayLength = runwayLength
	site.MaxFlightHeight = maxFlightHeight
	site.CompatibleModels = compatibleModels
	site.ImageURL = imageURL
	site.UpdatedAt = time.Now()
	return s.repo.UpdateSite(ctx, site)
}

func (s *TestSiteService) Book(ctx context.Context, siteID, userID, purpose, contactName, contactPhone string, startTime, endTime time.Time, bookingType, model, licenseURL, teamName string, peopleCount int, equipmentList, qualificationURL, equipmentNote, timeSlots string) (domain.TestSiteBooking, error) {
	// 时间范围校验：结束时间必须晚于开始时间（与 EventService 同款校验）
	if !endTime.After(startTime) {
		return domain.TestSiteBooking{}, errors.New("结束时间必须晚于开始时间")
	}
	// 站点存在校验：防对幽灵站点造预约
	if _, err := s.repo.FindByID(ctx, siteID); err != nil {
		return domain.TestSiteBooking{}, fmt.Errorf("试飞场不存在")
	}
	// 冲突判定升级：pending/approved 均占位（此前只认 approved，
	// 两条 pending 重叠可先后被批准导致双重占用）
	unlock := lockByKey("testsite-book|" + siteID)
	defer unlock()
	bookings, err := s.repo.ListBookings(ctx, siteID)
	if err != nil {
		return domain.TestSiteBooking{}, err
	}
	for _, b := range bookings {
		if (b.Status == "approved" || b.Status == "pending") && !(endTime.Before(b.StartTime) || startTime.After(b.EndTime)) {
			return domain.TestSiteBooking{}, fmt.Errorf("time slot conflicted")
		}
	}
	bk := domain.TestSiteBooking{ID: nextID("tsbk"),
		SiteID: siteID, UserID: userID, Purpose: purpose,
		ContactName: contactName, ContactPhone: contactPhone,
		StartTime: startTime, EndTime: endTime, Status: "pending", CreatedAt: time.Now(),
		BookingType: bookingType, Model: model, LicenseURL: licenseURL,
		TeamName: teamName, PeopleCount: peopleCount,
		EquipmentList: equipmentList, QualificationURL: qualificationURL,
		EquipmentNote: equipmentNote, TimeSlots: timeSlots}
	return s.repo.CreateBooking(ctx, bk)
}
func (s *TestSiteService) ReviewBooking(ctx context.Context, bookingID, status, note string) (domain.TestSiteBooking, error) {
	// 审批复查冲突：approved 前重跑同站点时段冲突检测——
	// 此前审批直接改状态，两条重叠 pending 可先后被批准（双重占用）。
	if status == "approved" {
		b, err := s.repo.FindBookingByID(ctx, bookingID)
		if err != nil {
			return domain.TestSiteBooking{}, err
		}
		bookings, err := s.repo.ListBookings(ctx, b.SiteID)
		if err != nil {
			return domain.TestSiteBooking{}, err
		}
		for _, ob := range bookings {
			if ob.ID == bookingID || ob.Status == "rejected" || ob.Status == "cancelled" {
				continue
			}
			if !(b.EndTime.Before(ob.StartTime) || b.StartTime.After(ob.EndTime)) {
				return domain.TestSiteBooking{}, fmt.Errorf("该时段已有预约，审批冲突")
			}
		}
	}
	return s.repo.UpdateBookingStatus(ctx, bookingID, status, note)
}

// ListAllBookings 管理端全量预约记录（分页）。
func (s *TestSiteService) ListAllBookings(ctx context.Context, offset, limit int) ([]domain.TestSiteBooking, int, error) {
	return s.repo.ListAllBookings(ctx, offset, limit)
}
func (s *TestSiteService) ListBookings(ctx context.Context, siteID string) ([]domain.TestSiteBooking, error) {
	return s.repo.ListBookings(ctx, siteID)
}

// ListMyBookings 我的预约：当前用户提交的全部场地预约（最新在前）。
func (s *TestSiteService) ListMyBookings(ctx context.Context, userID string) ([]domain.TestSiteBooking, error) {
	return s.repo.ListBookingsByUser(ctx, userID)
}

func (s *TestSiteService) DeleteSite(ctx context.Context, id string) error {
	return s.repo.DeleteSite(ctx, id)
}

// ── Exhibition Service ──

type ExhibitionService struct {
	repo repository.ExhibitionRepository
}

func NewExhibitionService(r repository.ExhibitionRepository) *ExhibitionService {
	return &ExhibitionService{repo: r}
}

func (s *ExhibitionService) Create(ctx context.Context, title, category, description, location, organizer, coverURL string, startDate, endDate time.Time, boothCount int, boothPrice int64, status string) (domain.Exhibition, error) {
	if boothPrice < 0 || boothCount < 0 {
		return domain.Exhibition{}, errors.New("booth count/price cannot be negative")
	}
	if status == "" {
		status = "draft"
	}
	e := domain.Exhibition{ID: nextID("expo"),
		Title: title, Category: category, Description: description, Location: location,
		Organizer: organizer, CoverURL: coverURL, StartDate: startDate, EndDate: endDate,
		BoothCount: boothCount, BoothPrice: boothPrice, Status: status,
		CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(ctx, e)
}
func (s *ExhibitionService) List(ctx context.Context, page, pageSize int) ([]domain.Exhibition, int, error) {
	return s.repo.List(ctx, (page-1)*pageSize, pageSize)
}
func (s *ExhibitionService) Get(ctx context.Context, id string) (domain.Exhibition, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ExhibitionService) Update(ctx context.Context, id, title, category, description, location, organizer, coverURL string, startDate, endDate time.Time, boothCount int, boothPrice int64, status string) (domain.Exhibition, error) {
	if boothPrice < 0 || boothCount < 0 {
		return domain.Exhibition{}, errors.New("booth count/price cannot be negative")
	}
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Exhibition{}, err
	}
	e.Title = title
	e.Category = category
	e.Description = description
	e.Location = location
	e.Organizer = organizer
	e.CoverURL = coverURL
	e.StartDate = startDate
	e.EndDate = endDate
	e.BoothCount = boothCount
	e.BoothPrice = boothPrice
	e.Status = status
	e.UpdatedAt = time.Now()
	return s.repo.Update(ctx, e)
}

func (s *ExhibitionService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ExhibitionService) ApplyBooth(ctx context.Context, exhibitionID, exhibitorID, boothNumber, exhibitName, exhibitDesc string) (domain.ExhibitionBooth, error) {
	// 业务校验：展会存在 + 招募中才可申请 + 展位数未超 + 重复申请/展位号去重。
	// 此前无任何校验：booth_count 形同虚设、同一企业可重复占位、同一展位号可被多人认领。
	ex, err := s.repo.FindByID(ctx, exhibitionID)
	if err != nil {
		return domain.ExhibitionBooth{}, fmt.Errorf("展会不存在")
	}
	if ex.Status != "" && ex.Status != "recruiting" && ex.Status != "published" {
		return domain.ExhibitionBooth{}, fmt.Errorf("展会未开放展位申请 (status %s)", ex.Status)
	}
	// 座位判定：以"有效状态（applied/approved/paid）"展位数对比 booth_count。
	booths, err := s.repo.ListBooths(ctx, exhibitionID)
	if err != nil {
		return domain.ExhibitionBooth{}, err
	}
	counted := 0
	for _, b := range booths {
		if b.Status == "applied" || b.Status == "approved" || b.Status == "paid" {
			counted++
			if b.ExhibitorID == exhibitorID {
				return domain.ExhibitionBooth{}, fmt.Errorf("您已申请过该展会展位")
			}
			if boothNumber != "" && b.BoothNumber == boothNumber {
				return domain.ExhibitionBooth{}, fmt.Errorf("展位号 %s 已被占用", boothNumber)
			}
		}
	}
	if ex.BoothCount > 0 && counted >= ex.BoothCount {
		return domain.ExhibitionBooth{}, fmt.Errorf("展位已满")
	}
	b := domain.ExhibitionBooth{ID: nextID("exbk"),
		ExhibitionID: exhibitionID, ExhibitorID: exhibitorID, BoothNumber: boothNumber,
		ExhibitName: exhibitName, ExhibitDesc: exhibitDesc, Status: "applied", CreatedAt: time.Now()}
	return s.repo.CreateBooth(ctx, b)
}
func (s *ExhibitionService) ListBooths(ctx context.Context, exhibitionID string) ([]domain.ExhibitionBooth, error) {
	return s.repo.ListBooths(ctx, exhibitionID)
}
func (s *ExhibitionService) ReviewBooth(ctx context.Context, boothID, status string) (domain.ExhibitionBooth, error) {
	return s.repo.UpdateBoothStatus(ctx, boothID, status)
}
