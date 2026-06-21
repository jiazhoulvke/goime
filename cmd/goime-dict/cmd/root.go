package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Execute 是 goime-dict 的入口
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "goime-dict",
	Short: "GoIME 词库工具 — 编译、转换和管理词库",
	Long: `GoIME 词库工具，用于：

  build     编译纯文本词库或 Rime .dict.yaml 为二进制 .goime 索引
  import    导入多个词库并编译（支持合并去重）
  user      管理用户词库（export / import）

使用 goime-dict <command> --help 查看子命令详情。`,
}
