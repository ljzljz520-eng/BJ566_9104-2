package cafeteria

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/shopspring/decimal"
)

var (
	ErrInvalidRequest = errors.New("请求参数不合法")
	ErrUserNotFound   = errors.New("用户不存在")
	ErrMenuNotFound   = errors.New("菜品不存在")
	ErrOutOfStock     = errors.New("菜品库存不足")
	ErrInsufficient   = errors.New("余额不足")
	ErrOrderNotFound  = errors.New("订单不存在")
	ErrOrderCancelled = errors.New("订单已取消")
)

type MenuItem struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Price decimal.Decimal `json:"price"`
	Stock int             `json:"stock"`
}

type Order struct {
	ID         string          `json:"id"`
	RequestID  string          `json:"request_id"`
	UserID     string          `json:"user_id"`
	MenuItemID string          `json:"menu_item_id"`
	Quantity   int             `json:"quantity"`
	Total      decimal.Decimal `json:"total"`
	Status     string          `json:"status"`
}

type PlaceOrderRequest struct {
	RequestID  string `json:"request_id"`
	UserID     string `json:"user_id"`
	MenuItemID string `json:"menu_item_id"`
	Quantity   int    `json:"quantity"`
}

type Store struct {
	mu            sync.Mutex
	menu          map[string]MenuItem
	balances      map[string]decimal.Decimal
	orders        map[string]Order
	orderSequence []string
	requestIndex  map[string]string
	nextOrderID   int
}

// NewStore 创建固定的演示数据，保证每次启动的验收结果一致。
func NewStore() *Store {
	return &Store{
		menu: map[string]MenuItem{
			"rice-chicken":   {ID: "rice-chicken", Name: "香煎鸡排饭", Price: decimal.NewFromFloat(12.50), Stock: 20},
			"tomato-egg":     {ID: "tomato-egg", Name: "番茄鸡蛋面", Price: decimal.NewFromFloat(9.00), Stock: 15},
			"seasonal-fruit": {ID: "seasonal-fruit", Name: "时令水果盒", Price: decimal.NewFromFloat(6.50), Stock: 12},
		},
		balances: map[string]decimal.Decimal{
			"student-001": decimal.NewFromFloat(100.00),
			"student-002": decimal.NewFromFloat(35.00),
		},
		orders:       make(map[string]Order),
		requestIndex: make(map[string]string),
	}
}

// Menu 返回按编号排列的今日菜单，避免 map 遍历顺序影响展示和测试。
func (s *Store) Menu() []MenuItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]MenuItem, 0, len(s.menu))
	for _, item := range s.menu {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}

// Balance 查询用户余额。
func (s *Store) Balance(userID string) (decimal.Decimal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	balance, ok := s.balances[userID]
	if !ok {
		return decimal.Zero, ErrUserNotFound
	}
	return balance, nil
}

// PlaceOrder 完成校验、扣减库存和余额，并记录订单。
func (s *Store) PlaceOrder(req PlaceOrderRequest) (Order, error) {
	if req.RequestID == "" || req.UserID == "" || req.MenuItemID == "" || req.Quantity <= 0 {
		return Order{}, ErrInvalidRequest
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if orderID, ok := s.requestIndex[req.RequestID]; ok {
		if order, exists := s.orders[orderID]; exists {
			return order, nil
		}
	}

	userBalance, ok := s.balances[req.UserID]
	if !ok {
		return Order{}, ErrUserNotFound
	}
	item, ok := s.menu[req.MenuItemID]
	if !ok {
		return Order{}, ErrMenuNotFound
	}
	if item.Stock < req.Quantity {
		return Order{}, ErrOutOfStock
	}
	total := item.Price.Mul(decimal.NewFromInt(int64(req.Quantity)))
	if userBalance.LessThan(total) {
		return Order{}, ErrInsufficient
	}

	item.Stock -= req.Quantity
	s.menu[item.ID] = item
	s.balances[req.UserID] = userBalance.Sub(total)
	s.nextOrderID++
	order := Order{
		ID:         fmt.Sprintf("order-%04d", s.nextOrderID),
		RequestID:  req.RequestID,
		UserID:     req.UserID,
		MenuItemID: req.MenuItemID,
		Quantity:   req.Quantity,
		Total:      total,
		Status:     "confirmed",
	}
	s.orders[order.ID] = order
	s.orderSequence = append(s.orderSequence, order.ID)
	s.requestIndex[req.RequestID] = order.ID
	return order, nil
}

// Orders 返回用户订单，并按创建顺序排列。
func (s *Store) Orders(userID string) []Order {
	s.mu.Lock()
	defer s.mu.Unlock()
	orders := make([]Order, 0)
	for _, id := range s.orderSequence {
		order := s.orders[id]
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return orders
}

// CancelOrder 取消已确认订单并返还对应资源。
func (s *Store) CancelOrder(userID, orderID string) (Order, error) {
	if userID == "" || orderID == "" {
		return Order{}, ErrInvalidRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	order, ok := s.orders[orderID]
	if !ok || order.UserID != userID {
		return Order{}, ErrOrderNotFound
	}
	if order.Status == "cancelled" {
		return Order{}, ErrOrderCancelled
	}
	item := s.menu[order.MenuItemID]
	item.Stock += order.Quantity
	s.menu[item.ID] = item
	s.balances[userID] = s.balances[userID].Add(order.Total)
	order.Status = "cancelled"
	s.orders[order.ID] = order
	return order, nil
}
