// Package cli 中的 setup 指令：安裝 Native Messaging host manifest，純本機操作
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/i18n"
	"github.com/SDpower/browse-pilot-cli/internal/output"
	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

// setupCmd 為指定瀏覽器安裝 Native Messaging host manifest
var setupCmd = &cobra.Command{
	Use:  "setup <browser>",
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
				return fmt.Errorf(i18n.T("error.unsupported_setup_browser"), b)
			}
			browsers = []string{b}
		default:
			return errors.New(i18n.T("error.specify_browser_or_all"))
		}

		// 取得 bp_cli binary 的絕對路徑，寫入 manifest 中供瀏覽器呼叫
		bpPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf(i18n.T("error.get_bp_path"), err)
		}
		bpPath, err = filepath.Abs(bpPath)
		if err != nil {
			return fmt.Errorf(i18n.T("error.get_abs_path"), err)
		}

		// 逐一為各瀏覽器安裝 manifest
		for _, browser := range browsers {
			if err := installNMHost(browser, bpPath, f); err != nil {
				f.PrintError(i18n.T("error.nm_install"), browser, err)
				continue
			}
			f.PrintSuccess(i18n.T("setup.success"), browser)
		}

		return nil
	},
}

// installNMHost 為指定瀏覽器寫入 Native Messaging host manifest 檔案。
// bpPath 是 bp_cli binary 的絕對路徑，manifest 中需引用此路徑。
func installNMHost(browser, bpPath string, f *output.Formatter) error {
	nmPath := transport.NMHostPath(browser)
	if nmPath == "" {
		return fmt.Errorf(i18n.T("error.unsupported_browser_platform"), browser, runtime.GOOS)
	}

	// 建立目錄（若尚未存在）
	dir := filepath.Dir(nmPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf(i18n.T("error.create_dir"), err)
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
		return fmt.Errorf(i18n.T("error.serialize_manifest"), err)
	}

	if err := os.WriteFile(nmPath, data, 0o644); err != nil { //nolint:gosec // manifest 檔案需要讓瀏覽器讀取，不需嚴格限制權限
		return fmt.Errorf(i18n.T("error.write_manifest"), err)
	}

	f.PrintVerbose(i18n.T("setup.verbose_wrote"), nmPath)
	return nil
}

func init() {
	// 設定 Short 與 Long 描述
	setupCmd.Short = i18n.T("setup.short")
	setupCmd.Long = i18n.T("setup.long")

	// 為 setup 指令新增 --all flag
	setupCmd.Flags().Bool("all", false, i18n.T("setup.all_flag"))
	rootCmd.AddCommand(setupCmd)
}
