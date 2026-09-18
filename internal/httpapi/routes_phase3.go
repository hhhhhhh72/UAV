package httpapi

import "net/http"

func (s *Server) registerPhase3Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/training-courses/{id}/enroll", s.enrollCourse)
	mux.HandleFunc("POST /api/v1/training-courses/{id}/pay-and-enroll", s.payAndEnroll)
	mux.HandleFunc("GET /api/v1/training-courses/{id}/enrollments", s.listEnrollments)
	mux.HandleFunc("GET /api/v1/enrollments/mine", s.listMyEnrollments)
	mux.HandleFunc("PUT /api/v1/admin/enrollments/{id}", s.updateEnrollment)
	mux.HandleFunc("POST /api/v1/enrollments/{id}/complete", s.completeEnrollment)
	mux.HandleFunc("POST /api/v1/enrollments/{id}/review", s.reviewEnrollment)
	mux.HandleFunc("GET /api/v1/certificates/expiring", s.listExpiringCerts)
	mux.HandleFunc("GET /api/v1/inspections/expiring", s.listExpiringInspections)
	mux.HandleFunc("POST /api/v1/trade-orders", s.createTradeOrder)
	mux.HandleFunc("POST /api/v1/trade-orders/{id}/pay", s.payTradeOrder)
	// 卖家发货独立端点：必须带快递单号（旧路径 PATCH status=shipped 已被服务层封掉）
	mux.HandleFunc("POST /api/v1/trade-orders/{id}/ship", s.shipTradeOrder)
	mux.HandleFunc("PATCH /api/v1/trade-orders/{id}/status", s.updateTradeOrderStatus)
	mux.HandleFunc("POST /api/v1/trade-orders/{id}/aftersale", s.applyAftersale)
	mux.HandleFunc("POST /api/v1/trade-orders/{id}/aftersale/review", s.reviewAftersaleBySeller)
	// 退货退款流程：买家提交退货物流 → 卖家/管理员确认收到 → 退款
	mux.HandleFunc("POST /api/v1/trade-orders/{id}/aftersale/return", s.submitReturnShipment)
	mux.HandleFunc("POST /api/v1/trade-orders/{id}/aftersale/confirm-return", s.confirmReturnReceived)
	mux.HandleFunc("GET /api/v1/trade-orders/mine", s.listMyTradeOrders)
	mux.HandleFunc("GET /api/v1/admin/dashboard", s.adminDashboard)
	// Escrow (资金托管)
	mux.HandleFunc("POST /api/v1/escrow/deposit", s.escrowDeposit)
	mux.HandleFunc("POST /api/v1/escrow/freeze", s.escrowFreeze)
	mux.HandleFunc("POST /api/v1/escrow/release", s.escrowRelease)
	mux.HandleFunc("POST /api/v1/escrow/refund", s.escrowRefund)
	mux.HandleFunc("GET /api/v1/escrow/balance", s.escrowBalance)
	mux.HandleFunc("GET /api/v1/escrow/transactions", s.escrowTransactions)
	mux.HandleFunc("GET /api/v1/escrow/mine", s.escrowMine)
	mux.HandleFunc("GET /api/v1/admin/escrow/reconciliation", s.escrowReconciliation)
	// 线上充值（微信支付）。notify 是微信服务器回调：无 Bearer 令牌，
	// 已在 auth.go 的公开路径白名单里显式放行（见该处注释）。
	mux.HandleFunc("POST /api/v1/payments/wechat/prepay", s.wechatPayPrepay)
	mux.HandleFunc("POST /api/v1/payments/wechat/notify", s.wechatPayNotify)
	mux.HandleFunc("GET /api/v1/payments/mine", s.paymentMine)
	// 线上退款：发起/查询仅平台管理员（adminGate + handler 内显式角色判定两道）。
	mux.HandleFunc("POST /api/v1/admin/payments/refunds", s.adminCreateRefund)
	mux.HandleFunc("GET /api/v1/admin/payments/refunds", s.adminListRefunds)
	mux.HandleFunc("GET /api/v1/admin/payments/paid-orders", s.adminListPaidOrders)
	// refund-notify 是微信服务器回调：无 Bearer 令牌，已在 auth.go 公开路径白名单里显式放行。
	mux.HandleFunc("POST /api/v1/payments/wechat/refund-notify", s.wechatRefundNotify)
	// News (行业资讯)
	mux.HandleFunc("POST /api/v1/articles", s.createArticle)
	mux.HandleFunc("GET /api/v1/articles", s.listArticles)
	mux.HandleFunc("PUT /api/v1/articles/{id}", s.updateArticle)
	mux.HandleFunc("POST /api/v1/articles/{id}/publish", s.publishArticle)
	mux.HandleFunc("DELETE /api/v1/articles/{id}", s.deleteArticle)
	// Reviews + Venues
	mux.HandleFunc("POST /api/v1/reviews", s.submitReview)
	mux.HandleFunc("GET /api/v1/reviews", s.listReviews)
	mux.HandleFunc("GET /api/v1/admin/reviews", s.listAllReviews)
	mux.HandleFunc("POST /api/v1/admin/reviews/{id}/approve", s.approveReview)
	mux.HandleFunc("POST /api/v1/admin/reviews/{id}/reject", s.rejectReview)
	mux.HandleFunc("DELETE /api/v1/admin/reviews/{id}", s.deleteReview)
	mux.HandleFunc("POST /api/v1/admin/users", s.createUser)
	mux.HandleFunc("GET /api/v1/admin/users", s.listUsers)
	mux.HandleFunc("DELETE /api/v1/admin/users/{id}", s.deleteUser)
	mux.HandleFunc("POST /api/v1/admin/users/{id}/role", s.updateUserRole)
	mux.HandleFunc("POST /api/v1/admin/users/{id}/password", s.resetUserPassword)
	mux.HandleFunc("POST /api/v1/venues", s.createVenue)
	mux.HandleFunc("GET /api/v1/venues", s.listVenues)
	mux.HandleFunc("POST /api/v1/venues/{id}/book", s.bookVenue)
}
