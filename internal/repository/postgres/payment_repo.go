package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type paymentOrderRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewPaymentOrderRepository() repository.PaymentOrderRepository {
	return &paymentOrderRepo{pool: s.Pool()}
}

// 列顺序与 scanPaymentOrder 严格对齐（此前多处因漏列导致字段静默丢失）。
const paymentOrderColumns = `id, out_trade_no, user_id, amount_fen, channel, status, ` +
	`prepay_id, transaction_id, refunded_fen, paid_at, created_at, updated_at`

// scanPaymentOrder 读一行。paid_at 可空（未支付），故用指针承接再回填。
func scanPaymentOrder(row pgx.Row) (domain.PaymentOrder, error) {
	var o domain.PaymentOrder
	var paidAt *time.Time
	err := row.Scan(&o.ID, &o.OutTradeNo, &o.UserID, &o.AmountFen, &o.Channel, &o.Status,
		&o.PrepayID, &o.TransactionID, &o.RefundedFen, &paidAt, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return domain.PaymentOrder{}, err
	}
	if paidAt != nil {
		o.PaidAt = *paidAt
	}
	return o, nil
}

func (r *paymentOrderRepo) Create(ctx context.Context, o domain.PaymentOrder) (domain.PaymentOrder, error) {
	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}
	o.UpdatedAt = now
	_, err := r.pool.Exec(ctx,
		`INSERT INTO payment_orders (id, out_trade_no, user_id, amount_fen, channel, status, prepay_id, transaction_id, created_at, updated_at)`+
			` VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		o.ID, o.OutTradeNo, o.UserID, o.AmountFen, o.Channel, o.Status, o.PrepayID, o.TransactionID, o.CreatedAt, o.UpdatedAt)
	if err != nil {
		return domain.PaymentOrder{}, fmt.Errorf("create payment order %s: %w", o.OutTradeNo, err)
	}
	return o, nil
}

func (r *paymentOrderRepo) FindByOutTradeNo(ctx context.Context, outTradeNo string) (domain.PaymentOrder, bool, error) {
	o, err := scanPaymentOrder(r.pool.QueryRow(ctx,
		`SELECT `+paymentOrderColumns+` FROM payment_orders WHERE out_trade_no=$1`, outTradeNo))
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.PaymentOrder{}, false, nil
		}
		return domain.PaymentOrder{}, false, fmt.Errorf("find payment order %s: %w", outTradeNo, err)
	}
	return o, true, nil
}

// SetPrepayID 回填预支付会话标识。写失败不影响资金安全（它只是前端调起支付的凭据），
// 但调用方仍应记日志——否则线上出现「下单成功、前端调不起来」时无从查起。
func (r *paymentOrderRepo) SetPrepayID(ctx context.Context, outTradeNo, prepayID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payment_orders SET prepay_id=$1, updated_at=$2 WHERE out_trade_no=$3`,
		prepayID, time.Now(), outTradeNo)
	if err != nil {
		return fmt.Errorf("set prepay id for %s: %w", outTradeNo, err)
	}
	return nil
}

// MarkPaid 原子置 paid：WHERE status='created' 是防重复入账的第一道防线。
// 并发回调/微信重试/人为重放都只有一个能把 created 改成 paid，其余 RowsAffected=0。
// transaction_id 上另有唯一索引兜底（同一微信支付单号不可能挂到两个订单）。
func (r *paymentOrderRepo) MarkPaid(ctx context.Context, outTradeNo, transactionID string, paidAt time.Time) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE payment_orders SET status=$1, transaction_id=$2, paid_at=$3, updated_at=$4`+
			` WHERE out_trade_no=$5 AND status=$6`,
		domain.PaymentPaid, transactionID, paidAt, time.Now(), outTradeNo, domain.PaymentCreated)
	if err != nil {
		return false, fmt.Errorf("mark payment order %s paid: %w", outTradeNo, err)
	}
	return tag.RowsAffected() > 0, nil
}

func (r *paymentOrderRepo) ListByUser(ctx context.Context, userID string, limit int) ([]domain.PaymentOrder, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+paymentOrderColumns+` FROM payment_orders WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list payment orders for %s: %w", userID, err)
	}
	defer rows.Close()
	var out []domain.PaymentOrder
	for rows.Next() {
		o, err := scanPaymentOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan payment order: %w", err)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}
// ReserveRefund 原子占用退款额度。「累计退款不超过原单金额」由这条条件更新保证——
// 微信支持多次部分退款，应用层先查后写在并发下必然穿透。
func (r *paymentOrderRepo) ReserveRefund(ctx context.Context, outTradeNo string, amountFen int64) (bool, error) {
	if amountFen <= 0 {
		return false, fmt.Errorf("退款额度必须为正，当前 %d", amountFen)
	}
	tag, err := r.pool.Exec(ctx,
		`UPDATE payment_orders SET refunded_fen = refunded_fen + $1, updated_at = $2`+
			` WHERE out_trade_no = $3 AND status = $4 AND refunded_fen + $1 <= amount_fen`,
		amountFen, time.Now(), outTradeNo, domain.PaymentPaid)
	if err != nil {
		return false, fmt.Errorf("reserve refund for %s: %w", outTradeNo, err)
	}
	return tag.RowsAffected() > 0, nil
}

// ReleaseRefund 归还已占用的退款额度（微信侧明确拒绝退款时回退）。
// 条件 refunded_fen >= amount 防回退成负数。
func (r *paymentOrderRepo) ReleaseRefund(ctx context.Context, outTradeNo string, amountFen int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payment_orders SET refunded_fen = refunded_fen - $1, updated_at = $2`+
			` WHERE out_trade_no = $3 AND refunded_fen >= $1`,
		amountFen, time.Now(), outTradeNo)
	if err != nil {
		return fmt.Errorf("release refund for %s: %w", outTradeNo, err)
	}
	return nil
}

// ListPaid 列出已支付的充值单（管理端发起退款时选单用），按支付时间倒序。
func (r *paymentOrderRepo) ListPaid(ctx context.Context, limit int) ([]domain.PaymentOrder, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+paymentOrderColumns+` FROM payment_orders WHERE status = $1`+
			` ORDER BY paid_at DESC NULLS LAST, created_at DESC LIMIT $2`, domain.PaymentPaid, limit)
	if err != nil {
		return nil, fmt.Errorf("list paid payment orders: %w", err)
	}
	defer rows.Close()
	var out []domain.PaymentOrder
	for rows.Next() {
		o, err := scanPaymentOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan payment order: %w", err)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}
