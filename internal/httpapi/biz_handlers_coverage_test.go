package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 本文件为 httpapi 层低覆盖文件（community_listings_labour.go / escrow.go /
// upload.go / work_order.go / biz_handlers.go 未覆盖端点）的覆盖率补齐测试。
// 复用同包既有基建：newBizServer / authAs / doRaw / request。

// checkCode 断言响应状态码，失败时输出 method/path/code/body（前 200 字符）。
func checkCode(t *testing.T, method, path string, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		body := w.Body.String()
		if len(body) > 200 {
			body = body[:200]
		}
		t.Fatalf("%s %s: want %d, got %d, body: %s", method, path, want, w.Code, body)
	}
}

// idFromBody 从 { "data": { "id": ... } } 响应中取出 data.id。
func idFromBody(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var out struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("parse id from %s: %v", w.Body.String(), err)
	}
	if out.Data.ID == "" {
		t.Fatalf("empty id in response: %s", w.Body.String())
	}
	return out.Data.ID
}

// ---- 社区（posts / comments / reports） ----

func TestCommunityPostsPublishCommentsReports(t *testing.T) {
	app := newBizServer(t)
	authorTok := authAs(t, "user-1", domain.RoleIndividual)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 匿名发帖 → 401（POST /api/v1/posts 不在公开白名单）
	w := doRaw(app, http.MethodPost, "/api/v1/posts", `{"title":"匿名帖","content":"x"}`, "")
	checkCode(t, http.MethodPost, "/api/v1/posts (anonymous)", w, http.StatusUnauthorized)

	// 带 author token 发帖 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/posts",
		`{"title":"巡检经验","content":"分享作业心得","images":["/uploads/a.jpg"]}`, authorTok)
	checkCode(t, http.MethodPost, "/api/v1/posts", w, http.StatusCreated)
	postID := idFromBody(t, w)

	// 帖子列表 → 200（含刚发布的帖子）
	w = doRaw(app, http.MethodGet, "/api/v1/posts", "", authorTok)
	checkCode(t, http.MethodGet, "/api/v1/posts", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), postID) {
		t.Fatalf("GET /api/v1/posts should contain created post %s: %s", postID, w.Body.String())
	}

	// 管理员发布 → 200
	w = doRaw(app, http.MethodPost, "/api/v1/posts/"+postID+"/publish", "", adminTok)
	checkCode(t, http.MethodPost, "/api/v1/posts/{id}/publish", w, http.StatusOK)

	// 评论 → 201（createComment 从 body 读 post_id）
	w = doRaw(app, http.MethodPost, "/api/v1/posts/"+postID+"/comments",
		`{"post_id":"`+postID+`","content":"好帖"}`, authorTok)
	checkCode(t, http.MethodPost, "/api/v1/posts/{id}/comments", w, http.StatusCreated)

	// 评论列表 → 200
	w = doRaw(app, http.MethodGet, "/api/v1/comments?post_id="+postID, "", authorTok)
	checkCode(t, http.MethodGet, "/api/v1/comments", w, http.StatusOK)

	// 举报 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/reports",
		`{"resource_type":"post","resource_id":"`+postID+`","reason":"违规内容"}`, authorTok)
	checkCode(t, http.MethodPost, "/api/v1/reports", w, http.StatusCreated)

	// 管理员举报列表 → 200
	w = doRaw(app, http.MethodGet, "/api/v1/admin/reports", "", adminTok)
	checkCode(t, http.MethodGet, "/api/v1/admin/reports", w, http.StatusOK)
}

// ---- 二手（listings）/ 用工（labour-orders） ----

func TestCommunityListingsLabour(t *testing.T) {
	app := newBizServer(t)
	sellerTok := authAs(t, "user-1", domain.RoleEnterprise)
	quoterTok := authAs(t, "user-2", domain.RoleIndividual)

	// 发布二手 → 201
	w := doRaw(app, http.MethodPost, "/api/v1/listings",
		`{"title":"二手M300","description":"九成新","category":"drone","price_fen":800000}`, sellerTok)
	checkCode(t, http.MethodPost, "/api/v1/listings", w, http.StatusCreated)
	listingID := idFromBody(t, w)

	// 收藏 → 200
	w = doRaw(app, http.MethodPost, "/api/v1/listings/"+listingID+"/favorites", "", sellerTok)
	checkCode(t, http.MethodPost, "/api/v1/listings/{id}/favorites", w, http.StatusOK)

	// 发布用工订单 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/labour-orders",
		`{"title":"航测外业","description":"需要两名飞手","worker_count":2,"start_date":"2026-08-20T00:00:00Z","end_date":"2026-08-22T00:00:00Z","budget_fen":500000}`, sellerTok)
	checkCode(t, http.MethodPost, "/api/v1/labour-orders", w, http.StatusCreated)
	orderID := idFromBody(t, w)

	// 飞手报价 → 201（createLabourQuote 从 body 读 order_id）
	w = doRaw(app, http.MethodPost, "/api/v1/labour-orders/"+orderID+"/quote",
		`{"order_id":"`+orderID+`","amount_fen":400000,"proposal":"可承接","quoter_name":"飞手李"}`, quoterTok)
	checkCode(t, http.MethodPost, "/api/v1/labour-orders/{id}/quote", w, http.StatusCreated)

	// 雇主查报价 → 200（需 order_id 查询参数）
	w = doRaw(app, http.MethodGet, "/api/v1/labour-orders/quotes?order_id="+orderID, "", sellerTok)
	checkCode(t, http.MethodGet, "/api/v1/labour-orders/quotes", w, http.StatusOK)
}

// ---- 托管（escrow） ----

func TestEscrowFullCycle(t *testing.T) {
	app := newBizServer(t)
	userTok := authAs(t, "user-1", domain.RoleIndividual)

	// 反向：amount_fen<=0 → 400
	w := doRaw(app, http.MethodPost, "/api/v1/escrow/deposit", `{"amount_fen":0}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/escrow/deposit (zero)", w, http.StatusBadRequest)

	// 充值 → 201（handler 返回 Created）
	w = doRaw(app, http.MethodPost, "/api/v1/escrow/deposit", `{"amount_fen":100000}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/escrow/deposit", w, http.StatusCreated)

	// 余额 → 200
	w = doRaw(app, http.MethodGet, "/api/v1/escrow/balance", "", userTok)
	checkCode(t, http.MethodGet, "/api/v1/escrow/balance", w, http.StatusOK)

	// 冻结 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/escrow/freeze",
		`{"amount_fen":50000,"reference_type":"work_order","reference_id":"wo-1"}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/escrow/freeze", w, http.StatusCreated)

	// 解冻（release 给 user-2）→ 201
	w = doRaw(app, http.MethodPost, "/api/v1/escrow/release",
		`{"to_user":"user-2","amount_fen":50000,"reference_type":"work_order","reference_id":"wo-1"}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/escrow/release", w, http.StatusCreated)

	// 再次冻结 → 退款（refund 回到可用余额）→ 201
	w = doRaw(app, http.MethodPost, "/api/v1/escrow/freeze",
		`{"amount_fen":30000,"reference_type":"work_order","reference_id":"wo-2"}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/escrow/freeze (2nd)", w, http.StatusCreated)
	w = doRaw(app, http.MethodPost, "/api/v1/escrow/refund",
		`{"amount_fen":30000,"reference_type":"work_order","reference_id":"wo-2"}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/escrow/refund", w, http.StatusCreated)

	// 交易流水 → 200
	w = doRaw(app, http.MethodGet, "/api/v1/escrow/transactions", "", userTok)
	checkCode(t, http.MethodGet, "/api/v1/escrow/transactions", w, http.StatusOK)
}

// ---- 上传（upload） ----

// pngBytes 生成 1x1 真 PNG 字节（魔数校验会拒绝文本伪装）。
func pngBytes(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// multipartBody 构造 multipart/form-data 请求体（字段名 file）。
func multipartBody(t *testing.T, filename, contentType string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename=%q`, filename))
	h.Set("Content-Type", contentType)
	part, err := mw.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf, mw.FormDataContentType()
}

func TestUploadPNGAndMagicValidation(t *testing.T) {
	app := newBizServer(t)
	tok := authAs(t, "user-1", domain.RoleIndividual)

	// 1. 真 PNG + token → 200，url 前缀 /uploads/
	body, ct := multipartBody(t, "real.png", "image/png", pngBytes(t))
	r := httptest.NewRequest(http.MethodPost, "/api/v1/upload", body)
	r.Header.Set("Content-Type", ct)
	r.Header.Set("Authorization", tok)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	checkCode(t, http.MethodPost, "/api/v1/upload", w, http.StatusOK)
	var out struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("parse upload: %v", err)
	}
	if !strings.HasPrefix(out.Data.URL, "/uploads/") {
		t.Fatalf("upload url %q, want /uploads/ prefix", out.Data.URL)
	}

	// 2. 文本伪装成 image/png → 400（魔数校验）
	body, ct = multipartBody(t, "fake.png", "image/png", []byte("not a real png, just text"))
	r = httptest.NewRequest(http.MethodPost, "/api/v1/upload", body)
	r.Header.Set("Content-Type", ct)
	r.Header.Set("Authorization", tok)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	checkCode(t, http.MethodPost, "/api/v1/upload (fake png)", w, http.StatusBadRequest)

	// 3. 无 token → 401
	body, ct = multipartBody(t, "anon.png", "image/png", pngBytes(t))
	r = httptest.NewRequest(http.MethodPost, "/api/v1/upload", body)
	r.Header.Set("Content-Type", ct)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	checkCode(t, http.MethodPost, "/api/v1/upload (anonymous)", w, http.StatusUnauthorized)
}

// ---- 工单闭环（HTTP 全链路） ----

func TestWorkOrderHTTPFullCycle(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "enterprise-1", domain.RoleEnterprise)
	workerTok := authAs(t, "worker-1", domain.RoleIndividual)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 1. 企业建需求 → 201
	w := doRaw(app, http.MethodPost, "/api/v1/demands",
		`{"title":"电力巡检","contact":"13800000000","district":"渝北区","biz_type":"cable_inspection","description":"50km 巡检"}`, entTok)
	checkCode(t, http.MethodPost, "/api/v1/demands", w, http.StatusCreated)
	demandID := idFromBody(t, w)

	// 2. 管理员审批通过（/approve）→ 200
	w = doRaw(app, http.MethodPost, "/api/v1/admin/demands/"+demandID+"/approve", "", adminTok)
	checkCode(t, http.MethodPost, "/api/v1/admin/demands/{id}/approve", w, http.StatusOK)

	// 3. 飞手投意向 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/demands/"+demandID+"/intents",
		`{"intentor_name":"飞手小张","contact":"13900000000","remark":"可完成巡检"}`, workerTok)
	checkCode(t, http.MethodPost, "/api/v1/demands/{id}/intents", w, http.StatusCreated)
	intentID := idFromBody(t, w)

	// 4. 发布者确认接单 → 201（handler 返回 Created）
	w = doRaw(app, http.MethodPost, "/api/v1/demands/"+demandID+"/intents/"+intentID+"/accept",
		`{"amount_fen":100000}`, entTok)
	checkCode(t, http.MethodPost, "/api/v1/demands/{id}/intents/{intentID}/accept", w, http.StatusCreated)
	orderID := idFromBody(t, w)

	// 5. 我的订单（发布方 / 接单方）→ 200
	w = doRaw(app, http.MethodGet, "/api/v1/work-orders/mine", "", entTok)
	checkCode(t, http.MethodGet, "/api/v1/work-orders/mine (publisher)", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), orderID) {
		t.Fatalf("publisher mine should contain order %s: %s", orderID, w.Body.String())
	}
	w = doRaw(app, http.MethodGet, "/api/v1/work-orders/mine", "", workerTok)
	checkCode(t, http.MethodGet, "/api/v1/work-orders/mine (worker)", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), orderID) {
		t.Fatalf("worker mine should contain order %s: %s", orderID, w.Body.String())
	}

	// 6. 飞手开始作业 → 200
	w = doRaw(app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/start", "", workerTok)
	checkCode(t, http.MethodPost, "/api/v1/work-orders/{id}/start", w, http.StatusOK)

	// 7. 飞手确认完成 → 200
	w = doRaw(app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/complete",
		`{"result_photos":["/uploads/a.jpg"]}`, workerTok)
	checkCode(t, http.MethodPost, "/api/v1/work-orders/{id}/complete", w, http.StatusOK)

	// 8. 发布方验收 → 200
	w = doRaw(app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/accept", "", entTok)
	checkCode(t, http.MethodPost, "/api/v1/work-orders/{id}/accept", w, http.StatusOK)
}

// ---- 智能匹配 / 推荐 ----

func TestMatchAndRecommendations(t *testing.T) {
	app := newBizServer(t)
	tok := authAs(t, "user-1", domain.RoleIndividual)

	// 推荐 → 200
	w := doRaw(app, http.MethodGet, "/api/v1/recommendations?biz_type=cable_inspection&limit=5", "", tok)
	checkCode(t, http.MethodGet, "/api/v1/recommendations", w, http.StatusOK)

	// 匹配 → 200
	w = doRaw(app, http.MethodGet, "/api/v1/match?q=巡检", "", tok)
	checkCode(t, http.MethodGet, "/api/v1/match", w, http.StatusOK)

	// 反向：缺 q 参数 → 400
	w = doRaw(app, http.MethodGet, "/api/v1/match", "", tok)
	checkCode(t, http.MethodGet, "/api/v1/match (no q)", w, http.StatusBadRequest)
}

// ---- biz_handlers 其余端点 ----

func TestBizRemainingHandlers(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)
	userTok := authAs(t, "user-1", domain.RoleIndividual)

	// 产业资源列表 → 200
	w := doRaw(app, http.MethodGet, "/api/v1/industry-resources", "", userTok)
	checkCode(t, http.MethodGet, "/api/v1/industry-resources", w, http.StatusOK)

	// 管理员建资源 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/admin/industry-resources",
		`{"name":"无人机001","res_type":"drone","model":"M300","specs":"RTK+热成像","location":"重庆南岸","booking_info":"工作日9-18点","visibility_level":"public","price_fen":100000}`, adminTok)
	checkCode(t, http.MethodPost, "/api/v1/admin/industry-resources", w, http.StatusCreated)
	resID := idFromBody(t, w)

	// 资源预约 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/industry-resources/"+resID+"/book",
		`{"date":"2026-08-20","purpose":"certification","contact_name":"张三","contact_phone":"13800000001"}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/industry-resources/{id}/book", w, http.StatusCreated)

	// 管理员建活动 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/admin/events",
		`{"title":"产业论坛","event_type":"forum","description":"年度论坛","location":"博览中心","start_time":"2026-10-01T10:00:00Z","end_time":"2026-10-02T18:00:00Z","max_attendees":500}`, adminTok)
	checkCode(t, http.MethodPost, "/api/v1/admin/events", w, http.StatusCreated)
	eventID := idFromBody(t, w)

	// 活动报名 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/events/"+eventID+"/register",
		`{"name":"李四","phone":"13800000002","org":"企业A"}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/events/{id}/register", w, http.StatusCreated)

	// 活动报名列表（管理员）→ 200
	w = doRaw(app, http.MethodGet, "/api/v1/events/"+eventID+"/registrations", "", adminTok)
	checkCode(t, http.MethodGet, "/api/v1/events/{id}/registrations", w, http.StatusOK)

	// 管理员建赛事 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/admin/competitions",
		`{"title":"竞速赛","category":"racing","description":"FPV","location":"巴南","sponsor":"协会","start_date":"2026-09-01","end_date":"2026-09-03","max_teams":50}`, adminTok)
	checkCode(t, http.MethodPost, "/api/v1/admin/competitions", w, http.StatusCreated)
	compID := idFromBody(t, w)

	// 赛事报名 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/competitions/"+compID+"/register",
		`{"team_name":"闪电队","member_count":3,"contact_info":"13800000000","name":"张三","phone":"13800000000"}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/competitions/{id}/register", w, http.StatusCreated)

	// 研发难题广场 → 200
	w = doRaw(app, http.MethodGet, "/api/v1/rd-challenges", "", userTok)
	checkCode(t, http.MethodGet, "/api/v1/rd-challenges", w, http.StatusOK)

	// 项目申报 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/project-applications",
		`{"project_name":"无人机产业示范","category":"示范","description":"打造示范区","budget_fen":1000000}`, userTok)
	checkCode(t, http.MethodPost, "/api/v1/project-applications", w, http.StatusCreated)
}
