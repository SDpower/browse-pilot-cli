package transport

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"testing"
	"time"
)

// buildNMMessage 建立一個 length-prefixed JSON 訊息（用於測試）
func buildNMMessage(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("buildNMMessage marshal 失敗: %v", err)
	}
	buf := make([]byte, 4+len(data))
	binary.LittleEndian.PutUint32(buf[:4], uint32(len(data)))
	copy(buf[4:], data)
	return buf
}

// TestNMTransportReadWriteMessage 驗證 readMessage/writeMessage 的 length-prefix 格式
func TestNMTransportReadWriteMessage(t *testing.T) {
	// 準備測試資料
	original := map[string]string{"key": "value", "hello": "world"}
	originalData, _ := json.Marshal(original)

	// 用 buffer 模擬 stdin：寫入 length-prefixed 訊息
	buf := bytes.NewBuffer(buildNMMessage(t, original))

	// 建立 NMTransport，注入 reader
	tr := &NMTransport{
		config:  Config{},
		reader:  buf,
		writer:  &bytes.Buffer{},
		pending: make(map[string]chan *Response),
		done:    make(chan struct{}),
	}

	// 讀取訊息
	got, err := tr.readMessage()
	if err != nil {
		t.Fatalf("readMessage() 失敗: %v", err)
	}
	if string(got) != string(originalData) {
		t.Errorf("readMessage() = %q，預期 %q", got, originalData)
	}

	// 測試 writeMessage
	outBuf := &bytes.Buffer{}
	tr.writer = outBuf

	if err := tr.writeMessage(originalData); err != nil {
		t.Fatalf("writeMessage() 失敗: %v", err)
	}

	// 驗證輸出格式
	written := outBuf.Bytes()
	if len(written) < 4 {
		t.Fatalf("writeMessage 輸出太短: %d bytes", len(written))
	}
	length := binary.LittleEndian.Uint32(written[:4])
	if int(length) != len(originalData) {
		t.Errorf("length prefix = %d，預期 %d", length, len(originalData))
	}
	if string(written[4:]) != string(originalData) {
		t.Errorf("寫入的 JSON = %q，預期 %q", written[4:], originalData)
	}
}

// TestNMTransportSendReceive 驗證 Send 可正確發送並接收回應
func TestNMTransportSendReceive(t *testing.T) {
	// 使用 io.Pipe 模擬雙向通道
	// CLI → Extension: cliWriter → extReader
	// Extension → CLI: extWriter → cliReader
	cliReader, extWriter := io.Pipe()
	extReader, cliWriter := io.Pipe()

	cfg := Config{Timeout: 3 * time.Second}
	tr := NewNMTransport(cfg)
	tr.reader = cliReader
	tr.writer = cliWriter

	ctx := context.Background()
	if err := tr.Start(ctx); err != nil {
		t.Fatalf("Start() 失敗: %v", err)
	}
	defer tr.Close()

	// 模擬 Extension：讀取 request，回傳 response
	go func() {
		defer extReader.Close()
		defer extWriter.Close()

		// 讀取 length prefix
		lenBuf := make([]byte, 4)
		if _, err := io.ReadFull(extReader, lenBuf); err != nil {
			return
		}
		msgLen := binary.LittleEndian.Uint32(lenBuf)

		// 讀取 JSON
		msgBuf := make([]byte, msgLen)
		if _, err := io.ReadFull(extReader, msgBuf); err != nil {
			return
		}

		var req Request
		if err := json.Unmarshal(msgBuf, &req); err != nil {
			return
		}

		// 建立回應
		resp := Response{
			ID:     req.ID,
			Result: json.RawMessage(`{"status":"ok"}`),
		}
		respData, _ := json.Marshal(resp)

		// 寫入 length-prefixed 回應
		lenOut := make([]byte, 4)
		binary.LittleEndian.PutUint32(lenOut, uint32(len(respData)))
		extWriter.Write(lenOut)
		extWriter.Write(respData)
	}()

	req, err := NewRequest("getTab", nil)
	if err != nil {
		t.Fatalf("NewRequest() 失敗: %v", err)
	}

	sendCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
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

// TestNMTransportSendTimeout 驗證當無回應時 Send 因 ctx 逾時返回錯誤
func TestNMTransportSendTimeout(t *testing.T) {
	// 使用 bytes.Buffer 作為 writer（不阻塞），用 io.Pipe 的 reader 端阻塞讀取
	cliReader, _ := io.Pipe()
	writeBuf := &bytes.Buffer{}

	cfg := Config{Timeout: 5 * time.Second}
	tr := NewNMTransport(cfg)
	tr.reader = cliReader
	tr.writer = writeBuf

	ctx := context.Background()
	if err := tr.Start(ctx); err != nil {
		t.Fatalf("Start() 失敗: %v", err)
	}
	defer tr.Close()

	req, _ := NewRequest("navigate", nil)

	// 使用短逾時
	sendCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	_, err := tr.Send(sendCtx, req)
	if err == nil {
		t.Fatal("預期收到逾時錯誤，但 Send() 回傳 nil")
	}
}

// TestNMTransportLargeMessageChunking 驗證大訊息（接近 1MB 限制）的讀寫
func TestNMTransportLargeMessageChunking(t *testing.T) {
	// 建立接近但不超過 1MB 的訊息
	largeData := make([]byte, 900*1024) // 900KB
	for i := range largeData {
		largeData[i] = 'x'
	}

	// 用 buffer 模擬
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, uint32(len(largeData)))
	buf.Write(largeData)

	tr := &NMTransport{
		config:  Config{},
		reader:  buf,
		writer:  &bytes.Buffer{},
		pending: make(map[string]chan *Response),
		done:    make(chan struct{}),
	}

	got, err := tr.readMessage()
	if err != nil {
		t.Fatalf("readMessage() 大訊息失敗: %v", err)
	}
	if len(got) != len(largeData) {
		t.Errorf("讀取長度 = %d，預期 %d", len(got), len(largeData))
	}
}

// TestNMTransportClose 驗證 Close 後 IsConnected 回傳 false
func TestNMTransportClose(t *testing.T) {
	cliReader, _ := io.Pipe()
	writeBuf := &bytes.Buffer{}

	cfg := Config{}
	tr := NewNMTransport(cfg)
	tr.reader = cliReader
	tr.writer = writeBuf

	ctx := context.Background()
	tr.Start(ctx)

	if !tr.IsConnected() {
		t.Error("Start() 後 IsConnected() 應為 true")
	}

	tr.Close()

	if tr.IsConnected() {
		t.Error("Close() 後 IsConnected() 應為 false")
	}
	if tr.Type() != "native_messaging" {
		t.Errorf("Type() = %q，預期 \"native_messaging\"", tr.Type())
	}
}
