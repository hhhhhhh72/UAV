package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"drone-platform/internal/domain"
)

// 内容收藏（商品/服务能力/培训课程）接口覆盖测试：toggle + 我的收藏列表 + 未登录/不存在边界。

// favIDsFromBody 从 { "data": [ {...}, ... ] } 中取全部 id。
func favIDsFromBody(t *testing.T, w *httptest.ResponseRecorder) []string {
	t.Helper()
	var out struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("parse favorites response: %v body=%s", err, w.Body.String())
	}
	var ids []string
	for _, m := range out.Data {
		if id, ok := m["id"].(string); ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func TestContentFavoritesProductServiceCourse(t *testing.T) {
	app := newBizServer(t)
	userTok := authAs(t, "user-1", domain.RoleIndividual)

	// 未登录 toggle → 401
	w := doRaw(app, http.MethodPost, "/api/v1/products/p-1/favorite", `{"favorite":true}`, "")
	checkCode(t, http.MethodPost, "/api/v1/products/{id}/favorite (anonymous)", w, http.StatusUnauthorized)

	// 收藏不存在的商品 → 404（服务层校验存在性）
	w = doRaw(app, http.MethodPost, "/api/v1/products/nope/favorite", `{"favorite":true}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/products/nope/favorite", w, http.StatusNotFound)

	// 创建商品（企业角色）→ 收藏 → 我的收藏列表包含
	prodTok := authAs(t, "seller-1", domain.RoleEnterprise)
	w = doRaw(app, http.MethodPost, "/api/v1/products",
		`{"title":"大疆M350","description":"巡检设备","brand":"DJI","model":"M350","price_fen":8800000,"images":[]}`, prodTok)
	checkCode(t, http.MethodPost, "/api/v1/products", w, http.StatusCreated)
	prodID := idFromBody(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/products/"+prodID+"/favorite", `{"favorite":true}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/products/{id}/favorite", w, http.StatusOK)

	// 创建服务能力 → 收藏
	w = doRaw(app, http.MethodPost, "/api/v1/service-listings",
		`{"title":"航测服务","category":"测绘","description":"正射影像","region":"南岸区","price_fen":500000,"unit":"次"}`, prodTok)
	checkCode(t, http.MethodPost, "/api/v1/service-listings", w, http.StatusCreated)
	slID := idFromBody(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/service-listings/"+slID+"/favorite", `{"favorite":true}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/service-listings/{id}/favorite", w, http.StatusOK)

	// 创建课程 → 收藏
	w = doRaw(app, http.MethodPost, "/api/v1/training-courses",
		`{"title":"CAAC超视距驾驶员","cert_type":"caac","description":"执照培训","price_fen":1280000,
		  "start_date":"2026-09-01","end_date":"2026-09-30","org_name":"巡航科技"}`, prodTok)
	checkCode(t, http.MethodPost, "/api/v1/training-courses", w, http.StatusCreated)
	courseID := idFromBody(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/training-courses/"+courseID+"/favorite", `{"favorite":true}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/training-courses/{id}/favorite", w, http.StatusOK)

	// 未登录我的收藏 → 401（路由不在公开白名单，与需求收藏一致，不泄露数据）
	w = doRaw(app, http.MethodGet, "/api/v1/products/favorites/mine", "", "")
	checkCode(t, http.MethodGet, "/api/v1/products/favorites/mine (anonymous)", w, http.StatusUnauthorized)

	// 我的收藏列表：三类各含收藏项
	w = doRaw(app, http.MethodGet, "/api/v1/products/favorites/mine", "", userTok)
	checkCode(t, http.MethodGet, "/api/v1/products/favorites/mine", w, http.StatusOK)
	if ids := favIDsFromBody(t, w); len(ids) != 1 || ids[0] != prodID {
		t.Fatalf("product favorites should contain %s, got %v body=%s", prodID, ids, w.Body.String())
	}

	w = doRaw(app, http.MethodGet, "/api/v1/service-listings/favorites/mine", "", userTok)
	checkCode(t, http.MethodGet, "/api/v1/service-listings/favorites/mine", w, http.StatusOK)
	if ids := favIDsFromBody(t, w); len(ids) != 1 || ids[0] != slID {
		t.Fatalf("service listing favorites should contain %s, got %v body=%s", slID, ids, w.Body.String())
	}

	w = doRaw(app, http.MethodGet, "/api/v1/training-courses/favorites/mine", "", userTok)
	checkCode(t, http.MethodGet, "/api/v1/training-courses/favorites/mine", w, http.StatusOK)
	if ids := favIDsFromBody(t, w); len(ids) != 1 || ids[0] != courseID {
		t.Fatalf("course favorites should contain %s, got %v body=%s", courseID, ids, w.Body.String())
	}

	// 取消收藏（幂等：重复取消不报错）→ 列表为空
	w = doRaw(app, http.MethodPost, "/api/v1/products/"+prodID+"/favorite", `{"favorite":false}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/products/{id}/favorite (unfavorite)", w, http.StatusOK)
	w = doRaw(app, http.MethodPost, "/api/v1/products/"+prodID+"/favorite", `{"favorite":false}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/products/{id}/favorite (unfavorite idempotent)", w, http.StatusOK)
	w = doRaw(app, http.MethodGet, "/api/v1/products/favorites/mine", "", userTok)
	checkCode(t, http.MethodGet, "/api/v1/products/favorites/mine (after unfavorite)", w, http.StatusOK)
	if ids := favIDsFromBody(t, w); len(ids) != 0 {
		t.Fatalf("product favorites should be empty after unfavorite, got %v", ids)
	}
}
