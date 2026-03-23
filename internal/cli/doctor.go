// Package cli 中的 doctor 指令：本機離線檢查瀏覽器與 Extension 安裝狀態
package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/i18n"
	"github.com/SDpower/browse-pilot-cli/internal/output"
	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

// doctorCmd 執行本機環境診斷，不需要建立 transport 連線
var doctorCmd = &cobra.Command{
	Use: "doctor",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := output.NewFormatter(flagJSON, flagVerbose)

		// 偵測所有瀏覽器安裝狀態
		browsers := transport.DetectBrowsers()

		if flagJSON {
			return f.PrintJSON(map[string]any{
				"browsers": browsers,
			})
		}

		fmt.Fprintln(os.Stdout, i18n.T("doctor.header"))
		fmt.Fprintln(os.Stdout)

		green := color.New(color.FgGreen).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		for _, b := range browsers {
			name := b.Name

			// 顯示瀏覽器執行狀態
			if b.Running {
				fmt.Fprintf(os.Stdout, "%s %s\n", green("✓"), fmt.Sprintf(i18n.T("doctor.browser_running"), name))
			} else {
				fmt.Fprintf(os.Stdout, "%s %s\n", yellow("○"), fmt.Sprintf(i18n.T("doctor.browser_stopped"), name))
			}

			// 顯示 Native Messaging host 安裝狀態
			if b.HasNMHost {
				fmt.Fprintf(os.Stdout, "  %s %s\n", green("✓"), i18n.T("doctor.nm_installed"))
			} else {
				nmPath := transport.NMHostPath(name)
				fmt.Fprintf(os.Stdout, "  %s %s\n", red("✗"), i18n.T("doctor.nm_not_installed"))
				fmt.Fprintf(os.Stdout, "    %s\n", fmt.Sprintf(i18n.T("doctor.nm_install_hint"), name))
				fmt.Fprintf(os.Stdout, "    %s\n", fmt.Sprintf(i18n.T("doctor.nm_path"), nmPath))
			}
			fmt.Fprintln(os.Stdout)
		}

		if len(browsers) == 0 {
			f.PrintError("%s", i18n.T("doctor.no_browser"))
		}

		return nil
	},
}

func init() {
	// 設定 Short 描述
	doctorCmd.Short = i18n.T("doctor.short")
	rootCmd.AddCommand(doctorCmd)
}
