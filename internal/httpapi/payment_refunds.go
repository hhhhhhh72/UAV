package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"drone-platform/internal/domain"
	"drone-platform/internal/service"
)

// 线上退款（微信支付）handler 层。
//
// 发起退款**仅平台管理员**：钱要离开平台，属于资金动作，与 /admin/escrow/reconciliation
// 同一权限级别（协会管理员能看到内容，但不经手资金）。adminGate 已挡住非管理员，
// 这里再显式判一次角色——两道一起才拦得住"adminGate 白名单将来被放宽"这种改动。

// POST /api/v1/admin/payments/refunds — 对某一笔已支付的充值单发起退款。
func (s *Server) adminCreateRefund(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("仅平台管理员可发起退款"))
		return
	}
	if !s.refundSvc.Enabled() {
		fail(w, r, http.StatusConflict, service.ErrPaymentDisabled)
		return
	}
	var in struct {
		OutTradeNo string `json:"out_trade_no"`
		AmountFen  int64  `json:"amount_fen"`
		Reason     string `json:"reason"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&in); err != nil {
		fail(w, r, http.StatusBadRequest, errors.New("请求体不是合法 JSON"))
		return
	}
	rf, err := s.refundSvc.Initiate(r.Context(), in.OutTradeNo, in.AmountFen, in.Reason, a.ID)
	if err != nil {
		slog.Warn("wechat refund initiate failed", "out_trade_no", in.OutTradeNo, "amount_fen", in.AmountFen, "error", err)
		// 结果未知（微信已受理但响应丢失）必须是 502 而不是 400：它不是参数问题，
		// 而且**绝不能**让调用方以为"没发出去"从而重试发起——那样会重复退款。
		if strings.Contains(err.Error(), "退款结果未知") {
			fail(w, r, http.StatusBadGateway, errors.New("退款结果未知，已提交微信；请在退款列表中查询进度，勿重复发起"))
			return
		}
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	s.audit(r.Context(), a.ID, "payment_refund", "payment_refund", rf.OutRefundNo,
		fmt.Sprintf("%d分 针对充值单 %s（%s）", rf.AmountFen, rf.OutTradeNo, rf.Status))
	respond(w, r, http.StatusCreated, rf)
}

// GET /api/v1/admin/payments/refunds — 最近退款单（管理端核对进度用）。
func (s *Server) adminListRefunds(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("仅平台管理员可查看退款记录"))
		return
	}
	if !s.refundSvc.Enabled() {
		failureOrEmpty(w, r, "refunds")
		return
	}
	// out_trade_no 给了就只看这一笔充值单下的退款记录。
	if no := strings.TrimSpace(r.URL.Query().Get("out_trade_no")); no != "" {
		list, err := s.refundSvc.ListByOrder(r.Context(), no)
		if err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
		respond(w, r, http.StatusOK, map[string]any{"enabled": true, "items": list})
		return
	}
	_, pageSize := paginationFromQuery(r)
	list, err := s.refundSvc.ListRecent(r.Context(), pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]any{"enabled": true, "items": list})
}

// GET /api/v1/admin/payments/paid-orders — 已支付的充值单（发起退款时选单）。
func (s *Server) adminListPaidOrders(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("仅平台管理员可查看充值单"))
		return
	}
	if !s.refundSvc.Enabled() {
		failureOrEmpty(w, r, "orders")
		return
	}
	_, pageSize := paginationFromQuery(r)
	list, err := s.refundSvc.ListPaidOrders(r.Context(), pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]any{"enabled": true, "items": list})
}

// failureOrEmpty 未开通微信支付时的只读响应：
// 与 payments/mine 同一口径——**探测语义不该是 5xx**（那会被 fail() 脱敏成
// internal server error 并每次记 ERROR）。这里恒回 200 + enabled:false。
func failureOrEmpty(w http.ResponseWriter, r *http.Request, _ string) {
	respond(w, r, http.StatusOK, map[string]any{"enabled": false, "items": []any{}})
}

// POST /api/v1/payments/wechat/refund-notify — 微信退款结果通知（公网可达，无 Bearer 令牌）。
//
// 与支付回调同一约定：应答体是微信认的 {"code":"SUCCESS"}（第三方协议线格式），
// 且**不采信通知内容**——只取退款单号，随即主动查单确认。
func (s *Server) wechatRefundNotify(w http.ResponseWriter, r *http.Request) {
	if !s.refundSvc.Enabled() {
		writeWeChatNotify(w, http.StatusServiceUnavailable, "FAIL", "微信支付未开通")
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeWeChatNotify(w, http.StatusBadRequest, "FAIL", "读取报文失败")
		return
	}
	outRefundNo, err := s.refundSvc.NotifyRef(raw)
	if err != nil {
		slog.Warn("wechat refund notify decode failed", "error", err)
		writeWeChatNotify(w, http.StatusBadRequest, "FAIL", "通知报文无法解析")
		return
	}
	rf, settled, err := s.refundSvc.Confirm(r.Context(), outRefundNo)
	if err != nil {
		// 查单失败 / 出账失败：必须回非 2xx 让微信重试（Confirm 是幂等的）。
		slog.Error("wechat refund notify confirm failed", "out_refund_no", outRefundNo, "error", err)
		writeWeChatNotify(w, http.StatusInternalServerError, "FAIL", "退款确认失败")
		return
	}
	if settled {
		s.audit(r.Context(), rf.UserID, "payment_refund_settled", "payment_refund", rf.OutRefundNo,
			fmt.Sprintf("%s %d分", rf.Status, rf.AmountFen))
		s.notify(rf.UserID, "退款结果",
			fmt.Sprintf("您的充值退款 %s 元已%s。", yuanText(rf.AmountFen), refundStatusText(rf.Status)),
			"payment_refund", rf.OutRefundNo)
	}
	writeWeChatNotify(w, http.StatusOK, "SUCCESS", "成功")
}

// refundStatusText 退款状态的中文说明（站内通知用）。
func refundStatusText(status string) string {
	switch status {
	case domain.PaymentRefundSuccess:
		return "原路退回微信钱包"
	case domain.PaymentRefundClosed:
		return "未成功，金额已退回平台余额"
	case domain.PaymentRefundAbnormal:
		return "异常，客服会尽快处理"
	default:
		return "处理中"
	}
}