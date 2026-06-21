package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jiazhoulvke/goime/internal/config"
	"github.com/jiazhoulvke/goime/internal/dict"
	"github.com/jiazhoulvke/goime/internal/server"
)

func main() {
	configPath := flag.String("config", "", "Path to config (default: ~/.config/goime/goime.toml)")
	flag.Parse()

	var cfg *config.Config
	if *configPath != "" {
		var err error
		cfg, err = config.Load(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		home, _ := os.UserHomeDir()
		cfgPath := filepath.Join(home, ".config", "goime", "goime.toml")
		var err error
		cfg, err = config.Load(cfgPath)
		if err != nil {
			cfg = config.Default()
		}
	}

	logLevel := &slog.LevelVar{}
	switch cfg.General.LogLevel {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelInfo)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))
	slog.Info("starting goimed", "socket", cfg.SocketPath())

	// 确保词库目录存在
	buildDir := config.ExpandPath(cfg.Dict.BuildDir)
	os.MkdirAll(buildDir, 0755)
	userDictDir := filepath.Dir(config.ExpandPath(cfg.Dict.User))
	os.MkdirAll(userDictDir, 0755)
	if len(cfg.Dict.Static) > 0 {
		staticDir := filepath.Dir(config.ExpandPath(cfg.Dict.Static[0]))
		os.MkdirAll(staticDir, 0755)
	}

	// 加载静态词库（合并所有）
	idx := dict.NewIndex()
	loaded := false
	for _, rawPath := range cfg.Dict.Static {
		dictPath := config.ExpandPath(rawPath)
		srcInfo, err1 := os.Stat(dictPath)
		if err1 != nil {
			// 源文件不存在时，尝试直接加载已编译的 .goime
			slog.Debug("dict source not found, trying pre-built", "path", dictPath)
			base := filepath.Base(dictPath)
			candidates := []string{
				filepath.Join(buildDir, base+".goime"),
				filepath.Join(buildDir, strings.TrimSuffix(base, filepath.Ext(base))+".goime"),
			}
			for _, gp := range candidates {
				gi, err := os.Stat(gp)
				if err != nil || gi.Size() == 0 {
					continue
				}
				sub, err := dict.Load(gp)
				if err == nil {
					idx.Merge(sub)
					loaded = true
					slog.Info("loaded dictionary", "path", gp)
					break
				}
			}
			continue
		}
		dstPath := filepath.Join(buildDir, strings.TrimSuffix(filepath.Base(dictPath), filepath.Ext(dictPath))+".goime")
		if cfg.Dict.AutoBuild {
			dstInfo, err2 := os.Stat(dstPath)
			if err2 != nil || srcInfo.ModTime().After(dstInfo.ModTime()) {
				slog.Info("building dictionary", "src", dictPath, "dst", dstPath)
				if err := dict.Build(dictPath, dstPath); err != nil {
					slog.Error("build failed", "error", err)
					os.Exit(1)
				}
			}
		}
		sub, err := dict.Load(dstPath)
		if err != nil {
			slog.Error("load failed", "error", err)
			os.Exit(1)
		}
		idx.Merge(sub)
		loaded = true
		slog.Info("loaded dictionary", "path", dstPath)
	}

	// 兜底：扫描 build 目录中的 .goime 文件
	if !loaded {
		entries, err := os.ReadDir(buildDir)
		if err == nil {
			type sizedFile struct {
				path string
				size int64
			}
			var candidates []sizedFile
			for _, entry := range entries {
				if entry.IsDir() || filepath.Ext(entry.Name()) != ".goime" {
					continue
				}
				info, _ := entry.Info()
				if info != nil && info.Size() > 0 {
					candidates = append(candidates, sizedFile{
						path: filepath.Join(buildDir, entry.Name()),
						size: info.Size(),
					})
				}
			}
			sort.Slice(candidates, func(i, j int) bool {
				return candidates[i].size > candidates[j].size
			})
			for _, cf := range candidates {
				sub, err := dict.Load(cf.path)
				if err == nil {
					idx.Merge(sub)
					loaded = true
					slog.Info("loaded dictionary", "path", cf.path, "size", cf.size)
				}
			}
		}
	}
	if !loaded {
		slog.Warn("no dictionary loaded, IME will not produce candidates")
	}

	// 打开用户词库
	var userDict *dict.UserDict
	if cfg.UserDict.Enabled {
		userDBPath := config.ExpandPath(cfg.Dict.User)
		var err error
		userDict, err = dict.OpenUserDict(userDBPath)
		if err != nil {
			slog.Error("open user dict failed", "error", err)
			os.Exit(1)
		}
		defer userDict.Close()
		if cfg.UserDict.FreqDecay {
			userDict.DecayAll(cfg.UserDict.DecayRate)
		}
	}

	schemes := []string{"xiaohe", "fullpin"}
	defer idx.Close()
	srv, err := server.New(cfg, idx, userDict, schemes)
	if err != nil {
		slog.Error("create server failed", "error", err)
		os.Exit(1)
	}
	if err := srv.Listen(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
