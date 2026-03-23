package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatterPrintSuccess(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(false, false)
	f.SetWriter(&buf)
	f.PrintSuccess("操作完成")
	if !strings.Contains(buf.String(), "操作完成") {
		t.Error("應包含成功訊息")
	}
}

func TestFormatterPrintSuccessJSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(true, false)
	f.SetWriter(&buf)
	f.PrintSuccess("操作完成")
	if !strings.Contains(buf.String(), `"success"`) {
		t.Error("JSON 模式應輸出 success 欄位")
	}
}

func TestFormatterPrintError(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(false, false)
	f.SetWriter(&buf)
	f.PrintError("連線失敗: %s", "timeout")
	if !strings.Contains(buf.String(), "連線失敗: timeout") {
		t.Error("應包含錯誤訊息")
	}
}

func TestFormatterPrintVerboseOff(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(false, false) // verbose=false
	f.SetWriter(&buf)
	f.PrintVerbose("除錯資訊")
	if buf.Len() != 0 {
		t.Error("非 verbose 模式不應有輸出")
	}
}

func TestFormatterPrintVerboseOn(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(false, true) // verbose=true
	f.SetWriter(&buf)
	f.PrintVerbose("除錯資訊")
	if !strings.Contains(buf.String(), "除錯資訊") {
		t.Error("verbose 模式應有輸出")
	}
}

func TestFormatterPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(true, false)
	f.SetWriter(&buf)
	err := f.PrintJSON(map[string]string{"key": "值"})
	if err != nil {
		t.Fatalf("PrintJSON 失敗: %v", err)
	}
	if !strings.Contains(buf.String(), `"key"`) {
		t.Error("應包含 JSON key")
	}
}
