package protocol

// Request 客户端 → 服务端请求
type Request struct {
	Method   string   `json:"method"`
	Key      string   `json:"key,omitempty"`      // input 方法的按键
	Index    int      `json:"index,omitempty"`    // select 方法的索引
	Dir      string   `json:"dir,omitempty"`      // arrow/page 的方向
	Name     string   `json:"name,omitempty"`     // set_scheme 的方案名
	Version  int      `json:"version,omitempty"`  // hello 的协议版本
	Client   string   `json:"client,omitempty"`   // hello 的客户端标识
	PageSize int      `json:"page_size,omitempty"` // hello 客户端期望的分页大小
	Schemes  []string `json:"schemes,omitempty"`  // hello 客户端期望启用的方案列表
}

// Response 服务端 → 客户端响应
type Response struct {
	Type       string       `json:"type"`
	Text       string       `json:"text,omitempty"`         // commit/preedit 的文字
	Pos        int          `json:"pos,omitempty"`          // preedit 的光标位置
	Candidates *Candidates  `json:"candidates,omitempty"`   // candidates 列表
	PendingKey string       `json:"pending_key,omitempty"`  // commit 需透传的后续字符
	Version    int          `json:"version,omitempty"`      // welcome 的协议版本
	Schemes    []string     `json:"schemes,omitempty"`      // welcome 的可用方案
	Active     string       `json:"active,omitempty"`       // welcome 的当前方案
	PageSize   int          `json:"page_size,omitempty"`    // welcome 的每页候选数
	Message    string       `json:"message,omitempty"`      // error 的错误信息
}

// Candidates 候选词列表
type Candidates struct {
	List  []Candidate `json:"list"`
	Page  int         `json:"page"`
	Total int         `json:"total"`
}

// Candidate 单个候选词
type Candidate struct {
	Text   string `json:"text"`
	Code   string `json:"code"`
	Weight int    `json:"weight"`
}

// NewWelcome 创建握手指令响应
func NewWelcome(version int, schemes []string, active string, pageSize int) Response {
	return Response{
		Type:     "welcome",
		Version:  version,
		Schemes:  schemes,
		Active:   active,
		PageSize: pageSize,
	}
}

// NewPreedit 创建预编辑响应
func NewPreedit(text string, pos int) Response {
	return Response{
		Type: "preedit",
		Text: text,
		Pos:  pos,
	}
}

// NewCommit 创建上屏响应
func NewCommit(text string, pendingKey string) Response {
	return Response{
		Type:       "commit",
		Text:       text,
		PendingKey: pendingKey,
	}
}

// NewIdle 创建空闲响应
func NewIdle() Response {
	return Response{Type: "idle"}
}

// NewError 创建错误响应
func NewError(message string) Response {
	return Response{
		Type:    "error",
		Message: message,
	}
}
