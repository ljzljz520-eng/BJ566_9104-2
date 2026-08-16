package cafeteria

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPOrderingFlow(t *testing.T) {
	server := NewHTTPServer(NewStore())
	menuReq := httptest.NewRequest(http.MethodGet, "/api/menu", nil)
	menuResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(menuResp, menuReq)
	if menuResp.Code != http.StatusOK {
		t.Fatalf("menu status: %d", menuResp.Code)
	}
	var menu []MenuItem
	if err := json.Unmarshal(menuResp.Body.Bytes(), &menu); err != nil || len(menu) != 3 {
		t.Fatalf("menu response: %v", err)
	}

	payload, _ := json.Marshal(PlaceOrderRequest{RequestID: "req-http-001", UserID: "student-002", MenuItemID: "seasonal-fruit", Quantity: 2})
	placeReq := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(payload))
	placeResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(placeResp, placeReq)
	if placeResp.Code != http.StatusCreated {
		t.Fatalf("place status: %d body=%s", placeResp.Code, placeResp.Body.String())
	}
	var order Order
	if err := json.Unmarshal(placeResp.Body.Bytes(), &order); err != nil {
		t.Fatalf("order response: %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/orders?user_id=student-002", nil)
	listResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(listResp, listReq)
	var orders []Order
	if err := json.Unmarshal(listResp.Body.Bytes(), &orders); err != nil || len(orders) != 1 || orders[0].ID != order.ID {
		t.Fatalf("orders response: %v %+v", err, orders)
	}

	cancelReq := httptest.NewRequest(http.MethodPost, "/api/orders/"+order.ID+"/cancel?user_id=student-002", nil)
	cancelResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(cancelResp, cancelReq)
	if cancelResp.Code != http.StatusOK {
		t.Fatalf("cancel status: %d body=%s", cancelResp.Code, cancelResp.Body.String())
	}
}
