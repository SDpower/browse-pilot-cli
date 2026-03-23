// Package cli 中的 status 指令：嘗試連線 Extension 並顯示當前連線資訊
package cli

import (
	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/i18n"
	"github.com/SDpower/browse-pilot-cli/internal/output"
)

// statusCmd 顯示當前 Extension 連線狀態與版本資訊
var statusCmd = &cobra.Command{
	Use: "status",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := output.NewFormatter(flagJSON, flagVerbose)

		// 嘗試建立 transport 連線
		tr, err := getTransport()
		if err != nil {
			f.PrintError(i18n.T("error.no_connection"), err)
			if flagJSON {
				return f.PrintJSON(map[string]any{
					"connected": false,
					"error":     err.Error(),
				})
			}
			return nil // 狀態查詢不回傳 error，僅顯示訊息
		}
		defer tr.Close() //nolint:errcheck // Close 錯誤在 defer 中無法處理

		// 向 Extension 查詢狀態
		resp, err := sendCommand(tr, "get_status", nil)
		if err != nil {
			f.PrintError(i18n.T("error.status_query"), err)
			return nil
		}

		if resp.IsError() {
			f.PrintError("%s", resp.Error.Message)
			return nil
		}

		// 解析 Extension 回傳的狀態資訊
		var status struct {
			Connected   bool   `json:"connected"`
			Browser     string `json:"browser"`
			Version     string `json:"version"`
			ExtensionID string `json:"extensionId"`
		}
		if err := resp.ParseResult(&status); err != nil {
			f.PrintError(i18n.T("error.parse_response"), err)
			return nil
		}

		if flagJSON {
			return f.PrintJSON(status)
		}

		f.PrintSuccess("%s", i18n.T("status.connected"))
		f.PrintKeyValue("瀏覽器", status.Browser)
		f.PrintKeyValue("版本", status.Version)
		f.PrintKeyValue("Transport", tr.Type())
		f.PrintKeyValue("Extension ID", status.ExtensionID)
		return nil
	},
}

func init() {
	// 設定 Short 描述
	statusCmd.Short = i18n.T("status.short")
	rootCmd.AddCommand(statusCmd)
}
