// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// getCmd 是取得頁面或元素資訊的父指令
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "取得頁面或元素資訊",
	// 未提供子指令時顯示說明
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// getTitleCmd 取得當前頁面的標題
var getTitleCmd = &cobra.Command{
	Use:   "title",
	Short: "取得頁面標題",
	RunE: func(cmd *cobra.Command, args []string) error {
		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "get_title", nil)
		if err != nil {
			return err
		}

		f := getFormatter()
		if resp.IsError() {
			f.PrintError("%s", resp.Error.Message)
			return resp.Error
		}

		// JSON 模式：輸出包含 title 欄位的結構
		if flagJSON {
			var result struct {
				Title string `json:"title"`
			}
			if err := resp.ParseResult(&result); err != nil {
				return err
			}
			return f.PrintJSON(result)
		}

		// Human 模式：直接輸出 title 字串
		var result struct {
			Title string `json:"title"`
		}
		if err := resp.ParseResult(&result); err != nil {
			return err
		}
		fmt.Println(result.Title)
		return nil
	},
}

// getHtmlCmd 取得頁面或指定元素的 HTML 內容
var getHtmlCmd = &cobra.Command{
	Use:   "html",
	Short: "取得頁面/元素 HTML",
	RunE: func(cmd *cobra.Command, args []string) error {
		selector, _ := cmd.Flags().GetString("selector")

		// 若有指定 CSS 選擇器則只取該元素的 HTML
		params := map[string]any{}
		if selector != "" {
			params["selector"] = selector
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "get_html", params)
		if err != nil {
			return err
		}

		f := getFormatter()
		if resp.IsError() {
			f.PrintError("%s", resp.Error.Message)
			return resp.Error
		}

		// JSON 模式：輸出原始結果
		if flagJSON {
			return f.PrintJSON(json.RawMessage(resp.Result))
		}

		// Human 模式：直接輸出 HTML 字串
		var result struct {
			HTML string `json:"html"`
		}
		if err := resp.ParseResult(&result); err != nil {
			return err
		}
		fmt.Println(result.HTML)
		return nil
	},
}

// getTextCmd 取得指定索引元素的文字內容
var getTextCmd = &cobra.Command{
	Use:   "text <index>",
	Short: "取得元素文字內容",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("元素索引必須為整數，收到: %s", args[0])
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "get_text", map[string]any{"index": index})
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

		// Human 模式：直接輸出文字內容
		var result struct {
			Text string `json:"text"`
		}
		if err := resp.ParseResult(&result); err != nil {
			return err
		}
		fmt.Println(result.Text)
		return nil
	},
}

// getValueCmd 取得指定索引 input/textarea 元素的當前值
var getValueCmd = &cobra.Command{
	Use:   "value <index>",
	Short: "取得 input/textarea 值",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("元素索引必須為整數，收到: %s", args[0])
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "get_value", map[string]any{"index": index})
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

		// Human 模式：直接輸出欄位值
		var result struct {
			Value string `json:"value"`
		}
		if err := resp.ParseResult(&result); err != nil {
			return err
		}
		fmt.Println(result.Value)
		return nil
	},
}

// getAttributesCmd 取得指定索引元素的所有 HTML 屬性
var getAttributesCmd = &cobra.Command{
	Use:   "attributes <index>",
	Short: "取得元素所有屬性",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("元素索引必須為整數，收到: %s", args[0])
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "get_attributes", map[string]any{"index": index})
		if err != nil {
			return err
		}

		f := getFormatter()
		if resp.IsError() {
			f.PrintError("%s", resp.Error.Message)
			return resp.Error
		}

		// JSON 模式：輸出原始屬性結構
		if flagJSON {
			return f.PrintJSON(json.RawMessage(resp.Result))
		}

		// Human 模式：以 key=value 逐行列出屬性
		var result struct {
			Attributes map[string]string `json:"attributes"`
		}
		if err := resp.ParseResult(&result); err != nil {
			return err
		}
		for k, v := range result.Attributes {
			fmt.Printf("%s=%s\n", k, v)
		}
		return nil
	},
}

// getBboxCmd 取得指定索引元素的 bounding box 座標
var getBboxCmd = &cobra.Command{
	Use:   "bbox <index>",
	Short: "取得元素 bounding box",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("元素索引必須為整數，收到: %s", args[0])
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "get_bbox", map[string]any{"index": index})
		if err != nil {
			return err
		}

		f := getFormatter()
		if resp.IsError() {
			f.PrintError("%s", resp.Error.Message)
			return resp.Error
		}

		// JSON 模式：輸出原始 bounding box 結構
		if flagJSON {
			return f.PrintJSON(json.RawMessage(resp.Result))
		}

		// Human 模式：以格式化方式輸出座標與尺寸
		var result struct {
			X      float64 `json:"x"`
			Y      float64 `json:"y"`
			Width  float64 `json:"width"`
			Height float64 `json:"height"`
		}
		if err := resp.ParseResult(&result); err != nil {
			return err
		}
		fmt.Printf("x=%.1f y=%.1f width=%.1f height=%.1f\n",
			result.X, result.Y, result.Width, result.Height)
		return nil
	},
}

func init() {
	// getHtmlCmd 的 --selector flag
	getHtmlCmd.Flags().String("selector", "", "CSS 選擇器（留空則取全頁 HTML）")

	// 組裝子指令
	getCmd.AddCommand(getTitleCmd)
	getCmd.AddCommand(getHtmlCmd)
	getCmd.AddCommand(getTextCmd)
	getCmd.AddCommand(getValueCmd)
	getCmd.AddCommand(getAttributesCmd)
	getCmd.AddCommand(getBboxCmd)
	rootCmd.AddCommand(getCmd)
}
