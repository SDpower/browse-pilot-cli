package transport

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// nmMaxMessageSize 是 Native Messaging 協議規定的單則訊息大小上限（1MB）。
const nmMaxMessageSize = 1024 * 1024

// NMTransport 是基於 Native Messaging 協議的 Transport 實作。
// 透過 stdin/stdout 與 Chrome/Edge Extension 進行通訊。
// 訊息格式：[4 bytes uint32 LE length][JSON bytes]
type NMTransport struct {
	// config 儲存建立時傳入的設定
	config Config

	// reader 是訊息輸入來源，預設為 os.Stdin
	reader io.Reader

	// writer 是訊息輸出目標，預設為 os.Stdout
	writer io.Writer

	// connected 表示 transport 是否已啟動並可使用
	connected bool
	connMu    sync.Mutex

	// pending 儲存等待回應的 request channel，以 request ID 為 key
	pending   map[string]chan *Response
	pendingMu sync.Mutex

	// done 用於通知 readLoop 結束
	done chan struct{}
}

// NewNMTransport 建立一個新的 NMTransport 實例，預設使用 os.Stdin/os.Stdout。
func NewNMTransport(cfg Config) *NMTransport {
	return &NMTransport{
		config:  cfg,
		reader:  os.Stdin,
		writer:  os.Stdout,
		pending: make(map[string]chan *Response),
		done:    make(chan struct{}),
	}
}

// readMessage 從 reader 讀取一條 length-prefixed JSON 訊息。
// 格式：[uint32 LE length][JSON bytes]
func (t *NMTransport) readMessage() ([]byte, error) {
	// 讀取 4 bytes 的訊息長度（little-endian）
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(t.reader, lenBuf); err != nil {
		return nil, fmt.Errorf("讀取訊息長度失敗: %w", err)
	}
	msgLen := binary.LittleEndian.Uint32(lenBuf)

	// 檢查訊息大小不超過 1MB 限制
	if msgLen > nmMaxMessageSize {
		return nil, fmt.Errorf("訊息長度 %d 超過 Native Messaging 1MB 限制", msgLen)
	}

	// 讀取實際的 JSON 訊息本體
	msgBuf := make([]byte, msgLen)
	if _, err := io.ReadFull(t.reader, msgBuf); err != nil {
		return nil, fmt.Errorf("讀取訊息本體失敗: %w", err)
	}

	return msgBuf, nil
}

// writeMessage 將 JSON 訊息以 length-prefixed 格式寫入 writer。
// 格式：[uint32 LE length][JSON bytes]
func (t *NMTransport) writeMessage(data []byte) error {
	if len(data) > nmMaxMessageSize {
		return fmt.Errorf("訊息長度 %d 超過 Native Messaging 1MB 限制", len(data))
	}

	// 寫入 4 bytes 的訊息長度（little-endian）
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(data)))
	if _, err := t.writer.Write(lenBuf); err != nil {
		return fmt.Errorf("寫入訊息長度失敗: %w", err)
	}

	// 寫入 JSON 訊息本體
	if _, err := t.writer.Write(data); err != nil {
		return fmt.Errorf("寫入訊息本體失敗: %w", err)
	}

	return nil
}

// Start 啟動 Native Messaging transport，開始監聽 stdin 輸入。
// Start 是非阻塞的，readLoop 在背景 goroutine 執行。
func (t *NMTransport) Start(ctx context.Context) error {
	t.connMu.Lock()
	t.connected = true
	t.connMu.Unlock()

	go t.readLoop()
	return nil
}

// readLoop 持續從 stdin 讀取 length-prefixed JSON 訊息。
// 將解析後的 Response 分派到對應的 pending channel。
// 當 stdin 關閉或讀取錯誤時退出。
func (t *NMTransport) readLoop() {
	defer func() {
		t.connMu.Lock()
		t.connected = false
		t.connMu.Unlock()

		// 通知所有 pending requests 連線已斷開
		t.failAllPending(&RPCError{
			Code:    ErrNativeMessagingError,
			Message: "Native Messaging stdin 已關閉",
		})

		// 關閉 done channel
		select {
		case <-t.done:
		default:
			close(t.done)
		}
	}()

	for {
		data, err := t.readMessage()
		if err != nil {
			// stdin 關閉或讀取錯誤，退出迴圈
			return
		}

		var resp Response
		if err := json.Unmarshal(data, &resp); err != nil {
			// 無法解析的訊息，忽略並繼續讀取
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
func (t *NMTransport) failAllPending(rpcErr *RPCError) {
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

// Send 發送 JSON-RPC 請求至 Extension（透過 stdout），並阻塞等待回應。
// 若 ctx 在收到回應前逾時或取消，回傳對應的 context 錯誤。
func (t *NMTransport) Send(ctx context.Context, req *Request) (*Response, error) {
	t.connMu.Lock()
	if !t.connected {
		t.connMu.Unlock()
		return nil, &RPCError{
			Code:    ErrNativeMessagingError,
			Message: "Native Messaging transport 尚未啟動",
		}
	}
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

	// 序列化請求為 JSON
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化請求失敗: %w", err)
	}

	// 以 length-prefixed 格式寫入 stdout
	if err := t.writeMessage(data); err != nil {
		return nil, &RPCError{
			Code:    ErrNativeMessagingError,
			Message: fmt.Sprintf("發送訊息失敗: %v", err),
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

// Close 停止 Native Messaging transport 並釋放資源。
func (t *NMTransport) Close() error {
	t.connMu.Lock()
	t.connected = false
	t.connMu.Unlock()

	// 清理所有 pending requests
	t.failAllPending(&RPCError{
		Code:    ErrNativeMessagingError,
		Message: "Native Messaging transport 已關閉",
	})

	// 嘗試關閉 done channel
	select {
	case <-t.done:
		// 已關閉
	default:
		close(t.done)
	}

	return nil
}

// IsConnected 回傳 transport 是否處於連線狀態。
func (t *NMTransport) IsConnected() bool {
	t.connMu.Lock()
	defer t.connMu.Unlock()
	return t.connected
}

// Type 回傳 transport 的類型識別字串。
func (t *NMTransport) Type() string {
	return "native_messaging"
}
