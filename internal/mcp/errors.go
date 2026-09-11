package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

// toolErrorEnvelope 是 MCP tool 執行失敗時固定回傳的安全 JSON 格式。
type toolErrorEnvelope struct {
	OK    bool            `json:"ok"`
	Error structuredError `json:"error"`
}

type structuredError struct {
	Code      int    `json:"code"`
	Name      string `json:"name"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
	Action    string `json:"action"`
	Data      any    `json:"data"`
}

func marshalToolError(err error, browser string, port int) string {
	structured := normalizeToolError(err, browser, port)
	data, marshalErr := json.Marshal(toolErrorEnvelope{OK: false, Error: structured})
	if marshalErr != nil {
		// 結構僅包含可序列化欄位；保留固定後備值，避免洩漏原始錯誤。
		return `{"ok":false,"error":{"code":-32000,"name":"ExtensionError","message":"工具執行失敗","retryable":false,"action":"顯示此錯誤並停止操作","data":{}}}`
	}
	return string(data)
}

func normalizeToolError(err error, browser string, port int) structuredError {
	rpcErr := &transport.RPCError{
		Code:    transport.ErrExtensionError,
		Message: "工具執行失敗",
	}

	var source *transport.RPCError
	if errors.As(err, &source) && isKnownToolErrorCode(source.Code) {
		if source.Code == transport.ErrExtensionError {
			// ExtensionError 可能包入 stack、路徑或頁面內容，僅回傳固定安全訊息。
			rpcErr = &transport.RPCError{
				Code:    transport.ErrExtensionError,
				Message: "Extension 執行工具時發生錯誤",
			}
		} else {
			rpcErr = source
		}
	} else if errors.Is(err, contextDeadlineExceeded) {
		rpcErr = &transport.RPCError{
			Code:    transport.ErrTimeoutError,
			Message: "等待 Extension 回應逾時",
		}
	}

	data := decodeErrorData(rpcErr.Data)
	if data == nil {
		data = defaultErrorData(rpcErr.Code, browser, port)
	}

	retryable, action := errorGuidance(rpcErr.Code, browser, port)
	return structuredError{
		Code:      rpcErr.Code,
		Name:      transport.ErrorName(rpcErr.Code),
		Message:   rpcErr.Message,
		Retryable: retryable,
		Action:    action,
		Data:      data,
	}
}

func isKnownToolErrorCode(code int) bool {
	switch code {
	case transport.ErrExtensionError,
		transport.ErrConnectionError,
		transport.ErrTimeoutError,
		transport.ErrElementNotFound,
		transport.ErrTabNotFound,
		transport.ErrInjectionError,
		transport.ErrPermissionError,
		transport.ErrStaleElement,
		transport.ErrBrowserNotFound,
		transport.ErrNativeMessagingError,
		transport.ErrInvalidParams:
		return true
	default:
		return false
	}
}

func decodeErrorData(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var data any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil
	}
	return data
}

func defaultErrorData(code int, browser string, port int) map[string]any {
	data := map[string]any{}
	if code == transport.ErrConnectionError || code == transport.ErrTimeoutError || code == transport.ErrBrowserNotFound {
		if browser != "" {
			data["browser"] = browser
		}
		if port > 0 && browser == "firefox" {
			data["port"] = port
		}
	}
	return data
}

func errorGuidance(code int, browser string, port int) (retryable bool, action string) {
	switch code {
	case transport.ErrConnectionError:
		browserName := browserDisplayName(browser)
		if browser == "firefox" && port > 0 {
			return true, fmt.Sprintf("確認 %s 暫時擴充套件已載入，且連接埠為 %d，然後重試", browserName, port)
		}
		return true, fmt.Sprintf("確認 %s Extension 已載入並可連線，然後重試", browserName)
	case transport.ErrTimeoutError:
		return true, "確認 Extension 與頁面狀態，然後重試"
	case transport.ErrElementNotFound:
		return true, "重新呼叫 bp_state，並使用新的元素索引"
	case transport.ErrTabNotFound:
		return true, "呼叫 bp_tabs list，重新確認分頁"
	case transport.ErrInjectionError:
		return true, "重新載入一般網頁；若為瀏覽器內建頁面則停止操作"
	case transport.ErrPermissionError:
		return false, "說明缺少的 Extension 權限並停止操作"
	case transport.ErrStaleElement:
		return true, "重新呼叫 bp_state，不得沿用舊索引"
	case transport.ErrBrowserNotFound:
		return true, "啟動 Firefox 並載入暫時擴充套件，然後重試"
	case transport.ErrNativeMessagingError:
		return true, "Chrome 或 Edge 請執行 bp_cli setup，然後重試"
	case transport.ErrInvalidParams:
		return false, "修正工具參數後重新呼叫"
	default:
		return false, "顯示安全錯誤訊息並停止操作"
	}
}

func browserDisplayName(browser string) string {
	switch strings.ToLower(browser) {
	case "firefox":
		return "Firefox"
	case "chrome":
		return "Chrome"
	case "edge":
		return "Edge"
	default:
		return "瀏覽器"
	}
}

// contextDeadlineExceeded 以私有變數保留 errors.Is 判斷，避免將 context 細節寫入輸出。
var contextDeadlineExceeded = context.DeadlineExceeded

func invalidParamsError(message string) error {
	return &transport.RPCError{Code: transport.ErrInvalidParams, Message: message}
}
