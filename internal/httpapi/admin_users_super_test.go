package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 用户列表只含真实账号。
//
// 回归：listUsers 曾硬编码一行 {"id":"admin","role":"platform_admin"} 的「超级管理员」占位——
// 数据库里没有这个账号、也没有密码，永远登录不了，却让运营以为平台有个超管，
// 还让删除保护（service 里 if id == "admin"）去保护一个幽灵。
func TestUserListHasNoPhantomAdminRow(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)
	w := doRaw(app, http.MethodGet, "/api/v1/admin/users?page=1&page_size=100", "", adminTok)
	if w.Code != http.StatusOK {
		t.Fatalf("list users: %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse list: %v", err)
	}
	if len(resp.Data) == 0 {
		t.Fatal("列表不应为空（种子用户至少 4 个）")
	}
	for _, row := range resp.Data {
		if id, _ := row["id"].(string); id == "admin" {
			t.Fatalf("列表里不应再有幽灵 admin 行：%v", row)
		}
		if _, ok := row["is_super_admin"]; !ok {
			t.Fatalf("每行都应带 is_super_admin 标记（前端据此显示超管标签/禁用按钮）：%v", row)
		}
		// 未配置 SUPER_ADMIN_PHONE 的测试服里不该冒出超管
		if v, _ := row["is_super_admin"].(bool); v {
			t.Fatalf("未配置 SUPER_ADMIN_PHONE 时不应标记超管：%v", row)
		}
	}
}
