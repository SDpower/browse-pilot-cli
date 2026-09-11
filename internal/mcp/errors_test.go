package mcp

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

func TestErrorGuidanceTable(t *testing.T) {
	tests := []struct {
		code      int
		name      string
		retryable bool
		actionHas string
	}{
		{transport.ErrConnectionError, "ConnectionError", true, "9222"},
		{transport.ErrTimeoutError, "TimeoutError", true, "頁面狀態"},
		{transport.ErrElementNotFound, "ElementNotFound", true, "bp_state"},
		{transport.ErrTabNotFound, "TabNotFound", true, "bp_tabs"},
		{transport.ErrInjectionError, "InjectionError", true, "內建頁面"},
		{transport.ErrPermissionError, "PermissionError", false, "權限"},
		{transport.ErrStaleElement, "StaleElement", true, "舊索引"},
		{transport.ErrBrowserNotFound, "BrowserNotFound", true, "Firefox"},
		{transport.ErrNativeMessagingError, "NativeMessagingError", true, "bp_cli setup"},
		{transport.ErrInvalidParams, "InvalidParams", false, "修正工具參數"},
		{transport.ErrExtensionError, "ExtensionError", false, "停止操作"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rpcErr := &transport.RPCError{Code: test.code, Message: "測試錯誤"}
			got := normalizeToolError(rpcErr, "firefox", 9222)
			if got.Code != test.code || got.Name != test.name || got.Retryable != test.retryable {
				t.Errorf("錯誤映射不正確: %+v", got)
			}
			if !strings.Contains(got.Action, test.actionHas) {
				t.Errorf("action %q 應包含 %q", got.Action, test.actionHas)
			}
		})
	}
}

func TestUnknownRPCErrorMapsToSafeExtensionError(t *testing.T) {
	err := &transport.RPCError{
		Code:    -31999,
		Message: "stack trace /Users/private/project",
		Data:    json.RawMessage(`{"cookie":"secret"}`),
	}

	encoded := marshalToolError(err, "firefox", 9222)
	if strings.Contains(encoded, "/Users/private") || strings.Contains(encoded, "secret") || strings.Contains(encoded, "stack trace") {
		t.Fatalf("未知 RPC 錯誤洩漏敏感資訊: %s", encoded)
	}

	var envelope toolErrorEnvelope
	if unmarshalErr := json.Unmarshal([]byte(encoded), &envelope); unmarshalErr != nil {
		t.Fatalf("無法解析錯誤 JSON: %v", unmarshalErr)
	}
	if envelope.Error.Code != transport.ErrExtensionError || envelope.Error.Name != "ExtensionError" || envelope.Error.Retryable {
		t.Errorf("未知 RPC 錯誤映射不正確: %+v", envelope.Error)
	}
}

func TestKnownRPCErrorPreservesFalsyData(t *testing.T) {
	err := &transport.RPCError{
		Code:    transport.ErrInvalidParams,
		Message: "參數不可為 false",
		Data:    json.RawMessage(`false`),
	}
	got := normalizeToolError(err, "firefox", 9222)
	value, ok := got.Data.(bool)
	if !ok || value {
		t.Errorf("應完整保留 false data，實際為 %#v", got.Data)
	}
}
