// Package python 管理持久的 Python 子程序 session，
// 提供 browser 物件代理讓 Python 腳本可操作瀏覽器。
package python

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// BrowserCallback 是 Python 呼叫瀏覽器操作時的回調函式型別。
// requestJSON 為 Python 傳入的 JSON 字串，包含 method 與 params。
type BrowserCallback func(ctx context.Context, requestJSON string) (any, error)

// Session 管理一個持久的 Python 子程序
type Session struct {
	cmd             *exec.Cmd
	stdin           io.WriteCloser
	stdout          *bufio.Reader
	mu              sync.Mutex
	running         bool
	browserCallback BrowserCallback
}

// NewSession 建立並啟動一個 Python session。
// 若 pythonPath 為空，預設使用 "python3"。
func NewSession(pythonPath string) (*Session, error) {
	if pythonPath == "" {
		pythonPath = "python3"
	}

	// 以 -u（unbuffered）模式啟動，使用自訂 REPL 協議：
	// 輸入: __BP_EXEC__<json_encoded_code>
	// 輸出: __BP_RESULT__<json_encoded_result>
	// 或:   __BP_ERROR__<json_encoded_error>
	// 或:   __BP_CALL__<json_encoded_browser_call>（Python 需呼叫瀏覽器時）
	cmd := exec.Command(pythonPath, "-u", "-c", pythonBootstrap)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("無法取得 stdin pipe: %w", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("無法取得 stdout pipe: %w", err)
	}

	// stderr 接到 /dev/null（避免 Python 錯誤訊息污染輸出協議）
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("啟動 Python 失敗: %w", err)
	}

	s := &Session{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  bufio.NewReader(stdoutPipe),
		running: true,
	}

	// 等待 Python 啟動就緒信號
	line, err := s.stdout.ReadString('\n')
	if err != nil {
		s.Close()
		return nil, fmt.Errorf("Python 啟動失敗: %w", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(line), "__BP_READY__") {
		s.Close()
		return nil, fmt.Errorf("Python 啟動異常: %s", line)
	}

	return s, nil
}

// pythonBootstrap 是注入 Python 子程序的啟動腳本。
// 建立 browser 代理物件並進入自訂 REPL 迴圈。
const pythonBootstrap = `
import sys, json, traceback

class BrowserProxy:
    """browser 物件代理，透過 CLI 協議操作瀏覽器"""

    def _call(self, method, params=None):
        """向 CLI 發送瀏覽器指令並等待結果"""
        req = json.dumps({"method": method, "params": params or {}})
        print("__BP_CALL__" + req, flush=True)
        line = input()
        if line.startswith("__BP_CALL_RESULT__"):
            return json.loads(line[len("__BP_CALL_RESULT__"):])
        elif line.startswith("__BP_CALL_ERROR__"):
            raise Exception(json.loads(line[len("__BP_CALL_ERROR__"):]))
        return None

    def navigate(self, url):
        return self._call("navigate", {"url": url})

    def state(self):
        return self._call("get_state")

    def click(self, index):
        return self._call("click", {"index": index})

    def input(self, index, text):
        return self._call("input_text", {"index": index, "text": text})

    def type(self, text):
        return self._call("type_text", {"text": text})

    def keys(self, keys):
        return self._call("send_keys", {"keys": keys})

    def select(self, index, value):
        return self._call("select_option", {"index": index, "value": value})

    def screenshot(self, full=False):
        return self._call("screenshot", {"full": full})

    def eval(self, code):
        return self._call("eval_js", {"code": code})

    def wait(self, type, value, timeout=30000):
        key = "selector" if type == "selector" else "text" if type == "text" else "pattern"
        return self._call("wait_" + type, {key: value, "timeout": timeout})

    def get(self, what, index=None, selector=None):
        params = {}
        if index is not None:
            params["index"] = index
        if selector is not None:
            params["selector"] = selector
        return self._call("get_" + what, params)

    def tabs(self):
        return self._call("get_tabs")

    def cookies(self, url=None):
        params = {}
        if url:
            params["url"] = url
        return self._call("get_cookies", params)

browser = BrowserProxy()
_user_vars = {}

print("__BP_READY__", flush=True)

while True:
    try:
        line = input()
        if not line.startswith("__BP_EXEC__"):
            continue

        code = json.loads(line[len("__BP_EXEC__"):])

        exec_globals = {"browser": browser, "json": json, "_user_vars": _user_vars, **_user_vars}
        exec(code, exec_globals)

        # 更新使用者變數（排除內建和保留名稱）
        for k, v in exec_globals.items():
            if not k.startswith("_") and k not in ("browser", "json", "__builtins__"):
                _user_vars[k] = v

        # 嘗試對程式碼求值以取得最後表達式的結果
        # 字串直接輸出，其他型別使用 repr
        try:
            result = eval(code, {"browser": browser, "json": json, "_user_vars": _user_vars, **_user_vars})
            if isinstance(result, str):
                print("__BP_RESULT__" + json.dumps({"value": result}), flush=True)
            else:
                print("__BP_RESULT__" + json.dumps({"value": repr(result)}), flush=True)
        except SyntaxError:
            print("__BP_RESULT__" + json.dumps({"value": None}), flush=True)

    except EOFError:
        break
    except Exception as e:
        print("__BP_ERROR__" + json.dumps({"error": str(e), "traceback": traceback.format_exc()}), flush=True)
`

// ExecResult 是 Python 執行結果
type ExecResult struct {
	// Value 是 Python 最後一個表達式的 repr 字串，若無則為 nil
	Value any `json:"value"`
}

// SetBrowserCallback 設定瀏覽器操作回調函式。
// 當 Python 程式碼呼叫 browser.xxx() 時，會觸發此回調。
func (s *Session) SetBrowserCallback(cb BrowserCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.browserCallback = cb
}

// Execute 在 Python session 中執行程式碼，回傳執行結果。
func (s *Session) Execute(ctx context.Context, code string) (*ExecResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil, fmt.Errorf("Python session 未啟動")
	}

	// 將程式碼以 JSON 編碼後寫入 Python stdin
	encoded, _ := json.Marshal(code)
	if _, err := fmt.Fprintf(s.stdin, "__BP_EXEC__%s\n", encoded); err != nil {
		return nil, fmt.Errorf("寫入 Python stdin 失敗: %w", err)
	}

	// 持續讀取輸出，直到收到結果或錯誤
	for {
		// 檢查 context 是否已取消
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line, err := s.stdout.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("讀取 Python 回應失敗: %w", err)
		}
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "__BP_RESULT__"):
			// 執行成功，解碼結果
			payload := line[len("__BP_RESULT__"):]
			var result ExecResult
			if err := json.Unmarshal([]byte(payload), &result); err != nil {
				return nil, fmt.Errorf("解碼執行結果失敗: %w", err)
			}
			return &result, nil

		case strings.HasPrefix(line, "__BP_ERROR__"):
			// 執行期間發生例外
			payload := line[len("__BP_ERROR__"):]
			var errResult struct {
				Error     string `json:"error"`
				Traceback string `json:"traceback"`
			}
			if err := json.Unmarshal([]byte(payload), &errResult); err != nil {
				return nil, fmt.Errorf("Python 執行錯誤（解碼失敗）: %s", payload)
			}
			return nil, fmt.Errorf("%s", errResult.Error)

		case strings.HasPrefix(line, "__BP_CALL__"):
			// Python 需要呼叫瀏覽器操作
			requestJSON := line[len("__BP_CALL__"):]
			if s.browserCallback != nil {
				result, callErr := s.browserCallback(ctx, requestJSON)
				if callErr != nil {
					// 通知 Python 呼叫失敗
					errMsg := mustJSON(callErr.Error())
					fmt.Fprintf(s.stdin, "__BP_CALL_ERROR__%s\n", errMsg) //nolint:errcheck
				} else {
					// 將結果回傳給 Python
					resultStr := mustJSON(result)
					fmt.Fprintf(s.stdin, "__BP_CALL_RESULT__%s\n", resultStr) //nolint:errcheck
				}
			} else {
				// 尚未設定 browserCallback
				fmt.Fprintf(s.stdin, "__BP_CALL_ERROR__%s\n", mustJSON("browser 物件未初始化")) //nolint:errcheck
			}
			// 繼續等待下一行輸出
		}
	}
}

// GetVars 列出 Python session 中的使用者自訂變數（名稱 -> repr 字串）。
func (s *Session) GetVars(ctx context.Context) (map[string]string, error) {
	result, err := s.Execute(ctx, `json.dumps({k: repr(v) for k, v in _user_vars.items()})`)
	if err != nil {
		return nil, err
	}
	if result == nil || result.Value == nil {
		return map[string]string{}, nil
	}
	str, ok := result.Value.(string)
	if !ok {
		return map[string]string{}, nil
	}
	var vars map[string]string
	if err := json.Unmarshal([]byte(str), &vars); err != nil {
		return nil, fmt.Errorf("解碼變數列表失敗: %w", err)
	}
	return vars, nil
}

// Close 關閉 Python session，終止子程序並釋放資源。
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false
	if s.stdin != nil {
		s.stdin.Close()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
		s.cmd.Wait() //nolint:errcheck
	}
	return nil
}

// IsRunning 回傳 Python session 是否仍在運行中。
func (s *Session) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// mustJSON 將任意值序列化為 JSON 字串，若失敗則回傳 "null"。
func mustJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(data)
}
