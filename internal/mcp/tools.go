// Package mcp 的 tools 子模組，定義所有瀏覽器操作 MCP tool。
package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

// RegisterAllTools 向 MCP server 註冊所有 browse-pilot 瀏覽器操作 tool。
// 每個 tool 對應一個可從 AI 呼叫的瀏覽器指令。
func RegisterAllTools(s *Server) {
	// 導航至指定 URL
	s.RegisterTool(&Tool{
		Name:        "bp_navigate",
		Description: "導航瀏覽器至指定 URL",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url": map[string]any{"type": "string", "description": "目標 URL"},
			},
			"required": []string{"url"},
		},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			return s.callExtensionRaw(ctx, "navigate", params)
		},
	})

	// 取得當前頁面狀態
	s.RegisterTool(&Tool{
		Name:        "bp_state",
		Description: "取得當前頁面 URL、標題、瀏覽器類型及可互動元素列表（含索引）",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			return s.callExtensionRaw(ctx, "get_state", nil)
		},
	})

	// 點擊元素
	s.RegisterTool(&Tool{
		Name:        "bp_click",
		Description: "點擊元素（使用 bp_state 的索引）",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"index": map[string]any{"type": "integer", "description": "元素索引（來自 bp_state）"},
			},
			"required": []string{"index"},
		},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			return s.callExtensionRaw(ctx, "click", params)
		},
	})

	// 點擊元素並輸入文字
	s.RegisterTool(&Tool{
		Name:        "bp_input",
		Description: "點擊元素並輸入文字",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"index": map[string]any{"type": "integer", "description": "元素索引"},
				"text":  map[string]any{"type": "string", "description": "要輸入的文字"},
			},
			"required": []string{"index", "text"},
		},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			return s.callExtensionRaw(ctx, "input_text", params)
		},
	})

	// 對當前焦點元素輸入文字
	s.RegisterTool(&Tool{
		Name:        "bp_type",
		Description: "對當前焦點元素輸入文字",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"text": map[string]any{"type": "string", "description": "要輸入的文字"},
			},
			"required": []string{"text"},
		},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			return s.callExtensionRaw(ctx, "type_text", params)
		},
	})

	// 截取頁面截圖
	s.RegisterTool(&Tool{
		Name:        "bp_screenshot",
		Description: "截取當前頁面截圖，回傳 base64 編碼圖片",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"full": map[string]any{"type": "boolean", "description": "全頁截圖", "default": false},
			},
		},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			return s.callExtensionRaw(ctx, "screenshot", params)
		},
	})

	// 執行 JavaScript
	s.RegisterTool(&Tool{
		Name:        "bp_eval",
		Description: "在頁面 context 執行 JavaScript 並回傳結果",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"code": map[string]any{"type": "string", "description": "要執行的 JavaScript 程式碼"},
			},
			"required": []string{"code"},
		},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			return s.callExtensionRaw(ctx, "eval_js", params)
		},
	})

	// 捲動頁面
	s.RegisterTool(&Tool{
		Name:        "bp_scroll",
		Description: "捲動頁面",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"direction": map[string]any{"type": "string", "enum": []string{"up", "down"}},
				"amount":    map[string]any{"type": "integer", "description": "捲動像素數"},
			},
			"required": []string{"direction"},
		},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			return s.callExtensionRaw(ctx, "scroll", params)
		},
	})

	// 送出鍵盤事件
	s.RegisterTool(&Tool{
		Name:        "bp_keys",
		Description: "送出鍵盤事件，例如 'Enter'、'Ctrl+a'、'Tab'",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"keys": map[string]any{"type": "string", "description": "按鍵組合"},
			},
			"required": []string{"keys"},
		},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			return s.callExtensionRaw(ctx, "send_keys", params)
		},
	})

	// 等待頁面條件成立
	s.RegisterTool(&Tool{
		Name:        "bp_wait",
		Description: "等待頁面上的條件成立",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"type":    map[string]any{"type": "string", "enum": []string{"selector", "text", "url"}},
				"value":   map[string]any{"type": "string", "description": "CSS 選擇器、文字或 URL pattern"},
				"state":   map[string]any{"type": "string", "enum": []string{"visible", "hidden"}, "default": "visible"},
				"timeout": map[string]any{"type": "integer", "default": 30000},
			},
			"required": []string{"type", "value"},
		},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			var p struct {
				Type    string `json:"type"`
				Value   string `json:"value"`
				State   string `json:"state"`
				Timeout int    `json:"timeout"`
			}
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("解析 bp_wait 參數失敗: %w", err)
			}

			// 根據等待類型選擇方法名稱
			method := "wait_" + p.Type // wait_selector, wait_text, wait_url
			extParams := map[string]any{"timeout": p.Timeout}
			switch p.Type {
			case "selector":
				extParams["selector"] = p.Value
				if p.State != "" {
					extParams["state"] = p.State
				}
			case "text":
				extParams["text"] = p.Value
			case "url":
				extParams["pattern"] = p.Value
			}
			rawParams, _ := json.Marshal(extParams)
			return s.callExtensionRaw(ctx, method, rawParams)
		},
	})

	// 取得頁面或元素資訊
	s.RegisterTool(&Tool{
		Name:        "bp_get",
		Description: "取得頁面或元素的資訊",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"what":     map[string]any{"type": "string", "enum": []string{"title", "html", "text", "value", "attributes", "bbox"}},
				"index":    map[string]any{"type": "integer", "description": "元素索引（text/value/attributes/bbox 需要）"},
				"selector": map[string]any{"type": "string", "description": "CSS 選擇器（html 可選）"},
			},
			"required": []string{"what"},
		},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			var p struct {
				What     string `json:"what"`
				Index    *int   `json:"index"`
				Selector string `json:"selector"`
			}
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("解析 bp_get 參數失敗: %w", err)
			}

			method := "get_" + p.What

			extParams := map[string]any{}
			if p.Index != nil {
				extParams["index"] = *p.Index
			}
			if p.Selector != "" {
				extParams["selector"] = p.Selector
			}
			rawParams, _ := json.Marshal(extParams)
			return s.callExtensionRaw(ctx, method, rawParams)
		},
	})

	// 管理瀏覽器 cookies
	s.RegisterTool(&Tool{
		Name:        "bp_cookies",
		Description: "管理瀏覽器 cookies",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{"type": "string", "enum": []string{"get", "set", "clear"}},
				"url":    map[string]any{"type": "string"},
				"name":   map[string]any{"type": "string"},
				"value":  map[string]any{"type": "string"},
				"domain": map[string]any{"type": "string"},
			},
			"required": []string{"action"},
		},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			var p struct {
				Action string `json:"action"`
			}
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("解析 bp_cookies 參數失敗: %w", err)
			}

			var method string
			switch p.Action {
			case "get":
				method = "get_cookies"
			case "set":
				method = "set_cookie"
			case "clear":
				method = "clear_cookies"
			default:
				return nil, fmt.Errorf("未知的 cookie action: %s", p.Action)
			}
			return s.callExtensionRaw(ctx, method, params)
		},
	})

	// 管理瀏覽器分頁
	s.RegisterTool(&Tool{
		Name:        "bp_tabs",
		Description: "列出、切換或關閉瀏覽器分頁",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{"type": "string", "enum": []string{"list", "switch", "close"}},
				"index":  map[string]any{"type": "integer"},
			},
			"required": []string{"action"},
		},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			var p struct {
				Action string `json:"action"`
				Index  *int   `json:"index"`
			}
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("解析 bp_tabs 參數失敗: %w", err)
			}

			var method string
			switch p.Action {
			case "list":
				method = "get_tabs"
			case "switch":
				method = "switch_tab"
			case "close":
				method = "close_tab"
			default:
				return nil, fmt.Errorf("未知的 tab action: %s", p.Action)
			}
			extParams := map[string]any{}
			if p.Index != nil {
				extParams["index"] = *p.Index
			}
			rawParams, _ := json.Marshal(extParams)
			return s.callExtensionRaw(ctx, method, rawParams)
		},
	})

	// 上傳檔案
	s.RegisterTool(&Tool{
		Name:        "bp_upload",
		Description: "上傳檔案至 file input 元素",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"index": map[string]any{"type": "integer", "description": "file input 元素索引"},
				"path":  map[string]any{"type": "string", "description": "檔案路徑"},
			},
			"required": []string{"index", "path"},
		},
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			return s.callExtensionRaw(ctx, "upload_file", params)
		},
	})

	// 批次註冊無需特殊參數處理的互動 tool
	for _, t := range []struct {
		name, desc, method string
	}{
		{"bp_back", "上一頁", "go_back"},
		{"bp_forward", "下一頁", "go_forward"},
		{"bp_reload", "重新載入當前頁面", "reload"},
		{"bp_hover", "滑鼠移入元素", "hover"},
		{"bp_dblclick", "雙擊元素", "dblclick"},
		{"bp_rightclick", "右鍵點擊元素", "rightclick"},
		{"bp_select", "選擇下拉選單選項", "select_option"},
	} {
		schema := map[string]any{"type": "object", "properties": map[string]any{}}

		// 需要 index 參數的操作 tool
		if t.method == "hover" || t.method == "dblclick" || t.method == "rightclick" {
			schema["properties"] = map[string]any{
				"index": map[string]any{"type": "integer", "description": "元素索引"},
			}
			schema["required"] = []string{"index"}
		}

		// 下拉選單選項需要 index 和 value
		if t.method == "select_option" {
			schema["properties"] = map[string]any{
				"index": map[string]any{"type": "integer", "description": "元素索引"},
				"value": map[string]any{"type": "string", "description": "選項值"},
			}
			schema["required"] = []string{"index", "value"}
		}

		s.RegisterTool(&Tool{
			Name:        t.name,
			Description: t.desc,
			InputSchema: schema,
			Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
				return s.callExtensionRaw(ctx, t.method, params)
			},
		})
	}
}

// callExtensionRaw 直接將 json.RawMessage 參數傳送至 Extension，
// 避免 transport.NewRequest 造成的雙重序列化問題。
// 若 params 為 nil，則不帶參數傳送。
func (s *Server) callExtensionRaw(ctx context.Context, method string, params json.RawMessage) (any, error) {
	if s.transport == nil || !s.transport.IsConnected() {
		return nil, fmt.Errorf("Extension 未連線")
	}

	id := generateID()

	// 直接建構 Request，避免 NewRequest 對 json.RawMessage 雙重序列化
	req := &transport.Request{
		ID:     id,
		Method: method,
		Params: params,
	}

	resp, err := s.transport.Send(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("%s", resp.Error.Message)
	}

	var result any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("解析回應失敗: %w", err)
	}
	return result, nil
}

// generateID 產生一個唯一的請求 ID。
func generateID() string {
	// 使用 uuid 套件產生唯一識別碼
	return fmt.Sprintf("%d", generateSeq())
}

// seq 是請求 ID 計數器（簡易實作，正式場景可改用 uuid）
var seqChan = func() chan int {
	ch := make(chan int)
	go func() {
		i := 1
		for {
			ch <- i
			i++
		}
	}()
	return ch
}()

// generateSeq 回傳遞增的序列號作為請求 ID。
func generateSeq() int {
	return <-seqChan
}
