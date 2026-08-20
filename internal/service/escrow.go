package service

import (
	"context"
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type EscrowService struct {
	repo repository.EscrowRepository
}

func NewEscrowService(repo repository.EscrowRepository) *EscrowService {
	return &EscrowService{repo: repo}
}

// newTx 构造一条资金流水（状态恒为 completed，写入与余额调整同事务原子提交）。
func newTx(userID, counterparty, txType, refType, refID string, amountFen int64) domain.EscrowTransaction {
	return domain.EscrowTransaction{
		ID: nextID("escrow"), FromUser: userID, ToUser: counterparty,
		AmountFen: amountFen, TxType: txType, ReferenceType: refType, ReferenceID: refID,
		Status: "completed", CreatedAt: time.Now(),
	}
}

func (s *EscrowService) Deposit(ctx context.Context, userID string, amountFen int64) (domain.EscrowTransaction, error) {
	if amountFen <= 0 {
		return domain.EscrowTransaction{}, fmt.Errorf("amount must be positive")
	}
	tx := newTx("system", userID, "deposit", "", "", amountFen)
	return s.repo.Deposit(ctx, userID, amountFen, tx)
}

func (s *EscrowService) Freeze(ctx context.Context, userID string, amountFen int64, refType, refID string) (domain.EscrowTransaction, error) {
	if amountFen <= 0 {
		return domain.EscrowTransaction{}, fmt.Errorf("amount must be positive")
	}
	tx := newTx(userID, "escrow", "freeze", refType, refID, amountFen)
	return s.repo.Freeze(ctx, userID, amountFen, tx)
}

func (s *EscrowService) Release(ctx context.Context, fromUser, toUser string, amountFen int64, refType, refID string) (domain.EscrowTransaction, error) {
	if amountFen <= 0 {
		return domain.EscrowTransaction{}, fmt.Errorf("amount must be positive")
	}
	tx := newTx(fromUser, toUser, "release", refType, refID, amountFen)
	return s.repo.Release(ctx, fromUser, toUser, amountFen, tx)
}

func (s *EscrowService) Refund(ctx context.Context, userID string, amountFen int64, refType, refID string) (domain.EscrowTransaction, error) {
	if amountFen <= 0 {
		return domain.EscrowTransaction{}, fmt.Errorf("amount must be positive")
	}
	tx := newTx("escrow", userID, "refund", refType, refID, amountFen)
	return s.repo.Refund(ctx, userID, amountFen, tx)
}

func (s *EscrowService) Balance(ctx context.Context, userID string) (domain.EscrowAccount, error) {
	return s.repo.GetAccount(ctx, userID)
}

func (s *EscrowService) Transactions(ctx context.Context, userID string) ([]domain.EscrowTransaction, error) {
	return s.repo.ListTransactions(ctx, userID)
}

// HasReleased 报告 userID 对 (refType, refID) 是否已有完成的 release 流水。
// completeEnrollment 幂等重试用：completed 报名重试时先查此判定"学费是否已释放"，
// 已释放则跳过 Release（防重复释放），未释放则补齐。
func (s *EscrowService) HasReleased(ctx context.Context, userID, refType, refID string) (bool, error) {
	return s.repo.HasReleased(ctx, userID, refType, refID)
}

// RefundOrphanFreezes 自动补偿：找出"冻结了但业务记录不存在"的孤儿冻结并退回余额。
// 场景：payAndEnroll 先冻结后报名，进程在两步之间崩溃 → 资金滞留 frozen。
// olderThan 过滤最近刚冻结的正常窗口（避免误伤刚发起、报名尚未落库的请求）；
// 每次最多处理 limit 条，返回处理条数。
func (s *EscrowService) RefundOrphanFreezes(ctx context.Context, refType string, olderThan time.Time, limit int) (int, error) {
	orphans, err := s.repo.ListOrphanFreezes(ctx, refType, olderThan, limit)
	if err != nil {
		return 0, err
	}
	refunded := 0
	for _, tx := range orphans {
		if tx.AmountFen <= 0 {
			continue
		}
		if _, err := s.repo.Refund(ctx, tx.FromUser, tx.AmountFen, newTx("escrow", tx.FromUser, "refund", tx.ReferenceType, tx.ReferenceID, tx.AmountFen)); err != nil {
			// 单条失败不阻断其余（余额可能已变动）；记录后继续
			continue
		}
		refunded++
	}
	return refunded, nil
}
