package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

func parseArticle(t *testing.T, w *httptest.ResponseRecorder) domain.Article {
	t.Helper()
	var out struct {
		Data domain.Article `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("parse article response: %v (%s)", err, w.Body.String())
	}
	return out.Data
}

func parseArticleItems(t *testing.T, w *httptest.ResponseRecorder) []domain.Article {
	t.Helper()
	var out struct {
		Data []domain.Article `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("parse article list: %v (%s)", err, w.Body.String())
	}
	return out.Data
}

// TestNewsAdminAuthorPinnedRoundTrip 对齐：create/update 必须接收 author/is_pinned，
// 摘要必须为纯文本（富文本正文不得把 <p> 残缺标签带出）。
func TestNewsAdminAuthorPinnedRoundTrip(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-2", domain.RoleAssociationAdmin)

	body := `{"title":"低空新政要点","content":"<p>政策<strong>要点</strong>：空域开放。</p>","category":"low_altitude_policy","source":"协会","author":"产业发展部","is_pinned":true}`
	w := doRaw(app, http.MethodPost, "/api/v1/articles", body, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/articles", w, http.StatusCreated)
	art := parseArticle(t, w)
	if art.Author != "产业发展部" {
		t.Fatalf("create author: got %q, want 产业发展部", art.Author)
	}
	if !art.IsPinned {
		t.Fatalf("create is_pinned: got false, want true")
	}
	if strings.Contains(art.Summary, "<") {
		t.Fatalf("summary must be plain text, got %q", art.Summary)
	}
	if !strings.Contains(art.Summary, "政策") || !strings.Contains(art.Summary, "要点") {
		t.Fatalf("summary should keep stripped text, got %q", art.Summary)
	}

	// 管理端列表往返
	w = doRaw(app, http.MethodGet, "/api/v1/admin/articles", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/admin/articles", w, http.StatusOK)
	found := false
	for _, it := range parseArticleItems(t, w) {
		if it.ID == art.ID {
			found = true
			if it.Author != "产业发展部" || !it.IsPinned {
				t.Fatalf("admin list roundtrip: author=%q is_pinned=%v", it.Author, it.IsPinned)
			}
		}
	}
	if !found {
		t.Fatalf("created article %s not in admin list", art.ID)
	}

	// 更新：作者/置顶可改，摘要重新生成
	w = doRaw(app, http.MethodPut, "/api/v1/articles/"+art.ID,
		`{"title":"低空新政要点","content":"<p>修订稿</p>","category":"low_altitude_policy","source":"协会","author":"理事会","is_pinned":false}`, adminTok)
	assertStatus(t, http.MethodPut, "/api/v1/articles/"+art.ID, w, http.StatusOK)
	upd := parseArticle(t, w)
	if upd.Author != "理事会" || upd.IsPinned {
		t.Fatalf("update: author=%q is_pinned=%v, want 理事会/false", upd.Author, upd.IsPinned)
	}
	if strings.Contains(upd.Summary, "<") || upd.Summary != "修订稿" {
		t.Fatalf("update summary: got %q, want 修订稿", upd.Summary)
	}
}

// TestNewsPinnedFirstPublic 对齐：公开列表置顶优先（is_pinned DESC, created_at DESC）。
func TestNewsPinnedFirstPublic(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-2", domain.RoleAssociationAdmin)

	w := doRaw(app, http.MethodPost, "/api/v1/articles",
		`{"title":"普通资讯","content":"普通正文A","category":"low_altitude_policy","source":"协会"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/articles", w, http.StatusCreated)
	a := parseArticle(t, w)
	doRaw(app, http.MethodPost, "/api/v1/articles/"+a.ID+"/publish", "", adminTok)

	w = doRaw(app, http.MethodPost, "/api/v1/articles",
		`{"title":"置顶资讯","content":"置顶正文B","category":"low_altitude_policy","source":"协会","is_pinned":true}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/articles", w, http.StatusCreated)
	b := parseArticle(t, w)
	doRaw(app, http.MethodPost, "/api/v1/articles/"+b.ID+"/publish", "", adminTok)

	w = doRaw(app, http.MethodGet, "/api/v1/articles", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/articles", w, http.StatusOK)
	pub := parseArticleItems(t, w)
	if len(pub) < 2 {
		t.Fatalf("want >=2 published articles, got %d", len(pub))
	}
	if pub[0].ID != b.ID {
		t.Fatalf("pinned first: got %s (title=%s), want %s", pub[0].ID, pub[0].Title, b.ID)
	}
	if pub[1].ID != a.ID {
		t.Fatalf("second should be non-pinned %s, got %s", a.ID, pub[1].ID)
	}
}

// TestNewsDelete 对齐：管理员可删除（草稿/已发布）；非管理员 403；重复删除 404。
func TestNewsDelete(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-2", domain.RoleAssociationAdmin)
	userTok := authAs(t, "enterprise-1", domain.RoleEnterprise)

	w := doRaw(app, http.MethodPost, "/api/v1/articles",
		`{"title":"待删资讯","content":"正文","category":"low_altitude_policy","source":"协会"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/articles", w, http.StatusCreated)
	art := parseArticle(t, w)

	w = doRaw(app, http.MethodDelete, "/api/v1/articles/"+art.ID, "", userTok)
	assertStatus(t, http.MethodDelete, "/api/v1/articles/"+art.ID, w, http.StatusForbidden)

	w = doRaw(app, http.MethodDelete, "/api/v1/articles/"+art.ID, "", adminTok)
	assertStatus(t, http.MethodDelete, "/api/v1/articles/"+art.ID, w, http.StatusOK)

	w = doRaw(app, http.MethodDelete, "/api/v1/articles/"+art.ID, "", adminTok)
	assertStatus(t, http.MethodDelete, "/api/v1/articles/"+art.ID, w, http.StatusNotFound)

	w = doRaw(app, http.MethodGet, "/api/v1/admin/articles", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/admin/articles", w, http.StatusOK)
	for _, it := range parseArticleItems(t, w) {
		if it.ID == art.ID {
			t.Fatalf("deleted article %s still listed", art.ID)
		}
	}
}
