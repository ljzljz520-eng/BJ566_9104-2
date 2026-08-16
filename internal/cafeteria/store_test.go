package cafeteria

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestOrderLifecycle(t *testing.T) {
	store := NewStore()
	order, err := store.PlaceOrder(PlaceOrderRequest{RequestID: "req-life-001", UserID: "student-001", MenuItemID: "tomato-egg", Quantity: 2})
	if err != nil {
		t.Fatalf("place order: %v", err)
	}
	if order.ID != "order-0001" || order.Total.Cmp(decimal.NewFromFloat(18)) != 0 || order.Status != "confirmed" {
		t.Fatalf("unexpected order: %+v", order)
	}
	balance, err := store.Balance("student-001")
	if err != nil || balance.Cmp(decimal.NewFromFloat(82)) != 0 {
		t.Fatalf("unexpected balance: %s, %v", balance, err)
	}
	if got := store.Orders("student-001"); len(got) != 1 {
		t.Fatalf("unexpected order count: %d", len(got))
	}
	cancelled, err := store.CancelOrder("student-001", order.ID)
	if err != nil || cancelled.Status != "cancelled" {
		t.Fatalf("cancel order: %+v, %v", cancelled, err)
	}
	balance, err = store.Balance("student-001")
	if err != nil || balance.Cmp(decimal.NewFromFloat(100)) != 0 {
		t.Fatalf("unexpected refunded balance: %s, %v", balance, err)
	}
}

func TestDuplicateRequestReturnsOriginalResult(t *testing.T) {
	store := NewStore()
	req := PlaceOrderRequest{RequestID: "req-idempotent-001", UserID: "student-001", MenuItemID: "tomato-egg", Quantity: 1}
	first, err := store.PlaceOrder(req)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	second, err := store.PlaceOrder(req)
	if err != nil {
		t.Fatalf("duplicate request: %v", err)
	}
	if second != first {
		t.Errorf("duplicate result differs: first=%+v second=%+v", first, second)
	}
	balance, err := store.Balance(req.UserID)
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if balance.Cmp(decimal.NewFromFloat(91)) != 0 {
		t.Errorf("duplicate request changed balance: %s", balance)
	}
	if got := store.Orders(req.UserID); len(got) != 1 {
		t.Errorf("duplicate request created %d orders", len(got))
	}
}
