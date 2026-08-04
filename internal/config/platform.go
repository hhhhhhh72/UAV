package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"drone-platform/internal/domain"
)

type PlatformConfig struct {
	Banners      []domain.Banner         `json:"banners"`
	Notices      []string                `json:"notices"`
	QuickEntries []domain.HomeQuickEntry `json:"quick_entries"`
	// MatchFeeRate 撮合服务费率（百分比，2 = 2%）；0 表示未启用收费（功能方案修订版 v2 第八章）
	MatchFeeRate float64 `json:"match_fee_rate"`
	// MatchFeeNote 撮合服务费说明文案（面向会员展示）
	MatchFeeNote string `json:"match_fee_note"`
}

// defaultMatchFeeNote 撮合服务费默认说明（未启用时的展示文案）。
const defaultMatchFeeNote = "撮合服务费：供需对接成功后按费率收取，会员享折扣（费率待协会确认后启用）"

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
		// 商业化费率默认关闭（0 = 未启用），待协会确认费率后由管理端开启
		MatchFeeRate: 0,
		MatchFeeNote: defaultMatchFeeNote,
	}
)

func init() {
	if data, err := os.ReadFile("platform_config.json"); err == nil {
		var cfg PlatformConfig
		if json.Unmarshal(data, &cfg) == nil {
			// 旧配置文件可能缺少新版费率字段：回填默认说明，避免空文案
			if cfg.MatchFeeNote == "" {
				cfg.MatchFeeNote = defaultMatchFeeNote
			}
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
	if err != nil {
		return fmt.Errorf("marshal platform config: %w", err)
	}
	return os.WriteFile("platform_config.json", data, 0644)
}
