package i18n

import "testing"

func TestMatchLocale(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"en_US.UTF-8", "en"},
		{"en", "en"},
		{"zh_TW.UTF-8", "zh-Hant"},
		{"zh_TW", "zh-Hant"},
		{"zh_CN", "zh-Hans"},
		{"zh-Hant-TW", "zh-Hant"},
		{"ja_JP.UTF-8", "ja"},
		{"ko_KR", "ko"},
		{"de_DE.UTF-8", "de"},
		{"fr_FR", "fr"},
		{"pt_BR.UTF-8", "pt-BR"},
		{"es_ES", "es"},
		{"it_IT", "it"},
		{"id_ID", "id"},
		{"hi_IN", "hi"},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := matchLocale(tt.input)
		if got != tt.expected {
			t.Errorf("matchLocale(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
