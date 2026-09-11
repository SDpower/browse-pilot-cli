// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/SDpower/browse-pilot-cli/internal/i18n"
	"github.com/SDpower/browse-pilot-cli/internal/output"
	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

const cliVersion = "0.2.0"

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
	// flagMCPHTTP 是否啟動 Streamable HTTP MCP server
	flagMCPHTTP bool
	// flagMCPPort 是 Streamable HTTP MCP server 的 loopback 埠號
	flagMCPPort int
	// flagNativeMessaging 是否以 Native Messaging host 模式啟動
	flagNativeMessaging bool
	// flagSession 連線 session 名稱，預設 "default"
	flagSession string
)

// rootCmd 是 bp_cli 指令的根節點
var rootCmd = &cobra.Command{
	Use:     "bp_cli",
	Version: cliVersion,
	// 根指令依 flag 選擇啟動模式，否則顯示說明
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagNativeMessaging {
			return runNativeMessagingHost()
		}
		if flagMCPHTTP {
			return runMCPHTTPServer()
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
	// 設定 rootCmd 的 Short 與 Long 描述
	rootCmd.Short = i18n.T("root.short")
	rootCmd.Long = i18n.T("root.long")

	// 瀏覽器選擇 flag
	rootCmd.PersistentFlags().StringVar(
		&flagBrowser,
		"browser",
		"auto",
		i18n.T("flag.browser"),
	)

	// WebSocket 埠號 flag
	rootCmd.PersistentFlags().IntVar(
		&flagPort,
		"port",
		9222,
		i18n.T("flag.port"),
	)

	// JSON 輸出 flag
	rootCmd.PersistentFlags().BoolVar(
		&flagJSON,
		"json",
		false,
		i18n.T("flag.json"),
	)

	// 逾時時間 flag
	rootCmd.PersistentFlags().IntVar(
		&flagTimeout,
		"timeout",
		30000,
		i18n.T("flag.timeout"),
	)

	// 詳細日誌 flag
	rootCmd.PersistentFlags().BoolVar(
		&flagVerbose,
		"verbose",
		false,
		i18n.T("flag.verbose"),
	)

	// Streamable HTTP MCP server 模式 flag
	rootCmd.PersistentFlags().BoolVar(
		&flagMCPHTTP,
		"mcp-http",
		false,
		i18n.T("flag.mcp"),
	)

	rootCmd.PersistentFlags().IntVar(
		&flagMCPPort,
		"mcp-port",
		8931,
		"Streamable HTTP MCP server 埠號",
	)

	// Native Messaging host 模式 flag
	rootCmd.PersistentFlags().BoolVar(
		&flagNativeMessaging,
		"native-messaging",
		false,
		i18n.T("flag.native_messaging"),
	)

	// Session 名稱 flag
	rootCmd.PersistentFlags().StringVar(
		&flagSession,
		"session",
		"default",
		i18n.T("flag.session"),
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

// GetMCPHTTP 取得 --mcp-http flag 的值。
func GetMCPHTTP() bool {
	return flagMCPHTTP
}

// GetMCPPort 取得 --mcp-port flag 的值。
func GetMCPPort() int {
	return flagMCPPort
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
	cfg.Browser = browser

	var tr transport.Transport
	switch browser {
	case "firefox":
		tr = transport.NewWSTransport(cfg)
	case "chrome", "edge":
		tr = transport.NewNMTransport(cfg)
	default:
		return nil, fmt.Errorf(i18n.T("error.unsupported_browser"), browser)
	}

	if err := tr.Start(context.Background()); err != nil {
		return nil, err
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
