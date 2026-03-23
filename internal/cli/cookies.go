// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/i18n"
)

// cookiesCmd 是 Cookie 管理的父指令
var cookiesCmd = &cobra.Command{
	Use: "cookies",
	// 未提供子指令時顯示說明
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// cookiesGetCmd 取得目前頁面的 cookies
var cookiesGetCmd = &cobra.Command{
	Use: "get",
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
			fmt.Println(i18n.T("cookies.none"))
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
	Use:  "set <name> <value>",
	Args: cobra.ExactArgs(2),
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
		f.PrintSuccess(i18n.T("cookies.set.success"), args[0])
		return nil
	},
}

// cookiesClearCmd 清除 cookies，可指定特定 URL
var cookiesClearCmd = &cobra.Command{
	Use: "clear",
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
			f.PrintSuccess(i18n.T("cookies.clear.url_success"), url)
		} else {
			f.PrintSuccess("%s", i18n.T("cookies.clear.all_success"))
		}
		return nil
	},
}

// cookiesExportCmd 將所有 cookies 匯出為 JSON 檔
var cookiesExportCmd = &cobra.Command{
	Use:  "export <file>",
	Args: cobra.ExactArgs(1),
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
			return fmt.Errorf(i18n.T("error.parse_cookies"), unmarshalErr)
		}

		// 以 pretty-printed JSON 格式寫入檔案
		data, err := json.MarshalIndent(cookies, "", "  ")
		if err != nil {
			return fmt.Errorf(i18n.T("error.serialize"), err)
		}
		if err := os.WriteFile(filePath, data, 0o644); err != nil {
			return fmt.Errorf(i18n.T("error.write_file"), err)
		}

		f.PrintSuccess(i18n.T("cookies.export.success"), len(cookies), filePath)
		return nil
	},
}

// cookiesImportCmd 從 JSON 檔匯入 cookies
var cookiesImportCmd = &cobra.Command{
	Use:  "import <file>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// 讀取並解析 JSON 檔案
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf(i18n.T("error.read_file"), err)
		}

		var cookies []map[string]any
		if unmarshalErr := json.Unmarshal(data, &cookies); unmarshalErr != nil {
			return fmt.Errorf(i18n.T("error.json_parse"), unmarshalErr)
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
				f.PrintVerbose(i18n.T("cookies.import.verbose_fail"), cookie["name"])
				continue
			}
			successCount++
		}

		f.PrintSuccess(i18n.T("cookies.import.success"), successCount)
		if failCount > 0 {
			f.PrintWarning(i18n.T("cookies.import.fail_warning"), failCount)
		}
		return nil
	},
}

func init() {
	// 設定各指令的 Short 描述
	cookiesCmd.Short = i18n.T("cookies.short")
	cookiesGetCmd.Short = i18n.T("cookies.get.short")
	cookiesSetCmd.Short = i18n.T("cookies.set.short")
	cookiesClearCmd.Short = i18n.T("cookies.clear.short")
	cookiesExportCmd.Short = i18n.T("cookies.export.short")
	cookiesImportCmd.Short = i18n.T("cookies.import.short")

	// cookiesGetCmd 的 flags
	cookiesGetCmd.Flags().String("url", "", i18n.T("cookies.get.url_flag"))

	// cookiesSetCmd 的 flags
	cookiesSetCmd.Flags().String("domain", "", i18n.T("cookies.set.domain_flag"))
	cookiesSetCmd.Flags().Bool("secure", false, i18n.T("cookies.set.secure_flag"))
	cookiesSetCmd.Flags().String("same-site", "", i18n.T("cookies.set.same_site_flag"))

	// cookiesClearCmd 的 flags
	cookiesClearCmd.Flags().String("url", "", i18n.T("cookies.clear.url_flag"))

	// 組裝子指令
	cookiesCmd.AddCommand(cookiesGetCmd)
	cookiesCmd.AddCommand(cookiesSetCmd)
	cookiesCmd.AddCommand(cookiesClearCmd)
	cookiesCmd.AddCommand(cookiesExportCmd)
	cookiesCmd.AddCommand(cookiesImportCmd)
	rootCmd.AddCommand(cookiesCmd)
}
