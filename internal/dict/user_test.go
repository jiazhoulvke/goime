package dict

import (
    "os"
    "testing"
)

func TestUserDict(t *testing.T) {
    f, err := os.CreateTemp("", "user-*.db")
    if err != nil {
        t.Fatal(err)
    }
    f.Close()
    defer os.Remove(f.Name())

    ud, err := OpenUserDict(f.Name())
    if err != nil {
        t.Fatalf("OpenUserDict: %v", err)
    }
    defer ud.Close()

    // Add a user word
    if err := ud.AddUserWord("nihaoshijie", "你好世界", 100); err != nil {
        t.Fatalf("AddUserWord: %v", err)
    }

    // Query it back
    entries := ud.GetUserWords("nihaoshijie")
    if len(entries) != 1 {
        t.Fatalf("expected 1 entry, got %d", len(entries))
    }
    if entries[0].Text != "你好世界" {
        t.Errorf("expected 你好世界, got %s", entries[0].Text)
    }

    // GetFreq
    freq, err := ud.GetFreq("shuru", "输入")
    if err != nil {
        t.Fatal(err)
    }
    if freq != 0 {
        t.Errorf("expected 0 freq for untracked word, got %d", freq)
    }

    // IncrementFreq
    if err := ud.IncrementFreq("shuru", "输入"); err != nil {
        t.Fatal(err)
    }
    freq, _ = ud.GetFreq("shuru", "输入")
    if freq != 1 {
        t.Errorf("expected 1, got %d", freq)
    }
}

func TestUserDictDecay(t *testing.T) {
    f, err := os.CreateTemp("", "user-*.db")
    if err != nil {
        t.Fatal(err)
    }
    f.Close()
    defer os.Remove(f.Name())

    ud, err := OpenUserDict(f.Name())
    if err != nil {
        t.Fatal(err)
    }
    defer ud.Close()

    ud.IncrementFreq("shuru", "输入")
    freqBefore, _ := ud.GetFreq("shuru", "输入")
    if freqBefore != 1 {
        t.Fatalf("expected 1, got %d", freqBefore)
    }

    ud.DecayAll(0.5)
    freqAfter, _ := ud.GetFreq("shuru", "输入")
    if freqAfter != 0 {
        t.Errorf("expected 0 after 0.5 decay on 1, got %d", freqAfter)
    }
}
