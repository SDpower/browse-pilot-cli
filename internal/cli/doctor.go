// Package cli 中的 doctor 指令：本機離線檢查瀏覽器與 Extension 安裝狀態
package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/output"
	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

// doctorCmd 執行本機環境診斷，不需要建立 transport 連線
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "檢查瀏覽器、Extension、連線狀態",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := output.NewFormatter(flagJSON, flagVerbose)

		// 偵測所有瀏覽器安裝狀態
		browsers := transport.DetectBrowsers()

		if flagJSON {
			return f.PrintJSON(map[string]any{
				"browsers": browsers,
			})
		}

		fmt.Fprintln(os.Stdout, "=== Browse Pilot Doctor ===")
		fmt.Fprintln(os.Stdout)

		green := color.New(color.FgGreen).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		for _, b := range browsers {
			name := b.Name

			// 顯示瀏覽器執行狀態
			if b.Running {
				fmt.Fprintf(os.Stdout, "%s %s — 執行中\n", green("✓"), name)
			} else {
				fmt.Fprintf(os.Stdout, "%s %s — 未執行\n", yellow("○"), name)
			}

			// 顯示 Native Messaging host 安裝狀態
			if b.HasNMHost {
				fmt.Fprintf(os.Stdout, "  %s Native Messaging host 已安裝\n", green("✓"))
			} else {
				nmPath := transport.NMHostPath(name)
				fmt.Fprintf(os.Stdout, "  %s Native Messaging host 未安裝\n", red("✗"))
				fmt.Fprintf(os.Stdout, "    執行 `bp_cli setup %s` 安裝\n", name)
				fmt.Fprintf(os.Stdout, "    路徑: %s\n", nmPath)
			}
			fmt.Fprintln(os.Stdout)
		}

		if len(browsers) == 0 {
			f.PrintError("未偵測到任何瀏覽器")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
