package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPPort != 8099 {
		t.Fatalf("expected 8099, got %d", cfg.HTTPPort)
	}
	if filepath.Base(cfg.DataDir) != "data" {
		t.Fatalf("unexpected data dir %s", cfg.DataDir)
	}
}

func TestLoadEnvPort(t *testing.T) {
	t.Setenv("RASGUI_HTTP_PORT", "9011")
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPPort != 9011 {
		t.Fatalf("expected 9011, got %d", cfg.HTTPPort)
	}
	_, _ = os.Stat(cfg.DataDir)
}
