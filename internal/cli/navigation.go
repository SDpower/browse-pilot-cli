// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// openCmd 導航至指定 URL
var openCmd = &cobra.Command{
	Use:   "open <url>",
	Short: "導航至指定 URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "navigate", map[string]string{"url": args[0]})
		if err != nil {
			return err
		}

		f := getFormatter()
		if resp.IsError() {
			f.PrintError("導航失敗: %s", resp.Error.Message)
			return resp.Error
		}

		// 解析回傳結果
		var result struct {
			Success bool   `json:"success"`
			URL     string `json:"url"`
			Title   string `json:"title"`
		}
		if err := resp.ParseResult(&result); err != nil {
			return err
		}

		if flagJSON {
			return f.PrintJSON(result)
		}
		f.PrintSuccess("已開啟 %s", result.URL)
		if result.Title != "" {
			f.PrintKeyValue("Title", result.Title)
		}
		return nil
	},
}

// backCmd 返回上一頁
var backCmd = &cobra.Command{
	Use:   "back",
	Short: "上一頁",
	RunE: func(cmd *cobra.Command, args []string) error {
		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "go_back", nil)
		if err != nil {
			return err
		}

		f := getFormatter()
		if resp.IsError() {
			f.PrintError("%s", resp.Error.Message)
			return resp.Error
		}
		if flagJSON {
			return f.PrintJSON(json.RawMessage(resp.Result))
		}
		f.PrintSuccess("已返回上一頁")
		return nil
	},
}

// forwardCmd 前進至下一頁
var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "下一頁",
	RunE: func(cmd *cobra.Command, args []string) error {
		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "go_forward", nil)
		if err != nil {
			return err
		}

		f := getFormatter()
		if resp.IsError() {
			f.PrintError("%s", resp.Error.Message)
			return resp.Error
		}
		if flagJSON {
			return f.PrintJSON(json.RawMessage(resp.Result))
		}
		f.PrintSuccess("已前進至下一頁")
		return nil
	},
}

// reloadCmd 重新載入當前頁面
var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "重新載入當前頁面",
	RunE: func(cmd *cobra.Command, args []string) error {
		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "reload", nil)
		if err != nil {
			return err
		}

		f := getFormatter()
		if resp.IsError() {
			f.PrintError("%s", resp.Error.Message)
			return resp.Error
		}
		if flagJSON {
			return f.PrintJSON(json.RawMessage(resp.Result))
		}
		f.PrintSuccess("已重新載入")
		return nil
	},
}

// scrollCmd 捲動頁面
var scrollCmd = &cobra.Command{
	Use:   "scroll <up|down>",
	Short: "捲動頁面",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		direction := args[0]
		if direction != "up" && direction != "down" {
			return fmt.Errorf("方向只能是 up 或 down，收到: %s", direction)
		}

		amount, _ := cmd.Flags().GetInt("amount")
		params := map[string]any{"direction": direction}
		if amount > 0 {
			params["amount"] = amount
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "scroll", params)
		if err != nil {
			return err
		}

		f := getFormatter()
		if resp.IsError() {
			f.PrintError("%s", resp.Error.Message)
			return resp.Error
		}
		if flagJSON {
			return f.PrintJSON(json.RawMessage(resp.Result))
		}

		// 依方向輸出對應的中文訊息
		dirMap := map[string]string{"up": "向上", "down": "向下"}
		f.PrintSuccess("已%s捲動", dirMap[direction])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(backCmd)
	rootCmd.AddCommand(forwardCmd)
	rootCmd.AddCommand(reloadCmd)
	rootCmd.AddCommand(scrollCmd)

	// scrollCmd 的 --amount flag
	scrollCmd.Flags().Int("amount", 0, "捲動像素數")
}
