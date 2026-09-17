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
	ID            string    `json:"id"`
	OutTradeNo    string    `json:"out_trade_no"`
	UserID        string    `json:"user_id"`
	AmountFen     int64     `json:"amount_fen"`
	Channel       string    `json:"channel"`
	Status        string    `json:"status"`
	PrepayID      string    `json:"prepay_id"`
	TransactionID string    `json:"transaction_id"`
	PaidAt        time.Time `json:"paid_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
