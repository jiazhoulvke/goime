package engine

import (
    "os"
    "testing"

    "github.com/jiazhoulvke/goime/internal/dict"
)

func TestTranslatorPhrase(t *testing.T) {
    content := "shu1ru4 输入 100\nni3hao3 你好 200\n"
    src, err := os.CreateTemp("", "dict-*.txt")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(src.Name())
    if _, err := src.WriteString(content); err != nil {
        t.Fatal(err)
    }
    src.Close()

    dst, err := os.CreateTemp("", "dict-*.goime")
    if err != nil {
        t.Fatal(err)
    }
    dst.Close()
    defer os.Remove(dst.Name())

    if err := dict.Build(src.Name(), dst.Name()); err != nil {
        t.Fatalf("Build failed: %v", err)
    }
    idx, err := dict.Load(dst.Name())
    if err != nil {
        t.Fatalf("Load failed: %v", err)
    }

    tr := NewTranslator(idx, nil, 8)
    candidates := tr.Query([]string{"shu", "ru"})
    found := false
    for _, c := range candidates {
        if c.Text == "输入" {
            found = true
            break
        }
    }
    if !found {
        t.Errorf("expected '输入' in 'shuru' candidates, got %+v", candidates)
    }
}
