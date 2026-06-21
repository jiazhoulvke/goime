package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jiazhoulvke/goime/internal/dict"
	"github.com/spf13/cobra"
)

var (
	mergeRime  bool
	mergeOut   string
)

var mergeCmd = &cobra.Command{
	Use:   "merge [flags] <files...>",
	Short: "合并多个词库为一个 .goime 索引",
	Long: `合并多个词库源文件，去重后编译为一个 .goime 二进制索引。
同拼音同词时取最大权重。

  goime-dict merge a.txt b.txt -o merged.goime
  goime-dict merge --rime a.dict.yaml b.dict.yaml -o merged.goime

省略 -o 时，默认输出到 ~/.cache/goime/merged.goime。`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMerge(args)
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().BoolVar(&mergeRime, "rime", false, "输入为 Rime .dict.yaml 格式")
	mergeCmd.Flags().StringVarP(&mergeOut, "output", "o", "", "输出 .goime 文件路径（默认 ~/.cache/goime/merged.goime）")
}

func runMerge(srcFiles []string) error {
	outputPath := mergeOut
	if outputPath == "" {
		dir := defaultBuildDir()
		os.MkdirAll(dir, 0755)
		outputPath = filepath.Join(dir, "merged.goime")
	}

	var allEntries []dict.Entry
	for _, src := range srcFiles {
		var entries []dict.Entry
		var err error

		if mergeRime {
			entries, err = dict.ParseRimeFile(src)
		} else {
			entries, err = parsePlainTextFile(src)
		}
		if err != nil {
			return fmt.Errorf("parse %s: %w", src, err)
		}
		allEntries = append(allEntries, entries...)
		fmt.Fprintf(os.Stderr, "  loaded %s: %d entries\n", src, len(entries))
	}

	fmt.Fprintf(os.Stderr, "  merging and building %s ... ", outputPath)
	if err := dict.BuildFromEntries(allEntries, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "FAILED\n")
		return fmt.Errorf("build: %w", err)
	}
	fmt.Fprintf(os.Stderr, "done (%d total, dedup by max weight)\n", len(allEntries))
	return nil
}
