package cli

import "testing"

func TestRootCommandHasVersion(t *testing.T) {
	if rootCmd.Version != cliVersion || rootCmd.Version == "" {
		t.Errorf("bp_cli version 未正確設定: %q", rootCmd.Version)
	}
}

func TestRootCommandUsesHTTPMCPOnly(t *testing.T) {
	if rootCmd.PersistentFlags().Lookup("mcp-http") == nil {
		t.Fatal("缺少 --mcp-http flag")
	}
	if rootCmd.PersistentFlags().Lookup("mcp-port") == nil {
		t.Fatal("缺少 --mcp-port flag")
	}
	if rootCmd.PersistentFlags().Lookup("mcp") != nil {
		t.Fatal("舊的 stdio --mcp flag 應移除")
	}
	if GetMCPPort() != 8931 {
		t.Fatalf("預設 MCP HTTP 埠號應為 8931，實際為 %d", GetMCPPort())
	}
}
