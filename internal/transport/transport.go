// Package transport 定義 CLI 與瀏覽器 Extension 之間的通訊抽象層
package transport

import (
	"context"
	"time"
)

// Transport 是瀏覽器通訊的抽象介面。
// WebSocket（Firefox）和 Native Messaging（Chrome/Edge）各自實作此介面，
// 使上層指令邏輯無需關心底層通訊方式。
type Transport interface {
	// Start 啟動 transport。
	// WebSocket 模式：啟動 HTTP server 並等待 Extension 連線。
	// Native Messaging 模式：啟動 stdin/stdout 監聽迴圈。
	Start(ctx context.Context) error

	// Send 發送 JSON-RPC 請求至 Extension，並阻塞等待對應的回應。
	// 若在 ctx 逾時前未收到回應，回傳 context.DeadlineExceeded。
	Send(ctx context.Context, req *Request) (*Response, error)

	// Close 關閉 transport 連線並釋放相關資源。
	Close() error

	// IsConnected 回傳目前是否有活躍的 Extension 連線。
	IsConnected() bool

	// Type 回傳 transport 的類型識別字串。
	// 可能值："websocket"、"native_messaging"
	Type() string
}

// Config 是建立 Transport 時的共用設定結構。
type Config struct {
	// Port 是 WebSocket 伺服器的監聽埠號（僅 WebSocket 模式使用）
	Port int

	// Timeout 是單一指令的最大等待時間
	Timeout time.Duration

	// Verbose 若為 true，將輸出詳細的通訊日誌
	Verbose bool
}
