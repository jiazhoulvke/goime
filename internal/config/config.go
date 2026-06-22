package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/BurntSushi/toml"
)

// Config 主配置结构
type Config struct {
	General    GeneralConfig    `toml:"general"`
	Scheme     SchemeConfig     `toml:"scheme"`
	Dict       DictConfig       `toml:"dict"`
	Candidates CandidatesConfig `toml:"candidates"`
	Translator TranslatorConfig `toml:"translator"`
	UserDict   UserDictConfig   `toml:"user_dict"`
}

// GeneralConfig 通用配置
type GeneralConfig struct {
	LogLevel    string `toml:"log_level"`    // 日志级别: debug/info/warn/error
	SocketPath  string `toml:"socket_path"`  // Unix socket 路径，留空自动推导
	IdleTimeout string `toml:"idle_timeout"` // 空闲超时，如 "15m"
	Listen      string `toml:"listen"`       // "unix" 或 "tcp"，默认值平台相关
	Host        string `toml:"host"`         // TCP 主机地址（仅 TCP 模式有效）
	Port        int    `toml:"port"`         // TCP 端口（仅 TCP 模式有效）；0=随机端口
}

// SchemeConfig 输入方案配置
type SchemeConfig struct {
	Active string `toml:"active"` // 默认输入方案
	Dir    string `toml:"dir"`    // 方案文件目录
}

// DictConfig 词典配置
type DictConfig struct {
	Static    []string `toml:"static"`     // 静态词库文件列表
	User      string   `toml:"user"`       // 用户词库 SQLite 路径
	SyncFile  string   `toml:"sync_file"`  // 用户词库纯文本同步文件
	BuildDir  string   `toml:"build_dir"`  // 二进制索引构建目录
	AutoBuild bool     `toml:"auto_build"` // 自动构建开关
}

// CandidatesConfig 候选词配置
type CandidatesConfig struct {
	PageSize      int `toml:"page_size"`      // 每页候选数
	MaxCandidates int `toml:"max_candidates"` // 最大候选数
}

// TranslatorConfig 翻译器配置
type TranslatorConfig struct {
	MaxSyllables int `toml:"max_syllables"` // 最大匹配音节数
}

// UserDictConfig 用户词库配置
type UserDictConfig struct {
	Enabled       bool    `toml:"enabled"`         // 启用用户词库
	FreqDecay     bool    `toml:"freq_decay"`      // 启用词频衰减
	DecayRate     float64 `toml:"decay_rate"`      // 衰减率
	NewWordWeight int     `toml:"new_word_weight"` // 新自造词初始权重
}

// defaultListen 返回平台相关的默认监听类型。
func defaultListen() string {
	if runtime.GOOS == "windows" {
		return "tcp"
	}
	return "unix"
}

// Default 返回带默认值的配置
func Default() *Config {
	home := homeDir()
	return &Config{
		General: GeneralConfig{
			LogLevel:    "info",
			SocketPath:  "",
			IdleTimeout: "15m",
			Listen:      defaultListen(),
			Host:        "127.0.0.1",
			Port:        11527,
		},
		Scheme: SchemeConfig{
			Active: "xiaohe",
			Dir:    filepath.Join(home, ".config", "goime", "schemes"),
		},
		Dict: DictConfig{
			Static:    []string{filepath.Join(home, ".config", "goime", "dicts", "zhonghua.dict.txt")},
			User:      filepath.Join(home, ".config", "goime", "user_dict.db"),
			SyncFile:  filepath.Join(home, ".config", "goime", "user_dict.txt"),
			BuildDir:  filepath.Join(home, ".cache", "goime"),
			AutoBuild: true,
		},
		Candidates: CandidatesConfig{
			PageSize:      5,
			MaxCandidates: 100,
		},
		Translator: TranslatorConfig{
			MaxSyllables: 8,
		},
		UserDict: UserDictConfig{
			Enabled:       true,
			FreqDecay:     true,
			DecayRate:     0.99,
			NewWordWeight: 100,
		},
	}
}

// homeDir 获取用户主目录，失败时返回空字符串
func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// ExpandPath 展开路径中的 ~ 为用户主目录
func ExpandPath(path string) string {
	if path == "" {
		return path
	}
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

// Load 从 TOML 文件加载配置
func Load(path string) (*Config, error) {
	cfg := Default()
	_, err := toml.DecodeFile(path, cfg)
	if err != nil {
		return nil, fmt.Errorf("load config file: %w", err)
	}
	return cfg, nil
}

// loadBytes 从 TOML 字节数据加载配置（内部测试用）
func loadBytes(data []byte) (*Config, error) {
	cfg := Default()
	err := toml.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// SocketPath 返回 Unix socket 路径
func (c *Config) SocketPath() string {
	if c.General.SocketPath != "" {
		return c.General.SocketPath
	}
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir != "" {
		return runtimeDir + "/goime.sock"
	}
	// Termux 等环境使用 $TMPDIR
	if tmpDir := os.Getenv("TMPDIR"); tmpDir != "" {
		return tmpDir + "/goime-" + strconv.Itoa(os.Getuid()) + ".sock"
	}
	return "/tmp/goime-" + strconv.Itoa(os.Getuid()) + ".sock"
}
