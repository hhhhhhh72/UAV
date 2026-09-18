package domain

import "time"

// 支付订单状态：created → paid（终态，钱已真实到账并入托管金）。
// closed 用于超时关单/用户取消；failed 保留给微信侧明确失败。
const (
	PaymentCreated = "created"
	PaymentPaid    = "paid"
	PaymentClosed  = "closed"
	PaymentFailed  = "failed"
)

// PaymentOrder 线上支付订单。
//
// 安全要点：OutTradeNo 是**我们自己生成的商户订单号**，也是回调对账的唯一主键。
// 微信回调只带回它，user_id 与金额一律由本表反查得到——
// 绝不采信回调报文里的金额，否则伪造回调即可凭空充值。
//
// PrepayID 是微信下单返回的预支付会话标识，仅用于前端调起支付，不构成到账凭证；
// 到账唯一以 Status=paid + TransactionID 为准。
type PaymentOrder struct {
	ID            string `json:"id"`
	OutTradeNo    string `json:"out_trade_no"`
	UserID        string `json:"user_id"`
	AmountFen     int64  `json:"amount_fen"`
	Channel       string `json:"channel"`
	Status        string `json:"status"`
	PrepayID      string `json:"prepay_id"`
	TransactionID string `json:"transaction_id"`
	// RefundedFen 已退金额累计（含处理中的）。「累计退款不超过原单金额」由仓库层的
	// 条件更新保证（refunded_fen + delta <= amount_fen），不依赖应用层先查后写。
	RefundedFen int64     `json:"refunded_fen"`
	PaidAt      time.Time `json:"paid_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 退款单状态：created → processing → success（终态，钱已退出平台）。
// closed 表示微信侧退款未成功（已把冻结退回用户余额）；abnormal 需人工介入。
const (
	PaymentRefundCreated    = "created"
	PaymentRefundProcessing = "processing"
	PaymentRefundSuccess    = "success"
	PaymentRefundClosed     = "closed"
	PaymentRefundAbnormal   = "abnormal"
)

// PaymentRefund 线上退款单：针对**某一笔已支付的充值单**发起，支持多次部分退款。
//
// 为什么退款的对象是充值单而不是订单：平台的钱是以「充值进托管余额」的形式进来的，
// 用户微信付的那笔钱对应一条 PaymentOrder。微信退款 API 本就要求 out_trade_no + 累计
// 退款不超过原单金额，与这个模型正好对上。
//
// 与 escrow_transactions 的分工（两者必须同时发生，否则真钱退出去、平台余额还留着＝双花）：
//   created  → 已冻结用户托管余额（EscrowService.Freeze，余额不足则发起就失败）
//   success  → 冻结额被扣掉（EscrowService.Withdraw，钱真的离开平台）
//   closed   → 冻结额退回用户余额（EscrowService.Refund）
//   abnormal → 钱仍冻着，等人工处理
type PaymentRefund struct {
	ID          string    `json:"id"`
	OutRefundNo string    `json:"out_refund_no"`
	OutTradeNo  string    `json:"out_trade_no"`
	UserID      string    `json:"user_id"`
	AmountFen   int64     `json:"amount_fen"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"`
	RefundID    string    `json:"refund_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
