package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jiazhoulvke/goime/internal/dict"
	"github.com/spf13/cobra"
)

var (
	userDBPath string
)

var userCmd = &cobra.Command{
	Use:   "user <command>",
	Short: "管理用户词库",
	Long: `管理用户词库（SQLite 格式）。

  goime-dict user export <output.txt>            导出为纯文本
  goime-dict user export <output.txt> --db path  指定用户词库路径
  goime-dict user import <input.txt>             从纯文本导入
  goime-dict user import <input.txt> --db path   指定用户词库路径

用户词库默认路径：~/.config/goime/user_dict.db`,
}

var userExportCmd = &cobra.Command{
	Use:   "export <output.txt>",
	Short: "导出用户词库为纯文本",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		txtPath := args[0]
		dbPath := resolveUserDB(userDBPath)
		ud, err := dict.OpenUserDict(dbPath)
		if err != nil {
			return fmt.Errorf("open user dict: %w", err)
		}
		defer ud.Close()
		if err := ud.Export(txtPath); err != nil {
			return fmt.Errorf("export: %w", err)
		}
		fmt.Fprintf(os.Stderr, "  exported %s -> %s\n", dbPath, txtPath)
		return nil
	},
}

var userImportCmd = &cobra.Command{
	Use:   "import <input.txt>",
	Short: "从纯文本导入用户词库",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		txtPath := args[0]
		dbPath := resolveUserDB(userDBPath)
		ud, err := dict.OpenUserDict(dbPath)
		if err != nil {
			return fmt.Errorf("open user dict: %w", err)
		}
		defer ud.Close()
		if err := ud.Import(txtPath); err != nil {
			return fmt.Errorf("import: %w", err)
		}
		fmt.Fprintf(os.Stderr, "  imported %s -> %s\n", txtPath, dbPath)
		return nil
	},
}

func init() {
	defaultDB := defaultUserDBPath()

	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userExportCmd)
	userCmd.AddCommand(userImportCmd)

	userExportCmd.Flags().StringVar(&userDBPath, "db", defaultDB, "用户词库 SQLite 路径（默认 ~/.config/goime/user_dict.db）")
	userImportCmd.Flags().StringVar(&userDBPath, "db", defaultDB, "用户词库 SQLite 路径（默认 ~/.config/goime/user_dict.db）")
}

// defaultUserDBPath 返回默认用户词库路径：~/.config/goime/user_dict.db
func defaultUserDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "user_dict.db"
	}
	return filepath.Join(home, ".config", "goime", "user_dict.db")
}

// resolveUserDB 展开路径中的 ~ 为用户目录
func resolveUserDB(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}
