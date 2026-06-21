package cmd

import (
	"fmt"
	"os"

	"github.com/jiazhoulvke/goime/internal/dict"
	"github.com/spf13/cobra"
)

var buildRime bool

var buildCmd = &cobra.Command{
	Use:   "build [flags] <input> <output>",
	Short: "编译单个词库为二进制 .goime 索引",
	Long: `编译单个词库文件为二进制 .goime 索引。

支持 GoIME 纯文本格式（默认）和 Rime .dict.yaml 格式（--rime）：

  goime-dict build dict.txt dict.goime          编译纯文本
  goime-dict build --rime dict.dict.yaml dict.goime  编译 Rime 词库

纯文本格式：拼音<TAB>词语<TAB>权重（权重可选，默认 0）
多音字分别按读音存储。同拼音同词时权重取最大值。`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		src, dst := args[0], args[1]

		if buildRime {
			return buildFromRime(src, dst)
		}
		return dict.Build(src, dst)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVar(&buildRime, "rime", false, "输入为 Rime .dict.yaml 格式")
}

// buildFromRime 编译单个 Rime 词库为 .goime 二进制。
func buildFromRime(src, dst string) error {
	entries, err := dict.ParseRimeFile(src)
	if err != nil {
		return fmt.Errorf("parse rime file: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("no entries found in %s", src)
	}
	if err := dict.BuildFromEntries(entries, dst); err != nil {
		return fmt.Errorf("build: %w", err)
	}
	fmt.Fprintf(os.Stderr, "  built %s: %d entries\n", dst, len(entries))
	return nil
}
