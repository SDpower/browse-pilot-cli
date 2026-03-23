package transport

import (
	"encoding/json"
	"testing"
)

func TestNewRequest(t *testing.T) {
	// 測試建立 request
	req, err := NewRequest("navigate", map[string]string{"url": "https://example.com"})
	if err != nil {
		t.Fatalf("NewRequest 失敗: %v", err)
	}
	if req.ID == "" {
		t.Error("Request ID 不應為空")
	}
	if req.Method != "navigate" {
		t.Errorf("Method 應為 navigate, 實際為 %s", req.Method)
	}

	// 驗證 params 可反序列化
	var params map[string]string
	if err := json.Unmarshal(req.Params, &params); err != nil {
		t.Fatalf("反序列化 params 失敗: %v", err)
	}
	if params["url"] != "https://example.com" {
		t.Errorf("URL 不正確: %s", params["url"])
	}
}

func TestNewRequestNilParams(t *testing.T) {
	req, err := NewRequest("get_state", nil)
	if err != nil {
		t.Fatalf("NewRequest 失敗: %v", err)
	}
	if req.Params != nil {
		t.Error("nil params 應產生 nil RawMessage")
	}
}

func TestResponseParseResult(t *testing.T) {
	resp := &Response{
		ID:     "test-id",
		Result: json.RawMessage(`{"success":true,"url":"https://example.com"}`),
	}

	var result struct {
		Success bool   `json:"success"`
		URL     string `json:"url"`
	}
	if err := resp.ParseResult(&result); err != nil {
		t.Fatalf("ParseResult 失敗: %v", err)
	}
	if !result.Success {
		t.Error("success 應為 true")
	}
}

func TestResponseIsError(t *testing.T) {
	resp := &Response{
		ID: "test-id",
		Error: &RPCError{
			Code:    ErrElementNotFound,
			Message: "元素索引 99 不存在",
		},
	}
	if !resp.IsError() {
		t.Error("應偵測到錯誤")
	}
	if resp.Error.Code != ErrElementNotFound {
		t.Errorf("錯誤碼應為 %d", ErrElementNotFound)
	}
}

func TestRPCErrorImplementsError(t *testing.T) {
	e := &RPCError{Code: ErrTimeoutError, Message: "操作逾時"}
	var err error = e
	if err.Error() == "" {
		t.Error("Error() 不應回傳空字串")
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		code     int
		expected int
	}{
		{ErrConnectionError, 2},
		{ErrTimeoutError, 3},
		{ErrElementNotFound, 4},
		{ErrPermissionError, 5},
		{ErrBrowserNotFound, 6},
		{ErrExtensionError, 1},
	}
	for _, tt := range tests {
		got := ExitCode(&RPCError{Code: tt.code})
		if got != tt.expected {
			t.Errorf("ExitCode(%d) = %d, 預期 %d", tt.code, got, tt.expected)
		}
	}
}

func TestErrorName(t *testing.T) {
	if ErrorName(ErrParseError) != "ParseError" {
		t.Error("ErrorName 不正確")
	}
	if ErrorName(9999) != "UnknownError" {
		t.Error("未知錯誤碼應回傳 UnknownError")
	}
}

func TestRequestSerialization(t *testing.T) {
	req, _ := NewRequest("click", map[string]int{"index": 3})
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("序列化失敗: %v", err)
	}

	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("反序列化失敗: %v", err)
	}
	if decoded.Method != "click" {
		t.Errorf("Method 不一致: %s", decoded.Method)
	}
}
