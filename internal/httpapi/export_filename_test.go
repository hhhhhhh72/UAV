package httpapi_test

import (
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// CSV 导出的下载文件名要用中文「表名」：
//  1. 老客户端读 ASCII 的 filename（不丢兼容）；
//  2. 浏览器/Excel 读 RFC 5987 的 filename*=UTF-8''…（中文名，如「需求管理」）；
//  3. 未收录的资源退回资源名，不编造中文名。
func TestExportChineseFilename(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	cases := []struct {
		path string
		want string // 期望出现在 filename* 里的 URL 编码中文名
	}{
		{"/api/v1/admin/export/demands", "%E9%9C%80%E6%B1%82%E7%AE%A1%E7%90%86_"},
		{"/api/v1/admin/export/enterprises", "%E4%BC%81%E4%B8%9A%E7%AE%A1%E7%90%86_"},
		{"/api/v1/admin/export/competitions", "%E8%B5%9B%E4%BA%8B%E7%AE%A1%E7%90%86_"},
		{"/api/v1/admin/export/training-courses", "%E5%9F%B9%E8%AE%AD%E8%AF%BE%E7%A8%8B_"},
		{"/api/v1/admin/export/certified-pilots", "%E9%A3%9E%E6%89%8B%E8%AE%A4%E8%AF%81_"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			w := doRaw(app, http.MethodGet, tc.path, "", adminTok)
			if w.Code != http.StatusOK {
				t.Fatalf("export %s: %d %s", tc.path, w.Code, w.Body.String())
			}
			cd := w.Header().Get("Content-Disposition")
			if !strings.Contains(cd, "filename*=UTF-8''"+tc.want) {
				t.Fatalf("Content-Disposition 缺少中文表名：%q", cd)
			}
			if !strings.Contains(cd, ".csv") {
				t.Fatalf("Content-Disposition 缺少 .csv 后缀：%q", cd)
			}
		})
	}
}
