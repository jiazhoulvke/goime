package config

import (
	"runtime"
	"testing"
)

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

func TestDefaultListenConfig(t *testing.T) {
	cfg := Default()

	// Listen 必须有值
	if cfg.General.Listen == "" {
		t.Error("default Listen should not be empty")
	}
	// 平台相关的默认值
	if runtime.GOOS == "windows" {
		if cfg.General.Listen != "tcp" {
			t.Errorf("on windows expected listen=tcp, got %s", cfg.General.Listen)
		}
	} else {
		if cfg.General.Listen != "unix" {
			t.Errorf("on unix expected listen=unix, got %s", cfg.General.Listen)
		}
	}
	if cfg.General.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.General.Host)
	}
	if cfg.General.Port != 11527 {
		t.Errorf("expected port 11527, got %d", cfg.General.Port)
	}
}

func TestLoadConfigWithTCP(t *testing.T) {
	data := `
[general]
listen = "tcp"
host = "0.0.0.0"
port = 9999
`
	cfg, err := loadBytes([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.General.Listen != "tcp" {
		t.Errorf("expected tcp, got %s", cfg.General.Listen)
	}
	if cfg.General.Host != "0.0.0.0" {
		t.Errorf("expected 0.0.0.0, got %s", cfg.General.Host)
	}
	if cfg.General.Port != 9999 {
		t.Errorf("expected 9999, got %d", cfg.General.Port)
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
