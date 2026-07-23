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

func (s *EscrowService) Deposit(userID string, amountFen int64) (domain.EscrowTransaction, error) {
	acct, _ := s.repo.GetAccount(userID)
	acct.BalanceFen += amountFen
	acct.UpdatedAt = time.Now()
	if err := s.repo.UpsertAccount(acct); err != nil {
		return domain.EscrowTransaction{}, err
	}
	tx := domain.EscrowTransaction{
		ID: fmt.Sprintf("escrow-%d", time.Now().UnixNano()), FromUser: "system",
		ToUser: userID, AmountFen: amountFen, TxType: "deposit", Status: "completed", CreatedAt: time.Now(),
	}
	return s.repo.CreateTransaction(tx)
}

func (s *EscrowService) Freeze(userID string, amountFen int64, refType, refID string) (domain.EscrowTransaction, error) {
	acct, _ := s.repo.GetAccount(userID)
	if acct.BalanceFen < amountFen {
		return domain.EscrowTransaction{}, fmt.Errorf("insufficient balance: have %d, need %d", acct.BalanceFen, amountFen)
	}
	acct.BalanceFen -= amountFen
	acct.FrozenFen += amountFen
	acct.UpdatedAt = time.Now()
	if err := s.repo.UpsertAccount(acct); err != nil {
		return domain.EscrowTransaction{}, err
	}
	tx := domain.EscrowTransaction{
		ID: fmt.Sprintf("escrow-%d", time.Now().UnixNano()), FromUser: userID, ToUser: "escrow",
		AmountFen: amountFen, TxType: "freeze", ReferenceType: refType, ReferenceID: refID,
		Status: "completed", CreatedAt: time.Now(),
	}
	return s.repo.CreateTransaction(tx)
}

func (s *EscrowService) Release(fromUser, toUser string, amountFen int64, refType, refID string) (domain.EscrowTransaction, error) {
	fromAcct, _ := s.repo.GetAccount(fromUser)
	if fromAcct.FrozenFen < amountFen {
		return domain.EscrowTransaction{}, fmt.Errorf("insufficient frozen balance")
	}
	toAcct, _ := s.repo.GetAccount(toUser)
	fromAcct.FrozenFen -= amountFen
	fromAcct.UpdatedAt = time.Now()
	toAcct.BalanceFen += amountFen
	toAcct.UpdatedAt = time.Now()
	if err := s.repo.UpsertAccount(fromAcct); err != nil {
		return domain.EscrowTransaction{}, err
	}
	if err := s.repo.UpsertAccount(toAcct); err != nil {
		return domain.EscrowTransaction{}, err
	}
	tx := domain.EscrowTransaction{
		ID: fmt.Sprintf("escrow-%d", time.Now().UnixNano()), FromUser: fromUser, ToUser: toUser,
		AmountFen: amountFen, TxType: "release", ReferenceType: refType, ReferenceID: refID,
		Status: "completed", CreatedAt: time.Now(),
	}
	return s.repo.CreateTransaction(tx)
}

func (s *EscrowService) Refund(userID string, amountFen int64, refType, refID string) (domain.EscrowTransaction, error) {
	acct, _ := s.repo.GetAccount(userID)
	if acct.FrozenFen < amountFen {
		return domain.EscrowTransaction{}, fmt.Errorf("insufficient frozen balance")
	}
	acct.FrozenFen -= amountFen
	acct.BalanceFen += amountFen
	acct.UpdatedAt = time.Now()
	if err := s.repo.UpsertAccount(acct); err != nil {
		return domain.EscrowTransaction{}, err
	}
	tx := domain.EscrowTransaction{
		ID: fmt.Sprintf("escrow-%d", time.Now().UnixNano()), FromUser: "escrow", ToUser: userID,
		AmountFen: amountFen, TxType: "refund", ReferenceType: refType, ReferenceID: refID,
		Status: "completed", CreatedAt: time.Now(),
	}
	return s.repo.CreateTransaction(tx)
}

func (s *EscrowService) Balance(userID string) (domain.EscrowAccount, error) {
	return s.repo.GetAccount(userID)
}

func (s *EscrowService) Transactions(userID string) ([]domain.EscrowTransaction, error) {
	return s.repo.ListTransactions(userID)
}
