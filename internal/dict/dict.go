package dict

import (
	"encoding/binary"
	"syscall"
)

// Entry represents a single dictionary entry.
type Entry struct {
	Pinyin string
	Text   string
	Weight int
}

// Index provides lookup of entries by their tone-stripped pinyin key.
// 支持两种存储方式：
//   - mmap 模式：refs + mapped，词条懒解析
//   - 内存模式：data，词条预加载到 map（用于 Merge 结果）
type Index struct {
	data   map[string][]Entry // 内存模式（Merge 后使用）
	mapped []byte             // mmap 映射数据
	refs   map[string]entryRef // mmap 偏移索引
}

// NewIndex 创建一个空的 Index（内存模式）
func NewIndex() *Index {
	return &Index{data: make(map[string][]Entry)}
}

// entryRef 指向 mmap 数据中的一组词条
type entryRef struct {
	offset int
	count  int
}

// Lookup 查询拼音 key 对应的所有词条。
// mmap 模式下懒解析，内存模式下直接返回。
func (idx *Index) Lookup(key string) []Entry {
	if idx == nil {
		return nil
	}
	if idx.data != nil {
		return idx.data[key]
	}
	if idx.refs != nil {
		if ref, ok := idx.refs[key]; ok {
			return idx.getEntries(key, ref)
		}
	}
	return nil
}

// Merge 将另一个 Index 的所有条目合并到当前 Index。
// 合并后当前 Index 会切换为内存模式。
func (idx *Index) Merge(other *Index) {
	if other == nil {
		return
	}
	// 确保当前 Index 是内存模式
	if idx.data == nil {
		idx.data = make(map[string][]Entry)
		idx.mapped = nil
		idx.refs = nil
	}

	// 遍历 other 的所有条目合并过来
	other.visitAll(func(key string, entries []Entry) {
		existing := idx.data[key]
		if existing == nil {
			idx.data[key] = entries
			return
		}
		merged := make(map[string]int) // word → max weight
		for _, e := range existing {
			if w, ok := merged[e.Text]; !ok || e.Weight > w {
				merged[e.Text] = e.Weight
			}
		}
		for _, e := range entries {
			if w, ok := merged[e.Text]; !ok || e.Weight > w {
				merged[e.Text] = e.Weight
			}
		}
		result := make([]Entry, 0, len(merged))
		for word, weight := range merged {
			result = append(result, Entry{Pinyin: key, Text: word, Weight: weight})
		}
		idx.data[key] = result
	})
}

// visitAll 遍历 Index 中的所有词条，调用 fn。
// 对 mmap 模式会懒解析后传给 fn。
func (idx *Index) visitAll(fn func(key string, entries []Entry)) {
	if idx == nil {
		return
	}
	if idx.data != nil {
		for key, entries := range idx.data {
			fn(key, entries)
		}
		return
	}
	if idx.refs != nil {
		for key, ref := range idx.refs {
			fn(key, idx.getEntries(key, ref))
		}
	}
}

// getEntries 从 mmap 数据中解析指定位置的词条
func (idx *Index) getEntries(key string, ref entryRef) []Entry {
	if idx.mapped == nil {
		return nil
	}
	data := idx.mapped
	offset := ref.offset
	entries := make([]Entry, 0, ref.count)

	for i := 0; i < ref.count; i++ {
		if offset+2 > len(data) {
			break
		}
		textLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
		offset += 2
		if offset+textLen+4 > len(data) {
			break
		}
		text := string(data[offset : offset+textLen])
		offset += textLen
		weight := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
		entries = append(entries, Entry{Pinyin: key, Text: text, Weight: weight})
	}
	return entries
}

// Close 释放 mmap 内存
func (idx *Index) Close() {
	if idx.mapped != nil {
		syscall.Munmap(idx.mapped)
		idx.mapped = nil
	}
}
