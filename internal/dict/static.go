package dict

import "fmt"

// Load 加载二进制词库文件，返回 Index。
// Unix 平台使用 mmap 零拷贝加载，其他平台回退到 ReadFile。
func Load(path string) (*Index, error) {
	idx, err := mmapOpen(path)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", path, err)
	}
	return idx, nil
}
