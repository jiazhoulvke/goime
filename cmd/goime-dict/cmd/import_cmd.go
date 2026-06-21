package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jiazhoulvke/goime/internal/dict"
	"github.com/spf13/cobra"
)

var (
	importRime   bool
	importDir    string // .goime 输出目录
	importDicts string // .dict.txt 源文件目录
)

var importCmd = &cobra.Command{
	Use:   "import [flags] <files...>",
	Short: "导入词库并编译为 .goime（同时生成 .dict.txt 源文件）",
	Long: `导入一个或多个词库源文件，每个文件生成一对输出：
  • .dict.txt — GoIME 纯文本源文件（供 goimed 自动构建）
  • .goime    — 编译后的二进制索引

  goime-dict import --rime a.dict.yaml b.dict.yaml
  → ~/.config/goime/dicts/a.dict.txt
  → ~/.cache/goime/a.dict.goime

  goime-dict import --rime --dir ./build --dict-dir ./dicts *.dict.yaml`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runImport(args)
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().BoolVar(&importRime, "rime", false, "输入为 Rime .dict.yaml 格式")
	importCmd.Flags().StringVar(&importDir, "dir", defaultBuildDir(), ".goime 输出目录（默认 ~/.cache/goime/）")
	importCmd.Flags().StringVar(&importDicts, "dict-dir", defaultDictDir(), ".dict.txt 源文件目录（默认 ~/.config/goime/dicts/）")
}

// defaultBuildDir 返回默认构建目录 ~/.cache/goime/
func defaultBuildDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".cache", "goime")
}

// defaultDictDir 返回默认词库源文件目录 ~/.config/goime/dicts/
func defaultDictDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "goime", "dicts")
}

func runImport(srcFiles []string) error {
	os.MkdirAll(importDir, 0755)
	os.MkdirAll(importDicts, 0755)

	for _, src := range srcFiles {
		var entries []dict.Entry
		var err error

		if importRime {
			entries, err = dict.ParseRimeFile(src)
		} else {
			entries, err = parsePlainTextFile(src)
		}
		if err != nil {
			return fmt.Errorf("parse %s: %w", src, err)
		}
		if len(entries) == 0 {
			fmt.Fprintf(os.Stderr, "  skipped %s: no entries\n", src)
			continue
		}

		base := filepath.Base(src)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)

		// 1. 写 .dict.txt 源文件（供 goimed 自动构建）
		dictTxtPath := filepath.Join(importDicts, name+".txt")
		if err := saveDictTxt(entries, dictTxtPath); err != nil {
			return fmt.Errorf("save dict.txt %s: %w", dictTxtPath, err)
		}

		// 2. 编译 .goime 二进制索引
		goimePath := filepath.Join(importDir, name+".goime")
		if err := dict.BuildFromEntries(entries, goimePath); err != nil {
			return fmt.Errorf("build %s: %w", goimePath, err)
		}

		fmt.Fprintf(os.Stderr, "  %s: %d entries\n", src, len(entries))
		fmt.Fprintf(os.Stderr, "    → %s\n", dictTxtPath)
		fmt.Fprintf(os.Stderr, "    → %s\n", goimePath)
	}
	return nil
}

// saveDictTxt 将词条列表写为 GoIME 纯文本格式（.dict.txt）。
func saveDictTxt(entries []dict.Entry, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, e := range entries {
		if _, err := fmt.Fprintf(f, "%s\t%s\t%d\n", e.Pinyin, e.Text, e.Weight); err != nil {
			return err
		}
	}
	return nil
}

// parsePlainTextFile 解析 GoIME 纯文本词库文件。
func parsePlainTextFile(path string) ([]dict.Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []dict.Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if e, ok := dict.ParseLine(scanner.Text()); ok {
			entries = append(entries, e)
		}
	}
	return entries, scanner.Err()
}
