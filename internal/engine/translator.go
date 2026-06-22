package engine

import (
    "log/slog"
    "sort"

    "github.com/jiazhoulvke/goime/internal/dict"
)

// Translator 候选词生成器
type Translator struct {
    static       *dict.Index
    user         *dict.UserDict
    maxSyllables int
}

// NewTranslator 创建 Translator
func NewTranslator(static *dict.Index, user *dict.UserDict, maxSyllables int) *Translator {
    return &Translator{
        static:       static,
        user:         user,
        maxSyllables: maxSyllables,
    }
}

// Query 根据拼音音节序列查询候选词
func (t *Translator) Query(syllables []string) []dict.Entry {
    if len(syllables) == 0 {
        return nil
    }

    fullPinyin := ""
    for _, s := range syllables {
        fullPinyin += s
    }

    type scored struct {
        entry      dict.Entry
        sylCount   int // 匹配的音节数
        sylPos     int // 起始音节位置（-1 表示多音节匹配）
    }

    seen := make(map[string]bool)
    var results []scored

    addEntry := func(e dict.Entry, count int, pos int) {
        if t.user != nil {
            freq, err := t.user.GetFreq(e.Pinyin, e.Text)
            if err != nil {
                slog.Warn("GetFreq failed", "error", err, "pinyin", e.Pinyin, "word", e.Text)
            }
            e.Weight += freq
        }
        key := e.Pinyin + "\x00" + e.Text
        if !seen[key] {
            seen[key] = true
            results = append(results, scored{entry: e, sylCount: count, sylPos: pos})
        }
    }

    // 1. User words for full pinyin (全匹配)
    if t.user != nil {
        for _, e := range t.user.GetUserWords(fullPinyin) {
            addEntry(e, len(syllables), -1)
        }
    }

    // 2. Single syllables
    if t.static != nil {
        for idx, syl := range syllables {
            for _, e := range t.static.Lookup(syl) {
                addEntry(e, 1, idx)
            }
        }
    }

    // 3. Multi-syllable phrases
    for length := 2; length <= len(syllables) && length <= t.maxSyllables; length++ {
        for i := 0; i+length <= len(syllables); i++ {
            pinyin := ""
            for j := i; j < i+length; j++ {
                pinyin += syllables[j]
            }
            if t.static != nil {
                for _, e := range t.static.Lookup(pinyin) {
                    addEntry(e, length, i) // i = 起始音节位置
                }
            }
        }
    }

    // 排序：起始位置靠前的优先 → 同位置长匹配优先 → 同位置同长度按权重
    sort.Slice(results, func(i, j int) bool {
        if results[i].sylPos != results[j].sylPos {
            return results[i].sylPos < results[j].sylPos
        }
        if results[i].sylCount != results[j].sylCount {
            return results[i].sylCount > results[j].sylCount
        }
        return results[i].entry.Weight > results[j].entry.Weight
    })

    out := make([]dict.Entry, len(results))
    for i, s := range results {
        out[i] = s.entry
    }
    return out
}

// CommitSelections 合并选词历史写入用户词库（仅包含多个词时生效）
// selections 来自 Session，避免跨 session 共享状态
func (t *Translator) CommitSelections(selections []Selection, weight int) {
    if len(selections) < 2 || t.user == nil {
        return
    }
    pinyin := ""
    word := ""
    for _, s := range selections {
        pinyin += s.Pinyin
        word += s.Word
    }
    if err := t.user.AddUserWord(pinyin, word, weight); err != nil {
        slog.Warn("AddUserWord failed", "error", err, "pinyin", pinyin, "word", word)
    }
}

// Selection 选词历史条目（供外部包使用）
type Selection struct {
    Pinyin string
    Word   string
}
