package httpapi_test

import (
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// BUG-004 回归：同一评价人对同一目标，在**待审核期间**也必须拒绝重复提交。
//
// 历史缺陷：查重走的是 ListByTarget（对外展示口径，仅 approved），而新评价是 pending，
// 于是查重永远查空 —— 生产实测同一人同一订单连发 3 次全部 201、库里落 3 条。
//
// 夹具用 enterprise 而不是 order：2026-09-22 起 order 这一类补了"订单存在 + 是本人买的 +
// 状态 completed"的校验，用不存在的 torder-* 做查重夹具会被前置校验先挡成 404
//（那是 TestReviewOrderValidationAndReviewedFlag 的事）。查重逻辑与 target 类型无关，
// 换类型不影响这条回归的覆盖。
func TestReviewDuplicateSubmitRejected(t *testing.T) {
	app := newBizServer(t)
	body := []byte(`{"target_type":"enterprise","target_id":"ent-dup-1","rating":5,"content":"好评"}`)

	first := requestAs(t, app, http.MethodPost, "/api/v1/reviews", body, "user-1", domain.RoleIndividual)
	if first.Code != http.StatusCreated {
		t.Fatalf("首次评价应成功，实际 %d %s", first.Code, first.Body.String())
	}

	second := requestAs(t, app, http.MethodPost, "/api/v1/reviews", body, "user-1", domain.RoleIndividual)
	if second.Code != http.StatusConflict {
		t.Fatalf("待审核期间的重复评价应 409，实际 %d %s", second.Code, second.Body.String())
	}

	// 换个人评价同一目标是正常业务，不能被拦
	other := requestAs(t, app, http.MethodPost, "/api/v1/reviews", body, "user-2", domain.RoleIndividual)
	if other.Code != http.StatusCreated {
		t.Fatalf("他人评价同一目标应成功，实际 %d %s", other.Code, other.Body.String())
	}

	// 换目标：同一个人可以评价另一单
	otherTarget := []byte(`{"target_type":"enterprise","target_id":"ent-dup-2","rating":4,"content":"不错"}`)
	if w := requestAs(t, app, http.MethodPost, "/api/v1/reviews", otherTarget, "user-1", domain.RoleIndividual); w.Code != http.StatusCreated {
		t.Fatalf("同人不同目标应成功，实际 %d %s", w.Code, w.Body.String())
	}
}
