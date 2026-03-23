// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/output"
)

// stateCmd 列出頁面 URL、標題與可互動元素（帶索引）
var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "列出頁面 URL、標題、可互動元素（帶索引）",
	RunE: func(cmd *cobra.Command, args []string) error {
		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "get_state", nil)
		if err != nil {
			return err
		}

		if resp.IsError() {
			getFormatter().PrintError("%s", resp.Error.Message)
			return resp.Error
		}

		// 解析 get_state 回傳的頁面狀態
		var state output.StateResult
		if err := resp.ParseResult(&state); err != nil {
			return err
		}

		return output.PrintState(os.Stdout, &state, flagJSON)
	},
}

// screenshotCmd 截取目前頁面畫面
var screenshotCmd = &cobra.Command{
	Use:   "screenshot [path]",
	Short: "截圖",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		full, _ := cmd.Flags().GetBool("full")

		// 組建截圖參數
		params := map[string]any{}
		if full {
			params["full"] = true
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "screenshot", params)
		if err != nil {
			return err
		}

		if resp.IsError() {
			getFormatter().PrintError("%s", resp.Error.Message)
			return resp.Error
		}

		// 解析截圖回傳結果
		var result output.ScreenshotResult
		if err := resp.ParseResult(&result); err != nil {
			return err
		}

		// 若有提供路徑參數則存檔，否則輸出至 stdout
		path := ""
		if len(args) > 0 {
			path = args[0]
		}
		return output.PrintScreenshot(os.Stdout, &result, path, flagJSON)
	},
}

func init() {
	// 全頁截圖 flag
	screenshotCmd.Flags().Bool("full", false, "全頁截圖")

	rootCmd.AddCommand(stateCmd)
	rootCmd.AddCommand(screenshotCmd)
}
