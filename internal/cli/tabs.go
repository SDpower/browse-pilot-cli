// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// tabsCmd 列出所有開啟的分頁
var tabsCmd = &cobra.Command{
	Use:   "tabs",
	Short: "列出所有分頁",
	RunE: func(cmd *cobra.Command, args []string) error {
		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "get_tabs", nil)
		if err != nil {
			return err
		}

		f := getFormatter()
		if resp.IsError() {
			f.PrintError("%s", resp.Error.Message)
			return resp.Error
		}

		// 解析分頁清單
		var result struct {
			Tabs []struct {
				Index  int    `json:"index"`
				ID     int    `json:"id"`
				URL    string `json:"url"`
				Title  string `json:"title"`
				Active bool   `json:"active"`
			} `json:"tabs"`
		}
		if err := resp.ParseResult(&result); err != nil {
			return err
		}

		// JSON 模式：直接輸出原始結果
		if flagJSON {
			return f.PrintJSON(result)
		}

		// Human 模式：逐行輸出，active tab 以 * 標記
		for _, tab := range result.Tabs {
			marker := " "
			if tab.Active {
				marker = "*"
			}
			fmt.Printf("[%d] %s %q — %s\n", tab.Index, marker, tab.Title, tab.URL)
		}
		return nil
	},
}

// tabCmd 切換至指定索引的分頁
var tabCmd = &cobra.Command{
	Use:   "tab <index>",
	Short: "切換至指定分頁",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 將字串引數解析為整數索引
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("分頁索引必須為整數，收到: %s", args[0])
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "switch_tab", map[string]any{"index": index})
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
		f.PrintSuccess("已切換至分頁 %d", index)
		return nil
	},
}

// closeTabCmd 關閉指定索引的分頁，若未提供索引則關閉當前分頁
var closeTabCmd = &cobra.Command{
	Use:   "close-tab [index]",
	Short: "關閉分頁（預設當前）",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 組建請求參數，索引為選填
		params := map[string]any{}
		if len(args) == 1 {
			index, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("分頁索引必須為整數，收到: %s", args[0])
			}
			params["index"] = index
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "close_tab", params)
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
		if len(args) == 1 {
			f.PrintSuccess("已關閉分頁 %s", args[0])
		} else {
			f.PrintSuccess("已關閉當前分頁")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tabsCmd)
	rootCmd.AddCommand(tabCmd)
	rootCmd.AddCommand(closeTabCmd)
}
