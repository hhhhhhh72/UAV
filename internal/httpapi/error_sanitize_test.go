package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"drone-platform/internal/repository"
)

// TestSanitizeErrorMessageStripsStorageText 用例里的字符串是**从生产响应体里抄下来的原文**：
// 2026-09-18 用不存在的 id 打全量详情端点，42 个端点把 pgx 的错误串原样回给了客户端。
func TestSanitizeErrorMessageStripsStorageText(t *testing.T) {
	leaked := []string{
		"no rows in result set",
		"find message msg-1789462427870439304-3: no rows in result set",
		"product zzz-nonexistent-id not found: no rows in result set",
		"find work order zzz-nonexistent-id: no rows in result set",
		"course zzz-nonexistent-id: no rows in result set",
		"ERROR: duplicate key value violates unique constraint \"users_wechat_openid_key\" (SQLSTATE 23505)",
		"pq: relation \"users\" does not exist",
		"pgx: column \"foo\" does not exist",
		"dial tcp 127.0.0.1:5432: connect: connection refused",
		"context deadline exceeded",
	}
	for _, s := range leaked {
		got := sanitizeErrorMessage(s)
		if got == s {
			t.Errorf("存储层文本未被脱敏: %q", s)
		}
		low := strings.ToLower(got)
		for _, m := range storageErrorMarkers {
			if strings.Contains(low, m) {
				t.Errorf("脱敏后仍含特征词 %q：%q → %q", m, s, got)
			}
		}
	}

	// 业务文案必须原样保留：把「金额需在…之间」「未开通」这类可操作提示糊成
	// 通用错误，等于把可诊断的信息也一起丢了。
	keep := []string{
		"充值金额需在 100~5000000 分之间",
		"微信支付未开通",
		"商品不存在、不属于你，或当前状态不允许该操作",
		"amount_fen > 0 required",
		"只有订单双方可以查看",
	}
	for _, s := range keep {
		if got := sanitizeErrorMessage(s); got != s {
			t.Errorf("业务文案被误改：%q → %q", s, got)
		}
	}
}

// TestFailDoesNotLeakStorageTextOn4xx 回归的核心：fail() 的脱敏曾经只覆盖 5xx，
// 4xx 会把 pgx 原文原样透传。这里直接打 404 路径验证。
func TestFailDoesNotLeakStorageTextOn4xx(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusBadRequest, http.StatusForbidden, http.StatusConflict} {
		rec := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/messages/x", nil)
		fail(rec, r, status, errors.New("find message x: no rows in result set"))
		if rec.Code != status {
			t.Fatalf("状态码被改动：want %d got %d", status, rec.Code)
		}
		var body struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("解析响应: %v (%s)", err, rec.Body.String())
		}
		if strings.Contains(strings.ToLower(body.Error.Message), "no rows") {
			t.Fatalf("%d 仍然泄漏存储层文本：%s", status, rec.Body.String())
		}
		if body.Error.Message != "记录不存在或操作无法完成" {
			t.Fatalf("%d 应回中性文案，实际 %q", status, body.Error.Message)
		}
	}
}

// TestMutationErrorCodeClassifies 归属/不存在/故障三态必须分得开。
// 工单详情此前把它们全硬编码成 403 —— 数据库故障会显示成「无权限」，排障时严重误导。
func TestMutationErrorCodeClassifies(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"仓储翻译过的未找到", repository.ErrNotFound, http.StatusNotFound},
		{"pgx 原文未翻译", errors.New("find work order wo-1: no rows in result set"), http.StatusNotFound},
		{"英文 not found", errors.New("demand wo-1: record not found"), http.StatusNotFound},
		{"中文不存在", errors.New("研学项目不存在"), http.StatusNotFound},
		{"英文归属拒绝", errors.New("only the owner can edit"), http.StatusForbidden},
		{"中文归属拒绝", errors.New("只有订单双方可以查看"), http.StatusForbidden},
		{"权限不足", errors.New("admin permission required"), http.StatusForbidden},
		{"数据库故障", errors.New("dial tcp: connection refused"), http.StatusInternalServerError},
		{"空错误", nil, http.StatusInternalServerError},
	}
	for _, c := range cases {
		if got := mutationErrorCode(c.err); got != c.want {
			t.Errorf("%s：want %d got %d（err=%v）", c.name, c.want, got, c.err)
		}
	}
}
