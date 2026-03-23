package transport

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Request 是 CLI 發送至 Extension 的 JSON-RPC 2.0 請求訊息。
type Request struct {
	// ID 是唯一識別此請求的 UUID 字串，用於對應非同步回應
	ID string `json:"id"`

	// Method 是要呼叫的 JSON-RPC 方法名稱（如 "navigate"、"click"）
	Method string `json:"method"`

	// Params 是方法的參數，以原始 JSON 格式儲存以支援任意結構
	Params json.RawMessage `json:"params,omitempty"`
}

// Response 是 Extension 回傳給 CLI 的 JSON-RPC 2.0 回應訊息。
type Response struct {
	// ID 對應原始請求的 ID，用於配對請求與回應
	ID string `json:"id"`

	// Result 是成功執行後的回傳值，與 Error 互斥
	Result json.RawMessage `json:"result,omitempty"`

	// Error 是執行失敗時的錯誤物件，與 Result 互斥
	Error *RPCError `json:"error,omitempty"`
}

// RPCError 是 JSON-RPC 2.0 規範的錯誤物件。
type RPCError struct {
	// Code 是機器可讀的錯誤碼，負數為規範保留值
	Code int `json:"code"`

	// Message 是人類可讀的錯誤描述
	Message string `json:"message"`

	// Data 是附加的錯誤資訊，格式依錯誤類型而定
	Data json.RawMessage `json:"data,omitempty"`
}

// Error 實作 error 介面，使 RPCError 可直接作為 Go error 使用。
func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC 錯誤 %d: %s", e.Code, e.Message)
}

// 錯誤碼常數。
// -32700 至 -32600 為 JSON-RPC 2.0 規範保留；
// -32000 至 -32099 為本專案自定義錯誤碼。
const (
	// ErrParseError 表示收到的 JSON 無法解析
	ErrParseError = -32700

	// ErrInvalidRequest 表示請求格式不符合 JSON-RPC 2.0 規範
	ErrInvalidRequest = -32600

	// ErrMethodNotFound 表示指定的方法不存在
	ErrMethodNotFound = -32601

	// ErrInvalidParams 表示方法參數格式錯誤
	ErrInvalidParams = -32602

	// ErrExtensionError 表示 Extension 內部發生一般錯誤
	ErrExtensionError = -32000

	// ErrConnectionError 表示與 Extension 的連線中斷或無法建立
	ErrConnectionError = -32001

	// ErrTimeoutError 表示操作在逾時時間內未完成
	ErrTimeoutError = -32002

	// ErrElementNotFound 表示在頁面中找不到指定的 DOM 元素
	ErrElementNotFound = -32003

	// ErrTabNotFound 表示找不到指定的瀏覽器分頁
	ErrTabNotFound = -32004

	// ErrInjectionError 表示 content script 注入失敗
	ErrInjectionError = -32005

	// ErrPermissionError 表示缺乏執行操作所需的權限
	ErrPermissionError = -32006

	// ErrStaleElement 表示目標 DOM 元素已過期（頁面已更新）
	ErrStaleElement = -32007

	// ErrBrowserNotFound 表示找不到可連線的瀏覽器實例
	ErrBrowserNotFound = -32008

	// ErrNativeMessagingError 表示 Native Messaging 通道發生錯誤
	ErrNativeMessagingError = -32009
)

// ErrorName 將 JSON-RPC 錯誤碼轉換為人類可讀的名稱字串。
// 若錯誤碼未知，回傳 "UnknownError"。
func ErrorName(code int) string {
	switch code {
	case ErrParseError:
		return "ParseError"
	case ErrInvalidRequest:
		return "InvalidRequest"
	case ErrMethodNotFound:
		return "MethodNotFound"
	case ErrInvalidParams:
		return "InvalidParams"
	case ErrExtensionError:
		return "ExtensionError"
	case ErrConnectionError:
		return "ConnectionError"
	case ErrTimeoutError:
		return "TimeoutError"
	case ErrElementNotFound:
		return "ElementNotFound"
	case ErrTabNotFound:
		return "TabNotFound"
	case ErrInjectionError:
		return "InjectionError"
	case ErrPermissionError:
		return "PermissionError"
	case ErrStaleElement:
		return "StaleElement"
	case ErrBrowserNotFound:
		return "BrowserNotFound"
	case ErrNativeMessagingError:
		return "NativeMessagingError"
	default:
		return "UnknownError"
	}
}

// NewRequest 建立一個新的 JSON-RPC 請求，自動產生 UUID 作為請求 ID。
// params 可為任意可序列化的結構，傳入 nil 表示無參數。
func NewRequest(method string, params any) (*Request, error) {
	id := uuid.New().String()

	var rawParams json.RawMessage
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("序列化參數失敗: %w", err)
		}
		rawParams = data
	}

	return &Request{
		ID:     id,
		Method: method,
		Params: rawParams,
	}, nil
}

// IsError 回傳此 Response 是否為錯誤回應。
func (r *Response) IsError() bool {
	return r.Error != nil
}

// ParseResult 將回應的 Result 欄位反序列化至目標結構 v。
// 若回應本身是錯誤，直接回傳 RPCError。
func (r *Response) ParseResult(v any) error {
	if r.IsError() {
		return r.Error
	}
	return json.Unmarshal(r.Result, v)
}

// ExitCode 根據 RPCError 的錯誤碼，回傳對應的 CLI process exit code。
// nil 輸入回傳 0（成功）。
//
// Exit code 對照：
//   - 0: 成功
//   - 1: 一般錯誤
//   - 2: 連線錯誤
//   - 3: 逾時錯誤
//   - 4: 元素錯誤
//   - 5: 權限錯誤
//   - 6: 瀏覽器未找到
func ExitCode(rpcErr *RPCError) int {
	if rpcErr == nil {
		return 0
	}
	switch rpcErr.Code {
	case ErrConnectionError, ErrNativeMessagingError:
		return 2
	case ErrTimeoutError:
		return 3
	case ErrElementNotFound, ErrStaleElement:
		return 4
	case ErrPermissionError:
		return 5
	case ErrBrowserNotFound:
		return 6
	default:
		return 1
	}
}
