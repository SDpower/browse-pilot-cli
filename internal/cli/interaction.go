// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/i18n"
)

// clickCmd 點擊元素（支援索引或座標兩種模式）
var clickCmd = &cobra.Command{
	Use:  "click <index> 或 click <x> <y>",
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var params map[string]any
		if len(args) == 1 {
			// 以元素索引點擊
			index, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf(i18n.T("error.invalid_index"), args[0])
			}
			params = map[string]any{"index": index}
		} else {
			// 以座標點擊
			x, err1 := strconv.Atoi(args[0])
			y, err2 := strconv.Atoi(args[1])
			if err1 != nil || err2 != nil {
				return errors.New(i18n.T("error.invalid_coord"))
			}
			params = map[string]any{"x": x, "y": y}
		}
		return simpleCommand("click", params, i18n.T("interaction.click.success"))
	},
}

// typeCmd 對當前焦點元素輸入文字
var typeCmd = &cobra.Command{
	Use:  "type <text>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return simpleCommand("type_text", map[string]string{"text": args[0]}, i18n.T("interaction.type.success"))
	},
}

// inputCmd 點擊指定元素後輸入文字
var inputCmd = &cobra.Command{
	Use:  "input <index> <text>",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf(i18n.T("error.invalid_index"), args[0])
		}
		return simpleCommand("input_text", map[string]any{"index": index, "text": args[1]}, i18n.T("interaction.input.success"))
	},
}

// keysCmd 送出鍵盤事件
var keysCmd = &cobra.Command{
	Use:  "keys <keys>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return simpleCommand("send_keys", map[string]string{"keys": args[0]}, i18n.T("interaction.keys.success"))
	},
}

// selectCmd 選擇下拉選單選項
var selectCmd = &cobra.Command{
	Use:  "select <index> <value>",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf(i18n.T("error.invalid_index"), args[0])
		}
		return simpleCommand("select_option", map[string]any{"index": index, "value": args[1]}, i18n.T("interaction.select.success"))
	},
}

// hoverCmd 將滑鼠移入指定元素
var hoverCmd = &cobra.Command{
	Use:  "hover <index>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf(i18n.T("error.invalid_index"), args[0])
		}
		return simpleCommand("hover", map[string]any{"index": index}, i18n.T("interaction.hover.success"))
	},
}

// dblclickCmd 雙擊指定元素
var dblclickCmd = &cobra.Command{
	Use:  "dblclick <index>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf(i18n.T("error.invalid_index"), args[0])
		}
		return simpleCommand("dblclick", map[string]any{"index": index}, i18n.T("interaction.dblclick.success"))
	},
}

// rightclickCmd 右鍵點擊指定元素
var rightclickCmd = &cobra.Command{
	Use:  "rightclick <index>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf(i18n.T("error.invalid_index"), args[0])
		}
		return simpleCommand("rightclick", map[string]any{"index": index}, i18n.T("interaction.rightclick.success"))
	},
}

// uploadCmd 上傳檔案至 file input 元素
var uploadCmd = &cobra.Command{
	Use:  "upload <index> <path>",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf(i18n.T("error.invalid_index"), args[0])
		}
		filePath := args[1]

		// 檢查檔案是否存在
		if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
			return fmt.Errorf(i18n.T("error.file_not_found"), filePath)
		}

		// 取得絕對路徑，確保 extension 端可正確存取
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf(i18n.T("error.abs_path"), err)
		}

		return simpleCommand("upload_file", map[string]any{
			"index": index,
			"path":  absPath,
		}, fmt.Sprintf(i18n.T("interaction.upload.success"), filepath.Base(absPath)))
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
		return f.PrintJSON(resp.Result)
	}
	f.PrintSuccess("%s", successMsg)
	return nil
}

func init() {
	// 設定各指令的 Short 描述
	clickCmd.Short = i18n.T("interaction.click.short")
	typeCmd.Short = i18n.T("interaction.type.short")
	inputCmd.Short = i18n.T("interaction.input.short")
	keysCmd.Short = i18n.T("interaction.keys.short")
	selectCmd.Short = i18n.T("interaction.select.short")
	hoverCmd.Short = i18n.T("interaction.hover.short")
	dblclickCmd.Short = i18n.T("interaction.dblclick.short")
	rightclickCmd.Short = i18n.T("interaction.rightclick.short")
	uploadCmd.Short = i18n.T("interaction.upload.short")

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
