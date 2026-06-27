package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigOptional_ClaudeToolRename(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	configYAML := []byte(`
claude-tool-rename:
  call_omo_agent: callOmoAgent
  some_tool: SomeTool
`)
	if err := os.WriteFile(configPath, configYAML, 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := LoadConfigOptional(configPath, false)
	if err != nil {
		t.Fatalf("LoadConfigOptional() error = %v", err)
	}

	if got := len(cfg.ClaudeToolRename); got != 2 {
		t.Fatalf("ClaudeToolRename len = %d, want 2", got)
	}
	if got := cfg.ClaudeToolRename["call_omo_agent"]; got != "callOmoAgent" {
		t.Fatalf("ClaudeToolRename[call_omo_agent] = %q, want %q", got, "callOmoAgent")
	}
	if got := cfg.ClaudeToolRename["some_tool"]; got != "SomeTool" {
		t.Fatalf("ClaudeToolRename[some_tool] = %q, want %q", got, "SomeTool")
	}
}
