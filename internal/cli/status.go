// Package cli 中的 status 指令：嘗試連線 Extension 並顯示當前連線資訊
package cli

import (
	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/output"
)

// statusCmd 顯示當前 Extension 連線狀態與版本資訊
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "顯示當前連線資訊",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := output.NewFormatter(flagJSON, flagVerbose)

		// 嘗試建立 transport 連線
		tr, err := getTransport()
		if err != nil {
			f.PrintError("未連線: %v", err)
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
			f.PrintError("狀態查詢失敗: %v", err)
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
			f.PrintError("解析回應失敗: %v", err)
			return nil
		}

		if flagJSON {
			return f.PrintJSON(status)
		}

		f.PrintSuccess("已連線")
		f.PrintKeyValue("瀏覽器", status.Browser)
		f.PrintKeyValue("版本", status.Version)
		f.PrintKeyValue("Transport", tr.Type())
		f.PrintKeyValue("Extension ID", status.ExtensionID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
