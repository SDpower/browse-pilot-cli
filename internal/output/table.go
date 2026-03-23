package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

// StateElement 代表 get_state 回傳的單一可互動元素
type StateElement struct {
	Index       int    `json:"index"`
	Tag         string `json:"tag"`
	Type        string `json:"type,omitempty"`
	Role        string `json:"role,omitempty"`
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Selector    string `json:"selector"`
	Visible     bool   `json:"visible"`
}

// StateResult 代表 get_state 的完整回傳結果
type StateResult struct {
	URL      string         `json:"url"`
	Title    string         `json:"title"`
	Browser  string         `json:"browser,omitempty"`
	Elements []StateElement `json:"elements"`
}

// PrintState 以人類可讀的表格格式輸出 state
func PrintState(w io.Writer, state *StateResult, jsonMode bool) error {
	if jsonMode {
		return printJSON(w, state)
	}

	// 頁面資訊
	cyan := color.New(color.FgCyan).SprintFunc()
	if state.Browser != "" {
		fmt.Fprintf(w, "%s %s\n", cyan("Browser:"), state.Browser)
	}
	fmt.Fprintf(w, "%s %s\n", cyan("URL:"), state.URL)
	fmt.Fprintf(w, "%s %s\n", cyan("Title:"), state.Title)
	fmt.Fprintln(w)

	if len(state.Elements) == 0 {
		fmt.Fprintln(w, "（無可互動元素）")
		return nil
	}

	// 元素列表（簡潔格式，類似 browser-use-cli 風格）
	for _, el := range state.Elements {
		line := formatElement(el)
		fmt.Fprintln(w, line)
	}

	fmt.Fprintf(w, "\n共 %d 個可互動元素\n", len(state.Elements))
	return nil
}

// formatElement 格式化單一元素為一行
// 格式: [index] tag "name" value="..." placeholder="..."
func formatElement(el StateElement) string {
	var parts []string

	// 索引（用方括號框起）
	indexStr := fmt.Sprintf("[%d]", el.Index)

	// tag + type
	tagStr := el.Tag
	if el.Type != "" {
		tagStr += fmt.Sprintf("[%s]", el.Type)
	}
	if el.Role != "" {
		tagStr += fmt.Sprintf("(%s)", el.Role)
	}
	parts = append(parts, indexStr, tagStr)

	// name（截斷至 40 字元）
	if el.Name != "" {
		name := el.Name
		if len(name) > 40 {
			name = name[:37] + "..."
		}
		parts = append(parts, fmt.Sprintf(`"%s"`, name))
	}

	// value（如果有）
	if el.Value != "" {
		val := el.Value
		if len(val) > 30 {
			val = val[:27] + "..."
		}
		parts = append(parts, fmt.Sprintf("value=%q", val))
	}

	// placeholder（如果有且 name 為空）
	if el.Placeholder != "" && el.Name == "" {
		parts = append(parts, fmt.Sprintf("placeholder=%q", el.Placeholder))
	}

	return strings.Join(parts, " ")
}

// printJSON 將任意值序列化為縮排 JSON 並輸出至 writer
func printJSON(w io.Writer, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(data))
	return nil
}
