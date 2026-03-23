// Package cli 中的 session 指令：列出活躍連線 session 與關閉連線
package cli

import (
	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/i18n"
)

// sessionsCmd 列出目前活躍的 Extension 連線 session
var sessionsCmd = &cobra.Command{
	Use: "sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		tr, err := getTransport()
		if err != nil {
			f.PrintError(i18n.T("error.build_connection"), err)
			if flagJSON {
				return f.PrintJSON(map[string]any{"connected": false, "error": err.Error()})
			}
			return nil
		}
		defer tr.Close() //nolint:errcheck // Close 錯誤在 defer 中無法處理

		resp, err := sendCommand(tr, "get_status", nil)
		if err != nil {
			f.PrintError(i18n.T("error.session_query"), err)
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
			f.PrintError(i18n.T("error.parse_response"), err)
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
	Use: "close",
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		f := getFormatter()

		tr, err := getTransport()
		if err != nil {
			f.PrintError(i18n.T("error.build_connection"), err)
			return nil
		}

		if err := tr.Close(); err != nil {
			f.PrintError(i18n.T("error.close_connection"), err)
			return nil
		}

		if all {
			f.PrintSuccess("%s", i18n.T("session.close.all_success"))
		} else {
			f.PrintSuccess("%s", i18n.T("session.close.success"))
		}
		return nil
	},
}

func init() {
	// 設定各指令的 Short 描述
	sessionsCmd.Short = i18n.T("session.sessions.short")
	closeCmd.Short = i18n.T("session.close.short")

	// 為 close 指令新增 --all flag
	closeCmd.Flags().Bool("all", false, i18n.T("session.close.all_flag"))
	rootCmd.AddCommand(sessionsCmd)
	rootCmd.AddCommand(closeCmd)
}
