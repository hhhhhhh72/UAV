package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"drone-platform/internal/service"
)

// ── 线上充值：下单限频 ──
//
// 每成功调一次 prepay 都会向微信**真实创建一个支付单**（2 小时内有效），
// 同时在 payment_orders 落一行。没有约束时，一个登录账号可以在几秒内造出成千上万个
// 微信侧待支付订单——轻则灌爆本地表，重则被微信风控标记商户号（影响的是整个平台的收款）。
// 按「用户」维度限频：换 IP 不该绕过，共用出口 IP 的用户也不该互相拖累
// （与 trade_order_maintenance.go 的商城下单限频同一套取舍）。
const (
	// prepayWindow 充值下单限频窗口。
	prepayWindow = 10 * time.Minute
	// prepayMax 窗口内同一用户最多发起的充值单数。
	prepayMax = 10
	// payLimitMaxEntries 限频表条目上限（防内存 DoS）：超限清空重建。
	payLimitMaxEntries = 10000
)

// prepayAllowed 报告该用户在当前窗口内是否还能发起充值（含本次），并累计本次。
func (s *Server) prepayAllowed(userID string) bool {
	if s.payLimitEntries.Load() >= payLimitMaxEntries {
		s.payLimits.Range(func(k, _ any) bool { s.payLimits.Delete(k); return true })
		s.payLimitEntries.Store(0)
	}
	s.payLimitEntries.Add(1)
	v, _ := s.payLimits.LoadOrStore(userID, &regLimitLog{windowStart: time.Now()})
	log := v.(*regLimitLog)
	log.mu.Lock()
	defer log.mu.Unlock()
	now := time.Now()
	if now.Sub(log.windowStart) >= prepayWindow {
		log.windowStart = now
		log.count = 0
	}
	log.count++
	return log.count <= prepayMax
}

// 线上充值（微信支付）handler 层。
//
// 这一层的分寸：
//   - 金额区间、状态机、幂等全部在 PaymentService 里，这里只做「解析 → 调 Service → 响应」；
//   - 唯一的例外是回调应答体，见 wechatPayNotify 的注释。

// POST /api/v1/payments/wechat/prepay — 发起一笔充值，返回 wx.requestPayment 所需参数。
func (s *Server) wechatPayPrepay(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	// 限频必须在调用微信之前：它的目的就是"别把请求发到微信去"。
	if !s.prepayAllowed(a.ID) {
		fail(w, r, http.StatusTooManyRequests, errors.New("充值下单过于频繁，请稍后再试"))
		return
	}
	if !s.paymentSvc.Enabled() {
		// 「没开通」是预期状态，不是故障——所以**不能用 5xx**。
		// fail() 对 status>=500 会做两件副作用很大的事：把 message 统一脱敏成
		// "internal server error"（客户端再也读不到"微信支付未开通"），并且每次都记一条
		// ERROR 日志（预期状态刷满错误日志，监控里还算一次故障）。
		// 409 是 4xx，消息原样透传，语义是"服务器当前状态无法完成该请求"。
		fail(w, r, http.StatusConflict, service.ErrPaymentDisabled)
		return
	}
	var body struct {
		AmountFen int64 `json:"amount_fen"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&body); err != nil {
		fail(w, r, http.StatusBadRequest, errors.New("请求体不是合法 JSON"))
		return
	}
	// payer.openid 必须来自**库里的当前用户**，不接受前端传入——否则等于让调用方
	// 指定「谁付钱」，可以把别人的 openid 塞进来（他人被迫付款 / 资金归属错乱）。
	u, err := s.userRepo.FindByID(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("读取用户失败: %w", err))
		return
	}
	order, params, err := s.paymentSvc.Prepay(r.Context(), a.ID, u.WechatOpenID, body.AmountFen)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPaymentDisabled):
			fail(w, r, http.StatusConflict, err)
		case strings.Contains(err.Error(), "充值金额需在"):
			fail(w, r, http.StatusBadRequest, err)
		default:
			// 微信侧报错（签名/证书/参数）属于运维问题，对用户只暴露「稍后再试」，
			// 详情留给服务端日志——详细报错里可能含商户号等信息。
			slog.Warn("wechat prepay failed", "user", a.ID, "error", err)
			fail(w, r, http.StatusBadGateway, errors.New("发起支付失败，请稍后重试"))
		}
		return
	}
	s.audit(r.Context(), a.ID, "payment_prepay", "payment_order", order.ID, fmt.Sprintf("%d分", order.AmountFen))
	respond(w, r, http.StatusCreated, map[string]any{
		"out_trade_no": order.OutTradeNo,
		"amount_fen":   order.AmountFen,
		"pay_params":   params,
	})
}

// POST /api/v1/payments/wechat/notify — 微信支付结果通知（公网可达，无 Bearer 令牌）。
//
// 应答体是本文件唯一手写的 JSON，原因：微信只认 {"code":"SUCCESS"} 这一种成功应答，
// 用 respond() 包成 {"data":{...}} 会被判为失败并触发重试风暴，用 fail() 的
// {"error":{...}} 更是永远收不到成功。这是第三方协议的线格式，不是本项目 API。
//
// 安全模型：本端点**不采信任何通知内容**，只用它取一个商户订单号，随即主动查微信
// 「这笔到底付没付、付了多少」。伪造通知最多触发一次返回「未支付」的查询。
func (s *Server) wechatPayNotify(w http.ResponseWriter, r *http.Request) {
	if !s.paymentSvc.Enabled() {
		writeWeChatNotify(w, http.StatusServiceUnavailable, "FAIL", "微信支付未开通")
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeWeChatNotify(w, http.StatusBadRequest, "FAIL", "读取报文失败")
		return
	}
	outTradeNo, err := s.paymentSvc.NotifyOrderRef(raw)
	if err != nil {
		// 解密失败/格式不对：可能是伪造，也可能是 APIv3 密钥配错。回 400 让微信按策略重试。
		slog.Warn("wechat notify decode failed", "error", err)
		writeWeChatNotify(w, http.StatusBadRequest, "FAIL", "通知报文无法解析")
		return
	}
	order, credited, err := s.paymentSvc.Confirm(r.Context(), outTradeNo)
	if err != nil {
		// 查单失败 / 金额不符 / 入账失败：必须回非 2xx，让微信重试。
		// 金额不符尤其不能吞——它是篡改或串单的信号，要留日志等人看。
		slog.Error("wechat notify confirm failed", "out_trade_no", outTradeNo, "error", err)
		writeWeChatNotify(w, http.StatusInternalServerError, "FAIL", "到账确认失败")
		return
	}
	if credited {
		s.audit(r.Context(), order.UserID, "payment_paid", "payment_order", order.ID,
			fmt.Sprintf("%d分 微信单号 %s", order.AmountFen, order.TransactionID))
		s.notify(order.UserID, "充值到账",
			fmt.Sprintf("充值 %s 元已到账，可在「我的钱包」查看。", yuanText(order.AmountFen)),
			"payment_order", order.ID)
	}
	writeWeChatNotify(w, http.StatusOK, "SUCCESS", "成功")
}

// GET /api/v1/payments/mine — 我的充值记录，同时充当「真实支付是否已开通」的探测点。
//
// 未开通时恒回 200 + enabled:false，而不是 503：
//   - 这个端点的职责之一是让前端判断该走微信支付还是模拟通道，属于**探测**而非失败；
//   - 5xx 会被 fail() 脱敏 + 记 ERROR，客户端分不清"没开通"和"服务炸了"。
//
// 不返回 total：仓储层是「取最近 N 条」（ListByUser 带 limit），不是真分页。
// 此前把 len(orders) 当 total 传给前端，会让它以为还有下一页、无限翻。
func (s *Server) paymentMine(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if !s.paymentSvc.Enabled() {
		respond(w, r, http.StatusOK, map[string]any{"enabled": false, "items": []any{}})
		return
	}
	page, pageSize := paginationFromQuery(r)
	orders, err := s.paymentSvc.ListMine(r.Context(), a.ID, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]any{
		"enabled": true, "items": orders, "page": page, "page_size": pageSize,
	})
}

// writeWeChatNotify 按微信支付 APIv3 的线格式写回调应答（见 wechatPayNotify 注释）。
func writeWeChatNotify(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": message})
}

// yuanText 分 → 元文本，用于站内通知（1 位小数够读，不做金额展示的精确性依赖）。
func yuanText(fen int64) string {
	return fmt.Sprintf("%d.%02d", fen/100, fen%100)
}
