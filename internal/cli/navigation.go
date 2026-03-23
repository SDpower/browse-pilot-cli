// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/i18n"
)

// openCmd 導航至指定 URL
var openCmd = &cobra.Command{
	Use:  "open <url>",
	Args: cobra.ExactArgs(1),
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
			f.PrintError(i18n.T("nav.open.error"), resp.Error.Message)
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
		f.PrintSuccess(i18n.T("nav.open.success"), result.URL)
		if result.Title != "" {
			f.PrintKeyValue("Title", result.Title)
		}
		return nil
	},
}

// backCmd 返回上一頁
var backCmd = &cobra.Command{
	Use: "back",
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
			return f.PrintJSON(resp.Result)
		}
		f.PrintSuccess("%s", i18n.T("nav.back.success"))
		return nil
	},
}

// forwardCmd 前進至下一頁
var forwardCmd = &cobra.Command{
	Use: "forward",
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
			return f.PrintJSON(resp.Result)
		}
		f.PrintSuccess("%s", i18n.T("nav.forward.success"))
		return nil
	},
}

// reloadCmd 重新載入當前頁面
var reloadCmd = &cobra.Command{
	Use: "reload",
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
			return f.PrintJSON(resp.Result)
		}
		f.PrintSuccess("%s", i18n.T("nav.reload.success"))
		return nil
	},
}

// scrollCmd 捲動頁面
var scrollCmd = &cobra.Command{
	Use:  "scroll <up|down>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		direction := args[0]
		if direction != "up" && direction != "down" {
			return fmt.Errorf(i18n.T("nav.scroll.invalid_direction"), direction)
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
			return f.PrintJSON(resp.Result)
		}

		// 依方向輸出對應訊息
		if direction == "up" {
			f.PrintSuccess("%s", i18n.T("nav.scroll.up.success"))
		} else {
			f.PrintSuccess("%s", i18n.T("nav.scroll.down.success"))
		}
		return nil
	},
}

func init() {
	// 設定各指令的 Short 描述
	openCmd.Short = i18n.T("nav.open.short")
	backCmd.Short = i18n.T("nav.back.short")
	forwardCmd.Short = i18n.T("nav.forward.short")
	reloadCmd.Short = i18n.T("nav.reload.short")
	scrollCmd.Short = i18n.T("nav.scroll.short")

	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(backCmd)
	rootCmd.AddCommand(forwardCmd)
	rootCmd.AddCommand(reloadCmd)
	rootCmd.AddCommand(scrollCmd)

	// scrollCmd 的 --amount flag
	scrollCmd.Flags().Int("amount", 0, i18n.T("nav.scroll.amount"))
}
