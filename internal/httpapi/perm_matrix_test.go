package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 管理后台权限矩阵探针：对 /api/v1/admin/* 的每一条路由用 5 种身份各打一次，
// 记录真实状态码。
//
// 硬断言（越权即漏洞）：
//   - 匿名 → 必须 401/403（不能进 handler）
//   - individual / enterprise → 必须 401/403（adminGate 拦非管理员）
// 只做输出、供人工核对的：
//   - association_admin 被 403 而 platform_admin 能过的路由（= 代码里显式限定平台管理员的清单）
//   - platform_admin 也被 403 的路由（可能是业务规则 403，需要看具体语义）
//
// 备注：/api/v1/admin/token 是 dev 模式令牌签发端点，adminGate 显式豁免，单独标注。
func TestAdminPermissionMatrix(t *testing.T) {
	app := newBizServer(t)
	ids := []struct {
		name  string
		token string
	}{
		{"anonymous", ""},
		{"individual", auth(t, domain.RoleIndividual)},
		{"enterprise", auth(t, domain.RoleEnterprise)},
		{"association_admin", auth(t, domain.RoleAssociationAdmin)},
		{"platform_admin", auth(t, domain.RolePlatformAdmin)},
	}

	rows := make([]permRow, 0, len(adminRouteProbes))

	// 1) 先跑非管理员与协会管理员，最后跑平台管理员：把平台管理员造成的写副作用放到最后，
	//    避免它新建的数据影响其它身份的判定。
	for i, rt := range adminRouteProbes {
		p := permFillPath(rt.Path)
		r := permRow{Method: rt.Method, Path: rt.Path, Handler: rt.Handler, Status: map[string]int{}}
		for _, id := range ids {
			code, _ := permProbe(app, rt.Method, p, id.token, i*7+len(id.name))
			r.Status[id.name] = code
		}
		rows = append(rows, r)
	}

	// 2) 硬断言：越权即测试失败
	var violations []string
	for _, r := range rows {
		exempt := r.Path == "/api/v1/admin/token"
		if !exempt {
			if c := r.Status["anonymous"]; c != http.StatusUnauthorized && c != http.StatusForbidden {
				violations = append(violations, fmt.Sprintf("匿名可访问 %s %s → %d (%s)", r.Method, r.Path, c, r.Handler))
			}
			for _, role := range []string{"individual", "enterprise"} {
				if c := r.Status[role]; c != http.StatusUnauthorized && c != http.StatusForbidden {
					violations = append(violations, fmt.Sprintf("%s 可访问 %s %s → %d (%s)", role, r.Method, r.Path, c, r.Handler))
				}
			}
		}
	}

	// 3) 汇总：各国身份的状态码分布
	summary := map[string]map[int]int{}
	for _, id := range ids {
		summary[id.name] = map[int]int{}
		for _, r := range rows { summary[id.name][r.Status[id.name]]++ }
	}
	report := map[string]any{
		"total": len(rows),
		"summary": summary,
		"platform_only": permPlatformOnly(rows),
		"plat_403": permStatusList(rows, "platform_admin", http.StatusForbidden),
		"assoc_403": permStatusList(rows, "association_admin", http.StatusForbidden),
		"violations": violations,
		"rows": rows,
	}
	if out := os.Getenv("PERM_PROBE_OUT"); out != "" {
		b, _ := json.MarshalIndent(report, "", " ")
		if err := os.WriteFile(out, b, 0o644); err != nil { t.Fatalf("写报告失败: %v", err) }
		t.Logf("报告已写入 %s", out)
	}

	for _, id := range ids {
		codes := make([]string, 0, len(summary[id.name]))
		for c, n := range summary[id.name] { codes = append(codes, fmt.Sprintf("%d×%d", c, n)) }
		sort.Strings(codes)
		t.Logf("身份 %-18s → %s", id.name, strings.Join(codes, "  "))
	}
	t.Logf("association_admin 被 403 的路由 %d 条：", len(report["assoc_403"].([]string)))
	for _, s := range report["assoc_403"].([]string) { t.Logf("    %s", s) }
	t.Logf("platform_admin 也被 403 的路由 %d 条：", len(report["plat_403"].([]string)))
	for _, s := range report["plat_403"].([]string) { t.Logf("    %s", s) }

	if len(violations) > 0 {
		for _, v := range violations { t.Errorf("越权：%s", v) }
	}
}

// permFillPath 把 {id}/{resource} 这类路径参数换成探针值。
func permFillPath(p string) string {
	parts := strings.Split(p, "/")
	for i, s := range parts {
		if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
			if strings.Trim(s, "{}") == "resource" {
				parts[i] = "demands"
			} else {
				parts[i] = "perm-probe"
			}
		}
	}
	return strings.Join(parts, "/")
}

// permProbe 发一次请求；RemoteAddr 每条都不同，避开 100/s 的按 IP 限流对矩阵的干扰。
func permProbe(app http.Handler, method, path, token string, seq int) (int, string) {
	var body []byte
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		body = []byte("{}")
	}
	r := httptest.NewRequest(method, path, bytes.NewReader(body))
	r.RemoteAddr = fmt.Sprintf("198.51.100.%d:1234", seq%250+1)
	if token != "" {
		r.Header.Set("Authorization", token)
	}
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w.Code, w.Body.String()
}

// permRow 一条路由在 5 种身份下的真实状态码。
type permRow struct {
	Method  string         `json:"method"`
	Path    string         `json:"path"`
	Handler string         `json:"handler"`
	Status  map[string]int `json:"status"`
}

// permPlatformOnly 返回「协会管理员 403 但平台管理员不是 403」的路由：代码里显式限定平台管理员的清单。
func permPlatformOnly(rows []permRow) []string {
	out := []string{}
	for _, r := range rows {
		if r.Status["association_admin"] == http.StatusForbidden && r.Status["platform_admin"] != http.StatusForbidden {
			out = append(out, fmt.Sprintf("%s %s → %s", r.Method, r.Path, r.Handler))
		}
	}
	sort.Strings(out)
	return out
}

// permStatusList 返回某身份拿到指定状态码的路由。
func permStatusList(rows []permRow, who string, code int) []string {
	out := []string{}
	for _, r := range rows {
		if r.Status[who] == code {
			out = append(out, fmt.Sprintf("%s %s → %s", r.Method, r.Path, r.Handler))
		}
	}
	sort.Strings(out)
	return out
}
// TestAdminRouteProbeListFreshness 保证路由清单不会悄悄过期：
// 权限矩阵的价值取决于「清单 = 真实注册的路由」。任何人新增 /api/v1/admin/* 路由
// 都必须重新生成清单，否则新路由不会进入权限回归覆盖（绿灯是假的）。
func TestAdminRouteProbeListFreshness(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("读取包目录: %v", err)
	}
	re := regexp.MustCompile(`mux\.HandleFunc\(\s*"([A-Z]+)\s+(/api/v1/admin/[^"]+)"`)
	live := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		b, err := os.ReadFile(e.Name())
		if err != nil {
			t.Fatalf("读取 %s: %v", e.Name(), err)
		}
		for _, m := range re.FindAllStringSubmatch(string(b), -1) {
			live[m[1]+" "+m[2]] = true
		}
	}
	listed := map[string]bool{}
	for _, rt := range adminRouteProbes {
		listed[rt.Method+" "+rt.Path] = true
	}
	for k := range live {
		if !listed[k] {
			t.Errorf("路由清单过期：%s 已注册但不在 adminRouteProbes 里，请重新生成（.tools/gen-perm-routes.cjs）", k)
		}
	}
	for k := range listed {
		if !live[k] {
			t.Errorf("路由清单过期：%s 在清单里但已不存在，请重新生成（.tools/gen-perm-routes.cjs）", k)
		}
	}
}
