// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/output"
	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

// 全域 flag 變數
var (
	// flagBrowser 指定目標瀏覽器，支援 firefox/chrome/edge/auto
	flagBrowser string
	// flagPort WebSocket 伺服器埠號，預設 9222
	flagPort int
	// flagJSON 是否以 JSON 格式輸出結果
	flagJSON bool
	// flagTimeout 指令逾時時間（毫秒），預設 30000
	flagTimeout int
	// flagVerbose 是否啟用詳細日誌輸出
	flagVerbose bool
	// flagMCP 是否以 MCP server 模式啟動
	flagMCP bool
	// flagNativeMessaging 是否以 Native Messaging host 模式啟動
	flagNativeMessaging bool
	// flagSession 連線 session 名稱，預設 "default"
	flagSession string
)

// rootCmd 是 bp 指令的根節點
var rootCmd = &cobra.Command{
	Use:   "bp",
	Short: "跨瀏覽器自動化 CLI 工具",
	Long: `bp 是一個跨瀏覽器自動化 CLI 工具，
透過 WebExtension API 控制 Firefox/Chrome/Edge，
支援 WebSocket 及 Native Messaging 雙通道通訊。`,
	// 根指令依 flag 選擇啟動模式，否則顯示說明
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagNativeMessaging {
			return runNativeMessagingHost()
		}
		if flagMCP {
			return runMCPServer()
		}
		return cmd.Help()
	},
}

// Execute 執行根指令，所有子指令皆由此啟動
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// 瀏覽器選擇 flag
	rootCmd.PersistentFlags().StringVar(
		&flagBrowser,
		"browser",
		"auto",
		"目標瀏覽器（firefox/chrome/edge/auto）",
	)

	// WebSocket 埠號 flag
	rootCmd.PersistentFlags().IntVar(
		&flagPort,
		"port",
		9222,
		"WebSocket 伺服器埠號",
	)

	// JSON 輸出 flag
	rootCmd.PersistentFlags().BoolVar(
		&flagJSON,
		"json",
		false,
		"以 JSON 格式輸出結果",
	)

	// 逾時時間 flag
	rootCmd.PersistentFlags().IntVar(
		&flagTimeout,
		"timeout",
		30000,
		"指令逾時時間（毫秒）",
	)

	// 詳細日誌 flag
	rootCmd.PersistentFlags().BoolVar(
		&flagVerbose,
		"verbose",
		false,
		"啟用詳細日誌輸出",
	)

	// MCP server 模式 flag
	rootCmd.PersistentFlags().BoolVar(
		&flagMCP,
		"mcp",
		false,
		"以 MCP server 模式啟動",
	)

	// Native Messaging host 模式 flag
	rootCmd.PersistentFlags().BoolVar(
		&flagNativeMessaging,
		"native-messaging",
		false,
		"以 Native Messaging host 模式啟動（供 Chrome/Edge Extension 呼叫）",
	)

	// Session 名稱 flag
	rootCmd.PersistentFlags().StringVar(
		&flagSession,
		"session",
		"default",
		"連線 session 名稱",
	)
}

// GetBrowser 取得 --browser flag 的值
func GetBrowser() string {
	return flagBrowser
}

// GetPort 取得 --port flag 的值
func GetPort() int {
	return flagPort
}

// GetJSON 取得 --json flag 的值
func GetJSON() bool {
	return flagJSON
}

// GetTimeout 取得 --timeout flag 的值（毫秒）
func GetTimeout() int {
	return flagTimeout
}

// GetVerbose 取得 --verbose flag 的值
func GetVerbose() bool {
	return flagVerbose
}

// GetMCP 取得 --mcp flag 的值
func GetMCP() bool {
	return flagMCP
}

// GetNativeMessaging 取得 --native-messaging flag 的值
func GetNativeMessaging() bool { return flagNativeMessaging }

// GetSession 取得 --session flag 的值
func GetSession() string {
	return flagSession
}

// getTransport 根據當前 flag 建立並啟動 transport。
// 依瀏覽器類型選擇 WebSocket（Firefox）或 Native Messaging（Chrome/Edge）。
func getTransport() (transport.Transport, error) {
	cfg := transport.Config{
		Port:    flagPort,
		Timeout: time.Duration(flagTimeout) * time.Millisecond,
		Verbose: flagVerbose,
	}

	browser := flagBrowser
	if browser == "auto" {
		browser = transport.AutoDetectBrowser()
	}

	var tr transport.Transport
	switch browser {
	case "firefox":
		tr = transport.NewWSTransport(cfg)
	case "chrome", "edge":
		tr = transport.NewNMTransport(cfg)
	default:
		return nil, fmt.Errorf("不支援的瀏覽器: %s", browser)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := tr.Start(ctx); err != nil {
		// WebSocket 模式下 Start 逾時表示 Extension 尚未連入，
		// 但 server 仍在運行，Send 時會再次等待連線
		if flagVerbose {
			fmt.Fprintf(os.Stderr, "[transport] %v\n", err)
		}
	}
	return tr, nil
}

// sendCommand 發送 JSON-RPC 指令並回傳原始 response。
func sendCommand(tr transport.Transport, method string, params any) (*transport.Response, error) {
	req, err := transport.NewRequest(method, params)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(flagTimeout)*time.Millisecond)
	defer cancel()

	return tr.Send(ctx, req)
}

// getFormatter 建立輸出格式化器。
func getFormatter() *output.Formatter {
	return output.NewFormatter(flagJSON, flagVerbose)
}
