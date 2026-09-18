package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"drone-platform/internal/service"
	"drone-platform/internal/wechatpay"
)

// weChatPayGateway 把 internal/wechatpay 的客户端适配成 service.PaymentGateway。
//
// 适配层放 main 包是刻意的依赖方向选择：wechatpay 保持「叶子包」（不认识业务），
// service 保持「不认识微信」（只认识自己定义的出站端口），两边都不知道对方存在，
// 由装配处把它们接起来——测试里换一个桩就能整条链路跑通，不必打真实微信。
type weChatPayGateway struct{ c *wechatpay.Client }

func newWeChatPayGateway(c *wechatpay.Client) service.PaymentGateway { return &weChatPayGateway{c: c} }

func (g *weChatPayGateway) Prepay(ctx context.Context, in service.PrepayInput) (string, error) {
	return g.c.CreateJSAPIOrder(ctx, in.OutTradeNo, in.Description, in.AmountFen, in.PayerOpenID)
}

func (g *weChatPayGateway) PayParams(prepayID string) (map[string]string, error) {
	return g.c.PayParams(prepayID)
}

// Query 查单并翻译成 service 的支付结果。success_time 是 RFC3339；
// 解析失败不让整笔入账失败——时间只用于流水展示，到账与否由 trade_state 决定，
// 解析不出来就用零值，Service 会回落到「现在」。
func (g *weChatPayGateway) Query(ctx context.Context, outTradeNo string) (service.PaymentResult, error) {
	tx, err := g.c.QueryOrder(ctx, outTradeNo)
	if err != nil {
		return service.PaymentResult{}, err
	}
	res := service.PaymentResult{
		TradeState:    tx.TradeState,
		TransactionID: tx.TransactionID,
		AmountFen:     tx.Amount.Total,
	}
	if ts := strings.TrimSpace(tx.SuccessTime); ts != "" {
		if t, perr := time.Parse(time.RFC3339, ts); perr == nil {
			res.SuccessTime = t
		}
	}
	return res, nil
}

// DecodeNotify 用 APIv3 密钥解密回调，只取出商户订单号。
//
// 这里**故意**不返回金额、openid 等任何其它字段：调用方拿不到就无从误用，
// 「回调只做触发、以主动查询为准」因此是接口层面的约束，而不是靠注释提醒。
func (g *weChatPayGateway) DecodeNotify(body []byte) (string, error) {
	var env wechatpay.NotifyEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return "", fmt.Errorf("wechatpay: 回调报文不是合法 JSON: %w", err)
	}
	plain, err := g.c.DecryptResource(env)
	if err != nil {
		return "", err
	}
	var payload struct {
		OutTradeNo string `json:"out_trade_no"`
	}
	if err := json.Unmarshal(plain, &payload); err != nil {
		return "", fmt.Errorf("wechatpay: 回调明文不是合法 JSON: %w", err)
	}
	if strings.TrimSpace(payload.OutTradeNo) == "" {
		return "", fmt.Errorf("wechatpay: 回调缺少 out_trade_no")
	}
	return payload.OutTradeNo, nil
}
