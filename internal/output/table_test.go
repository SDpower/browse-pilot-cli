package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintStateHuman(t *testing.T) {
	var buf bytes.Buffer
	state := &StateResult{
		URL:   "https://example.com",
		Title: "測試頁面",
		Elements: []StateElement{
			{Index: 0, Tag: "input", Type: "text", Name: "帳號", Placeholder: "請輸入帳號"},
			{Index: 1, Tag: "button", Name: "登入"},
		},
	}
	err := PrintState(&buf, state, false)
	if err != nil {
		t.Fatalf("PrintState 失敗: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "[0]") {
		t.Error("應包含索引 [0]")
	}
	if !strings.Contains(output, "帳號") {
		t.Error("應包含元素名稱")
	}
	if !strings.Contains(output, "2 個") {
		t.Error("應包含元素數量")
	}
}

func TestPrintStateJSON(t *testing.T) {
	var buf bytes.Buffer
	state := &StateResult{
		URL:   "https://example.com",
		Title: "測試",
		Elements: []StateElement{
			{Index: 0, Tag: "button", Name: "確定"},
		},
	}
	err := PrintState(&buf, state, true)
	if err != nil {
		t.Fatalf("PrintState JSON 失敗: %v", err)
	}
	if !strings.Contains(buf.String(), `"url"`) {
		t.Error("JSON 應包含 url 欄位")
	}
}

func TestPrintStateEmpty(t *testing.T) {
	var buf bytes.Buffer
	state := &StateResult{URL: "about:blank", Title: ""}
	err := PrintState(&buf, state, false)
	if err != nil {
		t.Fatalf("PrintState 空頁面失敗: %v", err)
	}
	if !strings.Contains(buf.String(), "無可互動元素") {
		t.Error("空頁面應顯示提示")
	}
}

func TestFormatElementLongName(t *testing.T) {
	el := StateElement{
		Index: 5,
		Tag:   "a",
		Name:  "這是一個非常長的連結文字，超過四十個字元的部分應該被截斷顯示省略號",
	}
	line := formatElement(el)
	if !strings.Contains(line, "...") {
		t.Error("長名稱應被截斷")
	}
}
