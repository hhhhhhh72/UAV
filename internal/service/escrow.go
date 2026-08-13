package service

import (
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
		ID: fmt.Sprintf("escrow-%d", time.Now().UnixNano()), FromUser: userID, ToUser: counterparty,
		AmountFen: amountFen, TxType: txType, ReferenceType: refType, ReferenceID: refID,
		Status: "completed", CreatedAt: time.Now(),
	}
}

func (s *EscrowService) Deposit(userID string, amountFen int64) (domain.EscrowTransaction, error) {
	if amountFen <= 0 {
		return domain.EscrowTransaction{}, fmt.Errorf("amount must be positive")
	}
	tx := newTx("system", userID, "deposit", "", "", amountFen)
	return s.repo.Deposit(userID, amountFen, tx)
}

func (s *EscrowService) Freeze(userID string, amountFen int64, refType, refID string) (domain.EscrowTransaction, error) {
	if amountFen <= 0 {
		return domain.EscrowTransaction{}, fmt.Errorf("amount must be positive")
	}
	tx := newTx(userID, "escrow", "freeze", refType, refID, amountFen)
	return s.repo.Freeze(userID, amountFen, tx)
}

func (s *EscrowService) Release(fromUser, toUser string, amountFen int64, refType, refID string) (domain.EscrowTransaction, error) {
	if amountFen <= 0 {
		return domain.EscrowTransaction{}, fmt.Errorf("amount must be positive")
	}
	tx := newTx(fromUser, toUser, "release", refType, refID, amountFen)
	return s.repo.Release(fromUser, toUser, amountFen, tx)
}

func (s *EscrowService) Refund(userID string, amountFen int64, refType, refID string) (domain.EscrowTransaction, error) {
	if amountFen <= 0 {
		return domain.EscrowTransaction{}, fmt.Errorf("amount must be positive")
	}
	tx := newTx("escrow", userID, "refund", refType, refID, amountFen)
	return s.repo.Refund(userID, amountFen, tx)
}

func (s *EscrowService) Balance(userID string) (domain.EscrowAccount, error) {
	return s.repo.GetAccount(userID)
}

func (s *EscrowService) Transactions(userID string) ([]domain.EscrowTransaction, error) {
	return s.repo.ListTransactions(userID)
}
