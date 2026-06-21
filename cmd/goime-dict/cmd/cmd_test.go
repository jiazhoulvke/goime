package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jiazhoulvke/goime/internal/dict"
)

func TestBuildCommand(t *testing.T) {
	dir, err := os.MkdirTemp("", "goime-dict-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	src := filepath.Join(dir, "test.dict.txt")
	content := "shu1ru4 输入 100\nni3hao3 你好 200\n"
	if err := os.WriteFile(src, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(dir, "test.goime")
	if err := dict.Build(src, dst); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Errorf("output file not created: %s", dst)
	}
}

func TestImportRime(t *testing.T) {
	dir, err := os.MkdirTemp("", "goime-dict-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	rimeFile := filepath.Join(dir, "test.dict.yaml")
	rimeContent := `---
name: test
...
shu1ru4	输入	100
ni3hao3	你好	200
`
	if err := os.WriteFile(rimeFile, []byte(rimeContent), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := dict.ParseRimeFile(rimeFile)
	if err != nil {
		t.Fatalf("ParseRimeFile failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestImportMultiFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "goime-dict-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// 创建两个纯文本词库
	f1 := filepath.Join(dir, "d1.txt")
	f2 := filepath.Join(dir, "d2.txt")

	if err := os.WriteFile(f1, []byte("shu1ru4 输入 100\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte("ni3hao3 你好 200\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// 每个文件单独编译
	for _, src := range []string{f1, f2} {
		entries, err := parsePlainTextFile(src)
		if err != nil {
			t.Fatal(err)
		}
		out := filepath.Join(dir, filepath.Base(src)+".goime")
		if err := dict.BuildFromEntries(entries, out); err != nil {
			t.Fatalf("BuildFromEntries failed: %v", err)
		}
		if _, err := os.Stat(out); os.IsNotExist(err) {
			t.Errorf("output file not created: %s", out)
		}
	}

	// 验证生成了两个独立文件
	if _, err := os.Stat(filepath.Join(dir, "d1.txt.goime")); os.IsNotExist(err) {
		t.Error("d1.txt.goime not found")
	}
	if _, err := os.Stat(filepath.Join(dir, "d2.txt.goime")); os.IsNotExist(err) {
		t.Error("d2.txt.goime not found")
	}
}
