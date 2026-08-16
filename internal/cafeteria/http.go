package cafeteria

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type HTTPServer struct {
	store *Store
	mux   *http.ServeMux
}

func NewHTTPServer(store *Store) *HTTPServer {
	s := &HTTPServer{store: store, mux: http.NewServeMux()}
	s.mux.HandleFunc("/", s.handlePage)
	s.mux.HandleFunc("/api/menu", s.handleMenu)
	s.mux.HandleFunc("/api/orders", s.handleOrders)
	s.mux.HandleFunc("/api/orders/", s.handleOrderAction)
	s.mux.HandleFunc("/api/balance", s.handleBalance)
	return s
}

func (s *HTTPServer) Handler() http.Handler { return s.mux }

func (s *HTTPServer) handlePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	const page = `<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><title>食堂订餐</title><style>body{font-family:system-ui,sans-serif;max-width:760px;margin:2rem auto;padding:0 1rem;color:#24313d}h1{margin-bottom:.4rem}.hint{color:#637381}.menu{display:grid;grid-template-columns:repeat(auto-fit,minmax(210px,1fr));gap:12px;margin-top:1.4rem}.item{border:1px solid #d8e0e6;padding:14px;border-radius:6px}.price{font-weight:700;color:#0b6e4f}.stock{color:#637381;font-size:.9rem}</style></head><body><h1>今日食堂菜单</h1><p class="hint">固定演示数据，可通过 API 完成订餐、查询和取消。</p><div id="menu" class="menu">加载中...</div><script>fetch('/api/menu').then(r=>r.json()).then(items=>{document.getElementById('menu').innerHTML=items.map(i=>'<div class="item"><strong>'+i.name+'</strong><div class="price">￥'+i.price+'</div><div class="stock">剩余 '+i.stock+' 份</div></div>').join('')})</script></body></html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(page))
}

func (s *HTTPServer) handleMenu(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, s.store.Menu())
}

func (s *HTTPServer) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			writeError(w, http.StatusBadRequest, ErrInvalidRequest)
			return
		}
		writeJSON(w, http.StatusOK, s.store.Orders(userID))
	case http.MethodPost:
		var req PlaceOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, ErrInvalidRequest)
			return
		}
		order, err := s.store.PlaceOrder(req)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, order)
	default:
		methodNotAllowed(w)
	}
}

func (s *HTTPServer) handleOrderAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/cancel") {
		methodNotAllowed(w)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 4 || parts[0] != "api" || parts[1] != "orders" || parts[3] != "cancel" {
		writeError(w, http.StatusNotFound, ErrOrderNotFound)
		return
	}
	userID := r.URL.Query().Get("user_id")
	order, err := s.store.CancelOrder(userID, parts[2])
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *HTTPServer) handleBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	balance, err := s.store.Balance(r.URL.Query().Get("user_id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"user_id": r.URL.Query().Get("user_id"), "balance": balance.StringFixed(2)})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeStoreError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrMenuNotFound) || errors.Is(err, ErrOrderNotFound) {
		status = http.StatusNotFound
	}
	writeError(w, status, err)
}

func methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", "GET, POST")
	writeError(w, http.StatusMethodNotAllowed, errors.New("请求方法不支持"))
}
