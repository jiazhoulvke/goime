package server

import (
	"encoding/json"
	"net"
	"os"
	"testing"
	"time"

	"github.com/jiazhoulvke/goime/internal/config"
	"github.com/jiazhoulvke/goime/internal/protocol"
)

func TestServerHandshake(t *testing.T) {
	dir, err := os.MkdirTemp("", "goime-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	socketPath := dir + "/goime.sock"
	cfg := config.Default()
	cfg.General.Listen = "unix"
	cfg.General.SocketPath = socketPath

	srv, err := New(cfg, nil, nil, []string{"xiaohe", "fullpin"})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	go srv.Listen()
	defer srv.Close()
	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	req := protocol.Request{Method: "hello", Version: 1, Client: "test"}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	var resp protocol.Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if resp.Type != "welcome" {
		t.Errorf("expected welcome, got %s", resp.Type)
	}
}

func TestServerHandshakeTCP(t *testing.T) {
	cfg := config.Default()
	cfg.General.Listen = "tcp"
	cfg.General.Host = "127.0.0.1"
	cfg.General.Port = 0

	srv, err := New(cfg, nil, nil, []string{"xiaohe", "fullpin"})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	go srv.Listen()
	defer srv.Close()
	time.Sleep(100 * time.Millisecond)

	// Server should be listening on TCP — find the addr
	// (transport.Listen returns it, but Server doesn't expose it yet)
	// We'll read the listener address from the server's internal listener
	srv.mu.Lock()
	lnAddr := srv.ln.Addr().String()
	srv.mu.Unlock()

	conn, err := net.Dial("tcp", lnAddr)
	if err != nil {
		t.Fatalf("Dial TCP %s failed: %v", lnAddr, err)
	}
	defer conn.Close()

	req := protocol.Request{Method: "hello", Version: 1, Client: "test"}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	var resp protocol.Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if resp.Type != "welcome" {
		t.Errorf("expected welcome, got %s", resp.Type)
	}
}

func TestServerInputTCP(t *testing.T) {
	cfg := config.Default()
	cfg.General.Listen = "tcp"
	cfg.General.Host = "127.0.0.1"
	cfg.General.Port = 0

	srv, err := New(cfg, nil, nil, []string{"xiaohe"})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	go srv.Listen()
	defer srv.Close()
	time.Sleep(100 * time.Millisecond)

	srv.mu.Lock()
	lnAddr := srv.ln.Addr().String()
	srv.mu.Unlock()

	conn, err := net.Dial("tcp", lnAddr)
	if err != nil {
		t.Fatalf("Dial TCP %s failed: %v", lnAddr, err)
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

	resp := send(protocol.Request{Method: "input", Key: "a"})
	if resp.Type != "preedit" {
		t.Fatalf("expected preedit, got %s", resp.Type)
	}
	resp = send(protocol.Request{Method: "enter"})
	if resp.Type != "commit" || resp.Text != "a" {
		t.Fatalf("expected commit 'a', got %s %q", resp.Type, resp.Text)
	}
}

func TestServerInput(t *testing.T) {
	dir, err := os.MkdirTemp("", "goime-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	socketPath := dir + "/goime.sock"
	cfg := config.Default()
	cfg.General.Listen = "unix"
	cfg.General.SocketPath = socketPath

	srv, err := New(cfg, nil, nil, []string{"xiaohe"})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	go srv.Listen()
	defer srv.Close()
	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
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

	// Simple input + enter
	resp := send(protocol.Request{Method: "input", Key: "a"})
	if resp.Type != "preedit" {
		t.Fatalf("expected preedit, got %s", resp.Type)
	}
	resp = send(protocol.Request{Method: "enter"})
	if resp.Type != "commit" || resp.Text != "a" {
		t.Fatalf("expected commit 'a', got %s %q", resp.Type, resp.Text)
	}
}
