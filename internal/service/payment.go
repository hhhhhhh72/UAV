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

// 充值金额边界（分）。两端都是业务规则，所以放 Service 而不是 handler：
// 下限挡掉「1 分钱刷单」噪声订单（每单都产生一次微信请求与一行审计），
// 上限挡掉误操作与单笔风控敞口。
const (
	MinRechargeFen int64 = 100     // ¥1
	MaxRechargeFen int64 = 5000000 // ¥50,000
)

// ErrPaymentDisabled 微信支付未开通（未配置商户号等五项）。
// handler 据此回 503 而不是 500——「没开通」是预期状态，不是故障。
var ErrPaymentDisabled = errors.New("微信支付未开通")

// PrepayInput 下单入参。刻意不引用 wechatpay 包的类型：
// Service 只依赖下面的 PaymentGateway 接口，测试换成桩即可，不必打真实微信。
type PrepayInput struct {
	OutTradeNo  string
	Description string
	AmountFen   int64
	PayerOpenID string
}

// PaymentResult 网关查单结果——**唯一可信的到账依据**。
type PaymentResult struct {
	TradeState    string // 只有 "SUCCESS" 才入账
	TransactionID string // 微信支付单号，作为托管流水的 external_txn_id
	AmountFen     int64  // 微信侧实收金额，必须与本地订单逐分相等
	SuccessTime   time.Time
}

// PaymentGateway 支付网关（出站端口），由 internal/wechatpay 适配实现。
type PaymentGateway interface {
	Prepay(ctx context.Context, in PrepayInput) (prepayID string, err error)
	Query(ctx context.Context, outTradeNo string) (PaymentResult, error)
	PayParams(prepayID string) (map[string]string, error)
	// DecodeNotify 解析一次网关通知原文，只返回其中的商户订单号。
	// 实现方负责解密与格式校验；**通知里的金额与身份字段一律不得作为入账依据**。
	DecodeNotify(body []byte) (outTradeNo string, err error)
}

// NotifyOrderRef 从网关通知原文里取出商户订单号——只取这一个字段，其余一概不采信。
// 随后必须走 Confirm 主动查单，通知本身不构成到账依据。
func (s *PaymentService) NotifyOrderRef(body []byte) (string, error) {
	if !s.Enabled() {
		return "", ErrPaymentDisabled
	}
	if len(body) == 0 {
		return "", errors.New("通知报文为空")
	}
	return s.gw.DecodeNotify(body)
}

// PaymentService 线上充值：下单 → 到账确认。
//
// 安全模型（回调只做触发、以主动查询为准）见 internal/wechatpay 包注释。
// 这里额外落三条红线：
//  1. **金额只信查单结果**，并与本地订单 amount_fen 逐分比对，不等即拒绝入账；
//  2. 入账用户一律由**本地订单**反查，绝不采信回调报文里的任何身份字段；
//  3. 下单/标记/入账三层各自幂等，微信重试回调不会重复加钱。
type PaymentService struct {
	orders repository.PaymentOrderRepository
	escrow *EscrowService
	gw     PaymentGateway
	now    func() time.Time
}

func NewPaymentService(orders repository.PaymentOrderRepository, escrow *EscrowService, gw PaymentGateway) *PaymentService {
	return &PaymentService{orders: orders, escrow: escrow, gw: gw, now: time.Now}
}

// Enabled 微信支付是否可用（网关已装配）。未装配时 handler 统一回 503。
func (s *PaymentService) Enabled() bool { return s != nil && s.gw != nil && s.orders != nil && s.escrow != nil }

// NewOutTradeNo 生成商户订单号：RC + yyyymmddHHMMSS + 8 位随机 hex = 24 字符。
// 微信要求 6~32 位、只能是字母数字与 _-|*@，本格式满足；带时间前缀便于人工对账时排序。
func NewOutTradeNo(now time.Time) (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成商户订单号失败: %w", err)
	}
	return "RC" + now.Format("20060102150405") + hex.EncodeToString(b), nil
}

// Prepay 发起一笔充值：先落本地订单（created），再向微信下单，最后回填 prepay_id。
//
// 顺序刻意为「先落单再下单」：微信下单成功但本地落单失败会留下一个用户拿不到
// 调起参数的微信订单（2 小时后自动关闭），而反过来只会留下一行永远付不掉的
// created——后者对账时看得见、也无资金风险。宁可留可见的垃圾，不留不可见的悬空。
func (s *PaymentService) Prepay(ctx context.Context, userID, payerOpenID string, amountFen int64) (domain.PaymentOrder, map[string]string, error) {
	if !s.Enabled() {
		return domain.PaymentOrder{}, nil, ErrPaymentDisabled
	}
	if strings.TrimSpace(userID) == "" {
		return domain.PaymentOrder{}, nil, errors.New("下单用户不能为空")
	}
	if amountFen < MinRechargeFen || amountFen > MaxRechargeFen {
		return domain.PaymentOrder{}, nil, fmt.Errorf("充值金额需在 %d~%d 分之间", MinRechargeFen, MaxRechargeFen)
	}
	if strings.TrimSpace(payerOpenID) == "" {
		return domain.PaymentOrder{}, nil, errors.New("缺少微信 openid，无法发起小程序支付")
	}

	outTradeNo, err := NewOutTradeNo(s.now())
	if err != nil {
		return domain.PaymentOrder{}, nil, err
	}
	order := domain.PaymentOrder{
		ID: nextID("pay"), OutTradeNo: outTradeNo, UserID: userID, AmountFen: amountFen,
		Channel: domain.ChannelWeChat, Status: domain.PaymentCreated, CreatedAt: s.now(),
	}
	if _, err := s.orders.Create(ctx, order); err != nil {
		return domain.PaymentOrder{}, nil, fmt.Errorf("创建支付订单失败: %w", err)
	}

	prepayID, err := s.gw.Prepay(ctx, PrepayInput{
		OutTradeNo:  outTradeNo,
		Description: "平台账户充值",
		AmountFen:   amountFen,
		PayerOpenID: payerOpenID,
	})
	if err != nil {
		return domain.PaymentOrder{}, nil, err
	}
	// 回填失败不阻断下单：prepay_id 只是前端调起支付的凭据，付没付以查单为准。
	if err := s.orders.SetPrepayID(ctx, outTradeNo, prepayID); err == nil {
		order.PrepayID = prepayID
	}
	params, err := s.gw.PayParams(prepayID)
	if err != nil {
		return domain.PaymentOrder{}, nil, err
	}
	return order, params, nil
}

// Confirm 处理一次「可能已到账」的通知（微信回调 / 用户主动查单），返回订单与本次是否真的加了钱。
//
// 无论通知从哪来，都**先主动查单**：回调报文不作为到账依据，
// 因此伪造回调最多触发一次返回「未支付」的查询，不产生任何资金影响。
//
// 关于「已 paid 就不再入账」：这里刻意**不**在 MarkPaid 返回 false 时提前返回。
// 因为存在「订单已置 paid、但托管入账那一步失败」的中间态，微信重试时必须补上；
// 补账由 DepositFromChannel 的 (channel, external_txn_id) 查重保证只加一次钱。
func (s *PaymentService) Confirm(ctx context.Context, outTradeNo string) (domain.PaymentOrder, bool, error) {
	if !s.Enabled() {
		return domain.PaymentOrder{}, false, ErrPaymentDisabled
	}
	outTradeNo = strings.TrimSpace(outTradeNo)
	if outTradeNo == "" {
		return domain.PaymentOrder{}, false, errors.New("商户订单号为空")
	}
	order, found, err := s.orders.FindByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		return domain.PaymentOrder{}, false, fmt.Errorf("查本地支付订单失败: %w", err)
	}
	if !found {
		// 未知订单号：可能是攻击者伪造，也可能是同一商户号下的其它系统订单。一律不查不记。
		return domain.PaymentOrder{}, false, fmt.Errorf("支付订单不存在: %s", outTradeNo)
	}

	res, err := s.gw.Query(ctx, outTradeNo)
	if err != nil {
		return order, false, fmt.Errorf("查微信支付订单失败: %w", err)
	}
	if res.TradeState != "SUCCESS" {
		// 未支付/已关闭/已撤销/支付失败——一律不入账，且不算错误（用户可能还没付完）。
		return order, false, nil
	}
	// 红线：金额只信查单结果，且必须与本地订单逐分相等。
	if res.AmountFen != order.AmountFen {
		return order, false, fmt.Errorf("金额不符：微信侧 %d 分，本地订单 %d 分（订单 %s）", res.AmountFen, order.AmountFen, outTradeNo)
	}
	if strings.TrimSpace(res.TransactionID) == "" {
		return order, false, fmt.Errorf("微信侧已支付但未返回支付单号（订单 %s）", outTradeNo)
	}

	paidAt := res.SuccessTime
	if paidAt.IsZero() {
		paidAt = s.now()
	}
	// 第一道防线：并发回调只有一个能把 created 改成 paid。
	if _, err := s.orders.MarkPaid(ctx, outTradeNo, res.TransactionID, paidAt); err != nil {
		return order, false, fmt.Errorf("标记订单已支付失败: %w", err)
	}
	// 第二道防线：托管入账按 (wechat, 微信支付单号) 查重，重放不会再加钱。
	// 入账用户取本地订单的 user_id——回调报文里的身份字段一律不采信。
	_, credited, err := s.escrow.DepositFromChannel(ctx, order.UserID, order.AmountFen, domain.ChannelWeChat, res.TransactionID)
	if err != nil {
		return order, false, fmt.Errorf("充值入账失败: %w", err)
	}
	if credited {
		order.Status = domain.PaymentPaid
		order.TransactionID = res.TransactionID
		order.PaidAt = paidAt
	}
	return order, credited, nil
}

// ListMine 我的充值记录（倒序，limit<=0 用默认值）。
func (s *PaymentService) ListMine(ctx context.Context, userID string, limit int) ([]domain.PaymentOrder, error) {
	if !s.Enabled() {
		return nil, ErrPaymentDisabled
	}
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("用户不能为空")
	}
	return s.orders.ListByUser(ctx, userID, limit)
}
