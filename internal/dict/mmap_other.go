//go:build !linux && !darwin && !freebsd && !openbsd

package dict

import (
	"encoding/binary"
	"fmt"
	"os"
)

// mmapOpen 回退到 ReadFile（Windows 等不支持 mmap 的平台）
func mmapOpen(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	idx := &Index{
		refs: make(map[string]entryRef),
	}

	offset := 0
	for offset < len(data) {
		if offset+2 > len(data) {
			break
		}
		keyLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
		offset += 2
		if offset+keyLen > len(data) {
			break
		}
		key := string(data[offset : offset+keyLen])
		offset += keyLen

		if offset+4 > len(data) {
			break
		}
		count := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4

		// 记录偏移并跳过词条
		idx.refs[key] = entryRef{offset: offset, count: count}
		for i := 0; i < count; i++ {
			if offset+2 > len(data) {
				break
			}
			textLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
			offset += 2
			if offset+textLen+4 > len(data) {
				break
			}
			offset += textLen + 4
		}
	}

	// 非 Unix 平台需要保留 data 供 getEntries 使用
	idx.mapped = data
	return idx, nil
}

// Close 释放内存（非 Unix 平台为无操作）
func (idx *Index) Close() {
	idx.mapped = nil
}
