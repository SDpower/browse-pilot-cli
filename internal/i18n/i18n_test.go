package i18n

import "testing"

func TestTFallback(t *testing.T) {
	Init("en")
	// 存在的 key 應回傳英文翻譯，而非 key 本身
	v := T("root.short")
	if v == "root.short" {
		t.Error("root.short 應有翻譯，不應回傳 key 本身")
	}
}

func TestTMissing(t *testing.T) {
	Init("en")
	v := T("nonexistent.key.xxx")
	if v != "nonexistent.key.xxx" {
		t.Errorf("不存在的 key 應回傳 key 本身，got %q", v)
	}
}

func TestTf(t *testing.T) {
	Init("en")
	v := Tf("error.unsupported_browser", "safari")
	if v == "" {
		t.Error("Tf 不應回傳空字串")
	}
}

func TestInitZhHant(t *testing.T) {
	Init("zh-Hant")
	v := T("root.short")
	if v == "root.short" {
		t.Error("zh-Hant 應有翻譯")
	}
	// 確認回傳的是中文而非英文
	if v == "Cross-browser automation CLI tool" {
		t.Error("zh-Hant 不應回傳英文")
	}
}

func TestLocale(t *testing.T) {
	Init("ja")
	if Locale() != "ja" {
		t.Errorf("Locale() = %q, want ja", Locale())
	}
}

func TestInitUnsupported(t *testing.T) {
	Init("xx")
	// 不支援的語系應 fallback 到英文
	v := T("root.short")
	if v == "root.short" {
		t.Error("不支援的語系應 fallback 到英文")
	}
}
