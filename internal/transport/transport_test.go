package transport

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jiazhoulvke/goime/internal/config"
)

func TestDefaultSocketPath(t *testing.T) {
	path := DefaultSocketPath()
	if path == "" {
		t.Fatal("DefaultSocketPath() returned empty")
	}
}

func TestPortFilePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".cache", "goime", "goime.port")
	got := PortFilePath()
	if got != want {
		t.Errorf("PortFilePath() = %q, want %q", got, want)
	}
}

func TestWriteReadPortFile(t *testing.T) {
	// Use a temp dir to avoid cluttering real cache
	origHome := os.Getenv("HOME")
	origUserHome := os.Getenv("USERPROFILE") // Windows
	home := t.TempDir()
	os.Setenv("HOME", home)
	os.Setenv("USERPROFILE", home)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserHome)
	}()

	port := 12345
	if err := WritePortFile(port); err != nil {
		t.Fatalf("WritePortFile(%d) failed: %v", port, err)
	}

	got, err := ReadPortFile()
	if err != nil {
		t.Fatalf("ReadPortFile() failed: %v", err)
	}
	if got != port {
		t.Errorf("ReadPortFile() = %d, want %d", got, port)
	}
}

func TestListenTCP(t *testing.T) {
	cfg := config.Default()
	cfg.General.Listen = "tcp"
	cfg.General.Host = "127.0.0.1"
	cfg.General.Port = 0 // random port

	ln, addr, err := Listen(cfg)
	if err != nil {
		t.Fatalf("Listen TCP failed: %v", err)
	}
	defer ln.Close()

	if addr == "" {
		t.Fatal("Listen TCP returned empty addr")
	}

	// Verify we can connect
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Dial TCP %s failed: %v", addr, err)
	}
	conn.Close()
}

func TestListenUnix(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "goime.sock")
	cfg := config.Default()
	cfg.General.Listen = "unix"
	cfg.General.SocketPath = socketPath

	ln, addr, err := Listen(cfg)
	if err != nil {
		t.Fatalf("Listen Unix failed: %v", err)
	}
	defer ln.Close()

	if addr != socketPath {
		t.Errorf("Listen Unix addr = %q, want %q", addr, socketPath)
	}

	// Verify we can connect
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Dial Unix %s failed: %v", socketPath, err)
	}
	conn.Close()
}

func TestWritePortFileCreatesDir(t *testing.T) {
	// Override home to a deep temp path
	origHome := os.Getenv("HOME")
	origUserHome := os.Getenv("USERPROFILE")
	home := filepath.Join(t.TempDir(), "deep", "nested")
	os.Setenv("HOME", home)
	os.Setenv("USERPROFILE", home)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserHome)
	}()

	// PortFilePath should create the directory tree
	if err := WritePortFile(9999); err != nil {
		t.Fatalf("WritePortFile in deep dir failed: %v", err)
	}

	port, err := ReadPortFile()
	if err != nil {
		t.Fatalf("ReadPortFile failed: %v", err)
	}
	if port != 9999 {
		t.Errorf("got port %d, want 9999", port)
	}
}

func TestReadPortFileNotFound(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserHome := os.Getenv("USERPROFILE")
	os.Setenv("HOME", t.TempDir())
	os.Setenv("USERPROFILE", t.TempDir())
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserHome)
	}()

	_, err := ReadPortFile()
	if err == nil {
		t.Fatal("expected error for missing port file")
	}
}

func TestListenTCPPortZero(t *testing.T) {
	// Test that port 0 gets a random port
	cfg := config.Default()
	cfg.General.Listen = "tcp"
	cfg.General.Host = "127.0.0.1"
	cfg.General.Port = 0

	ln1, addr1, err := Listen(cfg)
	if err != nil {
		t.Fatalf("first Listen: %v", err)
	}
	defer ln1.Close()

	ln2, addr2, err := Listen(cfg)
	if err != nil {
		t.Fatalf("second Listen: %v", err)
	}
	defer ln2.Close()

	_, port1Str, _ := net.SplitHostPort(addr1)
	_, port2Str, _ := net.SplitHostPort(addr2)
	port1, _ := strconv.Atoi(port1Str)
	port2, _ := strconv.Atoi(port2Str)

	if port1 == port2 {
		t.Errorf("two listeners with port=0 got same port %d", port1)
	}
}
