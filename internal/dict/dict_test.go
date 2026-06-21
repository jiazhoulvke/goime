package dict

import (
    "os"
    "testing"
)

func TestBuildAndLookup(t *testing.T) {
    content := "shu1ru4 输入 100\nni3hao3 你好 200\nshi4jie4 世界 150\nhang2 行 80\nxing2 行 90\n"
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

    if err := Build(src.Name(), dst.Name()); err != nil {
        t.Fatalf("Build failed: %v", err)
    }
    idx, err := Load(dst.Name())
    if err != nil {
        t.Fatalf("Load failed: %v", err)
    }

    // Lookup by pinyin key (tone-stripped)
    entries := idx.Lookup("shuru")
    if len(entries) == 0 {
        t.Fatal("expected entries for shuru")
    }
    found := false
    for _, e := range entries {
        if e.Text == "输入" && e.Weight == 100 {
            found = true
            break
        }
    }
    if !found {
        t.Errorf("expected 输入 with weight 100 in shuru entries, got %+v", entries)
    }

    // Multi-syllable lookup
    entries = idx.Lookup("nihao")
    if len(entries) == 0 {
        t.Fatal("expected entries for nihao")
    }

    // Lookup non-existent
    entries = idx.Lookup("xxxxx")
    if len(entries) != 0 {
        t.Errorf("expected empty for xxxxx, got %d entries", len(entries))
    }
}

func TestMultiDictMerge(t *testing.T) {
    content1 := "shu1ru4 输入 100\n"
    content2 := "shu1ru4 输入 200\nni3hao3 你好 150\n"
    
    // Build two separate dicts
    buildDict := func(content string) (string, func()) {
        f, _ := os.CreateTemp("", "dict-*.txt")
        f.WriteString(content)
        f.Close()
        d, _ := os.CreateTemp("", "dict-*.goime")
        d.Close()
        Build(f.Name(), d.Name())
        os.Remove(f.Name())
        return d.Name(), func() { os.Remove(d.Name()) }
    }
    
    path1, clean1 := buildDict(content1)
    defer clean1()
    path2, clean2 := buildDict(content2)
    defer clean2()
    
    idx1, _ := Load(path1)
    idx2, _ := Load(path2)
    
    entries := idx1.Lookup("shuru")
    entries2 := idx2.Lookup("shuru")
    if len(entries2) == 0 {
        t.Fatal("expected entries from second dict")
    }
    // Verify second dict has higher weight
    maxWeight := 0
    for _, e := range entries2 {
        if e.Text == "输入" && e.Weight > maxWeight {
            maxWeight = e.Weight
        }
    }
    if maxWeight <= 100 {
        t.Errorf("expected weight > 100 from second dict, got %d", maxWeight)
    }
    _ = entries
}
