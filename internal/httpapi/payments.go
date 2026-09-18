package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"drone-platform/internal/service"
)

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
	if !s.paymentSvc.Enabled() {
		// 「没开通」是预期状态而非故障，回 503 让前端能给出确定文案，而不是 500 让人以为崩了。
		fail(w, r, http.StatusServiceUnavailable, service.ErrPaymentDisabled)
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
			fail(w, r, http.StatusServiceUnavailable, err)
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

// GET /api/v1/payments/mine — 我的充值记录。
func (s *Server) paymentMine(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if !s.paymentSvc.Enabled() {
		fail(w, r, http.StatusServiceUnavailable, service.ErrPaymentDisabled)
		return
	}
	page, pageSize := paginationFromQuery(r)
	orders, err := s.paymentSvc.ListMine(r.Context(), a.ID, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respondPage(w, r, orders, len(orders), page, pageSize)
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
