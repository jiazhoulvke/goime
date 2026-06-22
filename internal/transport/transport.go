package transport

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jiazhoulvke/goime/internal/config"
)

// Listen 根据配置创建监听器。
// 返回 (listener, 监听地址, error)。
// Unix 模式下监听地址为 socket 路径，TCP 模式下为 "host:port"。
func Listen(cfg *config.Config) (net.Listener, string, error) {
	if cfg.General.Listen == "tcp" {
		return listenTCP(cfg)
	}
	return listenUnix(cfg)
}

// listenTCP 创建 TCP 监听器
func listenTCP(cfg *config.Config) (net.Listener, string, error) {
	addr := net.JoinHostPort(cfg.General.Host, strconv.Itoa(cfg.General.Port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, "", fmt.Errorf("listen tcp %s: %w", addr, err)
	}
	return ln, ln.Addr().String(), nil
}

// listenUnix 创建 Unix socket 监听器
func listenUnix(cfg *config.Config) (net.Listener, string, error) {
	socketPath := cfg.SocketPath()
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return nil, "", fmt.Errorf("remove existing socket: %w", err)
	}
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, "", fmt.Errorf("listen unix %s: %w", socketPath, err)
	}
	if err := os.Chmod(socketPath, 0600); err != nil {
		ln.Close()
		return nil, "", fmt.Errorf("chmod socket: %w", err)
	}
	return ln, socketPath, nil
}

// DefaultSocketPath 返回默认的 Unix socket 路径。
func DefaultSocketPath() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir != "" {
		return filepath.Join(runtimeDir, "goime.sock")
	}
	if tmpDir := os.Getenv("TMPDIR"); tmpDir != "" {
		return filepath.Join(tmpDir, "goime-"+strconv.Itoa(os.Getuid())+".sock")
	}
	return filepath.Join("/tmp", "goime-"+strconv.Itoa(os.Getuid())+".sock")
}

// PortFilePath 返回端口文件路径。
func PortFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
		if home == "" {
			home = "/tmp"
		}
	}
	return filepath.Join(home, ".cache", "goime", "goime.port")
}

// WritePortFile 将端口号写入端口文件。
func WritePortFile(port int) error {
	path := PortFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create port file dir: %w", err)
	}
	data := strconv.Itoa(port)
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		return fmt.Errorf("write port file: %w", err)
	}
	return nil
}

// ReadPortFile 从端口文件读取端口号。
func ReadPortFile() (int, error) {
	data, err := os.ReadFile(PortFilePath())
	if err != nil {
		return 0, fmt.Errorf("read port file: %w", err)
	}
	port, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("parse port: %w", err)
	}
	return port, nil
}

// RemovePortFile 删除端口文件（服务器关闭时调用）。
func RemovePortFile() error {
	return os.Remove(PortFilePath())
}

// RemoveSocket 删除 Unix socket 文件。
func RemoveSocket(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
