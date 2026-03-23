// Package cli 中的 session 指令：列出活躍連線 session 與關閉連線
package cli

import (
	"github.com/spf13/cobra"
)

// sessionsCmd 列出目前活躍的 Extension 連線 session
var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "列出活躍的連線 session",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		tr, err := getTransport()
		if err != nil {
			f.PrintError("無法建立連線: %v", err)
			if flagJSON {
				return f.PrintJSON(map[string]any{"connected": false, "error": err.Error()})
			}
			return nil
		}
		defer tr.Close() //nolint:errcheck

		resp, err := sendCommand(tr, "get_status", nil)
		if err != nil {
			f.PrintError("查詢 session 失敗: %v", err)
			return nil
		}

		if resp.IsError() {
			f.PrintError("%s", resp.Error.Message)
			return nil
		}

		// 解析回傳的 session 資訊
		var status struct {
			Connected   bool   `json:"connected"`
			Browser     string `json:"browser"`
			SessionID   string `json:"sessionId"`
			ExtensionID string `json:"extensionId"`
		}
		if err := resp.ParseResult(&status); err != nil {
			f.PrintError("解析回應失敗: %v", err)
			return nil
		}

		if flagJSON {
			return f.PrintJSON(map[string]any{
				"sessions": []any{
					map[string]any{
						"id":          status.SessionID,
						"browser":     status.Browser,
						"extensionId": status.ExtensionID,
						"connected":   status.Connected,
						"transport":   tr.Type(),
					},
				},
			})
		}

		f.PrintKeyValue("Transport", tr.Type())
		f.PrintKeyValue("瀏覽器", status.Browser)
		f.PrintKeyValue("Session ID", status.SessionID)
		f.PrintKeyValue("Extension ID", status.ExtensionID)
		f.PrintKeyValue("已連線", status.Connected)
		return nil
	},
}

// closeCmd 關閉當前或所有 session 連線
var closeCmd = &cobra.Command{
	Use:   "close",
	Short: "關閉當前 session 連線",
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		f := getFormatter()

		tr, err := getTransport()
		if err != nil {
			f.PrintError("無法建立連線: %v", err)
			return nil
		}

		if err := tr.Close(); err != nil {
			f.PrintError("關閉連線時發生錯誤: %v", err)
			return nil
		}

		if all {
			f.PrintSuccess("已關閉所有 session 連線")
		} else {
			f.PrintSuccess("已關閉連線")
		}
		return nil
	},
}

func init() {
	// 為 close 指令新增 --all flag
	closeCmd.Flags().Bool("all", false, "關閉所有 session")
	rootCmd.AddCommand(sessionsCmd)
	rootCmd.AddCommand(closeCmd)
}
