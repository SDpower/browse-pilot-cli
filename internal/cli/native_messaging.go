// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

// runNativeMessagingHost 以 Native Messaging host 模式啟動。
// Chrome/Edge Extension 透過 runtime.connectNative() 呼叫此程序。
// 通訊協議：stdin/stdout，訊息格式為 length-prefixed JSON。
//
// 資料流向：
//
//	Extension → stdin (JSON-RPC request) → CLI 處理 → stdout (JSON-RPC response) → Extension
func runNativeMessagingHost() error {
	if flagVerbose {
		fmt.Fprintln(os.Stderr, "[NM Host] 啟動 Native Messaging host 模式")
	}

	for {
		// 從 stdin 讀取 length-prefixed 訊息
		msg, err := readNMMessage(os.Stdin)
		if err != nil {
			if err == io.EOF {
				// Extension 已關閉連線，正常結束
				if flagVerbose {
					fmt.Fprintln(os.Stderr, "[NM Host] stdin EOF，結束")
				}
				return nil
			}
			return fmt.Errorf("讀取 NM 訊息失敗: %w", err)
		}

		// 解析 JSON-RPC 請求
		var req transport.Request
		if err := json.Unmarshal(msg, &req); err != nil {
			// JSON 解析失敗，回傳 parse error
			resp := &transport.Response{
				Error: &transport.RPCError{
					Code:    transport.ErrParseError,
					Message: fmt.Sprintf("JSON 解析失敗: %s", err.Error()),
				},
			}
			if writeErr := writeNMResponse(os.Stdout, resp); writeErr != nil {
				return fmt.Errorf("寫入 NM 錯誤回應失敗: %w", writeErr)
			}
			continue
		}

		if flagVerbose {
			fmt.Fprintf(os.Stderr, "[NM Host] 收到: method=%s id=%s\n", req.Method, req.ID)
		}

		// 回傳就緒狀態確認（保持連線活躍）
		// 完整實作中，此處應路由至對應的指令處理器
		resp := &transport.Response{
			ID:     req.ID,
			Result: json.RawMessage(`{"status":"ready","mode":"native_messaging"}`),
		}
		if err := writeNMResponse(os.Stdout, resp); err != nil {
			return fmt.Errorf("寫入 NM 回應失敗: %w", err)
		}
	}
}

// readNMMessage 從 reader 讀取一條 Native Messaging 格式的訊息。
// 訊息格式：[4 bytes uint32 little-endian 長度][JSON bytes]
// 最大訊息大小為 1MB（Native Messaging 協議限制）。
func readNMMessage(r io.Reader) ([]byte, error) {
	// 讀取 4 bytes 的訊息長度（little-endian）
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err
	}

	msgLen := binary.LittleEndian.Uint32(lenBuf)
	if msgLen == 0 {
		return nil, fmt.Errorf("訊息長度為 0")
	}
	if msgLen > 1024*1024 { // 1MB 上限
		return nil, fmt.Errorf("訊息長度超過 1MB: %d bytes", msgLen)
	}

	// 讀取訊息本體
	msg := make([]byte, msgLen)
	if _, err := io.ReadFull(r, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

// writeNMResponse 將 response 以 Native Messaging 格式寫入 writer。
// 訊息格式：[4 bytes uint32 little-endian 長度][JSON bytes]
func writeNMResponse(w io.Writer, resp *transport.Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	// 寫入 4 bytes 的訊息長度（little-endian）
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(data)))
	if _, err := w.Write(lenBuf); err != nil {
		return err
	}

	// 寫入 JSON 本體
	_, err = w.Write(data)
	return err
}
