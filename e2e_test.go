// e2e_test.go — 端到端测试：启服务→发请求→校验候选项
// 运行：go test -v -run TestE2E ./...
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

func TestE2EPipeline(t *testing.T) {
	dir, err := os.MkdirTemp("", "goime-e2e-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// 1. 创建词库
	dictContent := "shu1ru4 输入 100\nni3hao3 你好 200\n"
	dictFile := filepath.Join(dir, "test.dict.txt")
	os.WriteFile(dictFile, []byte(dictContent), 0644)
	indexFile := filepath.Join(dir, "test.goime")
	dict.Build(dictFile, indexFile)
	idx, _ := dict.Load(indexFile)

	// 2. 启动服务
	cfg := config.Default()
	cfg.General.Listen = "unix"
	cfg.General.SocketPath = filepath.Join(dir, "goime.sock")
	srv, _ := server.New(cfg, idx, nil, []string{"xiaohe", "fullpin"})
	go srv.Listen()
	defer srv.Close()
	time.Sleep(100 * time.Millisecond)

	// 3. 连接
	conn, err := net.Dial("unix", cfg.General.SocketPath)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	send := func(req protocol.Request) protocol.Response {
		json.NewEncoder(conn).Encode(req)
		var resp protocol.Response
		json.NewDecoder(conn).Decode(&resp)
		return resp
	}

	// 4. 握手
	resp := send(protocol.Request{Method: "hello", Version: 1, Client: "test"})
	if resp.Type != "welcome" {
		t.Fatalf("expected welcome, got %s", resp.Type)
	}

	// 5. 输入小鹤编码 "uuru"（= shu ru = 输入）
	resp = send(protocol.Request{Method: "input", Key: "u"})
	resp = send(protocol.Request{Method: "input", Key: "u"})
	resp = send(protocol.Request{Method: "input", Key: "r"})
	resp = send(protocol.Request{Method: "input", Key: "u"})

	if resp.Type != "preedit" {
		t.Fatalf("expected preedit, got %s", resp.Type)
	}
	if resp.Candidates == nil {
		t.Fatal("expected candidates, got nil")
	}
	if len(resp.Candidates.List) == 0 {
		t.Fatal("expected non-empty candidates list")
	}

	t.Logf("preedit: %s", resp.Text)
	for i, c := range resp.Candidates.List {
		t.Logf("  candidate[%d]: %s (code=%s, weight=%d)", i, c.Text, c.Code, c.Weight)
	}

	// 6. 选第一个候选项
	resp = send(protocol.Request{Method: "select", Index: 0})
	if resp.Type != "commit" {
		t.Fatalf("expected commit, got %s", resp.Type)
	}
	t.Logf("commit: %s", resp.Text)
	if resp.Text != "输入" {
		t.Errorf("expected '输入', got '%s'", resp.Text)
	}
}
