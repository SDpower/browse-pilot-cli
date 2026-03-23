package cli

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"

	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

// TestReadNMMessage 測試正常讀取 length-prefixed 訊息。
func TestReadNMMessage(t *testing.T) {
	// 建立一條 length-prefixed 訊息
	msg := []byte(`{"id":"test","method":"get_status"}`)
	var buf bytes.Buffer
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(msg)))
	buf.Write(lenBuf)
	buf.Write(msg)

	result, err := readNMMessage(&buf)
	if err != nil {
		t.Fatalf("readNMMessage 失敗: %v", err)
	}

	var req transport.Request
	if err := json.Unmarshal(result, &req); err != nil {
		t.Fatalf("JSON 解析失敗: %v", err)
	}
	if req.Method != "get_status" {
		t.Errorf("method 應為 get_status，實際為 %s", req.Method)
	}
	if req.ID != "test" {
		t.Errorf("id 應為 test，實際為 %s", req.ID)
	}
}

// TestWriteNMResponse 測試將 Response 以正確格式寫入緩衝區。
func TestWriteNMResponse(t *testing.T) {
	resp := &transport.Response{
		ID:     "test-id",
		Result: json.RawMessage(`{"success":true}`),
	}

	var buf bytes.Buffer
	if err := writeNMResponse(&buf, resp); err != nil {
		t.Fatalf("writeNMResponse 失敗: %v", err)
	}

	// 驗證輸出格式
	data := buf.Bytes()
	if len(data) < 4 {
		t.Fatal("輸出資料太短（少於 4 bytes）")
	}

	// 驗證 length prefix 與實際 JSON 長度一致
	msgLen := binary.LittleEndian.Uint32(data[:4])
	if int(msgLen) != len(data)-4 {
		t.Errorf("length header 不一致: header=%d, actual=%d", msgLen, len(data)-4)
	}

	// 驗證 JSON 內容可正確解析
	var decoded transport.Response
	if err := json.Unmarshal(data[4:], &decoded); err != nil {
		t.Fatalf("JSON 解析失敗: %v", err)
	}
	if decoded.ID != "test-id" {
		t.Errorf("ID 不一致: 期望 test-id，實際 %s", decoded.ID)
	}
}

// TestReadNMMessageTooLarge 測試超過 1MB 限制時應回傳錯誤。
func TestReadNMMessageTooLarge(t *testing.T) {
	var buf bytes.Buffer
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, 2*1024*1024) // 2MB，超過限制
	buf.Write(lenBuf)

	_, err := readNMMessage(&buf)
	if err == nil {
		t.Error("超過 1MB 應回傳錯誤，但 err 為 nil")
	}
}

// TestReadNMMessageZeroLength 測試長度為 0 時應回傳錯誤。
func TestReadNMMessageZeroLength(t *testing.T) {
	var buf bytes.Buffer
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, 0)
	buf.Write(lenBuf)

	_, err := readNMMessage(&buf)
	if err == nil {
		t.Error("長度為 0 應回傳錯誤，但 err 為 nil")
	}
}

// TestReadNMMessageEOF 測試空輸入（EOF）時應回傳 io.ErrUnexpectedEOF 或 io.EOF。
func TestReadNMMessageEOF(t *testing.T) {
	var buf bytes.Buffer
	// 不寫入任何資料，模擬 EOF

	_, err := readNMMessage(&buf)
	if err == nil {
		t.Error("空輸入應回傳 EOF 錯誤，但 err 為 nil")
	}
}

// TestWriteNMResponseError 測試 writeNMResponse 能正確寫入錯誤回應。
func TestWriteNMResponseError(t *testing.T) {
	resp := &transport.Response{
		ID: "err-id",
		Error: &transport.RPCError{
			Code:    transport.ErrParseError,
			Message: "解析失敗",
		},
	}

	var buf bytes.Buffer
	if err := writeNMResponse(&buf, resp); err != nil {
		t.Fatalf("writeNMResponse 失敗: %v", err)
	}

	data := buf.Bytes()
	if len(data) < 4 {
		t.Fatal("輸出資料太短")
	}

	var decoded transport.Response
	if err := json.Unmarshal(data[4:], &decoded); err != nil {
		t.Fatalf("JSON 解析失敗: %v", err)
	}
	if decoded.Error == nil {
		t.Fatal("Error 欄位應不為 nil")
	}
	if decoded.Error.Code != transport.ErrParseError {
		t.Errorf("錯誤碼不一致: 期望 %d，實際 %d", transport.ErrParseError, decoded.Error.Code)
	}
}
