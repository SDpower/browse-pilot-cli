package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SDpower/browse-pilot-cli/internal/transport"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type stubTransport struct {
	send func(context.Context, *transport.Request) (*transport.Response, error)
}

func (s *stubTransport) Start(context.Context) error { return nil }
func (s *stubTransport) Send(ctx context.Context, req *transport.Request) (*transport.Response, error) {
	return s.send(ctx, req)
}
func (s *stubTransport) Close() error      { return nil }
func (s *stubTransport) IsConnected() bool { return true }
func (s *stubTransport) Type() string      { return "stub" }

func newHTTPTestServer(t *testing.T, tr transport.Transport) (*httptest.Server, *Server) {
	t.Helper()
	server := NewServer(tr, false)
	server.SetBrowserContext("firefox", 9222)
	RegisterAllTools(server)
	RegisterAllResources(server)
	httpServer := httptest.NewServer(server.HTTPHandler())
	t.Cleanup(httpServer.Close)
	return httpServer, server
}

func connectHTTPClient(t *testing.T, endpoint string) *sdkmcp.ClientSession {
	t.Helper()
	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "browse-pilot-test", Version: "1.0.0"}, nil)
	session, err := client.Connect(t.Context(), &sdkmcp.StreamableClientTransport{
		Endpoint: endpoint, DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("HTTP MCP 初始化失敗: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func decodeToolError(t *testing.T, result *sdkmcp.CallToolResult) toolErrorEnvelope {
	t.Helper()
	if !result.IsError || len(result.Content) != 1 {
		t.Fatalf("預期單一 tool error content，實際為 %#v", result)
	}
	text, ok := result.Content[0].(*sdkmcp.TextContent)
	if !ok {
		t.Fatalf("tool error content 應為文字，實際為 %T", result.Content[0])
	}
	var envelope toolErrorEnvelope
	if err := json.Unmarshal([]byte(text.Text), &envelope); err != nil {
		t.Fatalf("tool error text 不是 JSON: %v", err)
	}
	return envelope
}

func TestHTTPServerAllowsMultipleClients(t *testing.T) {
	httpServer, _ := newHTTPTestServer(t, &stubTransport{send: func(_ context.Context, req *transport.Request) (*transport.Response, error) {
		return &transport.Response{ID: req.ID, Result: json.RawMessage(`{"url":"https://example.com"}`)}, nil
	}})

	first := connectHTTPClient(t, httpServer.URL)
	second := connectHTTPClient(t, httpServer.URL)
	for index, session := range []*sdkmcp.ClientSession{first, second} {
		listed, err := session.ListTools(t.Context(), nil)
		if err != nil {
			t.Fatalf("第 %d 個 client 無法列出工具: %v", index+1, err)
		}
		if len(listed.Tools) == 0 {
			t.Fatalf("第 %d 個 client 未取得工具", index+1)
		}
		result, err := session.CallTool(t.Context(), &sdkmcp.CallToolParams{Name: "bp_state"})
		if err != nil || result.IsError {
			t.Fatalf("第 %d 個 client 呼叫 bp_state 失敗: result=%#v err=%v", index+1, result, err)
		}
	}
}

func TestHTTPServerRegistersToolsAndResources(t *testing.T) {
	httpServer, _ := newHTTPTestServer(t, nil)
	session := connectHTTPClient(t, httpServer.URL)

	tools, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("列出工具失敗: %v", err)
	}
	want := map[string]bool{"bp_navigate": false, "bp_state": false, "bp_click": false, "bp_screenshot": false, "bp_eval": false}
	for _, tool := range tools.Tools {
		if _, ok := want[tool.Name]; ok {
			want[tool.Name] = true
		}
		if tool.Name == "bp_state" && (tool.Annotations == nil || !tool.Annotations.ReadOnlyHint) {
			t.Error("bp_state 應標記為唯讀工具")
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("缺少必要工具: %s", name)
		}
	}

	resources, err := session.ListResources(t.Context(), nil)
	if err != nil {
		t.Fatalf("列出資源失敗: %v", err)
	}
	if len(resources.Resources) == 0 {
		t.Fatal("未註冊 MCP resources")
	}
}

func TestHTTPServerToolCallWithoutExtension(t *testing.T) {
	httpServer, _ := newHTTPTestServer(t, nil)
	session := connectHTTPClient(t, httpServer.URL)

	result, err := session.CallTool(t.Context(), &sdkmcp.CallToolParams{Name: "bp_state"})
	if err != nil {
		t.Fatalf("Extension 未連線應回傳 tool error，而非協議錯誤: %v", err)
	}
	envelope := decodeToolError(t, result)
	if envelope.OK || envelope.Error.Code != transport.ErrConnectionError || envelope.Error.Name != "ConnectionError" {
		t.Fatalf("連線錯誤識別不正確: %+v", envelope.Error)
	}
	if !envelope.Error.Retryable || envelope.Error.Action == "" {
		t.Fatalf("連線錯誤缺少可操作資訊: %+v", envelope.Error)
	}
}

func TestHTTPServerPreservesExtensionRPCError(t *testing.T) {
	rpcData := json.RawMessage(`{"index":7,"selector":"#submit"}`)
	httpServer, _ := newHTTPTestServer(t, &stubTransport{send: func(_ context.Context, req *transport.Request) (*transport.Response, error) {
		return &transport.Response{ID: req.ID, Error: &transport.RPCError{Code: transport.ErrElementNotFound, Message: "找不到指定元素", Data: rpcData}}, nil
	}})
	session := connectHTTPClient(t, httpServer.URL)

	result, err := session.CallTool(t.Context(), &sdkmcp.CallToolParams{Name: "bp_click", Arguments: map[string]any{"index": 7}})
	if err != nil {
		t.Fatalf("Extension RPCError 應回傳 tool error: %v", err)
	}
	envelope := decodeToolError(t, result)
	if envelope.Error.Code != transport.ErrElementNotFound || envelope.Error.Name != "ElementNotFound" || envelope.Error.Message != "找不到指定元素" {
		t.Fatalf("Extension 錯誤未完整保留: %+v", envelope.Error)
	}
	data, ok := envelope.Error.Data.(map[string]any)
	if !ok || data["index"] != float64(7) || data["selector"] != "#submit" {
		t.Fatalf("Extension 錯誤 data 未保留: %#v", envelope.Error.Data)
	}
}

func TestHTTPServerToolCallsUseIndependentTimeouts(t *testing.T) {
	var calls atomic.Int32
	httpServer, server := newHTTPTestServer(t, &stubTransport{send: func(ctx context.Context, _ *transport.Request) (*transport.Response, error) {
		calls.Add(1)
		<-ctx.Done()
		return nil, ctx.Err()
	}})
	server.SetRequestTimeout(30 * time.Millisecond)
	session := connectHTTPClient(t, httpServer.URL)

	for index := 1; index <= 2; index++ {
		startedAt := time.Now()
		result, err := session.CallTool(t.Context(), &sdkmcp.CallToolParams{Name: "bp_state"})
		if err != nil {
			t.Fatalf("第 %d 次呼叫應回傳 tool error: %v", index, err)
		}
		if elapsed := time.Since(startedAt); elapsed > 500*time.Millisecond {
			t.Fatalf("第 %d 次工具呼叫未依獨立 timeout 返回，耗時 %v", index, elapsed)
		}
		envelope := decodeToolError(t, result)
		if envelope.Error.Code != transport.ErrTimeoutError || !envelope.Error.Retryable {
			t.Fatalf("第 %d 次工具呼叫應回傳可重試 TimeoutError: %+v", index, envelope.Error)
		}
	}
	if calls.Load() != 2 {
		t.Fatalf("兩次工具呼叫未獨立執行，實際呼叫 %d 次", calls.Load())
	}
}

func TestHTTPServerUnknownErrorDoesNotLeakSensitiveDetails(t *testing.T) {
	server := NewServer(nil, false)
	server.RegisterTool(&Tool{Name: "unsafe", Description: "測試未知錯誤", InputSchema: map[string]any{"type": "object"}, Handler: func(context.Context, json.RawMessage) (any, error) {
		return nil, errors.New("stack trace /Volumes/private/project cookie=session-secret")
	}})
	httpServer := httptest.NewServer(server.HTTPHandler())
	t.Cleanup(httpServer.Close)
	session := connectHTTPClient(t, httpServer.URL)

	result, err := session.CallTool(t.Context(), &sdkmcp.CallToolParams{Name: "unsafe"})
	if err != nil {
		t.Fatalf("未知執行錯誤應回傳安全的 tool error: %v", err)
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if strings.Contains(text, "Volumes") || strings.Contains(text, "session-secret") || strings.Contains(text, "stack trace") {
		t.Fatalf("未知錯誤洩漏敏感細節: %s", text)
	}
	envelope := decodeToolError(t, result)
	if envelope.Error.Code != transport.ErrExtensionError || envelope.Error.Name != "ExtensionError" || envelope.Error.Retryable {
		t.Fatalf("未知錯誤映射不正確: %+v", envelope.Error)
	}
}

func TestHTTPServerRejectsCrossOriginRequests(t *testing.T) {
	httpServer, _ := newHTTPTestServer(t, nil)
	req, err := http.NewRequest(http.MethodPost, httpServer.URL, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Origin", "https://attacker.example")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("跨來源請求應被拒絕，實際狀態碼 %d", resp.StatusCode)
	}
}
