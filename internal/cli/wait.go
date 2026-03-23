// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"github.com/spf13/cobra"
)

// waitCmd 是等待頁面條件的父指令
var waitCmd = &cobra.Command{
	Use:   "wait",
	Short: "等待頁面條件",
	// 未提供子指令時顯示說明
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// waitSelectorCmd 等待指定 CSS 選擇器的元素出現或消失
var waitSelectorCmd = &cobra.Command{
	Use:   "selector <css>",
	Short: "等待元素出現",
	Args:  cobra.ExactArgs(1),
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
			f.PrintSuccess("元素已消失: %s", args[0])
		} else {
			f.PrintSuccess("元素已出現: %s", args[0])
		}
		return nil
	},
}

// waitTextCmd 等待頁面中出現指定文字
var waitTextCmd = &cobra.Command{
	Use:   "text <text>",
	Short: "等待文字出現",
	Args:  cobra.ExactArgs(1),
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
		f.PrintSuccess("文字已出現: %s", args[0])
		return nil
	},
}

// waitUrlCmd 等待頁面 URL 符合指定的 pattern
var waitUrlCmd = &cobra.Command{
	Use:   "url <pattern>",
	Short: "等待 URL 符合 pattern",
	Args:  cobra.ExactArgs(1),
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
		f.PrintSuccess("URL 已符合 pattern: %s", args[0])
		return nil
	},
}

func init() {
	// waitSelectorCmd 的 flags
	waitSelectorCmd.Flags().Bool("hidden", false, "等待元素消失")
	waitSelectorCmd.Flags().Int("timeout", 30000, "逾時時間（ms）")

	// waitTextCmd 的 flags
	waitTextCmd.Flags().Int("timeout", 30000, "逾時時間（ms）")

	// waitUrlCmd 的 flags
	waitUrlCmd.Flags().Int("timeout", 30000, "逾時時間（ms）")

	// 組裝子指令
	waitCmd.AddCommand(waitSelectorCmd)
	waitCmd.AddCommand(waitTextCmd)
	waitCmd.AddCommand(waitUrlCmd)
	rootCmd.AddCommand(waitCmd)
}
