package service

import (
	"context"
	"errors"
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
		Status: "completed", Channel: domain.ChannelInternal, CreatedAt: time.Now(),
	}
}

func (s *EscrowService) Deposit(ctx context.Context, userID string, amountFen int64) (domain.EscrowTransaction, error) {
	if amountFen <= 0 {
		return domain.EscrowTransaction{}, fmt.Errorf("amount must be positive")
	}
	tx := newTx("system", userID, "deposit", "", "", amountFen)
	return s.repo.Deposit(ctx, userID, amountFen, tx)
}

// DepositFromChannel 按渠道入账——所有入金的统一入口（内部记账与真实支付共用）。
//
//   - channel 为空按 internal 处理；internal* 为平台内部记账（管理员手工入账 / 自助模拟充值），
//     可以没有外部单号。
//   - 真实资金渠道（wechat，见 domain.IsRealChannel）必须带 externalTxnID：没有外部支付单号
//     就没有真实资金，一律拒绝入账，绝不允许凭空多出一笔钱。
//
// 幂等：真实渠道按 (channel, externalTxnID) 查重，已入账的直接返回原流水且不再加钱
// （微信支付回调会重试，重复入账＝印钞）。第二个返回值 created 表示本次是否真的加了钱。
// 并发场景：两个回调同时通过查重时，库里的唯一索引会挡住第二次写入，此时重查并按幂等成功返回。
func (s *EscrowService) DepositFromChannel(ctx context.Context, userID string, amountFen int64, channel, externalTxnID string) (domain.EscrowTransaction, bool, error) {
	if userID == "" {
		return domain.EscrowTransaction{}, false, errors.New("入账用户不能为空")
	}
	if amountFen <= 0 {
		return domain.EscrowTransaction{}, false, fmt.Errorf("amount must be positive")
	}
	if channel == "" {
		channel = domain.ChannelInternal
	}
	if domain.IsRealChannel(channel) && externalTxnID == "" {
		return domain.EscrowTransaction{}, false, fmt.Errorf("渠道 %s 为真实资金渠道，入账必须提供外部支付单号", channel)
	}
	if externalTxnID != "" {
		if existing, found, err := s.repo.FindByExternalTxn(ctx, channel, externalTxnID); err != nil {
			return domain.EscrowTransaction{}, false, fmt.Errorf("查重外部支付单号: %w", err)
		} else if found {
			return existing, false, nil
		}
	}
	tx := newTx("system", userID, "deposit", "", "", amountFen)
	tx.Channel = channel
	tx.ExternalTxnID = externalTxnID
	saved, err := s.repo.Deposit(ctx, userID, amountFen, tx)
	if err != nil {
		// 并发重入：唯一索引拒绝重复入账时，重查该支付单号并返回已有流水（钱只加了一次）。
		if externalTxnID != "" {
			if existing, found, ferr := s.repo.FindByExternalTxn(ctx, channel, externalTxnID); ferr == nil && found {
				return existing, false, nil
			}
		}
		return domain.EscrowTransaction{}, false, err
	}
	return saved, true, nil
}

// Reconcile 对账：列出 [from, to) 内指定渠道的流水并汇总入金（channel 为空＝不限渠道）。
// limit 为明细条数上限（<=0 用默认 500，最大 5000）。
func (s *EscrowService) Reconcile(ctx context.Context, channel string, from, to time.Time, limit int) (domain.EscrowReconcile, error) {
	txs, err := s.repo.ListByChannel(ctx, channel, from, to, limit)
	if err != nil {
		return domain.EscrowReconcile{}, err
	}
	res := domain.EscrowReconcile{Channel: channel, RealFunds: domain.IsRealChannel(channel), From: from, To: to, Transactions: txs}
	for _, tx := range txs {
		if tx.TxType == "deposit" && tx.Status == "completed" {
			res.DepositCount++
			res.DepositFen += tx.AmountFen
		}
	}
	return res, nil
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
	// 幂等保护：同 (fromUser, refType, refID) 已完成 release 时不再入账——
	// 并发完成报名/重试场景防机构双倍入账（此前只校验余额，付款方还有其它
	// 冻结资金时第二次释放仍会成功）。返回占位流水仅供调用方展示，不落库。
	if has, err := s.repo.HasReleased(ctx, fromUser, refType, refID); err == nil && has {
		return newTx(fromUser, toUser, "release", refType, refID, amountFen), nil
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

// Transfer 双方余额内转账（不涉及冻结）：货款已放给卖家后通过售后退款时，
// 只能从卖家余额扣回买家——余额不足返回 ErrInsufficientBalance，由调用方拒绝本次操作。
func (s *EscrowService) Transfer(ctx context.Context, fromUser, toUser string, amountFen int64, refType, refID string) (domain.EscrowTransaction, error) {
	if amountFen <= 0 {
		return domain.EscrowTransaction{}, fmt.Errorf("amount must be positive")
	}
	tx := newTx(fromUser, toUser, "transfer", refType, refID, amountFen)
	return s.repo.Transfer(ctx, fromUser, toUser, amountFen, tx)
}

// HasFrozen 报告 userID 对 (refType, refID) 是否冻结过资金——
// 订单取消/删除据此判断"这笔业务有没有钱要退"，无冻结则无需退款流水。
func (s *EscrowService) HasFrozen(ctx context.Context, userID, refType, refID string) (bool, error) {
	return s.repo.HasFrozen(ctx, userID, refType, refID)
}

// HasRefunded 报告 userID 对 (refType, refID) 是否已退款——退款幂等用。
func (s *EscrowService) HasRefunded(ctx context.Context, userID, refType, refID string) (bool, error) {
	return s.repo.HasRefunded(ctx, userID, refType, refID)
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
