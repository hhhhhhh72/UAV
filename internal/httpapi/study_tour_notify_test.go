package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"drone-platform/internal/domain"
)

// countMessages 统计某人收到的某标题通知条数（先等一小段，让异步通知落地）。
func countMessages(t *testing.T, app http.Handler, userID string, role domain.Role, title string) int {
	t.Helper()
	time.Sleep(150 * time.Millisecond)
	w := requestAs(t, app, http.MethodGet, "/api/v1/messages", nil, userID, role)
	if w.Code != http.StatusOK {
		t.Fatalf("list messages for %s: %d %s", userID, w.Code, w.Body.String())
	}
	var out struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("parse messages: %v", err)
	}
	n := 0
	for _, m := range out.Data {
		if m["title"] == title {
			n++
		}
	}
	return n
}

// 研学报名线的站内通知（与培训报名同款回归）。
//
// 回归背景：研学报名此前**一条通知都没有**——报名成功与管理端审核结果全程静默，
// 而学员正是在「我的报名-研学」tab 里看进度的（myenrollments.vue），只能自己反复刷新。
// 通知是异步发的（s.notify 起 goroutine），所以用 waitMessage 轮询等待。
func TestStudyTourEnrollmentNotifications(t *testing.T) {
	app := newBizServer(t)

	// 1. 管理员建一个招募中的研学活动（status=active 才允许报名）
	cw := requestAs(t, app, http.MethodPost, "/api/v1/admin/study-tours",
		[]byte(`{"title":"研学通知测试","status":"active","capacity":10,"start_date":"2026-10-01","end_date":"2026-10-03"}`),
		"admin-1", domain.RolePlatformAdmin)
	if cw.Code != http.StatusCreated {
		t.Fatalf("create study tour: %d %s", cw.Code, cw.Body.String())
	}
	var tour struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(cw.Body.Bytes(), &tour); err != nil {
		t.Fatalf("parse study tour: %v", err)
	}
	tourID := tour.Data.ID
	if tourID == "" {
		t.Fatalf("study tour id 为空: %s", cw.Body.String())
	}

	// 2. 学员报名 → 收到「研学报名成功」
	ew := requestAs(t, app, http.MethodPost, "/api/v1/study/tours/"+tourID+"/enroll",
		[]byte(`{"name":"测试学员","phone":"13900000000","adult_count":1,"child_count":0}`),
		"user-1", domain.RoleIndividual)
	if ew.Code != http.StatusCreated {
		t.Fatalf("enroll: %d %s", ew.Code, ew.Body.String())
	}
	waitMessage(t, app, "user-1", domain.RoleIndividual, "研学报名成功")

	var enroll struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(ew.Body.Bytes(), &enroll); err != nil {
		t.Fatalf("parse enrollment: %v", err)
	}
	enrollID := enroll.Data.ID

	// 3. 管理端审核通过 → 学员收到「研学报名审核结果」
	rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/study-tours/enrollments/"+enrollID+"/review",
		[]byte(`{"status":"approved"}`),
		"admin-1", domain.RolePlatformAdmin)
	if rw.Code != http.StatusOK {
		t.Fatalf("review: %d %s", rw.Code, rw.Body.String())
	}
	waitMessage(t, app, "user-1", domain.RoleIndividual, "研学报名审核结果")

	// 4. 幂等重复审核（同状态）不得再发一条一模一样的通知
	rw2 := requestAs(t, app, http.MethodPost, "/api/v1/admin/study-tours/enrollments/"+enrollID+"/review",
		[]byte(`{"status":"approved"}`),
		"admin-1", domain.RolePlatformAdmin)
	if rw2.Code != http.StatusOK {
		t.Fatalf("idempotent review: %d %s", rw2.Code, rw2.Body.String())
	}
	if n := countMessages(t, app, "user-1", domain.RoleIndividual, "研学报名审核结果"); n != 1 {
		t.Fatalf("幂等重审不该重复通知：want 1 条，got %d", n)
	}
}
