// Package cli 定義 bp CLI 的所有 Cobra 指令。
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SDpower/browse-pilot-cli/internal/mcp"
	"github.com/SDpower/browse-pilot-cli/internal/transport"
)

// runMCPHTTPServer 啟動單一共用的 Streamable HTTP MCP server。
func runMCPHTTPServer() error {
	browser := flagBrowser
	if browser == "auto" {
		browser = transport.AutoDetectBrowser()
	}
	if browser != "firefox" {
		return fmt.Errorf("Streamable HTTP MCP 目前僅支援 Firefox，實際為 %s", browser)
	}
	if flagMCPPort < 1 || flagMCPPort > 65535 {
		return fmt.Errorf("無效的 MCP HTTP 埠號: %d", flagMCPPort)
	}

	mcpAddress := fmt.Sprintf("127.0.0.1:%d", flagMCPPort)
	listener, err := net.Listen("tcp", mcpAddress)
	if err != nil {
		return fmt.Errorf("無法監聽 MCP HTTP 位址 %s: %w", mcpAddress, err)
	}
	defer listener.Close()

	cfg := transport.Config{
		Browser: browser,
		Port:    flagPort,
		Timeout: time.Duration(flagTimeout) * time.Millisecond,
		Verbose: flagVerbose,
	}
	tr := transport.NewWSTransport(cfg)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := tr.Start(ctx); err != nil {
		return err
	}
	defer tr.Close()

	server := mcp.NewServer(tr, flagVerbose)
	server.SetRequestTimeout(cfg.Timeout)
	server.SetBrowserContext(browser, flagPort)
	mcp.RegisterAllTools(server)
	mcp.RegisterAllResources(server)

	mux := http.NewServeMux()
	mux.Handle("/mcp", server.HTTPHandler())
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true, "extensionConnected": tr.IsConnected(), "browser": browser,
		})
	})
	httpServer := &http.Server{
		Addr:              mcpAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if flagVerbose {
		fmt.Fprintf(os.Stderr, "[MCP] Streamable HTTP 已監聽 http://%s/mcp\n", mcpAddress)
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- httpServer.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("關閉 MCP HTTP server 失敗: %w", err)
		}
		return nil
	case err := <-serveErr:
		if err == nil || err == http.ErrServerClosed {
			return nil
		}
		return fmt.Errorf("MCP HTTP server 失敗: %w", err)
	}
}
