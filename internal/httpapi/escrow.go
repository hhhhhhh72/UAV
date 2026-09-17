package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"drone-platform/internal/domain"
)

// requireEscrowAdmin 托管金写操作门禁：充值/冻结/释放/退款仅管理员可操作；
// 业务侧由服务端状态机内部调用（pay-and-enroll / completeEnrollment / 订单资金钩子），前端无公开调用。
// 此前任意登录用户可无限充值（deposit 无资金来源校验）再转账，属 P0 印钞漏洞。
// 注：入账渠道由服务端判定（internal_self / internal_admin），真实资金渠道只能由支付回调写入。
func requireEscrowAdmin(a domain.Actor) bool {
	return a.Role == domain.RolePlatformAdmin || a.Role == domain.RoleAssociationAdmin
}

// isPlatformAdmin 平台管理员判定：资金对账等平台级财务数据不对协会管理员开放。
func isPlatformAdmin(a domain.Actor) bool { return a.Role == domain.RolePlatformAdmin }

// GET /api/v1/admin/escrow/reconciliation — 托管金对账：按渠道汇总入金 + 全量流水明细。
// 查询参数 channel（默认 internal，all=不限渠道）、from/to（RFC3339 或 2006-01-02 日期）、limit。
// 真实资金渠道（wechat）的入金带外部支付单号，可与微信商户平台账单逐笔核对。
func (s *Server) escrowReconciliation(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || !isPlatformAdmin(a) {
		fail(w, r, http.StatusForbidden, errors.New("仅平台管理员可查看资金对账"))
		return
	}
	q := r.URL.Query()
	channel := q.Get("channel")
	if channel == "all" {
		channel = ""
	} else if channel == "" {
		channel = domain.ChannelInternal
	}
	from, err := parseReconcileTime(q.Get("from"), false)
	if err != nil {
		fail(w, r, http.StatusBadRequest, errors.New("from 格式应为 RFC3339 或 2006-01-02"))
		return
	}
	to, err := parseReconcileTime(q.Get("to"), true)
	if err != nil {
		fail(w, r, http.StatusBadRequest, errors.New("to 格式应为 RFC3339 或 2006-01-02"))
		return
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	res, err := s.escrowSvc.Reconcile(r.Context(), channel, from, to, limit)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, res)
}

// parseReconcileTime 解析对账区间参数：空串＝该端不限；支持 RFC3339 与 2006-01-02。
// 日期形式按自然日处理：起始取当天 00:00，结束取次日 00:00（左闭右开，含当天全天）。
func parseReconcileTime(raw string, isEnd bool) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	d, err := time.ParseInLocation("2006-01-02", raw, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	if isEnd {
		return d.AddDate(0, 0, 1), nil
	}
	return d, nil
}

// POST /api/v1/escrow/deposit
func (s *Server) escrowDeposit(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		AmountFen int64  `json:"amount_fen"`
		ToUser    string `json:"to_user"`
	}
	if err := decode(r, &in); err != nil || in.AmountFen <= 0 {
		fail(w, r, http.StatusBadRequest, errors.New("amount_fen > 0 required"))
		return
	}
	// 自充值（模拟托管通道）：登录用户仅可为自己入账；管理员可指定 to_user 代充。
	// 单笔上限 20_000_000 分（¥200000）：覆盖高单价课程/商品；模拟通道限额定闸，真实支付接入后由支付校验替代。
	if in.AmountFen > 20000000 {
		fail(w, r, http.StatusBadRequest, errors.New("单笔充值上限 200000 元"))
		return
	}
	target := in.ToUser
	if target == "" {
		target = a.ID
	} else if target != a.ID && !requireEscrowAdmin(a) {
		fail(w, r, http.StatusForbidden, errors.New("仅管理员可为他人的托管金入账"))
		return
	}
	// 渠道由服务端判定，不接受前端传入：管理员代充记 internal_admin（线下来款后补记），
	// 用户自助充值记 internal_self（模拟通道）。真实资金渠道（wechat）只能由支付回调入账，
	// 否则任何登录用户都能自己造一笔"微信已收款"的余额。
	channel := domain.ChannelInternalSelf
	if requireEscrowAdmin(a) {
		channel = domain.ChannelInternalAdmin
	}
	tx, created, err := s.escrowSvc.DepositFromChannel(r.Context(), target, in.AmountFen, channel, "")
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if created {
		s.audit(r.Context(), a.ID, "escrow_deposit", "escrow", tx.ID, channel)
	}
	respond(w, r, http.StatusCreated, tx)
}

// GET /api/v1/escrow/mine — 我的托管金：余额 + 近期流水（用户自查，admin 版 balance/transactions 保留）
func (s *Server) escrowMine(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	acc, err := s.escrowSvc.Balance(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	txs, err := s.escrowSvc.Transactions(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]any{"account": acc, "transactions": txs})
}

// POST /api/v1/escrow/freeze
func (s *Server) escrowFreeze(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if !requireEscrowAdmin(a) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		AmountFen     int64  `json:"amount_fen"`
		ReferenceType string `json:"reference_type"`
		ReferenceID   string `json:"reference_id"`
	}
	if err := decode(r, &in); err != nil || in.AmountFen <= 0 {
		fail(w, r, http.StatusBadRequest, errors.New("amount_fen > 0 required"))
		return
	}
	tx, err := s.escrowSvc.Freeze(r.Context(), a.ID, in.AmountFen, in.ReferenceType, in.ReferenceID)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	s.audit(r.Context(), a.ID, "escrow_freeze", in.ReferenceType, in.ReferenceID, "frozen")
	respond(w, r, http.StatusCreated, tx)
}

// POST /api/v1/escrow/release
func (s *Server) escrowRelease(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if !requireEscrowAdmin(a) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		ToUser        string `json:"to_user"`
		FromUser      string `json:"from_user"`
		AmountFen     int64  `json:"amount_fen"`
		ReferenceType string `json:"reference_type"`
		ReferenceID   string `json:"reference_id"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// 释放账户：默认管理员本人；可指定 from_user（平台代学员/商户结算给收款方）。
	src := in.FromUser
	if src == "" {
		src = a.ID
	}
	tx, err := s.escrowSvc.Release(r.Context(), src, in.ToUser, in.AmountFen, in.ReferenceType, in.ReferenceID)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	s.audit(r.Context(), a.ID, "escrow_release", in.ReferenceType, in.ReferenceID, "released")
	respond(w, r, http.StatusCreated, tx)
}

// POST /api/v1/escrow/refund
func (s *Server) escrowRefund(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if !requireEscrowAdmin(a) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		UserID        string `json:"user_id"`
		AmountFen     int64  `json:"amount_fen"`
		ReferenceType string `json:"reference_type"`
		ReferenceID   string `json:"reference_id"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// 默认解冻操作者自己的资金；带 user_id 时解冻指定用户的——这是「驳回报名自动退款失败」
	// 唯一的人工补救入口（escrow_release 本就允许管理员任意指定 from/to，权限级别一致）。
	// 天然防超额：仓储层是 frozen_fen >= amount 的条件 UPDATE，退不出去就报错，不会凭空造钱。
	target := in.UserID
	if target == "" {
		target = a.ID
	}
	tx, err := s.escrowSvc.Refund(r.Context(), target, in.AmountFen, in.ReferenceType, in.ReferenceID)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	s.audit(r.Context(), a.ID, "escrow_refund", in.ReferenceType, in.ReferenceID, "refunded:"+target)
	respond(w, r, http.StatusCreated, tx)
}

// GET /api/v1/escrow/balance
func (s *Server) escrowBalance(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	bal, err := s.escrowSvc.Balance(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, bal)
}

// GET /api/v1/escrow/transactions
func (s *Server) escrowTransactions(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	txs, err := s.escrowSvc.Transactions(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, txs)
}
