package dict

import (
	"os"
	"strings"
	"testing"
)

func TestImportRime(t *testing.T) {
	// 模拟一个 Rime .dict.yaml 文件
	rimeContent := `# Rime dictionary
# encoding: utf-8
---
name: rime_ice
version: "2024-01-01"
sort: by_weight
...
shu1ru4	输入	100
ni3hao3	你好	200
shi4jie4	世界	150
hang2	行	80
xing2	行	90
`

	src, err := os.CreateTemp("", "rime-*.dict.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(src.Name())
	if _, err := src.WriteString(rimeContent); err != nil {
		t.Fatal(err)
	}
	src.Close()

	dst, err := os.CreateTemp("", "goime-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	dst.Close()
	defer os.Remove(dst.Name())

	if err := ImportRime(src.Name(), dst.Name()); err != nil {
		t.Fatalf("ImportRime failed: %v", err)
	}

	// 读取输出并验证
	outContent, err := os.ReadFile(dst.Name())
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(outContent)), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(lines))
	}

	// 检查内容（保持原始顺序）
	expected := []string{
		"shu1ru4\t输入\t100",
		"ni3hao3\t你好\t200",
		"shi4jie4\t世界\t150",
		"hang2\t行\t80",
		"xing2\t行\t90",
	}
	for i, line := range lines {
		if line != expected[i] {
			t.Errorf("line %d:\n  got:  %q\n  want: %q", i, line, expected[i])
		}
	}
}

func TestImportRimeNoWeight(t *testing.T) {
	// 测试权重可选的情况
	rimeContent := `---
name: test
...
shu1ru4	输入
ni3hao3	你好	200
`

	src, err := os.CreateTemp("", "rime-*.dict.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(src.Name())
	if _, err := src.WriteString(rimeContent); err != nil {
		t.Fatal(err)
	}
	src.Close()

	dst, err := os.CreateTemp("", "goime-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	dst.Close()
	defer os.Remove(dst.Name())

	if err := ImportRime(src.Name(), dst.Name()); err != nil {
		t.Fatalf("ImportRime failed: %v", err)
	}

	outContent, err := os.ReadFile(dst.Name())
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(outContent)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(lines))
	}
	if lines[0] != "shu1ru4\t输入\t0" {
		t.Errorf("expected 'shu1ru4\t输入\t0', got %q", lines[0])
	}
}
