// Package cli 中的 setup 指令：安裝 Native Messaging host manifest，純本機操作
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/output"
	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

// setupCmd 為指定瀏覽器安裝 Native Messaging host manifest
var setupCmd = &cobra.Command{
	Use:   "setup <browser>",
	Short: "安裝 Native Messaging host manifest",
	Long: `為指定瀏覽器安裝 Native Messaging host manifest。
支援: firefox, chrome, edge, --all`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		f := output.NewFormatter(flagJSON, flagVerbose)

		// 決定要安裝的瀏覽器清單
		var browsers []string
		switch {
		case all:
			browsers = []string{"firefox", "chrome", "edge"}
		case len(args) == 1:
			b := args[0]
			if b != "firefox" && b != "chrome" && b != "edge" {
				return fmt.Errorf("不支援的瀏覽器: %s（支援 firefox/chrome/edge）", b)
			}
			browsers = []string{b}
		default:
			return fmt.Errorf("請指定瀏覽器或使用 --all")
		}

		// 取得 bp binary 的絕對路徑，寫入 manifest 中供瀏覽器呼叫
		bpPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("無法取得 bp 路徑: %w", err)
		}
		bpPath, err = filepath.Abs(bpPath)
		if err != nil {
			return fmt.Errorf("無法取得絕對路徑: %w", err)
		}

		// 逐一為各瀏覽器安裝 manifest
		for _, browser := range browsers {
			if err := installNMHost(browser, bpPath, f); err != nil {
				f.PrintError("安裝 %s NM host 失敗: %v", browser, err)
				continue
			}
			f.PrintSuccess("已安裝 %s Native Messaging host", browser)
		}

		return nil
	},
}

// installNMHost 為指定瀏覽器寫入 Native Messaging host manifest 檔案。
// bpPath 是 bp binary 的絕對路徑，manifest 中需引用此路徑。
func installNMHost(browser, bpPath string, f *output.Formatter) error {
	nmPath := transport.NMHostPath(browser)
	if nmPath == "" {
		return fmt.Errorf("不支援的瀏覽器/平台組合: %s/%s", browser, runtime.GOOS)
	}

	// 建立目錄（若尚未存在）
	dir := filepath.Dir(nmPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("建立目錄失敗: %w", err)
	}

	// 產生 manifest 內容（符合 Native Messaging 規範）
	manifest := map[string]any{
		"name":        "com.browse_pilot.host",
		"description": "Browse Pilot CLI native messaging host",
		"path":        bpPath,
		"type":        "stdio",
	}

	// Firefox 使用 allowed_extensions；Chrome/Edge 使用 allowed_origins
	switch browser {
	case "firefox":
		manifest["allowed_extensions"] = []string{"browse-pilot@localhost"}
	case "chrome", "edge":
		// 安裝時 extension ID 未知，使用 placeholder 待日後更新
		manifest["allowed_origins"] = []string{"chrome-extension://PLACEHOLDER_EXTENSION_ID/"}
	}

	// 序列化並寫入 JSON 檔
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 manifest 失敗: %w", err)
	}

	if err := os.WriteFile(nmPath, data, 0644); err != nil { //nolint:gosec
		return fmt.Errorf("寫入 manifest 失敗: %w", err)
	}

	f.PrintVerbose("已寫入: %s", nmPath)
	return nil
}

func init() {
	// 為 setup 指令新增 --all flag
	setupCmd.Flags().Bool("all", false, "為所有瀏覽器安裝")
	rootCmd.AddCommand(setupCmd)
}
