package python

import (
	"context"
	"os/exec"
	"testing"
)

// hasPython 檢查系統是否有 python3 可執行檔
func hasPython() bool {
	_, err := exec.LookPath("python3")
	return err == nil
}

func TestNewSession(t *testing.T) {
	if !hasPython() {
		t.Skip("python3 不可用，跳過測試")
	}

	s, err := NewSession("")
	if err != nil {
		t.Fatalf("建立 session 失敗: %v", err)
	}
	defer s.Close()

	if !s.IsRunning() {
		t.Error("session 應為 running 狀態")
	}
}

func TestSessionExecute(t *testing.T) {
	if !hasPython() {
		t.Skip("python3 不可用，跳過測試")
	}

	s, err := NewSession("")
	if err != nil {
		t.Fatalf("建立 session 失敗: %v", err)
	}
	defer s.Close()

	result, err := s.Execute(context.Background(), "1 + 1")
	if err != nil {
		t.Fatalf("執行失敗: %v", err)
	}
	if result == nil {
		t.Fatal("結果不應為 nil")
	}
}

func TestSessionExecuteMultiple(t *testing.T) {
	if !hasPython() {
		t.Skip("python3 不可用，跳過測試")
	}

	s, err := NewSession("")
	if err != nil {
		t.Fatalf("建立 session 失敗: %v", err)
	}
	defer s.Close()

	// 第一次執行：定義變數
	_, err = s.Execute(context.Background(), "x = 42")
	if err != nil {
		t.Fatalf("第一次執行失敗: %v", err)
	}

	// 第二次執行：讀取變數（驗證 session 持久性）
	result, err := s.Execute(context.Background(), "x")
	if err != nil {
		t.Fatalf("第二次執行失敗: %v", err)
	}
	if result == nil {
		t.Fatal("第二次執行結果不應為 nil")
	}
}

func TestSessionExecuteError(t *testing.T) {
	if !hasPython() {
		t.Skip("python3 不可用，跳過測試")
	}

	s, err := NewSession("")
	if err != nil {
		t.Fatalf("建立 session 失敗: %v", err)
	}
	defer s.Close()

	// 執行會拋出例外的程式碼
	_, err = s.Execute(context.Background(), "raise ValueError('測試錯誤')")
	if err == nil {
		t.Fatal("應回傳錯誤，但沒有")
	}
}

func TestSessionClose(t *testing.T) {
	if !hasPython() {
		t.Skip("python3 不可用，跳過測試")
	}

	s, err := NewSession("")
	if err != nil {
		t.Fatalf("建立 session 失敗: %v", err)
	}

	s.Close()
	if s.IsRunning() {
		t.Error("關閉後 session 不應為 running 狀態")
	}
}

func TestSessionExecuteAfterClose(t *testing.T) {
	if !hasPython() {
		t.Skip("python3 不可用，跳過測試")
	}

	s, err := NewSession("")
	if err != nil {
		t.Fatalf("建立 session 失敗: %v", err)
	}

	s.Close()

	// 關閉後執行應回傳錯誤
	_, err = s.Execute(context.Background(), "1 + 1")
	if err == nil {
		t.Fatal("關閉後執行應回傳錯誤")
	}
}

func TestSessionGetVars(t *testing.T) {
	if !hasPython() {
		t.Skip("python3 不可用，跳過測試")
	}

	s, err := NewSession("")
	if err != nil {
		t.Fatalf("建立 session 失敗: %v", err)
	}
	defer s.Close()

	// 定義變數
	_, err = s.Execute(context.Background(), "my_var = 'hello'")
	if err != nil {
		t.Fatalf("執行失敗: %v", err)
	}

	vars, err := s.GetVars(context.Background())
	if err != nil {
		t.Fatalf("GetVars 失敗: %v", err)
	}

	if _, ok := vars["my_var"]; !ok {
		t.Error("應包含 my_var 變數")
	}
}

func TestSessionCustomPythonPath(t *testing.T) {
	if !hasPython() {
		t.Skip("python3 不可用，跳過測試")
	}

	// 使用完整路徑
	path, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("無法取得 python3 路徑")
	}

	s, err := NewSession(path)
	if err != nil {
		t.Fatalf("以自訂路徑建立 session 失敗: %v", err)
	}
	defer s.Close()

	if !s.IsRunning() {
		t.Error("session 應為 running 狀態")
	}
}
