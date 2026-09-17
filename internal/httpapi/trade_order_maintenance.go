package httpapi

import (
	"context"
	"log/slog"
	"time"
)

// ── 商城订单：下单限频 ──
//
// 下单即把商品置 sold（支付前占位，防一物多卖）。此前没有任何频次约束，
// 任意登录用户都能对他人商品批量下单不付款，把供给大厅的商品全部锁死（拒绝供给）。
// 按"买家"维度限频：同一个人换 IP 不该绕过，共用出口 IP 的用户也不该互相拖累。
const (
	// orderCreateWindow 下单限频窗口。
	orderCreateWindow = 20 * time.Minute
	// orderCreateMax 窗口内同一买家最多成功下单数。
	orderCreateMax = 10
	// orderLimitMaxEntries 限频表条目上限（防内存 DoS）：超限清空重建（与注册限频同款粗糙策略）。
	orderLimitMaxEntries = 10000
)

// orderCreateAllowed 报告该买家在 orderCreateWindow 内是否还能下单（含本次），并累计本次。
func (s *Server) orderCreateAllowed(userID string) bool {
	if s.orderLimitEntries.Load() >= orderLimitMaxEntries {
		s.orderLimits.Range(func(k, _ any) bool { s.orderLimits.Delete(k); return true })
		s.orderLimitEntries.Store(0)
	}
	s.orderLimitEntries.Add(1)
	v, _ := s.orderLimits.LoadOrStore(userID, &regLimitLog{windowStart: time.Now()})
	log := v.(*regLimitLog)
	log.mu.Lock()
	defer log.mu.Unlock()
	now := time.Now()
	if now.Sub(log.windowStart) >= orderCreateWindow {
		log.windowStart = now
		log.count = 0
	}
	log.count++
	return log.count <= orderCreateMax
}

// ── 商城订单：超时维护后台任务 ──

const (
	// orderUnpaidTimeout 支付超时：下单后这么久仍未付款即自动取消并恢复商品为可售。
	orderUnpaidTimeout = 20 * time.Minute
	// orderAutoConfirmAfter 自动确认收货：发货后这么久买家仍未确认，视为已收货并放款给卖家。
	// 没有它，买家不点"确认收货"，货款就永久停在冻结里（孤儿冻结补偿按"业务行不存在"
	// 判定，救不了这类订单行存在的场景）。
	orderAutoConfirmAfter = 7 * 24 * time.Hour
	// orderSweepInterval 扫描周期：超时判定是"分钟"级需求，5 分钟一扫兼顾及时与开销。
	orderSweepInterval = 5 * time.Minute
	// orderSweepFirstDelay 启动后首次扫描延迟：避开启动高峰，也让本地开发很快看到效果。
	orderSweepFirstDelay = 2 * time.Minute
	// orderSweepBatch 单轮每类订单的处理上限（分批，避免一次拉太多）。
	orderSweepBatch = 100
	// orphanSoldGrace 孤儿已售商品的宽限期：仅回收"标记为已售且超过这么久没有变动"的商品。
	// 宽限是为了不与「下单占位 → 创建订单」的正常时序竞争（那只有毫秒级），
	// 也避免把"已成交但订单记录被人工删除"的商品立刻放回货架。
	orphanSoldGrace = 24 * time.Hour
)

// StartTradeOrderMaintenance 启动商城订单的超时维护任务：
//   - 未付款超时 → 自动取消 + 恢复商品
//   - 已发货超时 → 自动确认收货 + 放款给卖家
//
// 与证书到期提醒、孤儿冻结补偿同一套做法：无退出通道的后台 goroutine + recover 兜底。
// 幂等性由"CAS 状态迁移 + settleMoney 内部查重"保证，进程重启/重复扫描都不会重复动钱。
func (s *Server) StartTradeOrderMaintenance() {
	if s == nil || s.tradeSvc == nil {
		return
	}
	go func() {
		defer func() { _ = recover() }()
		time.Sleep(orderSweepFirstDelay)
		for {
			s.sweepTradeOrders()
			time.Sleep(orderSweepInterval)
		}
	}()
}

// sweepTradeOrders 扫一轮：先关超时未付款单，再验收超时未确认单。
func (s *Server) sweepTradeOrders() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if n, err := s.tradeSvc.AutoCancelUnpaid(ctx, time.Now().Add(-orderUnpaidTimeout), orderSweepBatch); err != nil {
		slog.Warn("trade order sweep: auto cancel unpaid failed", "error", err)
	} else if n > 0 {
		slog.Info("trade order sweep: unpaid orders auto-cancelled", "count", n)
	}

	if n, err := s.tradeSvc.AutoConfirmShipped(ctx, time.Now().Add(-orderAutoConfirmAfter), orderSweepBatch); err != nil {
		slog.Warn("trade order sweep: auto confirm shipped failed", "error", err)
	} else if n > 0 {
		slog.Info("trade order sweep: shipped orders auto-confirmed", "count", n)
	}

	// 补偿③：孤儿已售商品回收——商品被下单占位后订单没落库/被删，会永久下架且无人察觉。
	if n, err := s.tradeSvc.RelistOrphanSoldProducts(ctx, time.Now().Add(-orphanSoldGrace), orderSweepBatch); err != nil {
		slog.Warn("trade order sweep: relist orphan sold failed", "error", err)
	} else if n > 0 {
		slog.Info("trade order sweep: orphan sold products relisted", "count", n)
	}
}
