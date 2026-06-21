package dict

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

// ImportRime 将 Rime .dict.yaml 格式的词库转换为 GoIME 纯文本格式。
func ImportRime(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open rime dict: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer out.Close()

	entries, err := ParseRimeFile(src)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if _, err := fmt.Fprintf(out, "%s\t%s\t%d\n", e.Pinyin, e.Text, e.Weight); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

// ParseRimeFile 解析 Rime .dict.yaml 文件，返回所有词条。
// 跳过 YAML 头（--- 到 ... 之间的内容）、注释和空行。
// 自动识别两种格式：
//
//	标准 Rime: 拼音<TAB>词语<TAB>权重  (shu1ru4<TAB>输入<TAB>100)
//	雾凇拼音:  词语<TAB>拼音<TAB>权重  (输入<TAB>shu1ru4<TAB>100)
func ParseRimeFile(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open rime dict: %w", err)
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	passedHeader := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// 跳过空行和注释
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// 跳过 YAML 头
		if !passedHeader {
			if trimmed == "..." {
				passedHeader = true
			}
			continue
		}

		if e, ok := ParseLineRime(line); ok {
			entries = append(entries, e)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read rime dict: %w", err)
	}

	return entries, nil
}

// ParseLineRime 解析一行 Rime 词条。
// 自动识别标准 Rime（拼音在前）和雾凇拼音（词语在前）两种格式。
func ParseLineRime(line string) (Entry, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return Entry{}, false
	}
	parts := strings.SplitN(line, "\t", 3)
	if len(parts) < 2 {
		// 尝试空格分隔
		parts = strings.SplitN(line, " ", 3)
	}
	if len(parts) < 2 {
		return Entry{}, false
	}

	col0 := strings.TrimSpace(parts[0])
	col1 := strings.TrimSpace(parts[1])
	weight := 0
	if len(parts) >= 3 {
		weight = parseWeight(strings.TrimSpace(parts[2]))
	}

	if col0 == "" || col1 == "" {
		return Entry{}, false
	}

	// 自动检测格式：首列是汉字则为 词语<TAB>拼音，否则为 拼音<TAB>词语
	if isChinese(col0) {
		// 雾凇格式：词语<TAB>拼音<TAB>权重
		return Entry{Pinyin: col1, Text: col0, Weight: weight}, true
	}
	// 标准格式：拼音<TAB>词语<TAB>权重
	return Entry{Pinyin: col0, Text: col1, Weight: weight}, true
}

// isChinese 判断字符串是否以中文字符开头
func isChinese(s string) bool {
	for _, r := range s {
		return unicode.Is(unicode.Han, r)
	}
	return false
}
