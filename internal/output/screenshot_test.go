package output

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveScreenshot(t *testing.T) {
	// 建立一個小的 base64 字串（不是真正的 PNG，但可測試 I/O）
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.png")

	// "hello" 的 base64
	err := SaveScreenshot("aGVsbG8=", path)
	if err != nil {
		t.Fatalf("SaveScreenshot 失敗: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("讀取檔案失敗: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("檔案內容不正確: %s", string(data))
	}
}

func TestSaveScreenshotInvalidBase64(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.png")
	err := SaveScreenshot("not-valid-base64!!!", path)
	if err == nil {
		t.Error("無效 base64 應回傳錯誤")
	}
}
