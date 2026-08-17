package service

import (
	"context"
	"math"
	"sort"
	"strings"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// MatchingService provides intelligent supply-demand matching.
type MatchingService struct {
	demandRepo repository.DemandRepository
}

func NewMatchingService(dr repository.DemandRepository) *MatchingService {
	return &MatchingService{demandRepo: dr}
}

// MatchResult is a scored demand recommendation.
type MatchResult struct {
	Demand  domain.Demand `json:"demand"`
	Score   float64       `json:"score"`
	Reasons []string      `json:"reasons"`
}

// Recommend returns demands ranked by relevance for a given user context.
// lat/lng are optional (0 means unknown). bizType is the user's preferred type.
func (s *MatchingService) Recommend(ctx context.Context, userID string, lat, lng float64, bizType, district string, limit int) ([]MatchResult, error) {
	demands, err := s.demandRepo.List(ctx, repository.DemandFilter{})
	if err != nil {
		return nil, err
	}

	results := make([]MatchResult, 0, len(demands))
	for _, d := range demands {
		if d.Status != domain.DemandPublished {
			continue
		}
		score, reasons := s.score(d, lat, lng, bizType, district)
		if score > 0 {
			results = append(results, MatchResult{Demand: d, Score: score, Reasons: reasons})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

// score computes a relevance score (0.0-1.0) for a demand given user context.
func (s *MatchingService) score(d domain.Demand, lat, lng float64, bizType, district string) (float64, []string) {
	var reasons []string
	score := 0.0
	weights := map[string]float64{
		"biz_type_match":    0.35,
		"district_match":    0.25,
		"distance_close":    0.25,
		"budget_reasonable": 0.15,
	}

	// 1. Business type match (highest weight).
	if bizType != "" && strings.EqualFold(string(d.BizType), bizType) {
		score += weights["biz_type_match"]
		reasons = append(reasons, "业务类型匹配")
	}

	// 2. District match.
	if district != "" && strings.Contains(d.District, district) {
		score += weights["district_match"]
		reasons = append(reasons, "区域匹配")
	}

	// 3. Distance score (Gaussian decay, max ~50km).
	if lat != 0 && lng != 0 && d.Latitude != 0 && d.Longitude != 0 {
		dist := haversineKm(lat, lng, d.Latitude, d.Longitude)
		if dist < 10 {
			score += weights["distance_close"]
			reasons = append(reasons, "距离近(<10km)")
		} else if dist < 30 {
			score += weights["distance_close"] * 0.5
			reasons = append(reasons, "距离适中(<30km)")
		} else if dist < 50 {
			score += weights["distance_close"] * 0.2
		}
	}

	// 4. Budget signals quality.
	if d.BudgetFen > 0 {
		score += weights["budget_reasonable"] * 0.5
	}

	// Round to 4 decimal places.
	score = math.Round(score*10000) / 10000
	return score, reasons
}

// haversineKm computes the great-circle distance in km between two points.
func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// SearchAndMatch combines keyword search with intelligent scoring.
func (s *MatchingService) SearchAndMatch(ctx context.Context, q string, lat, lng float64, bizType string, limit int) ([]MatchResult, error) {
	demands, err := s.demandRepo.Search(ctx, q)
	if err != nil {
		return nil, err
	}

	results := make([]MatchResult, 0, len(demands))
	for _, d := range demands {
		if d.Status != domain.DemandPublished {
			continue
		}
		score, reasons := s.score(d, lat, lng, bizType, "")
		// Boost by keyword match presence.
		score += 0.3
		if score > 1.0 {
			score = 1.0
		}
		score = math.Round(score*10000) / 10000
		results = append(results, MatchResult{Demand: d, Score: score, Reasons: reasons})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}
