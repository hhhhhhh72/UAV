package service

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// MatchingService provides intelligent supply-demand matching.
type MatchingService struct {
	demandRepo repository.DemandRepository // 需求列表 + 需求收藏（同接口）
	intentRepo repository.IntentRepository // 历史意向（画像数据源，可选注入）
	halfLife   time.Duration               // 需求新鲜度半衰期（14 天）
}

func NewMatchingService(dr repository.DemandRepository) *MatchingService {
	return &MatchingService{demandRepo: dr, halfLife: 14 * 24 * time.Hour}
}

// SetIntentRepo 注入历史意向数据源（画像）；不注入时画像仅含需求收藏，仍可工作。
func (s *MatchingService) SetIntentRepo(intent repository.IntentRepository) {
	s.intentRepo = intent
}

// MatchResult is a scored demand recommendation.
type MatchResult struct {
	Demand  domain.Demand `json:"demand"`
	Score   float64       `json:"score"`
	Reasons []string      `json:"reasons"`
}

// Recommend returns demands ranked by relevance for a given user context.
// lat/lng are optional (0 means unknown). bizType is the user's preferred type.
// 阶段 1：语义 2-gram 相似度 + 时间衰减 + 硬过滤（过期/取消）+ 历史行为画像。
func (s *MatchingService) Recommend(ctx context.Context, userID string, lat, lng float64, bizType, district string, limit int) ([]MatchResult, error) {
	demands, err := s.demandRepo.List(ctx, repository.DemandFilter{})
	if err != nil {
		return nil, err
	}
	profile := s.userProfile(ctx, userID)

	results := make([]MatchResult, 0, len(demands))
	for _, d := range demands {
		score, reasons := s.score(d, lat, lng, bizType, district, profile)
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

// userProfile 聚合用户画像文本：历史收藏需求 + 历史意向需求（标题/描述/类型/区域）。
// 每次调用实时聚合（当前量级可接受）；返回空串表示无画像（画像权重自动降级）。
func (s *MatchingService) userProfile(ctx context.Context, userID string) string {
	if userID == "" {
		return ""
	}
	var texts []string
	if s.intentRepo != nil {
		if intents, err := s.intentRepo.ListByIntentor(ctx, userID); err == nil {
			for _, it := range intents {
				if d, err := s.demandRepo.FindByID(ctx, it.DemandID); err == nil {
					texts = append(texts, d.Title, d.Description, string(d.BizType), d.District)
				}
			}
		}
	}
	if favs, err := s.demandRepo.ListFavoriteDemands(ctx, userID); err == nil {
		for _, d := range favs {
			texts = append(texts, d.Title, d.Description, string(d.BizType), d.District)
		}
	}
	return strings.Join(texts, " ")
}

// score computes a relevance score (0.0-1.0) for a demand given user context.
// 权重：语义 0.30 / 业务类型 0.15 / 区域 0.15 / 距离 0.20 / 时间 0.20。
// （旧版同权 0.35/0.25/0.25/0.15 且依赖字面全等，同义表达会得零分。）
func (s *MatchingService) score(d domain.Demand, lat, lng float64, bizType, district, profile string) (float64, []string) {
	// 硬过滤：过期（deadline 已过）一律排除（Deadline 为 YYYY-MM-DD 字符串，字典序即时间序）。
	if d.Status != domain.DemandPublished {
		return 0, nil
	}
	if d.Deadline != "" && d.Deadline < time.Now().Format("2006-01-02") {
		return 0, nil
	}

	var reasons []string
	score := 0.0
	hit := false // 至少一个语义维度命中才进入推荐（时间分只做排序，不做准入）

	// 1. 语义相似度（2-gram Dice）：需求文本 vs 用户画像（收藏/意向历史）。
	if profile != "" {
		sim := gramSim(shortText(d), profile)
		if sim > 0.35 {
			score += 0.30 * sim
			reasons = append(reasons, "与你关注过的需求相似")
			hit = true
		}
	}

	// 2. 业务类型匹配（字面，保留兜底）。
	if bizType != "" && strings.EqualFold(string(d.BizType), bizType) {
		score += 0.15
		reasons = append(reasons, "业务类型匹配")
		hit = true
	}

	// 3. 区域匹配。
	if district != "" && strings.Contains(d.District, district) {
		score += 0.15
		reasons = append(reasons, "区域匹配")
		hit = true
	}

	// 4. 距离（高斯衰减，sigma=20km）。
	if lat != 0 && lng != 0 && d.Latitude != 0 && d.Longitude != 0 {
		dist := haversineKm(lat, lng, d.Latitude, d.Longitude)
		gauss := math.Exp(-(dist * dist) / (2 * 20 * 20))
		score += 0.20 * gauss
		if gauss > 0.02 {
			hit = true
			if dist < 10 {
				reasons = append(reasons, "距离近(<10km)")
			} else if dist < 30 {
				reasons = append(reasons, "距离适中(<30km)")
			}
		}
	}

	// 5. 时间新鲜度（半衰期 14 天）：刚创建（age≈0）即满分 0.20；
	// CreatedAt 零值（历史数据/测试 seed）防御为满分，不误杀。
	created := d.CreatedAt
	if created.IsZero() {
		created = time.Now()
	}
	age := time.Since(created)
	if age >= 0 {
		score += 0.20 * math.Exp(-float64(age)/float64(s.halfLife)*math.Ln2)
		if age < 3*24*time.Hour {
			reasons = append(reasons, "新发布")
		}
	}

	if !hit {
		return 0, nil // 无任何语义匹配（类型/区域/距离/画像），仅时间分不推荐
	}
	if score > 1.0 {
		score = 1.0
	}
	score = math.Round(score*10000) / 10000
	return score, reasons
}

// shortText 需求匹配文本（标题/描述/类型/区域 聚合，用于语义相似度）。
func shortText(d domain.Demand) string {
	return strings.Join([]string{d.Title, d.Description, string(d.BizType), d.District}, " ")
}

// SearchAndMatch combines keyword search with intelligent scoring.
func (s *MatchingService) SearchAndMatch(ctx context.Context, q string, lat, lng float64, bizType string, limit int) ([]MatchResult, error) {
	demands, err := s.demandRepo.Search(ctx, q)
	if err != nil {
		return nil, err
	}

	results := make([]MatchResult, 0, len(demands))
	for _, d := range demands {
		score, reasons := s.score(d, lat, lng, bizType, "", "")
		// Keyword presence boost.
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

// gramN 字符 2-gram 集合（中文无分词库的无外部依赖方案：按字符 bi-gram，跳过空白/标点）。
func gramN(s string, n int) map[string]struct{} {
	set := make(map[string]struct{})
	runes := []rune(strings.ToLower(s))
	var acc []rune
	flush := func() {
		for i := 0; i+1 < len(acc); i++ {
			set[string(acc[i:i+2])] = struct{}{}
		}
		acc = acc[:0]
	}
	for _, r := range runes {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			acc = append(acc, r)
		} else {
			flush()
		}
	}
	flush()
	return set
}

// gramSim Dice 系数：2*|A&B|/(|A|+|B|)，范围 0~1。
func gramSim(a, b string) float64 {
	sa := gramN(a, 2)
	sb := gramN(b, 2)
	if len(sa) == 0 || len(sb) == 0 {
		return 0
	}
	inter := 0
	for g := range sa {
		if _, ok := sb[g]; ok {
			inter++
		}
	}
	return 2 * float64(inter) / float64(len(sa)+len(sb))
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
