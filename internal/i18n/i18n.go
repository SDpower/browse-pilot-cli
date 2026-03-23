// Package i18n 提供多語系翻譯功能，支援從 embed FS 載入 JSON 翻譯檔。
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed locales/*.json
var localesFS embed.FS

var (
	// messages 是目前語系的翻譯對應表
	messages map[string]string
	// fallback 是英文的翻譯對應表，找不到目標語系 key 時使用
	fallback map[string]string
	// mu 保護 messages、fallback 與 currentLocale 的並發存取
	mu            sync.RWMutex
	currentLocale string
)

func init() {
	locale := DetectLocale()
	Init(locale)
}

// Init 載入指定語系的翻譯檔，en 作為 fallback。
func Init(locale string) {
	mu.Lock()
	defer mu.Unlock()

	currentLocale = locale

	// 載入英文作為 fallback
	fallback = loadLocale("en")

	// 載入目標語系；若為英文則與 fallback 共用
	if locale != "en" {
		messages = loadLocale(locale)
	} else {
		messages = fallback
	}
}

// loadLocale 從 embed FS 讀取並解析指定語系的 JSON 翻譯檔。
// 若檔案不存在或解析失敗則回傳 nil。
func loadLocale(locale string) map[string]string {
	filename := fmt.Sprintf("locales/%s.json", locale)
	data, err := localesFS.ReadFile(filename)
	if err != nil {
		return nil
	}

	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	return m
}

// T 回傳指定 key 的翻譯字串。
// 查找順序：目前語系 → 英文 fallback → 回傳 key 本身。
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if messages != nil {
		if v, ok := messages[key]; ok {
			return v
		}
	}
	if fallback != nil {
		if v, ok := fallback[key]; ok {
			return v
		}
	}
	return key
}

// Tf 回傳翻譯後以 fmt.Sprintf 格式化的字串。
func Tf(key string, args ...any) string {
	return fmt.Sprintf(T(key), args...)
}

// Locale 回傳目前使用的語系代碼。
func Locale() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentLocale
}
