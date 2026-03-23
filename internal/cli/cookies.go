// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// cookiesCmd 是 Cookie 管理的父指令
var cookiesCmd = &cobra.Command{
	Use:   "cookies",
	Short: "Cookie 管理",
	// 未提供子指令時顯示說明
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// cookiesGetCmd 取得目前頁面的 cookies
var cookiesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "取得 cookies",
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("url")

		// 若有指定 URL 則加入篩選參數
		params := map[string]any{}
		if url != "" {
			params["url"] = url
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "get_cookies", params)
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
			return f.PrintJSON(resp.Result)
		}

		// Human 模式：解析 cookies 並逐行列出 name=value
		var result struct {
			Cookies []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"cookies"`
		}
		if err := resp.ParseResult(&result); err != nil {
			return err
		}

		if len(result.Cookies) == 0 {
			fmt.Println("（無 cookies）")
			return nil
		}
		for _, c := range result.Cookies {
			fmt.Printf("%s=%s\n", c.Name, c.Value)
		}
		return nil
	},
}

// cookiesSetCmd 設定一個 cookie
var cookiesSetCmd = &cobra.Command{
	Use:   "set <name> <value>",
	Short: "設定 cookie",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain, _ := cmd.Flags().GetString("domain")
		secure, _ := cmd.Flags().GetBool("secure")
		sameSite, _ := cmd.Flags().GetString("same-site")

		// 組建 cookie 參數
		params := map[string]any{
			"name":  args[0],
			"value": args[1],
		}
		if domain != "" {
			params["domain"] = domain
		}
		if secure {
			params["secure"] = true
		}
		if sameSite != "" {
			params["sameSite"] = sameSite
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "set_cookie", params)
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
		f.PrintSuccess("已設定 cookie: %s", args[0])
		return nil
	},
}

// cookiesClearCmd 清除 cookies，可指定特定 URL
var cookiesClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "清除 cookies",
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("url")

		// 若有指定 URL 則只清除該 URL 的 cookies
		params := map[string]any{}
		if url != "" {
			params["url"] = url
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		resp, err := sendCommand(tr, "clear_cookies", params)
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
		if url != "" {
			f.PrintSuccess("已清除 %s 的 cookies", url)
		} else {
			f.PrintSuccess("已清除所有 cookies")
		}
		return nil
	},
}

// cookiesExportCmd 將所有 cookies 匯出為 JSON 檔
var cookiesExportCmd = &cobra.Command{
	Use:   "export <file>",
	Short: "匯出 cookies 至 JSON 檔",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		// 取得所有 cookies
		resp, err := sendCommand(tr, "get_cookies", map[string]any{})
		if err != nil {
			return err
		}

		f := getFormatter()
		if resp.IsError() {
			f.PrintError("%s", resp.Error.Message)
			return resp.Error
		}

		// 解析回應中的 cookies 陣列
		var result struct {
			Cookies json.RawMessage `json:"cookies"`
		}
		if parseErr := resp.ParseResult(&result); parseErr != nil {
			return parseErr
		}

		var cookies []any
		if unmarshalErr := json.Unmarshal(result.Cookies, &cookies); unmarshalErr != nil {
			return fmt.Errorf("解析 cookies 失敗: %w", unmarshalErr)
		}

		// 以 pretty-printed JSON 格式寫入檔案
		data, err := json.MarshalIndent(cookies, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化失敗: %w", err)
		}
		if err := os.WriteFile(filePath, data, 0o644); err != nil {
			return fmt.Errorf("寫入檔案失敗: %w", err)
		}

		f.PrintSuccess("已匯出 %d 筆 cookies 至 %s", len(cookies), filePath)
		return nil
	},
}

// cookiesImportCmd 從 JSON 檔匯入 cookies
var cookiesImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "從 JSON 檔匯入 cookies",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// 讀取並解析 JSON 檔案
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("讀取檔案失敗: %w", err)
		}

		var cookies []map[string]any
		if unmarshalErr := json.Unmarshal(data, &cookies); unmarshalErr != nil {
			return fmt.Errorf("JSON 解析失敗: %w", unmarshalErr)
		}

		tr, err := getTransport()
		if err != nil {
			return err
		}
		defer tr.Close()

		f := getFormatter()
		successCount := 0
		failCount := 0

		// 逐一發送 set_cookie 至 extension
		for _, cookie := range cookies {
			resp, err := sendCommand(tr, "set_cookie", cookie)
			if err != nil || (resp != nil && resp.IsError()) {
				failCount++
				f.PrintVerbose("匯入失敗: %v", cookie["name"])
				continue
			}
			successCount++
		}

		f.PrintSuccess("已匯入 %d 筆 cookies", successCount)
		if failCount > 0 {
			f.PrintWarning("%d 筆匯入失敗", failCount)
		}
		return nil
	},
}

func init() {
	// cookiesGetCmd 的 flags
	cookiesGetCmd.Flags().String("url", "", "篩選指定 URL 的 cookies")

	// cookiesSetCmd 的 flags
	cookiesSetCmd.Flags().String("domain", "", "cookie 的 domain")
	cookiesSetCmd.Flags().Bool("secure", false, "secure cookie")
	cookiesSetCmd.Flags().String("same-site", "", "SameSite 屬性（Strict/Lax/None）")

	// cookiesClearCmd 的 flags
	cookiesClearCmd.Flags().String("url", "", "清除指定 URL 的 cookies")

	// 組裝子指令
	cookiesCmd.AddCommand(cookiesGetCmd)
	cookiesCmd.AddCommand(cookiesSetCmd)
	cookiesCmd.AddCommand(cookiesClearCmd)
	cookiesCmd.AddCommand(cookiesExportCmd)
	cookiesCmd.AddCommand(cookiesImportCmd)
	rootCmd.AddCommand(cookiesCmd)
}
