package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// TestServerInitialize 測試 MCP 初始化握手流程。
// 驗證 server 回傳正確的 protocolVersion 和 capabilities。
func TestServerInitialize(t *testing.T) {
	// 建立 server（無 transport，僅測試協議處理）
	s := NewServer(nil, false)

	var outBuf bytes.Buffer
	s.writer = &outBuf

	// 模擬 initialize request
	req := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}` + "\n"
	s.reader = bufio.NewReader(strings.NewReader(req))

	// 手動呼叫 handleRequest 測試協議處理邏輯
	var jrpcReq jsonRPCRequest
	if err := json.Unmarshal([]byte(req), &jrpcReq); err != nil {
		t.Fatalf("無法解析測試請求: %v", err)
	}
	s.handleRequest(context.Background(), &jrpcReq)

	// 驗證回應
	var resp jsonRPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("無法解析回應: %v", err)
	}

	// JSON 數字預設解析為 float64
	if resp.ID != float64(1) {
		t.Errorf("ID 不正確，期望 1，實際為 %v", resp.ID)
	}
	if resp.Error != nil {
		t.Errorf("不應有錯誤: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("result 不應為 nil")
	}

	// 驗證 protocolVersion
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result 應為 map")
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("protocolVersion 不正確: %v", result["protocolVersion"])
	}
}

// TestServerToolsList 測試 tools/list 方法。
// 驗證所有已註冊的 tool 都會出現在回應清單中。
func TestServerToolsList(t *testing.T) {
	s := NewServer(nil, false)
	RegisterAllTools(s)

	var outBuf bytes.Buffer
	s.writer = &outBuf

	req := jsonRPCRequest{JSONRPC: "2.0", ID: 2, Method: "tools/list"}
	s.handleRequest(context.Background(), &req)

	var resp jsonRPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("無法解析回應: %v", err)
	}

	if resp.Error != nil {
		t.Errorf("不應有錯誤: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result 應為 map")
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("tools 應為 array")
	}
	if len(tools) == 0 {
		t.Error("應有 tools 已被註冊")
	}

	// 確認包含核心 tool
	toolNames := make(map[string]bool)
	for _, ti := range tools {
		tm, ok := ti.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := tm["name"].(string); ok {
			toolNames[name] = true
		}
	}

	requiredTools := []string{"bp_navigate", "bp_state", "bp_click", "bp_screenshot", "bp_eval"}
	for _, name := range requiredTools {
		if !toolNames[name] {
			t.Errorf("缺少必要的 tool: %s", name)
		}
	}
}

// TestServerPing 測試 ping 方法。
// 驗證 server 能正常回應心跳請求。
func TestServerPing(t *testing.T) {
	s := NewServer(nil, false)
	var outBuf bytes.Buffer
	s.writer = &outBuf

	req := jsonRPCRequest{JSONRPC: "2.0", ID: 3, Method: "ping"}
	s.handleRequest(context.Background(), &req)

	var resp jsonRPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("無法解析回應: %v", err)
	}

	if resp.Error != nil {
		t.Errorf("ping 不應有錯誤: %v", resp.Error)
	}
}

// TestServerUnknownMethod 測試未知方法的錯誤處理。
// 驗證 server 回傳正確的 -32601 錯誤碼。
func TestServerUnknownMethod(t *testing.T) {
	s := NewServer(nil, false)
	var outBuf bytes.Buffer
	s.writer = &outBuf

	req := jsonRPCRequest{JSONRPC: "2.0", ID: 4, Method: "unknown/method"}
	s.handleRequest(context.Background(), &req)

	var resp jsonRPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("無法解析回應: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("未知方法應回傳錯誤")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("錯誤碼應為 -32601，實際為 %d", resp.Error.Code)
	}
}

// TestServerResourcesList 測試 resources/list 方法。
// 驗證所有已註冊的 resource 都會出現在回應清單中。
func TestServerResourcesList(t *testing.T) {
	s := NewServer(nil, false)
	RegisterAllResources(s)

	var outBuf bytes.Buffer
	s.writer = &outBuf

	req := jsonRPCRequest{JSONRPC: "2.0", ID: 5, Method: "resources/list"}
	s.handleRequest(context.Background(), &req)

	var resp jsonRPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("無法解析回應: %v", err)
	}

	if resp.Error != nil {
		t.Errorf("不應有錯誤: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result 應為 map")
	}
	resources, ok := result["resources"].([]any)
	if !ok {
		t.Fatal("resources 應為 array")
	}
	if len(resources) == 0 {
		t.Error("應有 resources 已被註冊")
	}
}

// TestServerInitializedNotification 測試 initialized 通知（無回應）。
// 驗證 server 不會對 initialized 通知回應任何內容。
func TestServerInitializedNotification(t *testing.T) {
	s := NewServer(nil, false)
	var outBuf bytes.Buffer
	s.writer = &outBuf

	// initialized 是 notification（無 ID），server 不應回傳任何內容
	req := jsonRPCRequest{JSONRPC: "2.0", Method: "initialized"}
	s.handleRequest(context.Background(), &req)

	if outBuf.Len() > 0 {
		t.Errorf("initialized 通知不應有回應，但收到: %s", outBuf.String())
	}
}

// TestServerInitializeWithoutTransport 驗證 MCP server 在 transport 未連線時
// 仍能正常回應 initialize 和 tools/list — 這是 MCP 模式的核心行為：
// server 先啟動回應協議握手，Extension 隨後連入。
func TestServerInitializeWithoutTransport(t *testing.T) {
	// transport 為 nil，模擬 Extension 尚未連入
	s := NewServer(nil, false)
	RegisterAllTools(s)
	RegisterAllResources(s)

	// 測試 initialize
	var outBuf bytes.Buffer
	s.writer = &outBuf

	initReq := jsonRPCRequest{JSONRPC: "2.0", ID: 1, Method: "initialize",
		Params: json.RawMessage(`{"protocolVersion":"2024-11-05"}`)}
	s.handleRequest(context.Background(), &initReq)

	var initResp jsonRPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &initResp); err != nil {
		t.Fatalf("initialize 回應解析失敗: %v", err)
	}
	if initResp.Error != nil {
		t.Errorf("initialize 不應有錯誤: %v", initResp.Error)
	}

	// 測試 tools/list（也不需要 transport）
	outBuf.Reset()
	listReq := jsonRPCRequest{JSONRPC: "2.0", ID: 2, Method: "tools/list"}
	s.handleRequest(context.Background(), &listReq)

	var listResp jsonRPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &listResp); err != nil {
		t.Fatalf("tools/list 回應解析失敗: %v", err)
	}
	if listResp.Error != nil {
		t.Errorf("tools/list 不應有錯誤: %v", listResp.Error)
	}
}

// TestServerToolCallWithoutTransport 驗證 transport 未連線時
// tool call 回傳「Extension 未連線」錯誤而非 panic。
func TestServerToolCallWithoutTransport(t *testing.T) {
	s := NewServer(nil, false)
	RegisterAllTools(s)

	var outBuf bytes.Buffer
	s.writer = &outBuf

	paramsJSON, _ := json.Marshal(map[string]any{
		"name":      "bp_navigate",
		"arguments": json.RawMessage(`{"url":"https://example.com"}`),
	})

	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      10,
		Method:  "tools/call",
		Params:  paramsJSON,
	}
	s.handleRequest(context.Background(), &req)

	var resp jsonRPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("回應解析失敗: %v", err)
	}

	// 應回傳 tool error（isError: true），而非 JSON-RPC error
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result 應為 map")
	}
	isError, _ := result["isError"].(bool)
	if !isError {
		t.Error("transport 未連線時 tool call 應回傳 isError: true")
	}
}

// TestServerToolsCallUnknownTool 測試呼叫未知 tool 時的錯誤處理。
func TestServerToolsCallUnknownTool(t *testing.T) {
	s := NewServer(nil, false)
	var outBuf bytes.Buffer
	s.writer = &outBuf

	paramsJSON, _ := json.Marshal(map[string]any{
		"name":      "nonexistent_tool",
		"arguments": map[string]any{},
	})

	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      6,
		Method:  "tools/call",
		Params:  paramsJSON,
	}
	s.handleRequest(context.Background(), &req)

	var resp jsonRPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("無法解析回應: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("呼叫未知 tool 應回傳錯誤")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("錯誤碼應為 -32602，實際為 %d", resp.Error.Code)
	}
}
