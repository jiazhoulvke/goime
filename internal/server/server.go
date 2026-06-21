package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"github.com/jiazhoulvke/goime/internal/config"
	"github.com/jiazhoulvke/goime/internal/dict"
	"github.com/jiazhoulvke/goime/internal/engine"
	"github.com/jiazhoulvke/goime/internal/protocol"
)

// Server Unix Socket 服务器
type Server struct {
	cfg        *config.Config
	schemes    []string
	spellers   map[string]engine.Speller
	static     *dict.Index
	user       *dict.UserDict
	translator *engine.Translator
	ln         net.Listener
	mu         sync.Mutex
	sessions   map[net.Conn]*Session
	lastActive time.Time
	closed     chan struct{}
}

// New 创建服务器
func New(cfg *config.Config, static *dict.Index, user *dict.UserDict, schemes []string) (*Server, error) {
	translator := engine.NewTranslator(static, user, cfg.Translator.MaxSyllables)
	spellers := map[string]engine.Speller{
		"xiaohe":  engine.NewXiaoheSpeller(),
		"fullpin": engine.NewFullPinyinSpeller(),
	}
	return &Server{
		cfg:        cfg,
		schemes:    schemes,
		spellers:   spellers,
		static:     static,
		user:       user,
		translator: translator,
		sessions:   make(map[net.Conn]*Session),
		lastActive: time.Now(),
		closed:     make(chan struct{}),
	}, nil
}

// Listen 监听 Unix socket 并接受连接
func (s *Server) Listen() error {
	socketPath := s.cfg.SocketPath()
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove existing socket: %w", err)
	}

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", socketPath, err)
	}
	if err := os.Chmod(socketPath, 0600); err != nil {
		ln.Close()
		return fmt.Errorf("chmod socket: %w", err)
	}
	s.ln = ln

	slog.Info("server listening", "socket", socketPath)

	// 启动空闲超时监控
	go s.idleLoop()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-s.closed:
				return nil
			default:
				return err
			}
		}

		session := NewSession(s.cfg.Scheme.Active)
		s.mu.Lock()
		s.sessions[conn] = session
		s.mu.Unlock()

		go s.handleConn(conn, session)
	}
}

// Close 关闭服务器
func (s *Server) Close() error {
	close(s.closed)
	if s.ln != nil {
		s.ln.Close()
	}
	socketPath := s.cfg.SocketPath()
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// handleConn 处理单个客户端连接
func (s *Server) handleConn(conn net.Conn, session *Session) {
	defer func() {
		conn.Close()
		s.mu.Lock()
		delete(s.sessions, conn)
		s.translator.CommitSelections(s.cfg.UserDict.NewWordWeight)
		s.mu.Unlock()
	}()

	decoder := json.NewDecoder(conn)
	for {
		var req protocol.Request
		if err := decoder.Decode(&req); err != nil {
			slog.Debug("decode error", "error", err)
			return
		}

		s.mu.Lock()
		responses := s.handleRequest(req, session)
		s.lastActive = time.Now()
		s.mu.Unlock()

		for _, resp := range responses {
			if err := s.send(conn, resp); err != nil {
				slog.Debug("send error", "error", err)
				return
			}
		}
	}
}

// handleRequest 处理单个请求，返回一个或多个响应
func (s *Server) handleRequest(req protocol.Request, session *Session) []protocol.Response {
	switch req.Method {
	case "hello":
		// 取客户端请求与服务器的交集
		activeScheme := session.Scheme()
		pageSize := s.cfg.Candidates.PageSize
		schemes := s.schemes

		if req.PageSize > 0 {
			pageSize = req.PageSize
		}
		session.SetPageSize(pageSize)

		if len(req.Schemes) > 0 {
			// 取交集：客户端期望的 ∩ 服务器支持的
			serverSet := make(map[string]bool)
			for _, sch := range s.schemes {
				serverSet[sch] = true
			}
			var inter []string
			for _, sch := range req.Schemes {
				if serverSet[sch] {
					inter = append(inter, sch)
				}
			}
			if len(inter) > 0 {
				schemes = inter
				// 如果当前方案不在交集中，切换到第一个
				found := false
				for _, sch := range inter {
					if sch == activeScheme {
						found = true
						break
					}
				}
				if !found {
					activeScheme = inter[0]
					session.SetScheme(activeScheme)
				}
			}
		}

		return []protocol.Response{protocol.NewWelcome(1, schemes, activeScheme, pageSize)}

	case "input":
		if len(req.Key) != 1 || req.Key[0] < 'a' || req.Key[0] > 'z' {
			return []protocol.Response{protocol.NewError("invalid key: " + req.Key)}
		}
		session.Append(req.Key)
		return []protocol.Response{s.buildInputResponse(session)}

	case "enter":
		text := session.Buffer()
		session.Reset()
		s.translator.ClearSelections()
		if text == "" {
			return []protocol.Response{protocol.NewIdle()}
		}
		return []protocol.Response{protocol.NewCommit(text, "")}

	case "escape":
		session.Reset()
		s.translator.ClearSelections()
		return []protocol.Response{protocol.NewIdle()}

	case "backspace":
		session.Backspace()
		return []protocol.Response{s.buildInputResponse(session)}

	case "space":
		if session.Buffer() == "" {
			return []protocol.Response{protocol.NewIdle()}
		}
		return s.handleSelect(session, 0)

	case "select":
		return s.handleSelect(session, req.Index)

	case "arrow":
		return []protocol.Response{s.handleArrow(session, req.Dir)}

	case "page":
		return []protocol.Response{s.handlePage(session, req.Dir)}

	case "commit_preedit":
		text := session.Buffer()
		session.Reset()
		s.translator.ClearSelections()
		return []protocol.Response{protocol.NewCommit(text, "")}

	case "set_scheme":
		session.SetScheme(req.Name)
		return []protocol.Response{protocol.NewWelcome(1, s.schemes, session.Scheme(), session.PageSize())}

	case "config":
		// 客户端随时更新配置：分页大小、启用方案
		if req.PageSize > 0 {
			session.SetPageSize(req.PageSize)
		}
		schemes := s.schemes
		activeScheme := session.Scheme()
		if len(req.Schemes) > 0 {
			serverSet := make(map[string]bool)
			for _, sch := range s.schemes {
				serverSet[sch] = true
			}
			var inter []string
			for _, sch := range req.Schemes {
				if serverSet[sch] {
					inter = append(inter, sch)
				}
			}
			if len(inter) > 0 {
				schemes = inter
				found := false
				for _, sch := range inter {
					if sch == activeScheme {
						found = true
						break
					}
				}
				if !found {
					activeScheme = inter[0]
					session.SetScheme(activeScheme)
				}
			}
		}
		session.SetPageSize(session.PageSize()) // 确保 >0，否则默认 5
		if session.PageSize() <= 0 {
			session.SetPageSize(5)
		}
		return []protocol.Response{protocol.NewWelcome(1, schemes, activeScheme, session.PageSize())}

	default:
		return []protocol.Response{protocol.NewError("unknown method: " + req.Method)}
	}
}

// handleSelect 根据 index 从候选中选词上屏
// 支持部分提交：选词后剩余输入继续显示为候选
func (s *Server) handleSelect(session *Session, index int) []protocol.Response {
	candidates := session.Candidates()
	if index < 0 || index >= len(candidates) {
		text := session.Buffer()
		session.Reset()
		s.translator.ClearSelections()
		if text == "" {
			return []protocol.Response{protocol.NewIdle()}
		}
		return []protocol.Response{protocol.NewCommit(text, "")}
	}

	selected := candidates[index]
	session.AppendSelection(selected.Pinyin, selected.Text)
	s.translator.AppendSelection(selected.Pinyin, selected.Text)

	consumed := selected.ConsumedChars
	buf := session.Buffer()
	remain := ""
	if consumed > 0 && consumed <= len(buf) {
		remain = buf[consumed:]
	}

	commitResp := protocol.NewCommit(selected.Text, "")

	if remain == "" {
		session.Clear()
		s.translator.CommitSelections(s.cfg.UserDict.NewWordWeight)
		return []protocol.Response{commitResp}
	}

	// 部分提交：先发 commit，再发剩余输入的 preedit+candidates
	session.SetBuffer(remain)
	session.SetCandidates(nil)
	session.SetPage(0)
	nextResp := s.buildInputResponse(session)

	return []protocol.Response{commitResp, nextResp}
}

// handleArrow 处理方向键（←/→ 移动光标，↑/↓ 翻页）
func (s *Server) handleArrow(session *Session, dir string) protocol.Response {
	switch dir {
	case "up", "prev":
		session.SetPage(session.Page() - 1)
	case "down", "next":
		session.SetPage(session.Page() + 1)
	default:
		// left/right: 移动光标（MVP 暂不实现）
	}
	return s.buildCandidatesResponse(session)
}

// handlePage 处理翻页
func (s *Server) handlePage(session *Session, dir string) protocol.Response {
	switch dir {
	case "next", "down":
		session.SetPage(session.Page() + 1)
	case "prev", "up":
		session.SetPage(session.Page() - 1)
	}
	return s.buildCandidatesResponse(session)
}

// buildInputResponse 根据 session 构建输入响应
// 执行完整引擎管线：Speller → Segmentor → Translator，返回 preedit + candidates
func (s *Server) buildInputResponse(session *Session) protocol.Response {
	buf := session.Buffer()
	if buf == "" {
		session.SetCandidates(nil)
		return protocol.NewIdle()
	}

	// 1. Speller: 按键码 → 拼音音节
	speller, ok := s.spellers[session.Scheme()]
	if !ok {
		return protocol.NewPreedit(buf, len(buf))
	}
	pinyins := speller.ToPinyin(buf)
	if len(pinyins) == 0 {
		session.SetCandidates(nil)
		return protocol.NewPreedit(buf, len(buf))
	}

	// 2. Segmentor: 拼音 → 音节切分 → 查候选
	allResults := s.queryCandidates(pinyins)

	// 3. 存储候选项到 session
	session.SetCandidates(allResults)

	// 4. 构建响应
	resp := protocol.NewPreedit(buf, len(buf))
	if len(allResults) > 0 {
		pageSize := session.PageSize()
		total := len(allResults)
		pages := (total + pageSize - 1) / pageSize
		if pages == 0 {
			pages = 1
		}
		start := 0
		end := start + pageSize
		if end > total {
			end = total
		}
		clist := make([]protocol.Candidate, 0, end-start)
		for _, c := range allResults[start:end] {
			clist = append(clist, protocol.Candidate{
				Text:   c.Text,
				Code:   c.Pinyin,
				Weight: c.Weight,
			})
		}
		resp.Candidates = &protocol.Candidates{
			List:  clist,
			Page:  0,
			Total: pages,
		}
	} else {
		slog.Debug("no candidates found", "buf", buf, "speller", session.Scheme())
	}
	return resp
}

// queryCandidates 将拼音音节送入 Translator 查询候选词
func (s *Server) queryCandidates(pinyins []string) []CandidateResult {
	pinyinStr := ""
	for _, p := range pinyins {
		pinyinStr += p
	}

	segmentations := engine.Segment(pinyinStr)
	if len(segmentations) == 0 {
		segmentations = [][]string{pinyins}
	}

	seen := make(map[string]bool)
	var results []CandidateResult

	for _, seg := range segmentations {
		entries := s.translator.Query(seg)
		for _, e := range entries {
			if !seen[e.Text] {
				seen[e.Text] = true
				consumed := calcConsumed(pinyins, seg, e.Pinyin)
				results = append(results, CandidateResult{
					Text:          e.Text,
					Pinyin:        e.Pinyin,
					Weight:        e.Weight,
					ConsumedChars: consumed,
				})
			}
		}
	}
	return results
}

// calcConsumed 计算该候选词消耗的输入字符数。
// 双拼：在 pinyins 中找匹配，每个 pinyin 映射 2 个输入字符。
// 全拼：在 segments 中找匹配，按实际拼音长度计算。
func calcConsumed(pinyins, seg []string, entryPinyin string) int {
	// 全拼：pinyins 只有一个元素（整个输入串）
	if len(pinyins) == 1 {
		concat := ""
		for _, s := range seg {
			concat += s
			if concat == entryPinyin {
				return len(concat)
			}
		}
		return len(entryPinyin)
	}
	// 双拼：每个 pinyin 对应 2 个输入字符
	matched := 0
	concat := ""
	for _, p := range pinyins {
		concat += p
		matched++
		if concat == entryPinyin {
			return matched * 2
		}
		if len(concat) > len(entryPinyin) {
			break
		}
	}
	return 2
}

// buildCandidatesResponse 根据 session 当前分页返回候选词响应
func (s *Server) buildCandidatesResponse(session *Session) protocol.Response {
	buf := session.Buffer()
	candidates := session.Candidates()
	resp := protocol.NewPreedit(buf, len(buf))
	if len(candidates) > 0 {
		page := session.Page()
		pageSize := session.PageSize()
		total := len(candidates)
		pages := (total + pageSize - 1) / pageSize
		if pages == 0 {
			pages = 1
		}
		if page >= pages {
			page = pages - 1
		}
		start := page * pageSize
		end := start + pageSize
		if end > total {
			end = total
		}
		clist := make([]protocol.Candidate, 0, end-start)
		for _, c := range candidates[start:end] {
			clist = append(clist, protocol.Candidate{
				Text:   c.Text,
				Code:   c.Pinyin,
				Weight: c.Weight,
			})
		}
		resp.Candidates = &protocol.Candidates{
			List:  clist,
			Page:  page,
			Total: pages,
		}
	}
	return resp
}

func (s *Server) send(conn net.Conn, resp protocol.Response) error {
	return json.NewEncoder(conn).Encode(resp)
}

// idleLoop 空闲超时检查
func (s *Server) idleLoop() {
	timeout, err := time.ParseDuration(s.cfg.General.IdleTimeout)
	if err != nil || timeout <= 0 {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			idle := time.Since(s.lastActive)
			noSessions := len(s.sessions) == 0
			s.mu.Unlock()

			if noSessions && idle > timeout {
				slog.Info("shutting down due to idle timeout", "idle", idle)
				s.Close()
				return
			}
		case <-s.closed:
			return
		}
	}
}
