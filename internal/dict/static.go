package dict

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
)

// Load reads a binary dictionary file using mmap for zero-copy loading.
func Load(path string) (*Index, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}
	if info.Size() == 0 {
		return NewIndex(), nil
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, int(info.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmap: %w", err)
	}

	idx := &Index{
		mapped: data,
		refs:   make(map[string]entryRef),
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

	return idx, nil
}
