package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSTransport 是基於 WebSocket 的 Transport 實作。
// CLI 端作為 WebSocket server，Browser Extension 作為 client 連入。
// 每次只允許一個 Extension 連線（單連線模式）。
type WSTransport struct {
	// config 儲存建立時傳入的設定
	config Config

	// server 是底層 HTTP server 實例
	server *http.Server

	// upgrader 負責將 HTTP 連線升級為 WebSocket
	upgrader websocket.Upgrader

	// conn 是目前作用中的 WebSocket 連線
	conn   *websocket.Conn
	connMu sync.Mutex

	// connected 表示目前是否有活躍的 Extension 連線
	connected bool

	// pending 儲存等待回應的 request channel，以 request ID 為 key
	pending   map[string]chan *Response
	pendingMu sync.Mutex

	// connReady 在 Extension 連線後關閉，用來喚醒所有等待中的 Send。
	// 連線中斷後會替換成新的 channel，供下一次重連使用。
	connReady chan struct{}
}

// NewWSTransport 建立一個新的 WSTransport 實例。
func NewWSTransport(cfg Config) *WSTransport {
	return &WSTransport{
		config: cfg,
		upgrader: websocket.Upgrader{
			// 允許所有 Origin，Extension 的 origin 可能為 moz-extension:// 等
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		pending:   make(map[string]chan *Response),
		connReady: make(chan struct{}),
	}
}

// Start 同步綁定監聽位址並啟動 WebSocket HTTP server。
// 成功綁定後立即回傳，不等待 Extension 連入；若連接埠已被占用則直接失敗。
func (t *WSTransport) Start(_ context.Context) error {
	addr := fmt.Sprintf("127.0.0.1:%d", t.config.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", t.handleWS)

	t.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("無法監聽 WebSocket 位址 %s: %w", addr, err)
	}

	// 綁定成功後才在背景服務，確保 Start 回傳 nil 時監聽已就緒。
	go func() {
		if err := t.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			// server 非預期關閉，記錄錯誤
			_ = err
		}
	}()

	if t.config.Verbose {
		fmt.Fprintf(os.Stderr, "[WS] 已監聽 ws://%s，等待 Extension 連線\n", addr)
	}
	return nil
}

// handleWS 處理 WebSocket 升級請求。
// 同一時間只允許一個連線：若已有連線，先關閉舊連線再接受新連線。
func (t *WSTransport) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := t.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// 升級失敗（如普通 HTTP 請求），upgrader 已自動回傳錯誤回應
		return
	}

	t.connMu.Lock()
	// 換線期間阻止新的 Send 取得連線，並先讓舊連線的未完成請求失敗。
	if t.conn != nil {
		_ = t.conn.Close()
		t.failAllPending(&RPCError{
			Code:    ErrConnectionError,
			Message: "連線已被新 Extension 取代",
			Data:    errorData(t.config),
		})
	}
	t.conn = conn
	t.connected = true
	select {
	case <-t.connReady:
	default:
		close(t.connReady)
	}
	t.connMu.Unlock()

	// 啟動讀取迴圈
	go t.readLoop(conn)
}

// readLoop 持續讀取 WebSocket 訊息，將 Response 分派到對應的 pending channel。
// 當連線關閉時，清理所有未完成的 pending requests。
func (t *WSTransport) readLoop(conn *websocket.Conn) {
	defer func() {
		t.connMu.Lock()
		// 只有在 conn 仍是目前連線時才標記為斷線
		isCurrent := t.conn == conn
		if isCurrent {
			t.connected = false
			t.conn = nil
			t.connReady = make(chan struct{})
		}
		t.connMu.Unlock()

		if isCurrent {
			// 僅目前連線斷開時清理請求，避免舊 readLoop 影響新連線。
			t.failAllPending(&RPCError{
				Code:    ErrConnectionError,
				Message: "WebSocket 連線已斷開",
				Data:    errorData(t.config),
			})
		}
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			// 連線關閉或讀取錯誤，退出迴圈
			return
		}

		var resp Response
		if err := json.Unmarshal(data, &resp); err != nil {
			// 無法解析的訊息，忽略並繼續
			continue
		}

		// 查找對應的 pending channel 並分派回應
		t.pendingMu.Lock()
		ch, ok := t.pending[resp.ID]
		if ok {
			delete(t.pending, resp.ID)
		}
		t.pendingMu.Unlock()

		if ok {
			select {
			case ch <- &resp:
			default:
				// channel 已滿或已被關閉，忽略
			}
		}
	}
}

// failAllPending 將錯誤回應發送給所有等待中的 pending requests。
func (t *WSTransport) failAllPending(rpcErr *RPCError) {
	t.pendingMu.Lock()
	defer t.pendingMu.Unlock()

	for id, ch := range t.pending {
		delete(t.pending, id)
		resp := &Response{
			ID:    id,
			Error: rpcErr,
		}
		select {
		case ch <- resp:
		default:
		}
	}
}

// Send 發送 JSON-RPC 請求至已連線的 Extension，並阻塞等待回應。
// 若目前無連線，會等待 Extension 連入（受 ctx 逾時控制）。
// 若 ctx 在收到回應前逾時或取消，回傳對應的 context 錯誤。
func (t *WSTransport) Send(ctx context.Context, req *Request) (*Response, error) {
	// 等待連線就緒
	for {
		t.connMu.Lock()
		if t.connected && t.conn != nil {
			conn := t.conn
			t.connMu.Unlock()
			// 建立回應 channel 並加入 pending map
			ch := make(chan *Response, 1)
			t.pendingMu.Lock()
			t.pending[req.ID] = ch
			t.pendingMu.Unlock()

			return t.sendAndWait(ctx, conn, req, ch)
		}
		ready := t.connReady
		t.connMu.Unlock()

		// 等待新連線或逾時。
		select {
		case <-ready:
			break
		case <-ctx.Done():
			return nil, &RPCError{
				Code:    ErrConnectionError,
				Message: fmt.Sprintf("等待 %s Extension 連線逾時", browserName(t.config.Browser)),
				Data:    errorData(t.config),
			}
		}
	}
}

func (t *WSTransport) sendAndWait(ctx context.Context, conn *websocket.Conn, req *Request, ch chan *Response) (*Response, error) {
	// 確保離開時清理 pending entry
	defer func() {
		t.pendingMu.Lock()
		delete(t.pending, req.ID)
		t.pendingMu.Unlock()
	}()

	// 序列化並發送請求
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化請求失敗: %w", err)
	}

	t.connMu.Lock()
	writeErr := conn.WriteMessage(websocket.TextMessage, data)
	t.connMu.Unlock()
	if writeErr != nil {
		return nil, &RPCError{
			Code:    ErrConnectionError,
			Message: "無法傳送訊息至 Extension",
			Data:    errorData(t.config),
		}
	}

	// 等待回應或逾時
	select {
	case resp := <-ch:
		return resp, nil
	case <-ctx.Done():
		return nil, &RPCError{
			Code:    ErrTimeoutError,
			Message: "等待 Extension 回應逾時",
			Data:    errorData(t.config),
		}
	}
}

// Close 關閉 WebSocket 連線及 HTTP server，並釋放所有資源。
func (t *WSTransport) Close() error {
	t.connMu.Lock()
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}
	t.connected = false
	t.connMu.Unlock()

	// 清理所有 pending requests
	t.failAllPending(&RPCError{
		Code:    ErrConnectionError,
		Message: "Transport 已關閉",
		Data:    errorData(t.config),
	})

	if t.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return t.server.Shutdown(ctx)
	}
	return nil
}

// IsConnected 回傳目前是否有活躍的 Extension 連線。
func (t *WSTransport) IsConnected() bool {
	t.connMu.Lock()
	defer t.connMu.Unlock()
	return t.connected
}

// Type 回傳 transport 的類型識別字串。
func (t *WSTransport) Type() string {
	return "websocket"
}

func browserName(browser string) string {
	switch browser {
	case "firefox":
		return "Firefox"
	case "chrome":
		return "Chrome"
	case "edge":
		return "Edge"
	default:
		return "瀏覽器"
	}
}
