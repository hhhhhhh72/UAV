package httpapi

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"drone-platform/internal/domain"
)

const (
	// certReminderDays 提前提醒窗口（天）：与 GET /api/v1/certificates/expiring 的默认口径一致。
	certReminderDays = 30
	// certReminderExpiredGrace 已过期的证书最多回溯多久还提醒：更久远的早已失去可操作性，
	// 补发只会变成噪音（GetExpiringCerts 本身对"已过期"没有下界）。
	certReminderExpiredGrace = 180 * 24 * time.Hour
	// certReminderInterval 扫描周期：证书到期是"天"级事件，一天一查足够。
	certReminderInterval = 24 * time.Hour
	// certReminderFirstDelay 启动后首次扫描延迟：避开启动高峰，也让本地开发能很快看到效果。
	certReminderFirstDelay = 2 * time.Minute
)

// StartCertExpiryReminder 启动"证书到期提醒"后台任务。
//
// 与 escrow 孤儿补偿、验证码清理同一套做法：无退出通道的后台 goroutine + recover 兜底。
// 幂等性由 certReminderSent 保证——同一张证书只提醒一次，进程重启与重复扫描都不会重复轰炸。
func (s *Server) StartCertExpiryReminder() {
	if s == nil || s.trainingSvc == nil || s.expirySvc == nil || s.msgSvc == nil {
		return
	}
	go func() {
		defer func() { _ = recover() }()
		time.Sleep(certReminderFirstDelay)
		for {
			s.scanExpiringCertificates()
			time.Sleep(certReminderInterval)
		}
	}()
}

// scanExpiringCertificates 扫一遍"已审核通过、且在提醒窗口内（含近期已过期）"的证书，
// 给持证人本人发一条站内消息。
func (s *Server) scanExpiringCertificates() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	certs, err := s.trainingSvc.ListAllCertificates(ctx)
	if err != nil {
		slog.Warn("cert expiry reminder: list certificates failed", "error", err)
		return
	}
	candidates := s.expirySvc.GetExpiringCerts(certs, certReminderDays)
	if len(candidates) == 0 {
		return
	}
	sent, skipped := 0, 0
	for _, c := range candidates {
		if c.ExpireDate.Before(time.Now().Add(-certReminderExpiredGrace)) {
			continue
		}
		if s.certReminderSent(ctx, c) {
			skipped++
			continue
		}
		if err := s.sendCertReminder(ctx, c); err != nil {
			slog.Warn("cert expiry reminder: send failed", "cert", c.ID, "user", c.UserID, "error", err)
			continue
		}
		sent++
	}
	if sent > 0 || skipped > 0 {
		slog.Info("cert expiry reminder scan done", "candidates", len(candidates), "sent", sent, "skipped", skipped)
	}
}

// certReminderSent 查重：该证书是否已经给本人发过提醒（已读未读都算）。
// 查重失败返回 true —— 宁可漏发一次，也不能把用户轰炸一遍。
func (s *Server) certReminderSent(ctx context.Context, c domain.Certificate) bool {
	msgs, err := s.msgSvc.ListForUser(ctx, c.UserID, false)
	if err != nil {
		slog.Warn("cert expiry reminder: dedupe lookup failed", "cert", c.ID, "error", err)
		return true
	}
	for _, m := range msgs {
		if m.ResourceType == "certificate" && m.ResourceID == c.ID {
			return true
		}
	}
	return false
}

// sendCertReminder 同步发送：本函数运行在后台 goroutine 内，没有请求上下文可继承，
// 同步反而让"发了几条"可测、可观测。
func (s *Server) sendCertReminder(ctx context.Context, c domain.Certificate) error {
	name := certDisplayName(c)
	expire := c.ExpireDate.Format("2006-01-02")
	var content string
	if days := int(time.Until(c.ExpireDate).Hours() / 24); days < 0 {
		content = fmt.Sprintf("您的《%s》已于 %s 过期，认证飞手资格会同时失效，请尽快换证并重新提交审核", name, expire)
	} else {
		content = fmt.Sprintf("您的《%s》将于 %s 到期（还剩 %d 天），请及时办理换证，以免影响认证飞手资格", name, expire, days)
	}
	_, err := s.msgSvc.Send(ctx, "system", c.UserID, "证书到期提醒", content, "certificate", c.ID)
	return err
}

// certDisplayName 证书展示名（httpapi 层自用，service 的同名函数未导出）。
func certDisplayName(c domain.Certificate) string {
	switch string(c.CertType) {
	case string(domain.CertCAAC):
		return "CAAC 无人机驾驶员执照"
	case string(domain.CertUTCDJI):
		return "DJI UTC 植保无人机驾驶证"
	case string(domain.CertGovLevel):
		return "政府职业技能等级证书"
	}
	if c.CertNumber != "" {
		return c.CertNumber
	}
	return "无人机驾驶证书"
}
