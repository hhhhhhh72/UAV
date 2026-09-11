package httpapi

import (
	"context"
	"log/slog"
	"time"

	"drone-platform/internal/service"
)

const (
	// userPurgeInterval 清除周期：缓冲期是"天"级事件，一天一查足够。
	userPurgeInterval = 24 * time.Hour
	// userPurgeFirstDelay 启动后首次扫描延迟：避开启动高峰。
	userPurgeFirstDelay = 2 * time.Minute
)

// StartUserPurge 启动"注销账号缓冲期到期自动清除"后台任务。
//
// 与证书到期提醒同一套做法：无退出通道的后台 goroutine + recover 兜底。
// 清除对象是"已注销且超过缓冲期（默认 7 天）"的账号行；被工单等外键引用的账号
// 由仓储跳过、下次再试；业务内容一律不动（46 张表以文本列记用户 ID，无外键级联）。
func (s *Server) StartUserPurge() {
	if s == nil || s.userSvc == nil {
		return
	}
	go func() {
		defer func() { _ = recover() }()
		time.Sleep(userPurgeFirstDelay)
		for {
			s.scanUserPurge()
			time.Sleep(userPurgeInterval)
		}
	}()
}

// scanUserPurge 扫一遍超过缓冲期的已注销账号并物理清除账号行。
func (s *Server) scanUserPurge() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	purged, err := s.userSvc.PurgeExpired(ctx)
	if err != nil {
		slog.Warn("user purge: scan failed", "purged", purged, "error", err)
		return
	}
	if purged > 0 {
		slog.Info("user purge scan done", "retention_days", int(service.UserPurgeRetention.Hours()/24), "purged", purged)
	}
}
