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

type paymentRefundRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewPaymentRefundRepository() repository.PaymentRefundRepository {
	return &paymentRefundRepo{pool: s.Pool()}
}

// 列顺序与 scanPaymentRefund 严格对齐（该文件上方的 payment_repo.go 有同款约定：
// 漏列会让字段静默丢失）。
const paymentRefundColumns = `id, out_refund_no, out_trade_no, user_id, amount_fen, ` +
	`reason, status, refund_id, created_at, updated_at`

func scanPaymentRefund(row pgx.Row) (domain.PaymentRefund, error) {
	var rf domain.PaymentRefund
	err := row.Scan(&rf.ID, &rf.OutRefundNo, &rf.OutTradeNo, &rf.UserID, &rf.AmountFen,
		&rf.Reason, &rf.Status, &rf.RefundID, &rf.CreatedAt, &rf.UpdatedAt)
	if err != nil {
		return domain.PaymentRefund{}, err
	}
	return rf, nil
}

func (r *paymentRefundRepo) Create(ctx context.Context, rf domain.PaymentRefund) (domain.PaymentRefund, error) {
	now := time.Now()
	if rf.CreatedAt.IsZero() {
		rf.CreatedAt = now
	}
	rf.UpdatedAt = now
	_, err := r.pool.Exec(ctx,
		`INSERT INTO payment_refunds (id, out_refund_no, out_trade_no, user_id, amount_fen, reason, status, refund_id, created_at, updated_at)`+
			` VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		rf.ID, rf.OutRefundNo, rf.OutTradeNo, rf.UserID, rf.AmountFen, rf.Reason, rf.Status, rf.RefundID, rf.CreatedAt, rf.UpdatedAt)
	if err != nil {
		return domain.PaymentRefund{}, fmt.Errorf("create payment refund %s: %w", rf.OutRefundNo, err)
	}
	return rf, nil
}

func (r *paymentRefundRepo) FindByOutRefundNo(ctx context.Context, outRefundNo string) (domain.PaymentRefund, bool, error) {
	rf, err := scanPaymentRefund(r.pool.QueryRow(ctx,
		`SELECT `+paymentRefundColumns+` FROM payment_refunds WHERE out_refund_no=$1`, outRefundNo))
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.PaymentRefund{}, false, nil
		}
		return domain.PaymentRefund{}, false, fmt.Errorf("find payment refund %s: %w", outRefundNo, err)
	}
	return rf, true, nil
}

// MarkStatus 原子推进状态：只有当前状态在 from 里（逗号分隔）才成功。
// 退款回调会重试，这是防重复处理的锚点——与 PaymentOrder.MarkPaid 同一手法。
// 用 `status = ANY(string_to_array($2, ','))` 而不是拼 SQL，避免把状态名拼进语句。
func (r *paymentRefundRepo) MarkStatus(ctx context.Context, outRefundNo, from, to, refundID string) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE payment_refunds`+
			` SET status=$1, refund_id=CASE WHEN $2 = '' THEN refund_id ELSE $2 END, updated_at=$3`+
			` WHERE out_refund_no=$4 AND status = ANY(string_to_array($5, ','))`,
		to, refundID, time.Now(), outRefundNo, from)
	if err != nil {
		return false, fmt.Errorf("mark payment refund %s as %s: %w", outRefundNo, to, err)
	}
	return tag.RowsAffected() > 0, nil
}

func (r *paymentRefundRepo) ListByOrder(ctx context.Context, outTradeNo string) ([]domain.PaymentRefund, error) {
	return r.list(ctx,
		`SELECT `+paymentRefundColumns+` FROM payment_refunds WHERE out_trade_no=$1 ORDER BY created_at DESC LIMIT 200`,
		outTradeNo)
}

func (r *paymentRefundRepo) ListRecent(ctx context.Context, limit int) ([]domain.PaymentRefund, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return r.list(ctx,
		`SELECT `+paymentRefundColumns+` FROM payment_refunds ORDER BY created_at DESC LIMIT $1`,
		limit)
}

func (r *paymentRefundRepo) list(ctx context.Context, sql string, arg any) ([]domain.PaymentRefund, error) {
	rows, err := r.pool.Query(ctx, sql, arg)
	if err != nil {
		return nil, fmt.Errorf("list payment refunds: %w", err)
	}
	defer rows.Close()
	var out []domain.PaymentRefund
	for rows.Next() {
		rf, err := scanPaymentRefund(rows)
		if err != nil {
			return nil, fmt.Errorf("scan payment refund: %w", err)
		}
		out = append(out, rf)
	}
	return out, rows.Err()
}