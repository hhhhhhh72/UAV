package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/httpapi"
)

// TestCreateMessageBroadcast verifies that an admin message with an empty
// receiver broadcasts to every admin user plus the requester (dev shadow admin).
func TestCreateMessageBroadcast(t *testing.T) {
	app := newBizServer(t)
	body := []byte(`{"sender_id":"system","receiver_id":"","title":"公告","content":"版本升级"}`)
	w := request(t, app, http.MethodPost, "/api/v1/admin/messages", body, domain.RolePlatformAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("broadcast create: %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Broadcast int `json:"broadcast"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// 广播发给全部平台用户（种子用户全员接收；req_actor 已在种子中）
	if resp.Data.Broadcast != len(commonUsersSeed) {
		t.Fatalf("broadcast count: want %d, got %d", len(commonUsersSeed), resp.Data.Broadcast)
	}

	// requester must see exactly 1 unread message
	w = request(t, app, http.MethodGet, "/api/v1/messages/unread-count", nil, domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("unread-count: %d", w.Code)
	}
	var unread struct {
		Data struct {
			Count int `json:"count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &unread); err != nil {
		t.Fatalf("decode unread: %v", err)
	}
	if unread.Data.Count != 1 {
		t.Fatalf("unread count: want 1, got %d", unread.Data.Count)
	}

	// 其他种子用户同样收到广播（全员触达）
	tokens, err := httpapi.NewTokenManager(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	otherToken, err := tokens.Issue(domain.Actor{ID: "user-2", Role: domain.RoleIndividual}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	otherReq := httptest.NewRequest(http.MethodGet, "/api/v1/messages/unread-count", nil)
	otherReq.Header.Set("Authorization", "Bearer "+otherToken)
	otherW := httptest.NewRecorder()
	app.ServeHTTP(otherW, otherReq)
	if otherW.Code != http.StatusOK {
		t.Fatalf("unread-count (other): %d", otherW.Code)
	}
	if err := json.Unmarshal(otherW.Body.Bytes(), &unread); err != nil {
		t.Fatalf("decode unread (other): %v", err)
	}
	if unread.Data.Count != 1 {
		t.Fatalf("other seeded user should receive broadcast, got %d", unread.Data.Count)
	}
}

// TestCreateMessageSingle verifies a direct message still works unchanged.
func TestCreateMessageSingle(t *testing.T) {
	app := newBizServer(t)
	body := []byte(`{"sender_id":"system","receiver_id":"user-2","title":"直发","content":"hello"}`)
	w := request(t, app, http.MethodPost, "/api/v1/admin/messages", body, domain.RolePlatformAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("single create: %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			ReceiverID string `json:"receiver_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.ReceiverID != "user-2" {
		t.Fatalf("receiver: want user-2, got %q", resp.Data.ReceiverID)
	}
}
