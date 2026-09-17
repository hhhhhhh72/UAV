package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 交付方式驱动下单是否需要收货地址（集成层）。
//
// 这条链路此前是断的：发布表单采集了「交付方式」，但 preview.vue 提交时丢弃，
// 服务端只能按商品类型猜——卖家选了「自提」，买家仍被要求填收货地址。
func TestOrderReceiverRequirementFollowsDelivery(t *testing.T) {
	app := newBizServer(t)

	// 管理端直接建两个已上架商品，唯一差别是交付方式
	mk := func(title, delivery string) string {
		t.Helper()
		w := requestAs(t, app, http.MethodPost, "/api/v1/admin/products",
			[]byte(`{"title":"`+title+`","prod_type":"drone","condition":"new","price_fen":100000,"status":"listed","delivery":"`+delivery+`"}`),
			"admin-1", domain.RolePlatformAdmin)
		if w.Code != http.StatusCreated {
			t.Fatalf("建商品(%s): %d %s", delivery, w.Code, w.Body.String())
		}
		var body struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("解析商品: %v", err)
		}
		return body.Data.ID
	}

	// ① 自提：不带收货信息也必须能下单
	pickupID := mk("自提整机", "pickup")
	w := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		[]byte(`{"product_id":"`+pickupID+`"}`), "buyer-1", domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("自提商品不应强制收货信息，实际 %d %s", w.Code, w.Body.String())
	}

	// ② 物流发货：不带收货信息必须被拒（400），且商品要回到可售
	logisticsID := mk("物流整机", "logistics")
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		[]byte(`{"product_id":"`+logisticsID+`"}`), "buyer-2", domain.RoleIndividual)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("物流发货缺收货信息应返回 400，实际 %d %s", w.Code, w.Body.String())
	}
	// 下单失败必须回滚商品占位，否则商品会滞留在 sold 再也卖不出去
	after := requestAs(t, app, http.MethodGet, "/api/v1/admin/products/"+logisticsID, nil, "admin-1", domain.RolePlatformAdmin)
	if !jsonContains(after.Body.String(), `"status":"listed"`) {
		t.Fatalf("下单被拒后商品必须回到在售，实际 %s", after.Body.String())
	}

	// ③ 补齐收货信息后下单成功
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		orderBody(logisticsID), "buyer-3", domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("带收货信息应下单成功，实际 %d %s", w.Code, w.Body.String())
	}
}
