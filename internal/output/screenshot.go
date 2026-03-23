package output

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ScreenshotResult 代表截圖回傳結果
type ScreenshotResult struct {
	Data   string `json:"data"` // base64 編碼的 PNG
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// SaveScreenshot 將 base64 截圖儲存至檔案
func SaveScreenshot(base64Data, path string) error {
	// 解碼 base64
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("base64 解碼失敗: %w", err)
	}

	// 確保目錄存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("建立目錄失敗: %w", err)
	}

	// 寫入檔案
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("儲存截圖失敗: %w", err)
	}

	return nil
}

// PrintScreenshot 根據模式輸出截圖
// 有 path: 儲存至檔案
// 無 path + json: 輸出 JSON（含 base64）
// 無 path + 非 json: 輸出 base64 字串
func PrintScreenshot(w io.Writer, result *ScreenshotResult, path string, jsonMode bool) error {
	if path != "" {
		if err := SaveScreenshot(result.Data, path); err != nil {
			return err
		}
		if jsonMode {
			return printJSON(w, map[string]any{
				"success": true,
				"path":    path,
			})
		}
		fmt.Fprintf(w, "✓ 截圖已儲存至 %s\n", path)
		return nil
	}

	if jsonMode {
		return printJSON(w, result)
	}

	// 非 JSON 模式且無路徑，輸出 base64
	fmt.Fprintln(w, result.Data)
	return nil
}
