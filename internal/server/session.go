package server

// Selection 选词历史条目
type Selection struct {
	Pinyin string
	Word   string
}

// CandidateResult 带拼音片段的候选项
type CandidateResult struct {
	Text          string   // 词语
	Pinyin        string   // 带声调拼音
	Weight        int      // 权重
	ConsumedChars int      // 该词消耗的输入字符数
}

// Session 单个编辑器连接的状态
type Session struct {
	buffer      string
	scheme      string
	selections  []Selection
	candidates  []CandidateResult
	page        int
	pageSize    int
}

// NewSession 创建新 session
func NewSession(scheme string) *Session {
	return &Session{scheme: scheme, pageSize: 5}
}

func (s *Session) Buffer() string                { return s.buffer }
func (s *Session) Scheme() string                { return s.scheme }
func (s *Session) Selections() []Selection       { return s.selections }
func (s *Session) Candidates() []CandidateResult { return s.candidates }
func (s *Session) Page() int                    { return s.page }

func (s *Session) SetScheme(name string) {
	s.scheme = name
	s.Clear()
}

func (s *Session) Append(key string) {
	s.buffer += key
}

func (s *Session) SetBuffer(buf string) {
	s.buffer = buf
}

func (s *Session) Backspace() {
	if len(s.buffer) > 0 {
		s.buffer = s.buffer[:len(s.buffer)-1]
	}
}

func (s *Session) Clear() {
	s.buffer = ""
	s.candidates = nil
	s.page = 0
}

func (s *Session) Reset() {
	s.Clear()
	s.selections = nil
}

func (s *Session) AppendSelection(pinyin, word string) {
	s.selections = append(s.selections, Selection{Pinyin: pinyin, Word: word})
}

func (s *Session) ClearSelections() {
	s.selections = nil
}

func (s *Session) HasSelection() bool {
	return len(s.selections) > 0
}

func (s *Session) SetCandidates(candidates []CandidateResult) {
	s.candidates = candidates
	s.page = 0
}

func (s *Session) SetPage(p int) {
	if p < 0 {
		p = 0
	}
	s.page = p
}

func (s *Session) SetPageSize(size int) {
	if size > 0 {
		s.pageSize = size
	}
}

func (s *Session) PageSize() int {
	return s.pageSize
}
