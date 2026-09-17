package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 培训报名线的站内通知。
//
// 回归背景：这条线此前**一处通知都没有** —— 报名成功、审核结果、管理端改状态、结业
// 全部静默，而企业入驻/需求/接单/求职/证书到期都有。学员只能自己反复去「我的报名」看进度；
// 付费报名还冻结了学费，连一张「钱动了」的凭证都没有。
//
// 通知是异步发的（s.notify 起 goroutine），所以用 waitMessage 轮询等待。
func TestEnrollmentNotifications(t *testing.T) {
	app := newBizServer(t)

	// 1. 管理员建一门已公开的免费课（免费省掉托管金充值；OrgID = 建课人 admin-1，
	//    这样机构那一侧的通知也能一并验到）
	dw := requestAs(t, app, http.MethodPost, "/api/v1/admin/training-courses",
		[]byte(`{"title":"通知测试课","price_fen":0,"status":"published","cert_type":"caac"}`),
		"admin-1", domain.RolePlatformAdmin)
	if dw.Code != http.StatusCreated {
		t.Fatalf("create course: %d %s", dw.Code, dw.Body.String())
	}
	var course struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(dw.Body.Bytes(), &course); err != nil {
		t.Fatalf("parse course: %v", err)
	}
	courseID := course.Data.ID

	// 2. 学员报名 → 学员收到「报名成功」，机构收到「新的报名待审核」
	ew := requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+courseID+"/enroll",
		[]byte(`{"name":"测试学员","phone":"13900000000","id_card":"500101199001011234"}`),
		"user-1", domain.RoleIndividual)
	if ew.Code != http.StatusCreated {
		t.Fatalf("enroll: %d %s", ew.Code, ew.Body.String())
	}
	waitMessage(t, app, "user-1", domain.RoleIndividual, "报名成功")
	waitMessage(t, app, "admin-1", domain.RolePlatformAdmin, "新的报名待审核")

	var enroll struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(ew.Body.Bytes(), &enroll); err != nil {
		t.Fatalf("parse enrollment: %v", err)
	}
	enrollID := enroll.Data.ID

	// 3. 机构审核通过 → 学员收到「报名审核结果」
	rw := requestAs(t, app, http.MethodPost, "/api/v1/enrollments/"+enrollID+"/review",
		[]byte(`{"action":"approve"}`),
		"admin-1", domain.RolePlatformAdmin)
	if rw.Code != http.StatusOK {
		t.Fatalf("review: %d %s", rw.Code, rw.Body.String())
	}
	waitMessage(t, app, "user-1", domain.RoleIndividual, "报名审核结果")

	// 4. 管理端再改状态 → 学员收到「报名状态更新」（approved → enrolled 是白名单内允许的）
	uw := requestAs(t, app, http.MethodPut, "/api/v1/admin/enrollments/"+enrollID,
		[]byte(`{"name":"测试学员","phone":"13900000000","status":"enrolled"}`),
		"admin-1", domain.RolePlatformAdmin)
	if uw.Code != http.StatusOK {
		t.Fatalf("update enrollment: %d %s", uw.Code, uw.Body.String())
	}
	waitMessage(t, app, "user-1", domain.RoleIndividual, "报名状态更新")

	// 5. 结业 → 学员收到「培训结业」（免费课无学费释放，但证书照发）
	cw := requestAs(t, app, http.MethodPost, "/api/v1/enrollments/"+enrollID+"/complete", nil,
		"admin-1", domain.RolePlatformAdmin)
	if cw.Code != http.StatusOK {
		t.Fatalf("complete: %d %s", cw.Code, cw.Body.String())
	}
	waitMessage(t, app, "user-1", domain.RoleIndividual, "培训结业")
}

// 付费报名被驳回：学费必须自动退回学员可用余额，通知也要如实说明。
//
// 回归背景：Review 此前全程不碰 escrow，而学员那份冻结**没有任何入口**能解开——
// 驳回的报名不能结业（completeEnrollment 只收 enrolled/paid/approved），
// 孤儿冻结补偿要求业务行不存在（行还在），POST /api/v1/escrow/refund 退的
// 又是操作者自己的冻结。学员看到「未通过审核」，学费却永久冻结。
func TestRejectedPaidEnrollmentRefundsFrozenTuition(t *testing.T) {
	app := newBizServer(t)

	const price = int64(300000)
	courseID := createPublishedCourse(t, app, "org-1", "退款口径测试课", price)
	fundEscrow(t, app, "student-1", 500000)

	w := requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+courseID+"/pay-and-enroll",
		[]byte(`{"name":"学员甲","phone":"13800000001"}`),
		"student-1", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay-and-enroll", w, http.StatusCreated)
	enrollID := dataID(t, w)

	// 报名后：学费从可用余额转入冻结
	if b, f := escrowBal(t, app, "student-1", domain.RoleIndividual); b != 200000 || f != 300000 {
		t.Fatalf("报名后应冻结学费：balance=%d frozen=%d want 200000/300000", b, f)
	}

	rw := requestAs(t, app, http.MethodPost, "/api/v1/enrollments/"+enrollID+"/review",
		[]byte(`{"action":"reject","reason":"资料不全"}`),
		"admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPost, ".../review(reject)", rw, http.StatusOK)

	// 驳回后：冻结清零，学费回到可用余额（既没被吞掉，也没被释放给机构）
	if b, f := escrowBal(t, app, "student-1", domain.RoleIndividual); b != 500000 || f != 0 {
		t.Fatalf("驳回后学费应退回可用余额：balance=%d frozen=%d want 500000/0", b, f)
	}

	msg := waitMessage(t, app, "student-1", domain.RoleIndividual, "报名审核结果")
	content, _ := msg["content"].(string)
	if !strings.Contains(content, "已退回平台账户余额") {
		t.Fatalf("驳回通知必须说明学费已退回，实际：%q", content)
	}
	if !strings.Contains(content, "¥3000.00") {
		t.Fatalf("驳回通知必须写出退款金额，实际：%q", content)
	}

	// rejected 是终态：改回 approved 必须被拒——否则「钱退了却通过审核」，
	// 之后结业释放学费时冻结里已无这笔钱，永久 500
	uw := requestAs(t, app, http.MethodPut, "/api/v1/admin/enrollments/"+enrollID,
		[]byte(`{"name":"学员甲","phone":"13800000001","status":"approved"}`),
		"admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPut, ".../enrollment(approved)", uw, http.StatusBadRequest)
}

// 自动退款失败时的人工补救入口：管理员带 user_id 解冻**指定用户**的资金。
//
// 场景：reviewEnrollment 驳回后的自动退款那一笔若失败（余额被并发改动等），
// 状态已是 rejected、钱却还冻着，学员和机构都没有入口能解开——
// 这个带 user_id 的退款接口是唯一的路（此前它只能退操作者自己的冻结）。
func TestAdminCanRefundAnotherUsersFrozenFunds(t *testing.T) {
	app := newBizServer(t)

	const price = int64(300000)
	courseID := createPublishedCourse(t, app, "org-1", "人工补退测试课", price)
	fundEscrow(t, app, "student-1", 500000)

	w := requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+courseID+"/pay-and-enroll",
		[]byte(`{"name":"学员甲","phone":"13800000001"}`),
		"student-1", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay-and-enroll", w, http.StatusCreated)

	if b, f := escrowBal(t, app, "student-1", domain.RoleIndividual); b != 200000 || f != 300000 {
		t.Fatalf("前置：应冻结 300000，实际 balance=%d frozen=%d", b, f)
	}

	// 非管理员不能借 user_id 退别人的钱
	fw := requestAs(t, app, http.MethodPost, "/api/v1/escrow/refund",
		[]byte(`{"user_id":"student-1","amount_fen":300000,"reference_type":"training_course","reference_id":"`+courseID+`"}`),
		"user-1", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../escrow/refund(非管理员)", fw, http.StatusForbidden)

	// 管理员指定 user_id → 解冻回到学员可用余额
	rw := requestAs(t, app, http.MethodPost, "/api/v1/escrow/refund",
		[]byte(`{"user_id":"student-1","amount_fen":300000,"reference_type":"training_course","reference_id":"`+courseID+`"}`),
		"admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPost, ".../escrow/refund(user_id)", rw, http.StatusCreated)

	if b, f := escrowBal(t, app, "student-1", domain.RoleIndividual); b != 500000 || f != 0 {
		t.Fatalf("人工补退后：balance=%d frozen=%d want 500000/0", b, f)
	}
}
