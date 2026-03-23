// Package mcp 的 resources 子模組，定義所有可讀取的瀏覽器狀態 MCP resource。
package mcp

import (
	"context"
	"encoding/json"
)

// RegisterAllResources 向 MCP server 註冊所有 browse-pilot 瀏覽器狀態 resource。
// resource 代表可讀取的瀏覽器即時狀態，例如當前頁面資訊或截圖。
func RegisterAllResources(s *Server) {
	// 當前頁面狀態 resource：包含 URL、標題和可互動元素列表
	s.RegisterResource(&Resource{
		URI:         "bp://state",
		Name:        "當前頁面狀態",
		Description: "頁面 URL、標題和可互動元素列表",
		MimeType:    "application/json",
		Handler: func(ctx context.Context) (string, error) {
			result, err := s.callExtensionRaw(ctx, "get_state", nil)
			if err != nil {
				return "", err
			}
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return "", err
			}
			return string(data), nil
		},
	})

	// 當前頁面截圖 resource：回傳 base64 編碼的 PNG 截圖
	s.RegisterResource(&Resource{
		URI:         "bp://screenshot",
		Name:        "當前頁面截圖",
		Description: "頁面截圖（base64 PNG）",
		MimeType:    "image/png",
		Handler: func(ctx context.Context) (string, error) {
			result, err := s.callExtensionRaw(ctx, "screenshot", nil)
			if err != nil {
				return "", err
			}
			// result 包含 {"data": "base64...", "width": ..., "height": ...}
			data, err := json.Marshal(result)
			if err != nil {
				return "", err
			}
			return string(data), nil
		},
	})
}
