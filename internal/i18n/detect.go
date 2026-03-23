// Package i18n — detect.go 負責偵測系統語系並對應到支援的語系代碼。
package i18n

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// supportedLocales 是原始語系字串與支援語系代碼的對應表。
var supportedLocales = map[string]string{
	"en":         "en",
	"de":         "de",
	"es":         "es",
	"fr":         "fr",
	"hi":         "hi",
	"id":         "id",
	"it":         "it",
	"ja":         "ja",
	"ko":         "ko",
	"pt":         "pt-BR",
	"pt_BR":      "pt-BR",
	"pt-BR":      "pt-BR",
	"zh":         "zh-Hans",
	"zh_CN":      "zh-Hans",
	"zh_Hans":    "zh-Hans",
	"zh-Hans":    "zh-Hans",
	"zh_TW":      "zh-Hant",
	"zh_Hant":    "zh-Hant",
	"zh-Hant":    "zh-Hant",
	"zh-Hant-TW": "zh-Hant",
}

// DetectLocale 偵測系統語系並回傳對應的支援語系代碼。
// macOS：AppleLocale → LANG/LC_MESSAGES → en
// 其他平台：LANG/LC_MESSAGES → en
func DetectLocale() string {
	// macOS 優先使用系統設定
	if runtime.GOOS == "darwin" {
		if locale := detectMacOSLocale(); locale != "" {
			return locale
		}
	}

	// 依序讀取環境變數
	for _, env := range []string{"LC_MESSAGES", "LANG"} {
		if v := os.Getenv(env); v != "" {
			if locale := matchLocale(v); locale != "" {
				return locale
			}
		}
	}

	return "en"
}

// detectMacOSLocale 透過 `defaults read -g AppleLocale` 取得 macOS 系統語系。
func detectMacOSLocale() string {
	out, err := exec.Command("defaults", "read", "-g", "AppleLocale").Output()
	if err != nil {
		return ""
	}
	raw := strings.TrimSpace(string(out))
	return matchLocale(raw)
}

// matchLocale 將原始語系字串（如 "zh_TW.UTF-8"）對應到支援的語系代碼。
// 若無對應則回傳空字串。
func matchLocale(raw string) string {
	if raw == "" {
		return ""
	}

	// 移除 .UTF-8 等編碼後綴
	raw = strings.Split(raw, ".")[0]
	// 移除 @modifier（如 @euro）
	raw = strings.Split(raw, "@")[0]

	// 完整比對
	if locale, ok := supportedLocales[raw]; ok {
		return locale
	}

	// 將 _ 替換為 - 後再比對（如 zh_Hant-TW → zh-Hant-TW）
	replaced := strings.ReplaceAll(raw, "_", "-")
	if locale, ok := supportedLocales[replaced]; ok {
		return locale
	}

	// 取 _ 分隔的語言前綴（如 zh_TW → zh）
	parts := strings.SplitN(raw, "_", 2)
	if len(parts) > 1 {
		if locale, ok := supportedLocales[parts[0]]; ok {
			return locale
		}
	}

	// 取 - 分隔的語言前綴（如 en-US → en）
	parts = strings.SplitN(raw, "-", 2)
	if len(parts) > 1 {
		if locale, ok := supportedLocales[parts[0]]; ok {
			return locale
		}
	}

	return ""
}
