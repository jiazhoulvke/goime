package config

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.General.LogLevel != "info" {
		t.Errorf("expected info, got %s", cfg.General.LogLevel)
	}
	if cfg.Scheme.Active != "xiaohe" {
		t.Errorf("expected xiaohe, got %s", cfg.Scheme.Active)
	}
	if cfg.Candidates.PageSize != 5 {
		t.Errorf("expected 5, got %d", cfg.Candidates.PageSize)
	}
}

func TestLoadConfig(t *testing.T) {
	data := `
[general]
log_level = "debug"
idle_timeout = "30m"

[scheme]
active = "fullpin"

[candidates]
page_size = 7
`
	cfg, err := loadBytes([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.General.LogLevel != "debug" {
		t.Errorf("expected debug, got %s", cfg.General.LogLevel)
	}
	if cfg.General.IdleTimeout != "30m" {
		t.Errorf("expected 30m, got %s", cfg.General.IdleTimeout)
	}
	if cfg.Scheme.Active != "fullpin" {
		t.Errorf("expected fullpin, got %s", cfg.Scheme.Active)
	}
	if cfg.Candidates.PageSize != 7 {
		t.Errorf("expected 7, got %d", cfg.Candidates.PageSize)
	}
	if cfg.SocketPath() == "" {
		t.Error("SocketPath() should not be empty")
	}
}
