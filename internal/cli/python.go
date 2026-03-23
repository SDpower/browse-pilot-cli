// Package cli — python 子指令：在持久 Python session 中執行程式碼，
// 透過 browser 物件操作瀏覽器。
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/i18n"
	"github.com/SDpower/browse-pilot-cli/internal/python"
)

// pythonSession 是全域持久的 Python 子程序 session。
// 跨指令呼叫共用同一個 session，直到 --reset 才重置。
var pythonSession *python.Session

// pythonCmd 定義 `bp python` 子指令
var pythonCmd = &cobra.Command{
	Use:  "python [code]",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fileFlag, _ := cmd.Flags().GetString("file")
		varsFlag, _ := cmd.Flags().GetBool("vars")
		resetFlag, _ := cmd.Flags().GetBool("reset")

		f := getFormatter()

		// --reset：重置 Python session
		if resetFlag {
			if pythonSession != nil {
				pythonSession.Close()
				pythonSession = nil
			}
			f.PrintSuccess("%s", i18n.T("python.reset.success"))
			return nil
		}

		// 確保 session 已啟動
		if pythonSession == nil || !pythonSession.IsRunning() {
			s, err := python.NewSession("")
			if err != nil {
				return fmt.Errorf(i18n.T("error.python_start"), err)
			}
			pythonSession = s

			// 設定 browser 操作回調：Python 呼叫 browser.xxx() 時透過 transport 轉發
			pythonSession.SetBrowserCallback(func(ctx context.Context, requestJSON string) (any, error) {
				var req struct {
					Method string         `json:"method"`
					Params map[string]any `json:"params"`
				}
				if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
					return nil, fmt.Errorf(i18n.T("error.python_decode_request"), err)
				}

				tr, err := getTransport()
				if err != nil {
					return nil, err
				}
				defer tr.Close()

				resp, err := sendCommand(tr, req.Method, req.Params)
				if err != nil {
					return nil, err
				}
				if resp.IsError() {
					return nil, fmt.Errorf("%s", resp.Error.Message)
				}

				var result any
				if err := json.Unmarshal(resp.Result, &result); err != nil {
					return nil, fmt.Errorf(i18n.T("error.python_decode_response"), err)
				}
				return result, nil
			})
		}

		// --vars：列出 session 變數
		if varsFlag {
			vars, err := pythonSession.GetVars(context.Background())
			if err != nil {
				return err
			}
			if flagJSON {
				return f.PrintJSON(vars)
			}
			if len(vars) == 0 {
				f.PrintInfo("%s", i18n.T("python.vars.none"))
				return nil
			}
			for k, v := range vars {
				f.PrintKeyValue(k, v)
			}
			return nil
		}

		// 取得要執行的程式碼
		var code string
		if fileFlag != "" {
			data, err := os.ReadFile(fileFlag)
			if err != nil {
				return fmt.Errorf(i18n.T("error.python_read_script"), err)
			}
			code = string(data)
		} else if len(args) > 0 {
			code = args[0]
		} else {
			return errors.New(i18n.T("error.python_no_code"))
		}

		// 以 timeout 執行程式碼
		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(flagTimeout)*time.Millisecond,
		)
		defer cancel()

		result, err := pythonSession.Execute(ctx, code)
		if err != nil {
			f.PrintError(i18n.T("error.python_exec"), err)
			return err
		}

		if flagJSON {
			return f.PrintJSON(result)
		}
		if result != nil && result.Value != nil {
			fmt.Fprintln(os.Stdout, result.Value)
		}
		return nil
	},
}

func init() {
	// 設定 Short 與 Long 描述
	pythonCmd.Short = i18n.T("python.short")
	pythonCmd.Long = i18n.T("python.long")

	pythonCmd.Flags().String("file", "", i18n.T("python.file_flag"))
	pythonCmd.Flags().Bool("vars", false, i18n.T("python.vars_flag"))
	pythonCmd.Flags().Bool("reset", false, i18n.T("python.reset_flag"))
	rootCmd.AddCommand(pythonCmd)
}
