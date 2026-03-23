// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// evalCmd 在當前頁面執行任意 JavaScript 程式碼並回傳結果
var evalCmd = &cobra.Command{
	Use:   "eval <code>",
	Short: "在頁面執行 JavaScript",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		params := map[string]any{
			"code": args[0],
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "eval_js", params)
		if err != nil {
			return err
		}

		f := getFormatter()
		if resp.IsError() {
			f.PrintError("%s", resp.Error.Message)
			return resp.Error
		}

		// JSON 模式：輸出包含 result 欄位的結構
		if flagJSON {
			var result struct {
				Result json.RawMessage `json:"result"`
			}
			if err := resp.ParseResult(&result); err != nil {
				// 若解析失敗，直接輸出原始 JSON
				return f.PrintJSON(resp.Result)
			}
			return f.PrintJSON(result)
		}

		// Human 模式：直接輸出執行結果字串
		var result struct {
			Result json.RawMessage `json:"result"`
		}
		if err := resp.ParseResult(&result); err != nil {
			// 若無法解析，輸出原始內容
			fmt.Println(string(resp.Result))
			return nil
		}
		// 若 result 是字串，去掉引號後輸出；否則輸出 JSON 表示
		var strVal string
		if err := json.Unmarshal(result.Result, &strVal); err == nil {
			fmt.Println(strVal)
		} else {
			fmt.Println(string(result.Result))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(evalCmd)
}
