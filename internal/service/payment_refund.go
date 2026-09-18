package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// refundRefType 退款单在托管金流水里的业务类型前缀。
// 用 (payment_refund, out_refund_no) 作为冻结/解冻/出账三条流水的业务单号，
// 天然与订单/课程那些 key 隔离，也不会互相压到唯一索引。
const refundRefType = "payment_refund"

// 微信退款状态（与 internal/wechatpay 的常量同名同值，但 Service 不依赖那个包）。
const (
	RefundStateSuccess    = "SUCCESS"
	RefundStateProcessing = "PROCESSING"
	RefundStateClosed     = "CLOSED"
	RefundStateAbnormal   = "ABNORMAL"
)

// RefundRequest 发起退款的入参。
type RefundRequest struct {
	OutTradeNo  string
	OutRefundNo string
	RefundFen   int64
	TotalFen    int64
	Reason      string
}

// RefundResult 退款单在微信侧的现状。
type RefundResult struct {
	RefundID  string
	Status    string
	RefundFen int64
}

// RefundRejectedError 网关**明确拒绝**了这笔退款（微信返回了业务错误码）。
//
// 与「网络超时/结果未知」必须区分开：前者可以安全回滚（解冻余额 + 归还退款额度），
// 后者**绝不能**回滚——微信可能已经受理了，回滚等于钱退出去而平台把余额又还给了用户。
type RefundRejectedError struct{ Reason string }

func (e *RefundRejectedError) Error() string { return e.Reason }

// RefundGateway 退款网关（出站端口），由 internal/wechatpay 适配实现。
type RefundGateway interface {
	CreateRefund(ctx context.Context, in RefundRequest) (RefundResult, error)
	QueryRefund(ctx context.Context, outRefundNo string) (RefundResult, error)
	// DecodeRefundNotify 解析退款回调原文，只返回退款单号。
	// 与支付回调同一约定：**回调只做触发**，金额与状态一律以主动查单为准。
	DecodeRefundNotify(body []byte) (outRefundNo string, err error)
}

// PaymentRefundService 线上退款：针对**某一笔已支付的充值单**发起，支持多次部分退款。
//
// 资金三件套（全部复用已有的 escrow 词汇，失败路径都是现成的）：
//   发起   → EscrowService.Freeze   ：余额→冻结。钱先锁住，退款在途时用户花不掉、
//                                     余额不足则**发起就失败**（不会出现"退了但扣不到"）
//   成功   → EscrowService.Withdraw ：冻结扣掉，钱真正离开平台
//   失败/关闭 → EscrowService.Refund ：冻结退回余额
//
// 三道幂等：
//   1. 退款额度由 payment_orders.refunded_fen 的条件更新保证（累计不超原单）
//   2. 状态推进是 CAS（MarkStatus WHERE status = ANY(...)），回调重试只处理一次
//   3. 出账流水由库级唯一索引 idx_escrow_once_per_ref 兜底（000119 起覆盖 withdraw）
type PaymentRefundService struct {
	orders  repository.PaymentOrderRepository
	refunds repository.PaymentRefundRepository
	escrow  *EscrowService
	gw      RefundGateway
	now     func() time.Time
}

func NewPaymentRefundService(orders repository.PaymentOrderRepository, refunds repository.PaymentRefundRepository, escrow *EscrowService, gw RefundGateway) *PaymentRefundService {
	return &PaymentRefundService{orders: orders, refunds: refunds, escrow: escrow, gw: gw, now: time.Now}
}

// Enabled 退款能力是否可用（网关已装配）。
func (s *PaymentRefundService) Enabled() bool {
	return s != nil && s.gw != nil && s.orders != nil && s.refunds != nil && s.escrow != nil
}

// NewOutRefundNo 生成商户退款单号：RF + yyyymmddHHMMSS + 8 位随机 hex（与订单号同格式）。
func NewOutRefundNo(now time.Time) (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成退款单号失败: %w", err)
	}
	return "RF" + now.Format("20060102150405") + hex.EncodeToString(b), nil
}

// Initiate 发起一笔退款（仅管理员入口，见 httpapi）。
func (s *PaymentRefundService) Initiate(ctx context.Context, outTradeNo string, amountFen int64, reason, operatorID string) (domain.PaymentRefund, error) {
	if !s.Enabled() {
		return domain.PaymentRefund{}, ErrPaymentDisabled
	}
	if amountFen <= 0 {
		return domain.PaymentRefund{}, errors.New("退款金额必须为正")
	}
	order, found, err := s.orders.FindByOutTradeNo(ctx, strings.TrimSpace(outTradeNo))
	if err != nil {
		return domain.PaymentRefund{}, fmt.Errorf("查充值单失败: %w", err)
	}
	if !found {
		return domain.PaymentRefund{}, fmt.Errorf("充值单不存在: %s", outTradeNo)
	}
	if order.Status != domain.PaymentPaid || order.TransactionID == "" {
		return domain.PaymentRefund{}, errors.New("只有已支付（且带微信支付单号）的充值单可以退款")
	}
	outRefundNo, err := NewOutRefundNo(s.now())
	if err != nil {
		return domain.PaymentRefund{}, err
	}

	// ① 占用退款额度：库级条件更新，累计不超过原单金额。
	ok, err := s.orders.ReserveRefund(ctx, order.OutTradeNo, amountFen)
	if err != nil {
		return domain.PaymentRefund{}, fmt.Errorf("占用退款额度失败: %w", err)
	}
	if !ok {
		return domain.PaymentRefund{}, fmt.Errorf("累计退款额将超过原单金额（已退 %d 分 / 原单 %d 分）", order.RefundedFen, order.AmountFen)
	}

	// ② 冻结用户托管余额：钱先锁住。余额不足则本次退款**发起就失败**，
	//    不会出现"微信退了但平台扣不到"的悬空状态。
	if _, err := s.escrow.Freeze(ctx, order.UserID, amountFen, refundRefType, outRefundNo); err != nil {
		if rerr := s.orders.ReleaseRefund(ctx, order.OutTradeNo, amountFen); rerr != nil {
			return domain.PaymentRefund{}, fmt.Errorf("冻结失败(%v) 且归还退款额度也失败(%v)，需人工核对", err, rerr)
		}
		return domain.PaymentRefund{}, fmt.Errorf("冻结用户余额失败（余额不足则无法退款）: %w", err)
	}

	// ③ 落退款单（created）。
	rf := domain.PaymentRefund{
		ID: nextID("refund"), OutRefundNo: outRefundNo, OutTradeNo: order.OutTradeNo,
		UserID: order.UserID, AmountFen: amountFen, Reason: strings.TrimSpace(reason),
		Status: domain.PaymentRefundCreated, CreatedAt: s.now(),
	}
	if _, err := s.refunds.Create(ctx, rf); err != nil {
		_ = s.rollbackInitiate(ctx, rf)
		return domain.PaymentRefund{}, fmt.Errorf("创建退款单失败: %w", err)
	}

	// ④ 通知微信。
	res, err := s.gw.CreateRefund(ctx, RefundRequest{
		OutTradeNo: order.OutTradeNo, OutRefundNo: outRefundNo,
		RefundFen: amountFen, TotalFen: order.AmountFen, Reason: rf.Reason,
	})
	if err != nil {
		var rejected *RefundRejectedError
		if errors.As(err, &rejected) {
			// 微信明确拒绝：可以安全回滚（解冻 + 归还额度）。
			_ = s.rollbackInitiate(ctx, rf)
			_, _ = s.refunds.MarkStatus(ctx, outRefundNo, domain.PaymentRefundCreated, domain.PaymentRefundClosed, "")
			rf.Status = domain.PaymentRefundClosed
			return rf, fmt.Errorf("微信拒绝了这笔退款: %s", rejected.Reason)
		}
		// 结果未知（超时/网络中断）：**绝不回滚**——微信可能已经受理。
		// 冻结与额度都保留，等退款回调或人工查单收口（Confirm 是幂等的）。
		return rf, fmt.Errorf("退款结果未知（已提交微信，请勿重复发起；在退款列表中查询进度）: %w", err)
	}
	_, _ = s.refunds.MarkStatus(ctx, outRefundNo, domain.PaymentRefundCreated, domain.PaymentRefundProcessing, res.RefundID)
	rf.Status = domain.PaymentRefundProcessing
	rf.RefundID = res.RefundID
	return rf, nil
}

// rollbackInitiate 发起阶段的回滚：解冻余额 + 归还退款额度（两步都做，各自失败都要暴露）。
func (s *PaymentRefundService) rollbackInitiate(ctx context.Context, rf domain.PaymentRefund) error {
	_, ferr := s.escrow.Refund(ctx, rf.UserID, rf.AmountFen, refundRefType, rf.OutRefundNo)
	rerr := s.orders.ReleaseRefund(ctx, rf.OutTradeNo, rf.AmountFen)
	if ferr != nil || rerr != nil {
		return fmt.Errorf("回滚退款发起失败: 解冻=%v 归还额度=%v", ferr, rerr)
	}
	return nil
}

// NotifyRef 从退款回调原文里取出退款单号——只取这一个字段，其余一概不采信。
func (s *PaymentRefundService) NotifyRef(body []byte) (string, error) {
	if !s.Enabled() {
		return "", ErrPaymentDisabled
	}
	if len(body) == 0 {
		return "", errors.New("通知报文为空")
	}
	return s.gw.DecodeRefundNotify(body)
}

// Confirm 处理一次「退款可能已结清」的通知（微信回调 / 人工查单）。
//
// 与支付回调同一模型：**通知只做触发，一律主动查单**——回调报文不构成到账/退款依据。
//
// 顺序刻意是「先出账、后置成功」而不是反过来：
//   Withdraw 由库级唯一索引保证幂等，MarkStatus 失败后下一次 Confirm 会重查到 SUCCESS、
//   再调一次 Withdraw（无副作用）后终于置成功——**自愈**。
//   反过来先置成功再出账，一旦出账失败就再也不会重试了（Confirm 见到 success 直接返回）。
func (s *PaymentRefundService) Confirm(ctx context.Context, outRefundNo string) (domain.PaymentRefund, bool, error) {
	if !s.Enabled() {
		return domain.PaymentRefund{}, false, ErrPaymentDisabled
	}
	outRefundNo = strings.TrimSpace(outRefundNo)
	if outRefundNo == "" {
		return domain.PaymentRefund{}, false, errors.New("退款单号为空")
	}
	rf, found, err := s.refunds.FindByOutRefundNo(ctx, outRefundNo)
	if err != nil {
		return domain.PaymentRefund{}, false, fmt.Errorf("查本地退款单失败: %w", err)
	}
	if !found {
		return domain.PaymentRefund{}, false, fmt.Errorf("退款单不存在: %s", outRefundNo)
	}
	if rf.Status == domain.PaymentRefundSuccess {
		return rf, false, nil // 已结清，幂等返回
	}

	res, err := s.gw.QueryRefund(ctx, outRefundNo)
	if err != nil {
		return rf, false, fmt.Errorf("查微信退款单失败: %w", err)
	}
	const openStatuses = domain.PaymentRefundCreated + "," + domain.PaymentRefundProcessing + "," + domain.PaymentRefundAbnormal

	switch res.Status {
	case RefundStateSuccess:
		// 先出账：钱已经真的退回用户微信钱包，平台这边必须同步减掉。
		if _, err := s.escrow.Withdraw(ctx, rf.UserID, rf.AmountFen, refundRefType, outRefundNo); err != nil {
			// 退款已成功但余额没扣掉 = 用户两头都有钱。不置成功，留待重试与人工介入。
			return rf, false, fmt.Errorf("退款已成功但扣减平台余额失败（需人工核账）: %w", err)
		}
		ok, err := s.refunds.MarkStatus(ctx, outRefundNo, openStatuses, domain.PaymentRefundSuccess, res.RefundID)
		if err != nil {
			return rf, false, fmt.Errorf("标记退款成功失败: %w", err)
		}
		if !ok {
			return rf, false, nil // 并发已处理
		}
		rf.Status = domain.PaymentRefundSuccess
		rf.RefundID = res.RefundID
		return rf, true, nil

	case RefundStateClosed:
		ok, err := s.refunds.MarkStatus(ctx, outRefundNo, openStatuses, domain.PaymentRefundClosed, res.RefundID)
		if err != nil {
			return rf, false, fmt.Errorf("标记退款关闭失败: %w", err)
		}
		if !ok {
			return rf, false, nil
		}
		// 钱没退出去：解冻 + 归还额度
		if rerr := s.rollbackInitiate(ctx, rf); rerr != nil {
			return rf, false, rerr
		}
		rf.Status = domain.PaymentRefundClosed
		return rf, true, nil

	case RefundStateAbnormal:
		if res.RefundID != "" {
			_, _ = s.refunds.MarkStatus(ctx, outRefundNo, openStatuses, domain.PaymentRefundAbnormal, res.RefundID)
		}
		rf.Status = domain.PaymentRefundAbnormal
		return rf, false, nil

	default: // PROCESSING 等中间态
		if rf.Status == domain.PaymentRefundCreated && res.RefundID != "" {
			_, _ = s.refunds.MarkStatus(ctx, outRefundNo, domain.PaymentRefundCreated, domain.PaymentRefundProcessing, res.RefundID)
			rf.Status = domain.PaymentRefundProcessing
			rf.RefundID = res.RefundID
		}
		return rf, false, nil
	}
}

// ListRecent 最近的退款单（管理端）。
func (s *PaymentRefundService) ListRecent(ctx context.Context, limit int) ([]domain.PaymentRefund, error) {
	if !s.Enabled() {
		return nil, ErrPaymentDisabled
	}
	return s.refunds.ListRecent(ctx, limit)
}

// ListByOrder 某笔充值单下的退款记录。
func (s *PaymentRefundService) ListByOrder(ctx context.Context, outTradeNo string) ([]domain.PaymentRefund, error) {
	if !s.Enabled() {
		return nil, ErrPaymentDisabled
	}
	return s.refunds.ListByOrder(ctx, strings.TrimSpace(outTradeNo))
}

// ListPaidOrders 已支付的充值单（管理端发起退款时选单）。
func (s *PaymentRefundService) ListPaidOrders(ctx context.Context, limit int) ([]domain.PaymentOrder, error) {
	if !s.Enabled() {
		return nil, ErrPaymentDisabled
	}
	return s.orders.ListPaid(ctx, limit)
}