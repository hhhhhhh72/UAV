package memory

import (
	"context"
	"fmt"
	"sort"
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
