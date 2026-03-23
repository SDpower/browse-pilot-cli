package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// Formatter 負責將結果以人類可讀或 JSON 格式輸出
type Formatter struct {
	writer   io.Writer
	jsonMode bool
	verbose  bool
}

// NewFormatter 建立新的 Formatter
func NewFormatter(jsonMode, verbose bool) *Formatter {
	return &Formatter{
		writer:   os.Stdout,
		jsonMode: jsonMode,
		verbose:  verbose,
	}
}

// SetWriter 設定輸出目標（主要供測試使用）
func (f *Formatter) SetWriter(w io.Writer) {
	f.writer = w
}

// IsJSON 回傳是否為 JSON 模式
func (f *Formatter) IsJSON() bool {
	return f.jsonMode
}

// PrintJSON 輸出 JSON（不論模式都以 JSON 輸出）
func (f *Formatter) PrintJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 序列化失敗: %w", err)
	}
	fmt.Fprintln(f.writer, string(data))
	return nil
}

// PrintResult 根據模式輸出結果
// jsonMode: 輸出 JSON
// 否則: 輸出 human-readable 文字
func (f *Formatter) PrintResult(result any) error {
	if f.jsonMode {
		return f.PrintJSON(result)
	}
	fmt.Fprintln(f.writer, result)
	return nil
}

// PrintSuccess 輸出成功訊息（綠色）
func (f *Formatter) PrintSuccess(format string, args ...any) {
	if f.jsonMode {
		f.PrintJSON(map[string]any{"success": true, "message": fmt.Sprintf(format, args...)}) //nolint:errcheck // PrintSuccess 本身不回傳 error，JSON 寫入失敗無法向上傳遞
		return
	}
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Fprintln(f.writer, green("✓"), fmt.Sprintf(format, args...))
}

// PrintError 輸出錯誤訊息（紅色）
func (f *Formatter) PrintError(format string, args ...any) {
	if f.jsonMode {
		f.PrintJSON(map[string]any{"error": fmt.Sprintf(format, args...)}) //nolint:errcheck // PrintError 本身不回傳 error，JSON 寫入失敗無法向上傳遞
		return
	}
	red := color.New(color.FgRed).SprintFunc()
	fmt.Fprintln(f.writer, red("✗"), fmt.Sprintf(format, args...))
}

// PrintWarning 輸出警告訊息（黃色）
func (f *Formatter) PrintWarning(format string, args ...any) {
	if f.jsonMode {
		// JSON 模式不輸出警告
		return
	}
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Fprintln(f.writer, yellow("⚠"), fmt.Sprintf(format, args...))
}

// PrintInfo 輸出一般資訊
func (f *Formatter) PrintInfo(format string, args ...any) {
	if f.jsonMode {
		return
	}
	fmt.Fprintln(f.writer, fmt.Sprintf(format, args...))
}

// PrintVerbose 只在 verbose 模式輸出（灰色）
func (f *Formatter) PrintVerbose(format string, args ...any) {
	if !f.verbose || f.jsonMode {
		return
	}
	dim := color.New(color.FgHiBlack).SprintFunc()
	fmt.Fprintln(f.writer, dim(fmt.Sprintf("[verbose] "+format, args...)))
}

// PrintKeyValue 輸出 key-value 對（label 用 cyan）
func (f *Formatter) PrintKeyValue(key string, value any) {
	if f.jsonMode {
		return
	}
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Fprintf(f.writer, "%s %v\n", cyan(key+":"), value)
}
