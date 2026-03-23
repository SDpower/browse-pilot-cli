package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// dialTestWS 嘗試連線至本地測試 WebSocket server，最多重試數次
func dialTestWS(t *testing.T, port int) *websocket.Conn {
	t.Helper()
	url := fmt.Sprintf("ws://127.0.0.1:%d/", port)
	var conn *websocket.Conn
	var err error
	for i := 0; i < 20; i++ {
		conn, _, err = websocket.DefaultDialer.Dial(url, nil)
		if err == nil {
			return conn
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("無法連線至測試 WebSocket server: %v", err)
	return nil
}

// startWithClient 在背景 goroutine 中啟動 client 連線，確保 Start 不會因等待連線而卡住
func startWithClient(t *testing.T, tr *WSTransport, port int) *websocket.Conn {
	t.Helper()
	var conn *websocket.Conn

	// 在背景啟動 client 連線（Start 會等待連入）
	done := make(chan struct{})
	go func() {
		conn = dialTestWS(t, port)
		close(done)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tr.Start(ctx); err != nil {
		t.Fatalf("Start() 失敗: %v", err)
	}

	<-done
	return conn
}

// TestWSTransportStartAndConnect 驗證 WSTransport 可啟動並接受連線
func TestWSTransportStartAndConnect(t *testing.T) {
	cfg := Config{Port: 19222, Timeout: 5 * time.Second}
	tr := NewWSTransport(cfg)

	conn := startWithClient(t, tr, 19222)
	defer tr.Close()
	defer conn.Close()

	if !tr.IsConnected() {
		t.Error("IsConnected() 應為 true")
	}
	if tr.Type() != "websocket" {
		t.Errorf("Type() = %q, 預期 \"websocket\"", tr.Type())
	}
}

// TestWSTransportSendReceive 驗證 Send 可正確發送並接收回應
func TestWSTransportSendReceive(t *testing.T) {
	cfg := Config{Port: 19223, Timeout: 5 * time.Second}
	tr := NewWSTransport(cfg)

	conn := startWithClient(t, tr, 19223)
	defer tr.Close()
	defer conn.Close()

	// 模擬 Extension：接收 request，回傳 response
	go func() {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var req Request
		if err := json.Unmarshal(msg, &req); err != nil {
			return
		}
		resp := Response{
			ID:     req.ID,
			Result: json.RawMessage(`{"url":"https://example.com"}`),
		}
		data, _ := json.Marshal(resp)
		conn.WriteMessage(websocket.TextMessage, data)
	}()

	req, err := NewRequest("navigate", map[string]string{"url": "https://example.com"})
	if err != nil {
		t.Fatalf("NewRequest() 失敗: %v", err)
	}

	sendCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := tr.Send(sendCtx, req)
	if err != nil {
		t.Fatalf("Send() 失敗: %v", err)
	}
	if resp.ID != req.ID {
		t.Errorf("回應 ID = %q，預期 %q", resp.ID, req.ID)
	}
	if resp.IsError() {
		t.Errorf("預期成功回應，但收到錯誤: %v", resp.Error)
	}
}

// TestWSTransportSendTimeout 驗證當無回應時 Send 會因 ctx 逾時而返回錯誤
func TestWSTransportSendTimeout(t *testing.T) {
	cfg := Config{Port: 19224, Timeout: 5 * time.Second}
	tr := NewWSTransport(cfg)

	conn := startWithClient(t, tr, 19224)
	defer tr.Close()
	defer conn.Close()

	// 建立連線但不回應任何訊息
	req, _ := NewRequest("navigate", nil)

	sendCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := tr.Send(sendCtx, req)
	if err == nil {
		t.Fatal("預期收到逾時錯誤，但 Send() 回傳 nil")
	}
}

// TestWSTransportSendNoConnection 驗證無連線時 Send 等待至逾時後回傳錯誤
func TestWSTransportSendNoConnection(t *testing.T) {
	cfg := Config{Port: 19225, Timeout: 5 * time.Second}
	tr := NewWSTransport(cfg)

	// 使用短逾時啟動 Start（不建立連線，預期逾時）
	startCtx, startCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer startCancel()
	// Start 逾時是正常的（無 Extension 連入）
	_ = tr.Start(startCtx)
	defer tr.Close()

	// 不建立任何連線，直接發送（Send 會等待連線直到逾時）
	req, _ := NewRequest("navigate", nil)

	sendCtx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	_, err := tr.Send(sendCtx, req)
	if err == nil {
		t.Fatal("預期收到連線錯誤，但 Send() 回傳 nil")
	}
}

// TestWSTransportReconnect 驗證舊連線斷開後，新連線可正常工作
func TestWSTransportReconnect(t *testing.T) {
	cfg := Config{Port: 19226, Timeout: 5 * time.Second}
	tr := NewWSTransport(cfg)

	conn1 := startWithClient(t, tr, 19226)
	defer tr.Close()

	// 第一次連線然後斷開
	conn1.Close()
	time.Sleep(100 * time.Millisecond)

	// 第二次連線，模擬 Extension 回應
	conn2 := dialTestWS(t, 19226)
	defer conn2.Close()

	go func() {
		_, msg, err := conn2.ReadMessage()
		if err != nil {
			return
		}
		var req Request
		if err := json.Unmarshal(msg, &req); err != nil {
			return
		}
		resp := Response{
			ID:     req.ID,
			Result: json.RawMessage(`true`),
		}
		data, _ := json.Marshal(resp)
		conn2.WriteMessage(websocket.TextMessage, data)
	}()

	time.Sleep(100 * time.Millisecond)

	req, _ := NewRequest("ping", nil)
	sendCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := tr.Send(sendCtx, req)
	if err != nil {
		t.Fatalf("重連後 Send() 失敗: %v", err)
	}
	if resp.ID != req.ID {
		t.Errorf("回應 ID = %q，預期 %q", resp.ID, req.ID)
	}
}

// TestWSTransportHTTPUpgrade 驗證非 WebSocket 請求的 HTTP 處理
func TestWSTransportHTTPUpgrade(t *testing.T) {
	cfg := Config{Port: 19227, Timeout: 5 * time.Second}
	tr := NewWSTransport(cfg)

	// Start 會等待連線，用短逾時讓它先跑起來
	startCtx, startCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer startCancel()
	_ = tr.Start(startCtx)
	defer tr.Close()

	// 普通 HTTP 請求應回傳 400 或類似錯誤（非 WebSocket upgrade）
	resp, err := http.Get("http://127.0.0.1:19227/")
	if err != nil {
		t.Fatalf("HTTP GET 失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Error("預期非 200 狀態碼，因為這是普通 HTTP 請求")
	}
}

// TestWSTransportStartWaitsForConnection 驗證 Start 會等待 Extension 連入
func TestWSTransportStartWaitsForConnection(t *testing.T) {
	cfg := Config{Port: 19228, Timeout: 5 * time.Second}
	tr := NewWSTransport(cfg)

	startDone := make(chan error, 1)

	// Start 在背景執行（會阻塞等待連線）
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		startDone <- tr.Start(ctx)
	}()

	// 等一下再連線，模擬 Extension 延遲連入
	time.Sleep(200 * time.Millisecond)
	conn := dialTestWS(t, 19228)
	defer conn.Close()
	defer tr.Close()

	// Start 應該在連線後回傳 nil
	err := <-startDone
	if err != nil {
		t.Fatalf("Start() 應在連線後成功，但回傳: %v", err)
	}
	if !tr.IsConnected() {
		t.Error("IsConnected() 應為 true")
	}
}

// TestWSTransportStartTimeout 驗證 Start 在無連線時正確逾時
func TestWSTransportStartTimeout(t *testing.T) {
	cfg := Config{Port: 19229, Timeout: 5 * time.Second}
	tr := NewWSTransport(cfg)
	defer tr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := tr.Start(ctx)
	if err == nil {
		t.Fatal("無連線時 Start() 應逾時回傳錯誤")
	}
}

// TestWSTransportSendWaitsForConnection 驗證 Send 在無連線時會等待連入
func TestWSTransportSendWaitsForConnection(t *testing.T) {
	cfg := Config{Port: 19230, Timeout: 5 * time.Second}
	tr := NewWSTransport(cfg)

	// Start 短逾時（server 啟動但不等連線）
	startCtx, startCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer startCancel()
	_ = tr.Start(startCtx) // 逾時正常
	defer tr.Close()

	// 在背景延遲連線，模擬 Extension
	go func() {
		time.Sleep(300 * time.Millisecond)
		conn := dialTestWS(t, 19230)
		defer conn.Close()

		// 讀取 request 並回傳 response
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var req Request
		json.Unmarshal(msg, &req)
		resp := Response{
			ID:     req.ID,
			Result: json.RawMessage(`{"success":true}`),
		}
		data, _ := json.Marshal(resp)
		conn.WriteMessage(websocket.TextMessage, data)
	}()

	// Send 應等待連線後才發送
	req, _ := NewRequest("test", nil)
	sendCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := tr.Send(sendCtx, req)
	if err != nil {
		t.Fatalf("Send() 等待連線後應成功，但回傳: %v", err)
	}
	if resp.IsError() {
		t.Errorf("預期成功回應，但收到錯誤: %v", resp.Error)
	}
}
