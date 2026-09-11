package cli

import "testing"

func TestRootCommandHasVersion(t *testing.T) {
	if rootCmd.Version != cliVersion || rootCmd.Version == "" {
		t.Errorf("bp_cli version 未正確設定: %q", rootCmd.Version)
	}
}
