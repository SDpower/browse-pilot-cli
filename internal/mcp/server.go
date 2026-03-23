// Package mcp 實作 MCP (Model Context Protocol) stdio server，
// 讓 Claude Code 等 AI 工具可透過 MCP 直接操控瀏覽器。
package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

// Server 是 MCP stdio server 主結構。
// 透過 stdin 讀取 JSON-RPC 2.0 請求，將操作轉發至瀏覽器 Extension，
// 再將結果透過 stdout 回傳給 MCP 客戶端。
type Server struct {
	// transport 是與瀏覽器 Extension 通訊的底層傳輸層
	transport transport.Transport

	// reader 從 stdin 讀取 MCP 客戶端的請求
	reader *bufio.Reader

	// writer 向 stdout 寫入回應給 MCP 客戶端
	writer io.Writer

	// tools 儲存已註冊的 MCP tool，以 name 為 key
	tools map[string]*Tool

	// resources 儲存已註冊的 MCP resource，以 URI 為 key
	resources map[string]*Resource

	// mu 保護 writer 的並發安全
	mu sync.Mutex

	// verbose 若為 true，將詳細日誌輸出至 stderr
	verbose bool
}

// Tool 定義一個 MCP tool（對應瀏覽器操作指令）。
type Tool struct {
	// Name 是 tool 的唯一識別名稱
	Name string `json:"name"`

	// Description 描述 tool 的功能，供 AI 模型理解用途
	Description string `json:"description"`

	// InputSchema 是 JSON Schema 格式的參數定義
	InputSchema map[string]any `json:"inputSchema"`

	// Handler 是 tool 的執行函數，接收 JSON 格式參數並回傳結果
	Handler func(ctx context.Context, params json.RawMessage) (any, error)
}

// Resource 定義一個 MCP resource（代表可讀取的瀏覽器狀態）。
type Resource struct {
	// URI 是 resource 的唯一識別符，例如 "bp://state"
	URI string `json:"uri"`

	// Name 是人類可讀的 resource 名稱
	Name string `json:"name"`

	// Description 描述 resource 的內容
	Description string `json:"description,omitempty"`

	// MimeType 是 resource 內容的 MIME 類型
	MimeType string `json:"mimeType,omitempty"`

	// Handler 是讀取 resource 的函數，回傳文字內容
	Handler func(ctx context.Context) (string, error)
}

// jsonRPCRequest 是 JSON-RPC 2.0 請求訊息格式。
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// jsonRPCResponse 是 JSON-RPC 2.0 回應訊息格式。
type jsonRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id,omitempty"`
	Result  any           `json:"result,omitempty"`
	Error   *jsonRPCError `json:"error,omitempty"`
}

// jsonRPCError 是 JSON-RPC 2.0 錯誤物件格式。
type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// NewServer 建立一個新的 MCP server。
// tr 是與瀏覽器 Extension 通訊的 transport，可為 nil（測試時使用）。
// verbose 控制是否輸出詳細的除錯日誌。
func NewServer(tr transport.Transport, verbose bool) *Server {
	return &Server{
		transport: tr,
		reader:    bufio.NewReader(os.Stdin),
		writer:    os.Stdout,
		tools:     make(map[string]*Tool),
		resources: make(map[string]*Resource),
		verbose:   verbose,
	}
}

// RegisterTool 向 server 註冊一個 MCP tool。
func (s *Server) RegisterTool(tool *Tool) {
	s.tools[tool.Name] = tool
}

// RegisterResource 向 server 註冊一個 MCP resource。
func (s *Server) RegisterResource(res *Resource) {
	s.resources[res.URI] = res
}

// Run 啟動 MCP server 主迴圈，持續從 stdin 讀取並處理 JSON-RPC 請求。
// 當 ctx 被取消或 stdin 關閉時，函數返回。
func (s *Server) Run(ctx context.Context) error {
	if s.verbose {
		fmt.Fprintln(os.Stderr, "[MCP] 啟動 MCP server (stdio)")
	}

	for {
		// 檢查 context 是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 從 stdin 讀取一行 JSON
		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("讀取 stdin 失敗: %w", err)
		}

		// 跳過空行
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		// 解析 JSON-RPC request
		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(nil, -32700, "JSON 解析失敗")
			continue
		}

		if s.verbose {
			fmt.Fprintf(os.Stderr, "[MCP] 收到: method=%s\n", req.Method)
		}

		// 根據 method 路由至對應的 handler
		s.handleRequest(ctx, &req)
	}
}

// handleRequest 根據 JSON-RPC method 將請求路由至對應的處理函數。
func (s *Server) handleRequest(ctx context.Context, req *jsonRPCRequest) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "initialized":
		// 客戶端確認通知，不需回應
	case "tools/list":
		s.handleToolsList(req)
	case "tools/call":
		s.handleToolsCall(ctx, req)
	case "resources/list":
		s.handleResourcesList(req)
	case "resources/read":
		s.handleResourcesRead(ctx, req)
	case "ping":
		s.sendResult(req.ID, map[string]string{})
	default:
		s.sendError(req.ID, -32601, fmt.Sprintf("未知方法: %s", req.Method))
	}
}

// handleInitialize 處理 MCP 初始化握手，回傳 server 能力宣告。
func (s *Server) handleInitialize(req *jsonRPCRequest) {
	s.sendResult(req.ID, map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools":     map[string]any{},
			"resources": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "browse-pilot",
			"version": "1.0.0",
		},
	})
}

// handleToolsList 回傳所有已註冊 tool 的清單。
func (s *Server) handleToolsList(req *jsonRPCRequest) {
	tools := make([]map[string]any, 0, len(s.tools))
	for _, t := range s.tools {
		tools = append(tools, map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"inputSchema": t.InputSchema,
		})
	}
	// 確保 null 不出現在回應中
	if tools == nil {
		tools = []map[string]any{}
	}
	s.sendResult(req.ID, map[string]any{"tools": tools})
}

// handleToolsCall 呼叫指定的 tool 並回傳執行結果。
func (s *Server) handleToolsCall(ctx context.Context, req *jsonRPCRequest) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "無效的參數格式")
		return
	}

	tool, ok := s.tools[params.Name]
	if !ok {
		s.sendError(req.ID, -32602, fmt.Sprintf("未知的 tool: %s", params.Name))
		return
	}

	result, err := tool.Handler(ctx, params.Arguments)
	if err != nil {
		// MCP 規範：tool 執行錯誤以 content 形式回傳，而非 JSON-RPC error
		s.sendResult(req.ID, map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": fmt.Sprintf("錯誤: %s", err.Error())},
			},
			"isError": true,
		})
		return
	}

	// 序列化成功結果為文字
	resultJSON, _ := json.Marshal(result)
	s.sendResult(req.ID, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(resultJSON)},
		},
	})
}

// handleResourcesList 回傳所有已註冊 resource 的清單。
func (s *Server) handleResourcesList(req *jsonRPCRequest) {
	resources := make([]map[string]any, 0, len(s.resources))
	for _, r := range s.resources {
		res := map[string]any{
			"uri":  r.URI,
			"name": r.Name,
		}
		if r.Description != "" {
			res["description"] = r.Description
		}
		if r.MimeType != "" {
			res["mimeType"] = r.MimeType
		}
		resources = append(resources, res)
	}
	// 確保 null 不出現在回應中
	if resources == nil {
		resources = []map[string]any{}
	}
	s.sendResult(req.ID, map[string]any{"resources": resources})
}

// handleResourcesRead 讀取指定 URI 的 resource 內容並回傳。
func (s *Server) handleResourcesRead(ctx context.Context, req *jsonRPCRequest) {
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "無效的參數")
		return
	}

	res, ok := s.resources[params.URI]
	if !ok {
		s.sendError(req.ID, -32002, fmt.Sprintf("未知的 resource: %s", params.URI))
		return
	}

	content, err := res.Handler(ctx)
	if err != nil {
		s.sendError(req.ID, -32000, err.Error())
		return
	}

	mimeType := "text/plain"
	if res.MimeType != "" {
		mimeType = res.MimeType
	}

	s.sendResult(req.ID, map[string]any{
		"contents": []map[string]any{
			{"uri": params.URI, "mimeType": mimeType, "text": content},
		},
	})
}

// sendResult 傳送 JSON-RPC 成功回應至 stdout。
func (s *Server) sendResult(id, result any) {
	s.send(&jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

// sendError 傳送 JSON-RPC 錯誤回應至 stdout。
func (s *Server) sendError(id any, code int, message string) {
	s.send(&jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &jsonRPCError{Code: code, Message: message},
	})
}

// send 序列化 jsonRPCResponse 並以換行結尾寫入 writer（stdout）。
// 使用 mutex 確保並發安全。
func (s *Server) send(resp *jsonRPCResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[MCP] JSON 序列化失敗: %v\n", err)
		return
	}

	// MCP stdio transport：每個訊息以換行結尾
	data = append(data, '\n')
	if _, err := s.writer.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "[MCP] 寫入回應失敗: %v\n", err)
	}
}
