// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

// clickCmd 點擊元素（支援索引或座標兩種模式）
var clickCmd = &cobra.Command{
	Use:   "click <index> 或 click <x> <y>",
	Short: "點擊元素",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var params map[string]any
		if len(args) == 1 {
			// 以元素索引點擊
			index, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("無效的元素索引: %s", args[0])
			}
			params = map[string]any{"index": index}
		} else {
			// 以座標點擊
			x, err1 := strconv.Atoi(args[0])
			y, err2 := strconv.Atoi(args[1])
			if err1 != nil || err2 != nil {
				return fmt.Errorf("無效的座標")
			}
			params = map[string]any{"x": x, "y": y}
		}
		return simpleCommand("click", params, "已點擊")
	},
}

// typeCmd 對當前焦點元素輸入文字
var typeCmd = &cobra.Command{
	Use:   "type <text>",
	Short: "對當前焦點元素輸入文字",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return simpleCommand("type_text", map[string]string{"text": args[0]}, "已輸入文字")
	},
}

// inputCmd 點擊指定元素後輸入文字
var inputCmd = &cobra.Command{
	Use:   "input <index> <text>",
	Short: "點擊元素後輸入文字",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("無效的元素索引: %s", args[0])
		}
		return simpleCommand("input_text", map[string]any{"index": index, "text": args[1]}, "已輸入文字")
	},
}

// keysCmd 送出鍵盤事件
var keysCmd = &cobra.Command{
	Use:   "keys <keys>",
	Short: "送出鍵盤事件",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return simpleCommand("send_keys", map[string]string{"keys": args[0]}, "已送出按鍵")
	},
}

// selectCmd 選擇下拉選單選項
var selectCmd = &cobra.Command{
	Use:   "select <index> <value>",
	Short: "選擇下拉選單選項",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("無效的元素索引: %s", args[0])
		}
		return simpleCommand("select_option", map[string]any{"index": index, "value": args[1]}, "已選擇選項")
	},
}

// hoverCmd 將滑鼠移入指定元素
var hoverCmd = &cobra.Command{
	Use:   "hover <index>",
	Short: "滑鼠移入元素",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("無效的元素索引: %s", args[0])
		}
		return simpleCommand("hover", map[string]any{"index": index}, "已 hover")
	},
}

// dblclickCmd 雙擊指定元素
var dblclickCmd = &cobra.Command{
	Use:   "dblclick <index>",
	Short: "雙擊元素",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("無效的元素索引: %s", args[0])
		}
		return simpleCommand("dblclick", map[string]any{"index": index}, "已雙擊")
	},
}

// rightclickCmd 右鍵點擊指定元素
var rightclickCmd = &cobra.Command{
	Use:   "rightclick <index>",
	Short: "右鍵點擊元素",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("無效的元素索引: %s", args[0])
		}
		return simpleCommand("rightclick", map[string]any{"index": index}, "已右鍵點擊")
	},
}

// uploadCmd 上傳檔案至 file input 元素
var uploadCmd = &cobra.Command{
	Use:   "upload <index> <path>",
	Short: "上傳檔案至 file input",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("無效的元素索引: %s", args[0])
		}
		filePath := args[1]

		// 檢查檔案是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("檔案不存在: %s", filePath)
		}

		// 取得絕對路徑，確保 extension 端可正確存取
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("無法取得絕對路徑: %w", err)
		}

		return simpleCommand("upload_file", map[string]any{
			"index": index,
			"path":  absPath,
		}, fmt.Sprintf("已上傳 %s", filepath.Base(absPath)))
	},
}

// simpleCommand 是簡單指令的共用實作。
// 發送指令 → 收到回應 → 輸出成功訊息（或 JSON）。
func simpleCommand(method string, params any, successMsg string) error {
	tr, err := getTransport()
	if err != nil {
		return err
	}
	defer tr.Close()

	resp, err := sendCommand(tr, method, params)
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
	f.PrintSuccess("%s", successMsg)
	return nil
}

func init() {
	rootCmd.AddCommand(clickCmd)
	rootCmd.AddCommand(typeCmd)
	rootCmd.AddCommand(inputCmd)
	rootCmd.AddCommand(keysCmd)
	rootCmd.AddCommand(selectCmd)
	rootCmd.AddCommand(hoverCmd)
	rootCmd.AddCommand(dblclickCmd)
	rootCmd.AddCommand(rightclickCmd)
	rootCmd.AddCommand(uploadCmd)
}
