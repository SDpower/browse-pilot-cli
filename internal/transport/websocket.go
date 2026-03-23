package transport

import (
	"context"
	"encoding/json"
	"fmt"
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

	// done 用於通知 readLoop 停止
	done chan struct{}

	// newConn 用於通知 handleWS 有新連線就緒
	newConn chan struct{}
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
		pending: make(map[string]chan *Response),
		done:    make(chan struct{}),
		newConn: make(chan struct{}, 1),
	}
}

// Start 啟動 WebSocket HTTP server，並等待 Extension 連入。
// 若在 ctx 逾時前 Extension 成功連線則回傳 nil，
// 逾時則回傳錯誤（但 server 仍持續運行，後續 Send 會再次等待連線）。
func (t *WSTransport) Start(ctx context.Context) error {
	addr := fmt.Sprintf("127.0.0.1:%d", t.config.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", t.handleWS)

	t.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// 在背景啟動 HTTP server
	go func() {
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// server 非預期關閉，記錄錯誤
			_ = err
		}
	}()

	// 等待 server 就緒
	time.Sleep(10 * time.Millisecond)

	if t.config.Verbose {
		fmt.Fprintf(os.Stderr, "[WS] 等待 Extension 連線 ws://%s ...\n", addr)
	}

	// 等待 Extension 連入或 ctx 逾時
	select {
	case <-t.newConn:
		if t.config.Verbose {
			fmt.Fprintln(os.Stderr, "[WS] Extension 已連線")
		}
		return nil
	case <-ctx.Done():
		// 逾時不關閉 server，回傳提示（Send 會再等待）
		return fmt.Errorf("等待 Extension 連線逾時（server 仍在 ws://%s 監聽中）", addr)
	}
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
	// 若已有舊連線，關閉它並清理 pending requests
	if t.conn != nil {
		t.conn.Close()
	}
	t.conn = conn
	t.connected = true
	// 重置 done channel 供新的 readLoop 使用
	t.done = make(chan struct{})
	t.connMu.Unlock()

	// 清理所有因舊連線斷開而懸掛的 pending requests
	t.failAllPending(&RPCError{
		Code:    ErrConnectionError,
		Message: "連線已被新 Extension 取代",
	})

	// 啟動讀取迴圈
	go t.readLoop(conn, t.done)

	// 通知有新連線（非阻塞）
	select {
	case t.newConn <- struct{}{}:
	default:
	}
}

// readLoop 持續讀取 WebSocket 訊息，將 Response 分派到對應的 pending channel。
// 當連線關閉時，清理所有未完成的 pending requests。
func (t *WSTransport) readLoop(conn *websocket.Conn, done chan struct{}) {
	defer func() {
		t.connMu.Lock()
		// 只有在 conn 仍是目前連線時才標記為斷線
		if t.conn == conn {
			t.connected = false
			t.conn = nil
		}
		t.connMu.Unlock()

		// 通知所有 pending requests 連線已斷開
		t.failAllPending(&RPCError{
			Code:    ErrConnectionError,
			Message: "WebSocket 連線已斷開",
		})

		// 關閉 done channel 通知外部迴圈結束
		select {
		case <-done:
		default:
			close(done)
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
			break
		}
		t.connMu.Unlock()

		// 等待新連線或逾時
		select {
		case <-t.newConn:
			continue
		case <-ctx.Done():
			return nil, &RPCError{
				Code:    ErrConnectionError,
				Message: "等待 Extension 連線逾時",
			}
		}
	}
	conn := t.conn
	t.connMu.Unlock()

	// 建立回應 channel 並加入 pending map
	ch := make(chan *Response, 1)
	t.pendingMu.Lock()
	t.pending[req.ID] = ch
	t.pendingMu.Unlock()

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
			Message: fmt.Sprintf("發送訊息失敗: %v", writeErr),
		}
	}

	// 等待回應或逾時
	select {
	case resp := <-ch:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
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
