package service

import (
	"math"

	"drone-platform/internal/config"
	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type HomeService struct {
	demandRepo     repository.DemandRepository
	enterpriseRepo repository.EnterpriseRepository
}

func NewHomeService(d repository.DemandRepository, e repository.EnterpriseRepository) *HomeService {
	return &HomeService{demandRepo: d, enterpriseRepo: e}
}

// HomeData aggregates all data needed by the mini-program home page.
type HomeData struct {
	City         string                  `json:"city"`
	Banners      []domain.Banner         `json:"banners"`
	QuickEntries []domain.HomeQuickEntry `json:"quick_entries"`
	HotDemands   []domain.Demand         `json:"hot_demands"`
	Notices      []string                `json:"notices"`
	Shops        []domain.Shop           `json:"shops"` // 商家列表：已审核企业（商家/企业合一）
}

// GetHome assembles the home page data with optional city and location.
func (s *HomeService) GetHome(city string, lat, lng float64) HomeData {
	if city == "" {
		city = "重庆"
	}

	// Banners and quick entries from dynamic config.
	cfg := config.GetPlatformConfig()
	banners := cfg.Banners
	entries := cfg.QuickEntries
	notices := cfg.Notices

	// Get published demands.
	demands, err := s.demandRepo.List(repository.DemandFilter{})
	if err != nil {
		demands = nil
	}

	// Sort by distance if coordinates provided.
	if lat != 0 && lng != 0 && len(demands) > 1 {
		sortByDistance(demands, lat, lng)
	}
	// Limit to 10 for home display.
	if len(demands) > 10 {
		demands = demands[:10]
	}

	// Sanitize public output.
	public := make([]domain.Demand, len(demands))
	for i, d := range demands {
		public[i] = sanitizeDemand(d)
	}

	// 商家列表 = 已审核企业（商家/企业合一：入驻审核通过即成为商家，供商家 Tab 展示）
	shops := []domain.Shop{}
	if s.enterpriseRepo != nil {
		ents, _, err := s.enterpriseRepo.ListByStatus(string(domain.EnterpriseApproved), 0, 100)
		if err == nil {
			for _, e := range ents {
				shops = append(shops, domain.Shop{
					ID:        e.ID,
					Name:      e.Name,
					OwnerID:   e.OwnerUserID,
					LogoURL:   e.Logo,
					IsMember:  e.IsMember,
					Status:    string(e.Status),
					CreatedAt: e.CreatedAt,
					UpdatedAt: e.UpdatedAt,
				})
			}
		}
	}

	return HomeData{City: city, Banners: banners, QuickEntries: entries, HotDemands: public, Notices: notices, Shops: shops}
}

func sortByDistance(demands []domain.Demand, lat, lng float64) {
	// Simple insertion sort by haversine approximation.
	dist := func(d domain.Demand) float64 {
		dlat := d.Latitude - lat
		dlng := d.Longitude - lng
		return math.Sqrt(dlat*dlat+dlng*dlng) * 111000 // approx meters
	}
	for i := 1; i < len(demands); i++ {
		for j := i; j > 0 && dist(demands[j]) < dist(demands[j-1]); j-- {
			demands[j], demands[j-1] = demands[j-1], demands[j]
		}
	}
}

func sanitizeDemand(d domain.Demand) domain.Demand {
	d.PublisherID = ""
	d.Contact = ""
	d.Latitude = 0
	d.Longitude = 0
	return d
}
