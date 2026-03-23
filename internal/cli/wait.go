// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/i18n"
)

// waitCmd 是等待頁面條件的父指令
var waitCmd = &cobra.Command{
	Use: "wait",
	// 未提供子指令時顯示說明
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// waitSelectorCmd 等待指定 CSS 選擇器的元素出現或消失
var waitSelectorCmd = &cobra.Command{
	Use:  "selector <css>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hidden, _ := cmd.Flags().GetBool("hidden")
		timeout, _ := cmd.Flags().GetInt("timeout")

		// 根據 --hidden flag 決定等待狀態
		state := "visible"
		if hidden {
			state = "hidden"
		}

		params := map[string]any{
			"selector": args[0],
			"state":    state,
			"timeout":  timeout,
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "wait_selector", params)
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
		if hidden {
			f.PrintSuccess(i18n.T("wait.selector.disappeared"), args[0])
		} else {
			f.PrintSuccess(i18n.T("wait.selector.appeared"), args[0])
		}
		return nil
	},
}

// waitTextCmd 等待頁面中出現指定文字
var waitTextCmd = &cobra.Command{
	Use:  "text <text>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetInt("timeout")

		params := map[string]any{
			"text":    args[0],
			"timeout": timeout,
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "wait_text", params)
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
		f.PrintSuccess(i18n.T("wait.text.success"), args[0])
		return nil
	},
}

// waitUrlCmd 等待頁面 URL 符合指定的 pattern
var waitUrlCmd = &cobra.Command{
	Use:  "url <pattern>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetInt("timeout")

		params := map[string]any{
			"pattern": args[0],
			"timeout": timeout,
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "wait_url", params)
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
		f.PrintSuccess(i18n.T("wait.url.success"), args[0])
		return nil
	},
}

func init() {
	// 設定各指令的 Short 描述
	waitCmd.Short = i18n.T("wait.short")
	waitSelectorCmd.Short = i18n.T("wait.selector.short")
	waitTextCmd.Short = i18n.T("wait.text.short")
	waitUrlCmd.Short = i18n.T("wait.url.short")

	// waitSelectorCmd 的 flags
	waitSelectorCmd.Flags().Bool("hidden", false, i18n.T("wait.selector.hidden_flag"))
	waitSelectorCmd.Flags().Int("timeout", 30000, i18n.T("wait.timeout_flag"))

	// waitTextCmd 的 flags
	waitTextCmd.Flags().Int("timeout", 30000, i18n.T("wait.timeout_flag"))

	// waitUrlCmd 的 flags
	waitUrlCmd.Flags().Int("timeout", 30000, i18n.T("wait.timeout_flag"))

	// 組裝子指令
	waitCmd.AddCommand(waitSelectorCmd)
	waitCmd.AddCommand(waitTextCmd)
	waitCmd.AddCommand(waitUrlCmd)
	rootCmd.AddCommand(waitCmd)
}
