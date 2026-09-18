package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- 线上支付订单（开发/测试用的内存实现，语义与 PG 版一致）----

type paymentOrderRepo struct {
	mu     sync.RWMutex
	orders map[string]domain.PaymentOrder // key: out_trade_no
}

func NewPaymentOrderRepository() repository.PaymentOrderRepository {
	return &paymentOrderRepo{orders: map[string]domain.PaymentOrder{}}
}

func (r *paymentOrderRepo) Create(ctx context.Context, o domain.PaymentOrder) (domain.PaymentOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.orders[o.OutTradeNo]; ok {
		return domain.PaymentOrder{}, fmt.Errorf("payment order %s already exists", o.OutTradeNo)
	}
	r.orders[o.OutTradeNo] = o
	return o, nil
}

func (r *paymentOrderRepo) FindByOutTradeNo(ctx context.Context, outTradeNo string) (domain.PaymentOrder, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.orders[outTradeNo]
	return o, ok, nil
}

// SetPrepayID 回填预支付会话标识（与 PG 版同语义）。
func (r *paymentOrderRepo) SetPrepayID(ctx context.Context, outTradeNo, prepayID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.orders[outTradeNo]
	if !ok {
		return fmt.Errorf("payment order %s not found", outTradeNo)
	}
	o.PrepayID = prepayID
	o.UpdatedAt = time.Now()
	r.orders[outTradeNo] = o
	return nil
}

// ReserveRefund 原子占用退款额度：仅当 remaining >= amount 时成功。
// 与 PG 版的条件更新同语义（PG: WHERE refunded_fen + $1 <= amount_fen）。
func (r *paymentOrderRepo) ReserveRefund(ctx context.Context, outTradeNo string, amountFen int64) (bool, error) {
	if amountFen <= 0 {
		return false, fmt.Errorf("退款额度必须为正，当前 %d", amountFen)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.orders[outTradeNo]
	if !ok || o.Status != domain.PaymentPaid || o.RefundedFen+amountFen > o.AmountFen {
		return false, nil
	}
	o.RefundedFen += amountFen
	o.UpdatedAt = time.Now()
	r.orders[outTradeNo] = o
	return true, nil
}

// ReleaseRefund 归还已占用的退款额度。
func (r *paymentOrderRepo) ReleaseRefund(ctx context.Context, outTradeNo string, amountFen int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.orders[outTradeNo]
	if !ok || o.RefundedFen < amountFen {
		return nil
	}
	o.RefundedFen -= amountFen
	o.UpdatedAt = time.Now()
	r.orders[outTradeNo] = o
	return nil
}

// ListPaid 列出已支付的充值单。
func (r *paymentOrderRepo) ListPaid(ctx context.Context, limit int) ([]domain.PaymentOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.PaymentOrder
	for _, o := range r.orders {
		if o.Status == domain.PaymentPaid {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PaidAt.After(out[j].PaidAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
// ---- 线上退款单（开发/测试用内存实现，语义与 PG 版一致）----

type paymentRefundRepo struct {
	mu      sync.RWMutex
	refunds map[string]domain.PaymentRefund // key: out_refund_no
}

func NewPaymentRefundRepository() repository.PaymentRefundRepository {
	return &paymentRefundRepo{refunds: map[string]domain.PaymentRefund{}}
}

func (r *paymentRefundRepo) Create(ctx context.Context, rf domain.PaymentRefund) (domain.PaymentRefund, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.refunds[rf.OutRefundNo]; ok {
		return domain.PaymentRefund{}, fmt.Errorf("refund %s already exists", rf.OutRefundNo)
	}
	now := time.Now()
	if rf.CreatedAt.IsZero() {
		rf.CreatedAt = now
	}
	rf.UpdatedAt = now
	r.refunds[rf.OutRefundNo] = rf
	return rf, nil
}

func (r *paymentRefundRepo) FindByOutRefundNo(ctx context.Context, outRefundNo string) (domain.PaymentRefund, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rf, ok := r.refunds[outRefundNo]
	return rf, ok, nil
}

// MarkStatus 条件更新：仅当前状态在 from 列表中才推进。
// 与 PG 版同款 CAS（WHERE status = ANY(...)），内存实现也照搬以免测试放过并发缺陷。
func (r *paymentRefundRepo) MarkStatus(ctx context.Context, outRefundNo, from, to, refundID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rf, ok := r.refunds[outRefundNo]
	if !ok {
		return false, nil
	}
	allowed := false
	for _, s := range strings.Split(from, ",") {
		if rf.Status == strings.TrimSpace(s) {
			allowed = true
			break
		}
	}
	if !allowed {
		return false, nil
	}
	rf.Status = to
	if refundID != "" {
		rf.RefundID = refundID
	}
	rf.UpdatedAt = time.Now()
	r.refunds[outRefundNo] = rf
	return true, nil
}

func (r *paymentRefundRepo) ListByOrder(ctx context.Context, outTradeNo string) ([]domain.PaymentRefund, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.PaymentRefund
	for _, rf := range r.refunds {
		if rf.OutTradeNo == outTradeNo {
			out = append(out, rf)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *paymentRefundRepo) ListRecent(ctx context.Context, limit int) ([]domain.PaymentRefund, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.PaymentRefund
	for _, rf := range r.refunds {
		out = append(out, rf)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// MarkPaid 条件更新：只有 created 能变 paid，重放/并发只有一个返回 true。
// 与 PG 版同款 CAS（WHERE status='created'），内存实现也照搬以免测试放过并发缺陷。
func (r *paymentOrderRepo) MarkPaid(ctx context.Context, outTradeNo, transactionID string, paidAt time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.orders[outTradeNo]
	if !ok || o.Status != domain.PaymentCreated {
		return false, nil
	}
	o.Status = domain.PaymentPaid
	o.TransactionID = transactionID
	o.PaidAt = paidAt
	o.UpdatedAt = time.Now()
	r.orders[outTradeNo] = o
	return true, nil
}

func (r *paymentOrderRepo) ListByUser(ctx context.Context, userID string, limit int) ([]domain.PaymentOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.PaymentOrder
	for _, o := range r.orders {
		if o.UserID == userID {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
