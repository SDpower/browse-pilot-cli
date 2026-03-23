// Package cli 定義 bp CLI 的所有 Cobra 指令
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SDpower/browse-pilot-cli/internal/mcp"
	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

// runMCPServer 以 MCP server 模式啟動。
// MCP 透過 stdio (stdin/stdout) 與 Claude Code 溝通，
// 同時需要建立 transport 連線至瀏覽器 Extension。
func runMCPServer() error {
	if flagVerbose {
		fmt.Fprintln(os.Stderr, "[MCP] 啟動 MCP server 模式")
	}

	// 建立 transport 連線至 Extension
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
		return fmt.Errorf("不支援的瀏覽器: %s", browser)
	}

	// 建立帶取消功能的 context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 啟動 transport（背景）
	// MCP 模式下 transport 連線失敗不應阻止 server 啟動，
	// server 仍然啟動並等待 Extension 連入
	if err := tr.Start(ctx); err != nil {
		if flagVerbose {
			fmt.Fprintf(os.Stderr, "[MCP] Transport 啟動提示: %v（等待 Extension 連線）\n", err)
		}
	}
	defer tr.Close()

	if flagVerbose {
		fmt.Fprintf(os.Stderr, "[MCP] 使用 %s transport（瀏覽器: %s）\n", tr.Type(), browser)
	}

	// 建立 MCP server 並註冊所有 tool 與 resource
	server := mcp.NewServer(tr, flagVerbose)
	mcp.RegisterAllTools(server)
	mcp.RegisterAllResources(server)

	// 處理 SIGINT/SIGTERM，確保優雅關閉
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		if flagVerbose {
			fmt.Fprintln(os.Stderr, "[MCP] 收到終止信號，關閉中...")
		}
		cancel()
	}()

	// 啟動 MCP server 主迴圈（阻塞直到 ctx 取消或 stdin 關閉）
	return server.Run(ctx)
}
