package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/jiazhoulvke/goime/internal/protocol"
)

func main() {
	socketFlag := flag.String("s", "", "Unix socket 路径（默认自动查找）")
	jsonOutput := flag.Bool("json", false, "JSON 格式输出（默认人类可读文本）")
	selectIdx := flag.Int("select", -1, "选词索引（-1 表示只查看不选）")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}
	input := args[0]

	// 连接 socket
	conn := dialSocket(*socketFlag)
	defer conn.Close()

	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)

	// 1. 握手
	send(enc, protocol.Request{Method: "hello", Version: 1, Client: "goimec-0.1"})
	var welcome protocol.Response
	if err := dec.Decode(&welcome); err != nil {
		fmt.Fprintf(os.Stderr, "Error: handshake failed: %v\n", err)
		os.Exit(1)
	}

	// 2. 逐个发送按键
	sent := 0
	for _, ch := range input {
		if ch < 'a' || ch > 'z' {
			continue
		}
		send(enc, protocol.Request{Method: "input", Key: string(ch)})
		sent++
	}

	// 3. 读取最终响应
	var resp protocol.Response
	for i := 0; i < sent; i++ {
		if err := dec.Decode(&resp); err != nil {
			fmt.Fprintf(os.Stderr, "Error: read response: %v\n", err)
			os.Exit(1)
		}
	}

	// 4. 选词
	if *selectIdx >= 0 && resp.Candidates != nil && *selectIdx < len(resp.Candidates.List) {
		send(enc, protocol.Request{Method: "select", Index: *selectIdx})
		var commitResp protocol.Response
		dec.Decode(&commitResp)
		resp = commitResp
	}

	// 5. 输出
	if *jsonOutput {
		outputJSON(resp)
	} else {
		outputText(resp, input, *selectIdx)
	}
}

// dialSocket 连接 goimed 的 Unix socket
// 优先级：-s 参数 > 配置文件的 socket_path > 自动尝试候选路径
func dialSocket(socketFlag string) net.Conn {
	// 1. -s 参数优先级最高
	if socketFlag != "" {
		conn, err := net.Dial("unix", socketFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: connect to %s: %v\n", socketFlag, err)
			fmt.Fprintln(os.Stderr, "Is goimed running?")
			os.Exit(1)
		}
		return conn
	}

	// 2. 读取配置文件
	cfgPath := configPath()
	if sock := socketFromConfig(cfgPath); sock != "" {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			return conn
		}
	}

	// 3. 自动尝试候选路径
	for _, candidate := range candidatePaths() {
		conn, err := net.Dial("unix", candidate)
		if err == nil {
			return conn
		}
	}

	fmt.Fprintf(os.Stderr, "Error: cannot connect to goimed\n")
	fmt.Fprintln(os.Stderr, "Tried: ")
	for _, p := range candidatePaths() {
		fmt.Fprintf(os.Stderr, "  %s\n", p)
	}
	if s := socketFromConfig(configPath()); s != "" {
		fmt.Fprintf(os.Stderr, "  (config) %s\n", s)
	}
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Is goimed running? Start with: goimed")
	os.Exit(1)
	return nil
}

// configPath 返回 GoIME 配置文件路径
func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "goime", "goime.toml")
}

// socketFromConfig 从配置文件中读取 socket_path
func socketFromConfig(path string) string {
	if path == "" {
		return ""
	}
	var cfg struct {
		General struct {
			SocketPath string `toml:"socket_path"`
		} `toml:"general"`
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return ""
	}
	return cfg.General.SocketPath
}

// candidatePaths 返回所有可能的 socket 候选路径
func candidatePaths() []string {
	var paths []string
	// 当前目录的 .goime.sock（项目内常用）
	if _, err := os.Stat(".goime.sock"); err == nil {
		abs, _ := filepath.Abs(".goime.sock")
		paths = append(paths, abs)
	}
	// XDG_RUNTIME_DIR
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir != "" {
		paths = append(paths, filepath.Join(runtimeDir, "goime.sock"))
	}
	// $TMPDIR/goime-$UID.sock (Termux 等)
	if tmpDir := os.Getenv("TMPDIR"); tmpDir != "" {
		paths = append(paths, filepath.Join(tmpDir, "goime-"+strconv.Itoa(os.Getuid())+".sock"))
	}
	// /tmp/goime-$UID.sock
	paths = append(paths, filepath.Join("/tmp", "goime-"+strconv.Itoa(os.Getuid())+".sock"))
	return paths
}

func send(enc *json.Encoder, req protocol.Request) {
	enc.Encode(req)
}

func outputJSON(resp protocol.Response) {
	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}

func outputText(resp protocol.Response, input string, selected int) {
	switch resp.Type {
	case "preedit":
		fmt.Printf("输入: %s\n", resp.Text)
		if resp.Candidates != nil {
			printCandidates(resp.Candidates)
		}

	case "candidates":
		printCandidates(resp.Candidates)

	case "commit":
		fmt.Printf("上屏: %s\n", resp.Text)
		if resp.PendingKey != "" {
			fmt.Printf("透传: %s\n", resp.PendingKey)
		}

	case "idle":
		fmt.Println("(空)")

	case "welcome":
		fmt.Printf("已连接: 版本=%d, 方案=%v, 当前=%s, page_size=%d\n",
			resp.Version, resp.Schemes, resp.Active, resp.PageSize)

	case "error":
		fmt.Printf("错误: %s\n", resp.Message)

	default:
		data, _ := json.Marshal(resp)
		fmt.Println(string(data))
	}
}

func printCandidates(c *protocol.Candidates) {
	if c == nil || len(c.List) == 0 {
		fmt.Println("(无候选词)")
		return
	}
	pageSize := len(c.List)
	fmt.Printf("候选词 (页 %d/%d, 每页 %d):\n", c.Page+1, c.Total, pageSize)
	for i, cand := range c.List {
		marker := " "
		if c.Page == 0 && i == 0 {
			marker = "→"
		}
		fmt.Printf("  %s[%d] %-6s %s (权重:%d)\n", marker, i, cand.Text, cand.Code, cand.Weight)
	}
}

func printUsage() {
	fmt.Println(`Usage: goimec [flags] <input>

模拟用户输入拼音，从 goimed 获取候选词。

Flags:
  -s string    Unix socket 路径（默认自动查找）
  -json        JSON 格式输出
  -select int  选词索引（-1 表示只查看不选）

连接优先级:
  1. -s 参数
  2. ~/.config/goime/goime.toml 中的 socket_path
  3. 当前目录的 .goime.sock
  4. $XDG_RUNTIME_DIR/goime.sock
  5. /tmp/goime-$UID.sock`)
}
