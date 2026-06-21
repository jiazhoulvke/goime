package dict

import (
	"bufio"
	"encoding/binary"
	"os"
	"sort"
	"strings"
)

// stripTones removes tone digits (1-5) and spaces from a pinyin string.
func stripTones(pinyin string) string {
	runes := make([]rune, 0, len(pinyin))
	for _, r := range pinyin {
		if r >= '1' && r <= '5' {
			continue
		}
		if r == ' ' {
			continue
		}
		runes = append(runes, r)
	}
	return string(runes)
}

// parseWeight 解析权重字符串（纯数字，遇到非数字停止）。
func parseWeight(s string) int {
	w := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			w = w*10 + int(c-'0')
		} else {
			break
		}
	}
	return w
}

// ParseLine 解析 GoIME 纯文本格式的一行，返回 Entry。
// 格式：拼音<TAB>词语<TAB>权重（权重可选，默认 0）
// 支持 Tab 或空格分隔。
func ParseLine(line string) (Entry, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return Entry{}, false
	}
	parts := strings.SplitN(line, "\t", 3)
	if len(parts) < 2 {
		parts = strings.SplitN(line, " ", 3)
	}
	if len(parts) < 2 {
		return Entry{}, false
	}
	pinyin := parts[0]
	text := parts[1]
	weight := 0
	if len(parts) >= 3 {
		weight = parseWeight(parts[2])
	}
	if pinyin == "" || text == "" {
		return Entry{}, false
	}
	return Entry{Pinyin: pinyin, Text: text, Weight: weight}, true
}

// writeBinary 将已按 key 分组的条目写入二进制文件。
// 二进制格式：
//
//	[2 bytes key_len][key_len bytes key][4 bytes count][entry data...]
//	Each entry: [2 bytes text_len][text_len bytes text][4 bytes weight]
//
// 所有多字节值均为 BigEndian。
func writeBinary(groups map[string][]Entry, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		entries := groups[key]
		keyBytes := []byte(key)
		keyLen := uint16(len(keyBytes))

		if err := binary.Write(out, binary.BigEndian, keyLen); err != nil {
			return err
		}
		if _, err := out.Write(keyBytes); err != nil {
			return err
		}
		if err := binary.Write(out, binary.BigEndian, uint32(len(entries))); err != nil {
			return err
		}
		for _, e := range entries {
			textBytes := []byte(e.Text)
			textLen := uint16(len(textBytes))
			if err := binary.Write(out, binary.BigEndian, textLen); err != nil {
				return err
			}
			if _, err := out.Write(textBytes); err != nil {
				return err
			}
			if err := binary.Write(out, binary.BigEndian, uint32(e.Weight)); err != nil {
				return err
			}
		}
	}
	return nil
}

// entriesToGroups 将 Entry 切片按去调拼音 key 分组，同 key 同词取最大权重。
func entriesToGroups(entries []Entry) map[string][]Entry {
	// 先用 map 去重：key → word → 最大权重
	type keyWord struct {
		key  string
		word string
	}
	best := make(map[keyWord]int)
	meta := make(map[keyWord]Entry) // 存完整 Entry（带原始拼音）

	for _, e := range entries {
		kw := keyWord{key: stripTones(e.Pinyin), word: e.Text}
		if e.Weight > best[kw] {
			best[kw] = e.Weight
			meta[kw] = e
		}
	}

	// 按 key 分组
	groups := make(map[string][]Entry)
	for kw, e := range meta {
		e.Weight = best[kw]
		groups[kw.key] = append(groups[kw.key], e)
	}
	return groups
}

// Build 读取纯文本词库文件并编译为二进制索引。
func Build(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if e, ok := ParseLine(scanner.Text()); ok {
			entries = append(entries, e)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	groups := entriesToGroups(entries)
	return writeBinary(groups, dst)
}

// BuildFromEntries 将多个来源的 Entry 合并（同拼音同词取最大权重）后直接编译为二进制。
func BuildFromEntries(entries []Entry, dst string) error {
	groups := entriesToGroups(entries)
	return writeBinary(groups, dst)
}
