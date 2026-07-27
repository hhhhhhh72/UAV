package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"drone-platform/internal/domain"
)

type PlatformConfig struct {
	Banners      []domain.Banner          `json:"banners"`
	Notices      []string                 `json:"notices"`
	QuickEntries []domain.HomeQuickEntry  `json:"quick_entries"`
}

var (
	platformMu  sync.RWMutex
	platformCfg = PlatformConfig{
		Banners: []domain.Banner{
			{ID: "banner-1", ImageURL: "https://images.unsplash.com/photo-1473968512647-3e447244af8f?w=800&q=80", LinkURL: "/pages/demand/list", SortOrder: 1},
			{ID: "banner-2", ImageURL: "https://images.unsplash.com/photo-1506947411487-a56738267384?w=800&q=80", LinkURL: "/pages/training/courses", SortOrder: 2},
		},
		Notices: []string{"无人机外卖配送服务正式上线", "新开通江心屿景区无人机物流航线"},
		QuickEntries: []domain.HomeQuickEntry{
			{Key: "demand", Name: "需求大厅"},
			{Key: "enterprise", Name: "企业入驻"},
			{Key: "community", Name: "同城社区"},
			{Key: "jobs", Name: "求职招聘"},
		},
	}
)

func init() {
	if data, err := os.ReadFile("platform_config.json"); err == nil {
		var cfg PlatformConfig
		if json.Unmarshal(data, &cfg) == nil {
			platformCfg = cfg
		}
	}
}

func GetPlatformConfig() PlatformConfig {
	platformMu.RLock()
	defer platformMu.RUnlock()
	return platformCfg
}

func SavePlatformConfig(cfg PlatformConfig) error {
	platformMu.Lock()
	platformCfg = cfg
	platformMu.Unlock()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil { return fmt.Errorf("marshal platform config: %w", err) }
	return os.WriteFile("platform_config.json", data, 0644)
}
