package goime

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jiazhoulvke/goime/internal/config"
	"github.com/jiazhoulvke/goime/internal/dict"
	"github.com/jiazhoulvke/goime/internal/protocol"
	"github.com/jiazhoulvke/goime/internal/server"
)

func TestIntegrationFullFlow(t *testing.T) {
	dir, err := os.MkdirTemp("", "goime-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	dictContent := "shu1ru4 输入 100\nni3hao3 你好 200\nshi4jie4 世界 150\n"
	dictFile := filepath.Join(dir, "test.dict.txt")
	if err := os.WriteFile(dictFile, []byte(dictContent), 0644); err != nil {
		t.Fatal(err)
	}
	indexFile := filepath.Join(dir, "test.goime")
	if err := dict.Build(dictFile, indexFile); err != nil {
		t.Fatalf("Build: %v", err)
	}
	idx, err := dict.Load(indexFile)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	userDB := filepath.Join(dir, "user.db")
	user, err := dict.OpenUserDict(userDB)
	if err != nil {
		t.Fatalf("OpenUserDict: %v", err)
	}
	defer user.Close()

	cfg := config.Default()
	cfg.General.Listen = "unix"
	cfg.General.SocketPath = filepath.Join(dir, "goime.sock")

	srv, err := server.New(cfg, idx, user, []string{"xiaohe", "fullpin"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	go srv.Listen()
	defer srv.Close()
	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("unix", cfg.General.SocketPath)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	send := func(req protocol.Request) protocol.Response {
		if err := json.NewEncoder(conn).Encode(req); err != nil {
			t.Fatalf("Encode: %v", err)
		}
		var resp protocol.Response
		if err := json.NewDecoder(conn).Decode(&resp); err != nil {
			t.Fatalf("Decode: %v", err)
		}
		return resp
	}

	// 1. Handshake
	resp := send(protocol.Request{Method: "hello", Version: 1, Client: "test"})
	if resp.Type != "welcome" {
		t.Fatalf("expected welcome, got %s", resp.Type)
	}

	// 2. Input + enter basic flow
	resp = send(protocol.Request{Method: "input", Key: "a"})
	if resp.Type != "preedit" {
		t.Fatalf("expected preedit, got %s", resp.Type)
	}
	resp = send(protocol.Request{Method: "enter"})
	if resp.Type != "commit" || resp.Text != "a" {
		t.Fatalf("expected commit 'a', got %s %q", resp.Type, resp.Text)
	}

	// 3. Escape empties buffer
	resp = send(protocol.Request{Method: "input", Key: "b"})
	if resp.Type != "preedit" {
		t.Fatalf("expected preedit, got %s", resp.Type)
	}
	resp = send(protocol.Request{Method: "escape"})
	if resp.Type != "idle" {
		t.Fatalf("expected idle, got %s", resp.Type)
	}

	// 4. Backspace
	resp = send(protocol.Request{Method: "input", Key: "c"})
	resp = send(protocol.Request{Method: "input", Key: "d"})
	if resp.Text != "cd" {
		t.Fatalf("expected preedit 'cd', got %s", resp.Text)
	}
	resp = send(protocol.Request{Method: "backspace"})
	if resp.Text != "c" {
		t.Fatalf("expected preedit 'c', got %s", resp.Text)
	}
	// Clear remaining buffer
	resp = send(protocol.Request{Method: "backspace"})
	if resp.Type != "idle" {
		t.Fatalf("expected idle after clearing buffer, got %s", resp.Type)
	}

	// 5. Space on empty buffer = idle
	resp = send(protocol.Request{Method: "space"})
	if resp.Type != "idle" {
		t.Fatalf("expected idle on empty buffer, got %s", resp.Type)
	}

	// 6. Commit preedit
	resp = send(protocol.Request{Method: "input", Key: "z"})
	if resp.Type != "preedit" {
		t.Fatalf("expected preedit, got %s", resp.Type)
	}
	resp = send(protocol.Request{Method: "commit_preedit"})
	if resp.Type != "commit" || resp.Text != "z" {
		t.Fatalf("expected commit 'z', got %s %q", resp.Type, resp.Text)
	}

	// 7. Idle state after commit_preedit
	resp = send(protocol.Request{Method: "space"})
	if resp.Type != "idle" {
		t.Fatalf("expected idle after commit_preedit, got %s", resp.Type)
	}
}
