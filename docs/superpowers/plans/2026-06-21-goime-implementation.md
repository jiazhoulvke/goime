# GoIME Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build GoIME — a Go-based server-side IME with Unix socket, supporting shuangpin (Xiaohe) and full pinyin with user word frequency and custom compound words.

**Architecture:** Three-layer Unix socket daemon: protocol layer → input engine (Speller → Segmentor → Translator) → dictionary engine (static mmap + SQLite user dict). Config-driven input schemes. Separate CLI tool for dictionary building.

**Tech Stack:** Go 1.26, `github.com/BurntSushi/toml` for config, `modernc.org/sqlite` for user dict, standard library for everything else.

---

### Task 1: Project scaffolding

**Files:**
- Create: `goime/go.mod`
- Create: `goime/Makefile`
- Create: `goime/cmd/goimed/main.go`
- Create: `goime/cmd/goime-dict/main.go`

- [ ] **Step 1: Initialize Go module**

Create `go.mod`:

```
module github.com/jiazhoulvke/goime

go 1.26
```

- [ ] **Step 2: Create directory structure**

Run:

```bash
mkdir -p goime/cmd/goimed goime/cmd/goime-dict goime/internal/server goime/internal/protocol goime/internal/pinyin goime/internal/engine goime/internal/dict goime/internal/config goime/configs/schemes goime/dicts
```

- [ ] **Step 3: Create stub main files**

`cmd/goimed/main.go`:

```go
package main

import "fmt"

func main() {
    fmt.Println("goimed: GoIME daemon starting...")
}
```

`cmd/goime-dict/main.go`:

```go
package main

import (
    "flag"
    "fmt"
    "os"
)

func main() {
    importRime := flag.Bool("rime", false, "Import from Rime .dict.yaml format")
    export := flag.Bool("export", false, "Export user dict to plain text")
    importUser := flag.Bool("import", false, "Import user dict from plain text")
    flag.Parse()

    if *importRime {
        fmt.Println("Importing Rime dictionary...")
    } else if *export {
        fmt.Println("Exporting user dictionary...")
    } else if *importUser {
        fmt.Println("Importing user dictionary...")
    } else {
        fmt.Println("Building dictionary index...")
    }
    os.Exit(0)
}
```

- [ ] **Step 4: Create minimal Makefile**

```
.PHONY: build clean

build:
	go build ./cmd/goimed
	go build ./cmd/goime-dict

clean:
	rm -f goimed goime-dict

test:
	go test ./...
```

- [ ] **Step 5: Verify build**

Run:

```bash
cd goime && go build ./cmd/goimed && go build ./cmd/goime-dict && go vet ./...
```

Expected: clean build, no errors.

- [ ] **Step 6: Commit**

```bash
git init && git add -A && git commit -m "feat: scaffold project structure"
```

---

### Task 2: Pinyin syllable table

**Files:**
- Create: `goime/internal/pinyin/syllables.go`
- Create: `goime/internal/pinyin/syllables_test.go`

- [ ] **Step 1: Write the failing test**

```go
package pinyin

import "testing"

func TestIsValidSyllable(t *testing.T) {
    tests := []struct {
        syllable string
        valid    bool
    }{
        {"shu", true},
        {"ru", true},
        {"ni", true},
        {"hao", true},
        {"zhuang", true},
        {"x", false},
        {"sh", false},
        {"", false},
        {"abcde", false},
    }
    for _, tc := range tests {
        got := IsValidSyllable(tc.syllable)
        if got != tc.valid {
            t.Errorf("IsValidSyllable(%q) = %v, want %v", tc.syllable, got, tc.valid)
        }
    }
}

func TestAllSyllablesCount(t *testing.T) {
    n := len(AllSyllables())
    if n < 400 || n > 420 {
        t.Errorf("expected ~410 syllables, got %d", n)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd goime && go test ./internal/pinyin/ -v -count=1
```

Expected: FAIL with "IsValidSyllable not defined".

- [ ] **Step 3: Implement syllable table**

```go
package pinyin

var allSyllables = map[string]bool{
    "a": true, "ai": true, "an": true, "ang": true, "ao": true,
    "ba": true, "bai": true, "ban": true, "bang": true, "bao": true, "bei": true,
    "ben": true, "beng": true, "bi": true, "bian": true, "biao": true, "bie": true,
    "bin": true, "bing": true, "bo": true, "bu": true,
    "ca": true, "cai": true, "can": true, "cang": true, "cao": true, "ce": true,
    "cen": true, "ceng": true, "cha": true, "chai": true, "chan": true, "chang": true,
    "chao": true, "che": true, "chen": true, "cheng": true, "chi": true, "chong": true,
    "chou": true, "chu": true, "chua": true, "chuai": true, "chuan": true, "chuang": true,
    "chui": true, "chun": true, "chuo": true, "ci": true, "cong": true, "cou": true,
    "cu": true, "cuan": true, "cui": true, "cun": true, "cuo": true,
    "da": true, "dai": true, "dan": true, "dang": true, "dao": true, "de": true,
    "den": true, "dei": true, "deng": true, "di": true, "dia": true, "dian": true,
    "diao": true, "die": true, "ding": true, "diu": true, "dong": true, "dou": true,
    "du": true, "duan": true, "dui": true, "dun": true, "duo": true,
    "e": true, "ei": true, "en": true, "eng": true, "er": true,
    "fa": true, "fan": true, "fang": true, "fei": true, "fen": true, "feng": true,
    "fo": true, "fou": true, "fu": true,
    "ga": true, "gai": true, "gan": true, "gang": true, "gao": true, "ge": true,
    "gei": true, "gen": true, "geng": true, "gong": true, "gou": true, "gu": true,
    "gua": true, "guai": true, "guan": true, "guang": true, "gui": true, "gun": true,
    "guo": true,
    "ha": true, "hai": true, "han": true, "hang": true, "hao": true, "he": true,
    "hei": true, "hen": true, "heng": true, "hong": true, "hou": true, "hu": true,
    "hua": true, "huai": true, "huan": true, "huang": true, "hui": true, "hun": true,
    "huo": true,
    "ji": true, "jia": true, "jian": true, "jiang": true, "jiao": true, "jie": true,
    "jin": true, "jing": true, "jiong": true, "jiu": true, "ju": true, "juan": true,
    "jue": true, "jun": true,
    "ka": true, "kai": true, "kan": true, "kang": true, "kao": true, "ke": true,
    "kei": true, "ken": true, "keng": true, "kong": true, "kou": true, "ku": true,
    "kua": true, "kuai": true, "kuan": true, "kuang": true, "kui": true, "kun": true,
    "kuo": true,
    "la": true, "lai": true, "lan": true, "lang": true, "lao": true, "le": true,
    "lei": true, "leng": true, "li": true, "lia": true, "lian": true, "liang": true,
    "liao": true, "lie": true, "lin": true, "ling": true, "liu": true, "lo": true,
    "long": true, "lou": true, "lu": true, "lv": true, "luan": true, "lve": true,
    "lun": true, "luo": true,
    "ma": true, "mai": true, "man": true, "mang": true, "mao": true, "me": true,
    "mei": true, "men": true, "meng": true, "mi": true, "mian": true, "miao": true,
    "mie": true, "min": true, "ming": true, "miu": true, "mo": true, "mou": true,
    "mu": true,
    "na": true, "nai": true, "nan": true, "nang": true, "nao": true, "ne": true,
    "nei": true, "nen": true, "neng": true, "ng": true, "ni": true, "nian": true,
    "niang": true, "niao": true, "nie": true, "nin": true, "ning": true, "niu": true,
    "nong": true, "nou": true, "nu": true, "nv": true, "nuan": true, "nve": true,
    "nuo": true,
    "o": true, "ou": true,
    "pa": true, "pai": true, "pan": true, "pang": true, "pao": true, "pei": true,
    "pen": true, "peng": true, "pi": true, "pian": true, "piao": true, "pie": true,
    "pin": true, "ping": true, "po": true, "pou": true, "pu": true,
    "qi": true, "qia": true, "qian": true, "qiang": true, "qiao": true, "qie": true,
    "qin": true, "qing": true, "qiong": true, "qiu": true, "qu": true, "quan": true,
    "que": true, "qun": true,
    "ran": true, "rang": true, "rao": true, "re": true, "ren": true, "reng": true,
    "ri": true, "rong": true, "rou": true, "ru": true, "ruan": true, "rui": true,
    "run": true, "ruo": true,
    "sa": true, "sai": true, "san": true, "sang": true, "sao": true, "se": true,
    "sen": true, "seng": true, "sha": true, "shai": true, "shan": true, "shang": true,
    "shao": true, "she": true, "shei": true, "shen": true, "sheng": true, "shi": true,
    "shou": true, "shu": true, "shua": true, "shuai": true, "shuan": true, "shuang": true,
    "shui": true, "shun": true, "shuo": true, "si": true, "song": true, "sou": true,
    "su": true, "suan": true, "sui": true, "sun": true, "suo": true,
    "ta": true, "tai": true, "tan": true, "tang": true, "tao": true, "te": true,
    "tei": true, "teng": true, "ti": true, "tian": true, "tiao": true, "tie": true,
    "ting": true, "tong": true, "tou": true, "tu": true, "tuan": true, "tui": true,
    "tun": true, "tuo": true,
    "wa": true, "wai": true, "wan": true, "wang": true, "wei": true, "wen": true,
    "weng": true, "wo": true, "wu": true,
    "xi": true, "xia": true, "xian": true, "xiang": true, "xiao": true, "xie": true,
    "xin": true, "xing": true, "xiong": true, "xiu": true, "xu": true, "xuan": true,
    "xue": true, "xun": true,
    "ya": true, "yan": true, "yang": true, "yao": true, "ye": true, "yi": true,
    "yin": true, "ying": true, "yo": true, "yong": true, "you": true, "yu": true,
    "yuan": true, "yue": true, "yun": true,
    "za": true, "zai": true, "zan": true, "zang": true, "zao": true, "ze": true,
    "zei": true, "zen": true, "zeng": true, "zha": true, "zhai": true, "zhan": true,
    "zhang": true, "zhao": true, "zhe": true, "zhen": true, "zheng": true, "zhi": true,
    "zhong": true, "zhou": true, "zhu": true, "zhua": true, "zhuai": true, "zhuan": true,
    "zhuang": true, "zhui": true, "zhun": true, "zhuo": true, "zi": true, "zong": true,
    "zou": true, "zu": true, "zuan": true, "zui": true, "zun": true, "zuo": true,
}

// IsValidSyllable 检查字符串是否为合法的拼音音节（无声调）
func IsValidSyllable(s string) bool {
    return allSyllables[s]
}

// AllSyllables 返回所有合法拼音音节的列表
func AllSyllables() []string {
    result := make([]string, 0, len(allSyllables))
    for s := range allSyllables {
        result = append(result, s)
    }
    return result
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd goime && go test ./internal/pinyin/ -v -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: add pinyin syllable table"
```

---

### Task 3: Protocol message types

**Files:**
- Create: `goime/internal/protocol/message.go`
- Create: `goime/internal/protocol/message_test.go`

- [ ] **Step 1: Write the failing test**

```go
package protocol

import (
    "encoding/json"
    "testing"
)

func TestRequestMarshal(t *testing.T) {
    req := Request{Method: "input", Key: "u"}
    data, err := json.Marshal(req)
    if err != nil {
        t.Fatal(err)
    }
    var got Request
    if err := json.Unmarshal(data, &got); err != nil {
        t.Fatal(err)
    }
    if got.Method != "input" || got.Key != "u" {
        t.Errorf("roundtrip failed: %+v", got)
    }
}

func TestResponseMarshal(t *testing.T) {
    resp := Response{
        Type: "candidates",
        Candidates: &Candidates{
            List: []Candidate{{Text: "输入", Code: "uuru", Weight: 100}},
            Page: 0, Total: 1,
        },
    }
    data, err := json.Marshal(resp)
    if err != nil {
        t.Fatal(err)
    }
    var got Response
    if err := json.Unmarshal(data, &got); err != nil {
        t.Fatal(err)
    }
    if got.Type != "candidates" || len(got.Candidates.List) != 1 {
        t.Errorf("roundtrip failed: %+v", got)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd goime && go test ./internal/protocol/ -v -count=1
```

Expected: FAIL.

- [ ] **Step 3: Implement protocol types**

```go
package protocol

// Request 客户端 → 服务端请求
type Request struct {
    Method  string `json:"method"`
    Key     string `json:"key,omitempty"`     // input 方法的按键
    Index   int    `json:"index,omitempty"`   // select 方法的索引
    Dir     string `json:"dir,omitempty"`     // arrow/page 的方向
    Name    string `json:"name,omitempty"`    // set_scheme 的方案名
    Version int    `json:"version,omitempty"` // hello 的协议版本
    Client  string `json:"client,omitempty"`  // hello 的客户端标识
}

// Response 服务端 → 客户端响应
type Response struct {
    Type        string       `json:"type"`
    Text        string       `json:"text,omitempty"`         // commit/preedit 的文字
    Pos         int          `json:"pos,omitempty"`          // preedit 的光标位置
    Candidates  *Candidates  `json:"candidates,omitempty"`   // candidates 列表
    PendingKey  string       `json:"pending_key,omitempty"`  // commit 需透传的后续字符
    Version     int          `json:"version,omitempty"`      // welcome 的协议版本
    Schemes     []string     `json:"schemes,omitempty"`      // welcome 的可用方案
    Active      string       `json:"active,omitempty"`       // welcome 的当前方案
    PageSize    int          `json:"page_size,omitempty"`    // welcome 的每页候选数
    Message     string       `json:"message,omitempty"`      // error 的错误信息
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

// NewCandidates 创建候选词响应
func NewCandidates(list []Candidate, page, total int) Response {
    return Response{
        Type: "candidates",
        Candidates: &Candidates{
            List:  list,
            Page:  page,
            Total: total,
        },
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

// NewPreedit 创建预编辑响应
func NewPreedit(text string, pos int) Response {
    return Response{
        Type: "preedit",
        Text: text,
        Pos:  pos,
    }
}

// NewError 创建错误响应
func NewError(msg string) Response {
    return Response{
        Type:    "error",
        Message: msg,
    }
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd goime && go test ./internal/protocol/ -v -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: add protocol message types"
```

---

### Task 4: Config loading

**Files:**
- Create: `goime/internal/config/config.go`
- Create: `goime/internal/config/config_test.go`
- Create: `goime/configs/goime.toml`
- Create: `goime/configs/schemes/xiaohe.toml`
- Create: `goime/configs/schemes/fullpin.toml`

- [ ] **Step 1: Write the failing test**

```go
package config

import (
    "os"
    "testing"
)

func TestLoadConfig(t *testing.T) {
    content := `
[general]
log_level = "debug"
socket_path = "/tmp/test.sock"
idle_timeout = "5m"

[scheme]
active = "xiaohe"
dir = "~/.config/goime/schemes/"

[dict]
static = ["test.dict.txt"]
user = "test.db"
sync_file = "test.txt"
build_dir = "/tmp/goime"
auto_build = true

[candidates]
page_size = 7
max_candidates = 50

[translator]
max_syllables = 6

[user_dict]
enabled = true
freq_decay = true
decay_rate = 0.95
new_word_weight = 200
`
    f, err := os.CreateTemp("", "goime-*.toml")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(f.Name())
    if _, err := f.WriteString(content); err != nil {
        t.Fatal(err)
    }
    f.Close()

    cfg, err := Load(f.Name())
    if err != nil {
        t.Fatalf("Load failed: %v", err)
    }
    if cfg.General.LogLevel != "debug" {
        t.Errorf("LogLevel = %q, want debug", cfg.General.LogLevel)
    }
    if cfg.Scheme.Active != "xiaohe" {
        t.Errorf("Active = %q, want xiaohe", cfg.Scheme.Active)
    }
    if cfg.UserDict.NewWordWeight != 200 {
        t.Errorf("NewWordWeight = %d, want 200", cfg.UserDict.NewWordWeight)
    }
}

func TestDefaultConfig(t *testing.T) {
    cfg := Default()
    if cfg.General.IdleTimeout != "15m" {
        t.Errorf("IdleTimeout = %q, want 15m", cfg.General.IdleTimeout)
    }
    if cfg.Candidates.PageSize != 5 {
        t.Errorf("PageSize = %d, want 5", cfg.Candidates.PageSize)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd goime && go test ./internal/config/ -v -count=1
```

Expected: FAIL.

- [ ] **Step 3: Implement config**

```go
package config

import (
    "os"
    "path/filepath"

    "github.com/BurntSushi/toml"
)

// Config 主配置结构，对应 goime.toml
type Config struct {
    General    General    `toml:"general"`
    Scheme     Scheme     `toml:"scheme"`
    Dict       Dict       `toml:"dict"`
    Candidates Candidates `toml:"candidates"`
    Translator Translator `toml:"translator"`
    UserDict   UserDict   `toml:"user_dict"`
}

// General 通用设置
type General struct {
    LogLevel    string `toml:"log_level"`
    SocketPath  string `toml:"socket_path"`
    IdleTimeout string `toml:"idle_timeout"`
}

// Scheme 输入方案设置
type Scheme struct {
    Active string `toml:"active"`
    Dir    string `toml:"dir"`
}

// Dict 词库设置
type Dict struct {
    Static     []string `toml:"static"`
    User       string   `toml:"user"`
    SyncFile   string   `toml:"sync_file"`
    BuildDir   string   `toml:"build_dir"`
    AutoBuild  bool     `toml:"auto_build"`
}

// Candidates 候选词设置
type Candidates struct {
    PageSize      int `toml:"page_size"`
    MaxCandidates int `toml:"max_candidates"`
}

// Translator 翻译器设置
type Translator struct {
    MaxSyllables int `toml:"max_syllables"`
}

// UserDict 用户词库设置
type UserDict struct {
    Enabled       bool    `toml:"enabled"`
    FreqDecay     bool    `toml:"freq_decay"`
    DecayRate     float64 `toml:"decay_rate"`
    NewWordWeight int     `toml:"new_word_weight"`
}

// Default 返回默认配置
func Default() *Config {
    home, _ := os.UserHomeDir()
    return &Config{
        General: General{
            LogLevel:    "info",
            SocketPath:  "",
            IdleTimeout: "15m",
        },
        Scheme: Scheme{
            Active: "xiaohe",
            Dir:    filepath.Join(home, ".config", "goime", "schemes"),
        },
        Dict: Dict{
            Static:    []string{filepath.Join(home, ".config", "goime", "dicts", "zhonghua.dict.txt")},
            User:      filepath.Join(home, ".config", "goime", "user_dict.db"),
            SyncFile:  filepath.Join(home, ".config", "goime", "user_dict.txt"),
            BuildDir:  filepath.Join(home, ".cache", "goime"),
            AutoBuild: true,
        },
        Candidates: Candidates{
            PageSize:      5,
            MaxCandidates: 100,
        },
        Translator: Translator{
            MaxSyllables: 8,
        },
        UserDict: UserDict{
            Enabled:       true,
            FreqDecay:     true,
            DecayRate:     0.99,
            NewWordWeight: 100,
        },
    }
}

// Load 从文件加载配置，不存在的字段使用默认值
func Load(path string) (*Config, error) {
    cfg := Default()
    if _, err := toml.DecodeFile(path, cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}

// SocketPath 获取实际 socket 路径
func (c *Config) SocketPath() string {
    if c.General.SocketPath != "" {
        return c.General.SocketPath
    }
    if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
        return filepath.Join(dir, "goime.sock")
    }
    return filepath.Join("/tmp", "goime-"+os.Getenv("USER")+".sock")
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd goime && go test ./internal/config/ -v -count=1
```

Expected: PASS.

- [ ] **Step 5: Create default config files**

`configs/goime.toml` — copy from the spec document config section (the annotated version).

`configs/schemes/xiaohe.toml`:

```toml
[speller]
type = "shuangpin"
alphabet = "abcdefghijklmnopqrstuvwxyz"
delimiter = "'"
algebra = [
  "erase/^xx$/",
  "derive/^([jqxy])u$/$1v/",
  "abbrev/^([a-z]).+$/$1/",
  "abbrev/^([zcs]h).+$/$1/",
  "derive/^([zcs])h/$1/",
  "derive/^([zcs])([^h])/$1h$2/",
  "xform/iu$/Q/",
  "xform/(.)ei$/$1W/",
  "xform/uan$/R/",
  "xform/[uv]e$/T/",
  "xform/un$/Y/",
  "xform/^sh/U/",
  "xform/^ch/I/",
  "xform/^zh/V/",
  "xform/uo$/O/",
  "xform/ie$/P/",
  "xform/i?ong$/S/",
  "xform/ing$|uai$/K/",
  "xform/(.)ai$/$1D/",
  "xform/(.)en$/$1F/",
  "xform/(.)eng$/$1G/",
  "xform/[iu]ang$/L/",
  "xform/(.)ang$/$1H/",
  "xform/ian$/M/",
  "xform/(.)an$/$1J/",
  "xform/(.)ou$/$1Z/",
  "xform/^([aoe])([ioun])$/$1$1$2/",
  "xform/^([aoe])(ng)?$/$1$1$2/",
]
```

`configs/schemes/fullpin.toml`:

```toml
[speller]
type = "pinyin"
alphabet = "abcdefghijklmnopqrstuvwxyz"
delimiter = "'"
algebra = []
```

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "feat: add config loading and default configs"
```

---

### Task 5: Dictionary builder

**Files:**
- Create: `goime/internal/dict/dict.go`
- Create: `goime/internal/dict/builder.go`
- Create: `goime/internal/dict/builder_test.go`

- [ ] **Step 1: Write the failing test**

```go
package dict

import (
    "os"
    "testing"
)

func TestBuildAndLoad(t *testing.T) {
    content := "shu1ru4 输入 100\nni3hao3 你好 200\n"
    src, err := os.CreateTemp("", "dict-*.txt")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(src.Name())
    if _, err := src.WriteString(content); err != nil {
        t.Fatal(err)
    }
    src.Close()

    dst, err := os.CreateTemp("", "dict-*.goime")
    if err != nil {
        t.Fatal(err)
    }
    dst.Close()
    defer os.Remove(dst.Name())

    if err := Build(src.Name(), dst.Name()); err != nil {
        t.Fatalf("Build failed: %v", err)
    }

    idx, err := Load(dst.Name())
    if err != nil {
        t.Fatalf("Load failed: %v", err)
    }

    entries := idx.Lookup("shuru")
    if len(entries) == 0 || entries[0].Text != "输入" {
        t.Errorf("expected '输入', got %+v", entries)
    }

    entries = idx.Lookup("nihao")
    if len(entries) == 0 || entries[0].Text != "你好" {
        t.Errorf("expected '你好', got %+v", entries)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd goime && go test ./internal/dict/ -v -count=1
```

Expected: FAIL.

- [ ] **Step 3: Implement dict types and builder**

```go
// dict.go
package dict

import (
    "sort"
)

// Entry 词典条目
type Entry struct {
    Text   string
    Weight int
    Tone   string
}

// Index 词典索引
type Index struct {
    entries []indexEntry
}

type indexEntry struct {
    pinyin string
    word   string
    weight int
    tone   string
}

// Lookup 根据去调拼音查找所有匹配条目
func (idx *Index) Lookup(pinyin string) []Entry {
    lo, hi := 0, len(idx.entries)
    for lo < hi {
        mid := (lo + hi) / 2
        if idx.entries[mid].pinyin < pinyin {
            lo = mid + 1
        } else {
            hi = mid
        }
    }
    if lo >= len(idx.entries) || idx.entries[lo].pinyin != pinyin {
        return nil
    }
    start := lo
    for lo < len(idx.entries) && idx.entries[lo].pinyin == pinyin {
        lo++
    }
    end := lo
    result := make([]Entry, 0, end-start)
    for i := start; i < end; i++ {
        result = append(result, Entry{
            Text:   idx.entries[i].word,
            Weight: idx.entries[i].weight,
            Tone:   idx.entries[i].tone,
        })
    }
    sort.Slice(result, func(i, j int) bool {
        return result[i].Weight > result[j].Weight
    })
    return result
}
```

```go
// builder.go
package dict

import (
    "bufio"
    "encoding/binary"
    "fmt"
    "os"
    "sort"
    "strings"
)

func stripTones(pinyin string) (plain string, tones string) {
    for _, r := range pinyin {
        if r >= '0' && r <= '9' {
            tones += string(r)
        } else {
            plain += string(r)
        }
    }
    return plain, tones
}

// Build 将纯文本词库编译为二进制索引文件
func Build(srcPath, dstPath string) error {
    f, err := os.Open(srcPath)
    if err != nil {
        return fmt.Errorf("open source: %w", err)
    }
    defer f.Close()

    type rawEntry struct {
        plain  string
        tones  string
        word   string
        weight int
    }

    var entries []rawEntry
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        parts := strings.Split(line, "\t")
        if len(parts) < 2 {
            continue
        }
        pinyin := parts[0]
        word := parts[1]
        weight := 0
        if len(parts) >= 3 {
            fmt.Sscanf(parts[2], "%d", &weight)
        }
        plain, tones := stripTones(pinyin)
        entries = append(entries, rawEntry{
            plain:  plain,
            tones:  tones,
            word:   word,
            weight: weight,
        })
    }
    if err := scanner.Err(); err != nil {
        return fmt.Errorf("scan: %w", err)
    }

    sort.Slice(entries, func(i, j int) bool {
        if entries[i].plain != entries[j].plain {
            return entries[i].plain < entries[j].plain
        }
        return entries[i].weight > entries[j].weight
    })

    out, err := os.Create(dstPath)
    if err != nil {
        return fmt.Errorf("create dst: %w", err)
    }
    defer out.Close()

    numEntries := uint32(len(entries))
    if err := binary.Write(out, binary.LittleEndian, numEntries); err != nil {
        return err
    }
    for _, e := range entries {
        pinyinBytes := []byte(e.plain)
        wordBytes := []byte(e.word)
        if err := binary.Write(out, binary.LittleEndian, uint16(len(pinyinBytes))); err != nil {
            return err
        }
        if _, err := out.Write(pinyinBytes); err != nil {
            return err
        }
        if err := binary.Write(out, binary.LittleEndian, uint16(len(wordBytes))); err != nil {
            return err
        }
        if _, err := out.Write(wordBytes); err != nil {
            return err
        }
        if err := binary.Write(out, binary.LittleEndian, int32(e.weight)); err != nil {
            return err
        }
    }
    return nil
}

// Load 从二进制索引文件加载词典
func Load(path string) (*Index, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read index: %w", err)
    }
    if len(data) < 4 {
        return nil, fmt.Errorf("invalid index file: too small")
    }
    numEntries := binary.LittleEndian.Uint32(data[:4])
    pos := 4
    entries := make([]indexEntry, 0, numEntries)
    for i := uint32(0); i < numEntries; i++ {
        if pos+2 > len(data) {
            return nil, fmt.Errorf("truncated at entry %d", i)
        }
        pinyinLen := int(binary.LittleEndian.Uint16(data[pos:]))
        pos += 2
        if pos+pinyinLen > len(data) {
            return nil, fmt.Errorf("truncated pinyin at entry %d", i)
        }
        pinyin := string(data[pos : pos+pinyinLen])
        pos += pinyinLen
        if pos+2 > len(data) {
            return nil, fmt.Errorf("truncated word len at entry %d", i)
        }
        wordLen := int(binary.LittleEndian.Uint16(data[pos:]))
        pos += 2
        if pos+wordLen > len(data) {
            return nil, fmt.Errorf("truncated word at entry %d", i)
        }
        word := string(data[pos : pos+wordLen])
        pos += wordLen
        if pos+4 > len(data) {
            return nil, fmt.Errorf("truncated weight at entry %d", i)
        }
        weight := int(int32(binary.LittleEndian.Uint32(data[pos:])))
        pos += 4
        entries = append(entries, indexEntry{
            pinyin: pinyin,
            word:   word,
            weight: weight,
        })
    }
    return &Index{entries: entries}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd goime && go test ./internal/dict/ -v -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: add dictionary builder and index loader"
```

---

### Task 6: User dict (SQLite)

**Files:**
- Create: `goime/internal/dict/user.go`
- Create: `goime/internal/dict/user_test.go`

- [ ] **Step 1: Write the failing test**

```go
package dict

import (
    "os"
    "testing"
)

func TestUserDict(t *testing.T) {
    f, err := os.CreateTemp("", "user-*.db")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(f.Name())
    f.Close()

    ud, err := OpenUserDict(f.Name())
    if err != nil {
        t.Fatalf("OpenUserDict failed: %v", err)
    }
    defer ud.Close()

    if err := ud.IncFreq("shuru", "输入"); err != nil {
        t.Fatalf("IncFreq failed: %v", err)
    }
    if err := ud.IncFreq("shuru", "输入"); err != nil {
        t.Fatalf("IncFreq failed: %v", err)
    }
    freq, _ := ud.GetFreq("shuru", "输入")
    if freq != 2 {
        t.Errorf("freq = %d, want 2", freq)
    }

    if err := ud.AddUserWord("nihaoshijie", "你好世界", 100); err != nil {
        t.Fatalf("AddUserWord failed: %v", err)
    }
    entries := ud.GetUserWords("nihaoshijie")
    if len(entries) == 0 || entries[0].Text != "你好世界" {
        t.Errorf("expected '你好世界', got %+v", entries)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd goime && go test ./internal/dict/ -v -count=1 -run TestUserDict
```

Expected: FAIL.

- [ ] **Step 3: Implement user dict**

```go
package dict

import (
    "database/sql"
    "fmt"

    _ "modernc.org/sqlite"
)

// UserDict 用户词库（SQLite WAL 模式）
type UserDict struct {
    db *sql.DB
}

// OpenUserDict 打开或创建用户词库
func OpenUserDict(path string) (*UserDict, error) {
    db, err := sql.Open("sqlite", path+"?_journal_mode=WAL")
    if err != nil {
        return nil, fmt.Errorf("open user dict: %w", err)
    }
    if err := migrate(db); err != nil {
        db.Close()
        return nil, fmt.Errorf("migrate: %w", err)
    }
    return &UserDict{db: db}, nil
}

func migrate(db *sql.DB) error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS word_freq (
            pinyin    TEXT,
            word      TEXT,
            frequency INTEGER NOT NULL DEFAULT 0,
            PRIMARY KEY (pinyin, word)
        )`,
        `CREATE TABLE IF NOT EXISTS user_words (
            pinyin    TEXT,
            word      TEXT,
            frequency INTEGER NOT NULL DEFAULT 0,
            created_at INTEGER NOT NULL DEFAULT (unixepoch()),
            PRIMARY KEY (pinyin, word)
        )`,
    }
    for _, q := range queries {
        if _, err := db.Exec(q); err != nil {
            return err
        }
    }
    return nil
}

// Close 关闭用户词库
func (u *UserDict) Close() error {
    return u.db.Close()
}

// IncFreq 增加静态词库词的频率
func (u *UserDict) IncFreq(pinyin, word string) error {
    _, err := u.db.Exec(
        `INSERT INTO word_freq (pinyin, word, frequency) VALUES (?, ?, 1)
         ON CONFLICT(pinyin, word) DO UPDATE SET frequency = frequency + 1`,
        pinyin, word,
    )
    return err
}

// GetFreq 获取静态词库词的频率
func (u *UserDict) GetFreq(pinyin, word string) (int, error) {
    var freq int
    err := u.db.QueryRow(
        `SELECT frequency FROM word_freq WHERE pinyin = ? AND word = ?`,
        pinyin, word,
    ).Scan(&freq)
    if err == sql.ErrNoRows {
        return 0, nil
    }
    return freq, err
}

// AddUserWord 新增或增加自造词频率
func (u *UserDict) AddUserWord(pinyin, word string, weight int) error {
    _, err := u.db.Exec(
        `INSERT INTO user_words (pinyin, word, frequency, created_at) VALUES (?, ?, ?, unixepoch())
         ON CONFLICT(pinyin, word) DO UPDATE SET frequency = frequency + 1`,
        pinyin, word, weight,
    )
    return err
}

// GetUserWords 获取自造词
func (u *UserDict) GetUserWords(pinyin string) []Entry {
    rows, err := u.db.Query(
        `SELECT word, frequency FROM user_words WHERE pinyin = ? ORDER BY frequency DESC`,
        pinyin,
    )
    if err != nil {
        return nil
    }
    defer rows.Close()
    var entries []Entry
    for rows.Next() {
        var word string
        var freq int
        if err := rows.Scan(&word, &freq); err == nil {
            entries = append(entries, Entry{Text: word, Weight: freq})
        }
    }
    return entries
}

// DecayAll 对所有词频进行衰减
func (u *UserDict) DecayAll(rate float64) error {
    _, err := u.db.Exec(
        `UPDATE word_freq SET frequency = CAST(ROUND(frequency * ?) AS INTEGER) WHERE frequency > 0`,
        rate,
    )
    if err != nil {
        return err
    }
    _, err = u.db.Exec(
        `UPDATE user_words SET frequency = CAST(ROUND(frequency * ?) AS INTEGER) WHERE frequency > 0`,
        rate,
    )
    return err
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd goime && go test ./internal/dict/ -v -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: add user dict (SQLite)"
```

---

### Task 7: goime-dict CLI

**Files:**
- Modify: `goime/cmd/goime-dict/main.go`

- [ ] **Step 1: Verify build**

No unit test needed — CLI verified manually.

- [ ] **Step 2: Implement goime-dict main**

```go
package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/jiazhoulvke/goime/internal/dict"
)

func main() {
    importRime := flag.Bool("rime", false, "Import from Rime .dict.yaml")
    export := flag.Bool("export", false, "Export user dict to plain text")
    importUser := flag.Bool("import", false, "Import user dict from plain text")
    src := flag.String("src", "", "Source file path")
    dst := flag.String("dst", "", "Destination file path")
    flag.Parse()

    if *importRime {
        if *src == "" || *dst == "" {
            fmt.Fprintln(os.Stderr, "Usage: goime-dict --rime --src input.dict.yaml --dst output.dict.txt")
            os.Exit(1)
        }
        fmt.Fprintf(os.Stderr, "Error: Rime import not yet implemented\n")
        os.Exit(1)
    }
    if *export {
        fmt.Fprintf(os.Stderr, "Error: export not yet implemented\n")
        os.Exit(1)
    }
    if *importUser {
        fmt.Fprintf(os.Stderr, "Error: import not yet implemented\n")
        os.Exit(1)
    }
    if *src == "" || *dst == "" {
        fmt.Fprintln(os.Stderr, "Usage: goime-dict --src dict.txt --dst dict.goime")
        os.Exit(1)
    }
    if err := dict.Build(*src, *dst); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Dictionary built: %s → %s\n", *src, *dst)
}
```

- [ ] **Step 3: Verify build**

```bash
cd goime && go build ./cmd/goime-dict
```

Expected: clean build.

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat: implement goime-dict CLI"
```

---

### Task 8: Speller with algebra rules

**Files:**
- Create: `goime/internal/engine/speller.go`
- Create: `goime/internal/engine/speller_test.go`

- [ ] **Step 1: Write the failing test**

```go
package engine

import (
    "testing"
)

func TestShuangpinSpeller(t *testing.T) {
    s, err := NewSpeller("shuangpin", nil)
    if err != nil {
        t.Fatalf("NewSpeller failed: %v", err)
    }
    result := s.ToPinyin("uuru")
    if len(result) == 0 || result[0] != "shuru" {
        t.Errorf("uuru → shuru, got %v", result)
    }
}

func TestFullPinyinSpeller(t *testing.T) {
    s, err := NewSpeller("pinyin", nil)
    if err != nil {
        t.Fatalf("NewSpeller failed: %v", err)
    }
    result := s.ToPinyin("shuru")
    if len(result) != 1 || result[0] != "shuru" {
        t.Errorf("full pinyin: expected ['shuru'], got %v", result)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd goime && go test ./internal/engine/ -v -count=1
```

Expected: FAIL.

- [ ] **Step 3: Implement Speller**

```go
package engine

import (
    "fmt"
    "regexp"
)

type algebraRule struct {
    pattern *regexp.Regexp
    repl    string
    typ     string
}

// Speller 将用户按键转换为拼音音节序列
type Speller struct {
    typ    string
    lookup map[string][]string
}

// NewSpeller 创建 Speller
func NewSpeller(typ string, rules []string) (*Speller, error) {
    s := &Speller{typ: typ}
    if typ == "pinyin" {
        return s, nil
    }
    s.lookup = buildShuangpinLookup()
    return s, nil
}

// ToPinyin 将输入码转换为拼音音节序列
func (s *Speller) ToPinyin(code string) []string {
    if s.typ == "pinyin" {
        return []string{code}
    }
    if s.lookup != nil {
        if p, ok := s.lookup[code]; ok {
            return p
        }
    }
    return nil
}

func buildShuangpinLookup() map[string][]string {
    return map[string][]string{
        "aa": {"a"}, "ai": {"ai"}, "an": {"an"}, "ah": {"ang"}, "ao": {"ao"},
        "ba": {"ba"}, "bd": {"bai"}, "bj": {"ban"}, "bh": {"bang"}, "bk": {"bao"},
        "bw": {"bei"}, "bf": {"ben"}, "bg": {"beng"}, "bi": {"bi"},
        "bm": {"bian"}, "bc": {"biao"}, "bp": {"bie"}, "bl": {"bin"},
        "by": {"bing"}, "bo": {"bo"}, "bu": {"bu"},
        "ca": {"ca"}, "cd": {"cai"}, "cj": {"can"}, "ch": {"cang"}, "ck": {"cao"},
        "ce": {"ce"}, "cf": {"cen"}, "cg": {"ceng"},
        "ua": {"cha"}, "ud": {"chai"}, "uj": {"chan"}, "uh": {"chang"}, "uk": {"chao"},
        "ue": {"che"}, "uf": {"chen"}, "ug": {"cheng"}, "ui": {"chi"},
        "us": {"chong"}, "uz": {"chou"}, "uu": {"chu"}, "uc": {"chua"},
        "ux": {"chuai"}, "ur": {"chuan"}, "ul": {"chuang"}, "uv": {"chui"},
        "uy": {"chun"}, "uo": {"chuo"}, "ci": {"ci"}, "cs": {"cong"},
        "cz": {"cou"}, "cu": {"cu"}, "cr": {"cuan"}, "cv": {"cui"},
        "cy": {"cun"}, "co": {"cuo"},
        "da": {"da"}, "dd": {"dai"}, "dj": {"dan"}, "dh": {"dang"}, "dk": {"dao"},
        "de": {"de"}, "dg": {"deng"}, "di": {"di"},
        "dm": {"dian"}, "dc": {"diao"}, "dp": {"die"}, "dy": {"ding"},
        "dn": {"diu"}, "ds": {"dong"}, "dz": {"dou"}, "du": {"du"},
        "dr": {"duan"}, "dv": {"dui"}, "dy": {"dun"}, "do": {"duo"},
        "ee": {"e"}, "ei": {"ei"}, "en": {"en"}, "eg": {"eng"}, "er": {"er"},
        "fa": {"fa"}, "fj": {"fan"}, "fh": {"fang"}, "fw": {"fei"},
        "ff": {"fen"}, "fg": {"feng"}, "fo": {"fo"}, "fz": {"fou"}, "fu": {"fu"},
        "ga": {"ga"}, "gd": {"gai"}, "gj": {"gan"}, "gh": {"gang"}, "gk": {"gao"},
        "ge": {"ge"}, "gw": {"gei"}, "gf": {"gen"}, "gg": {"geng"},
        "gs": {"gong"}, "gz": {"gou"}, "gu": {"gu"},
        "gc": {"gua"}, "gx": {"guai"}, "gr": {"guan"}, "gl": {"guang"},
        "gv": {"gui"}, "gy": {"gun"}, "go": {"guo"},
        "ha": {"ha"}, "hd": {"hai"}, "hj": {"han"}, "hh": {"hang"}, "hk": {"hao"},
        "he": {"he"}, "hw": {"hei"}, "hf": {"hen"}, "hg": {"heng"},
        "hs": {"hong"}, "hz": {"hou"}, "hu": {"hu"},
        "hc": {"hua"}, "hx": {"huai"}, "hr": {"huan"}, "hl": {"huang"},
        "hv": {"hui"}, "hy": {"hun"}, "ho": {"huo"},
        "ji": {"ji"}, "jb": {"jia"}, "jm": {"jian"}, "jl": {"jiang"}, "jc": {"jiao"},
        "jp": {"jie"}, "jl": {"jin"}, "jy": {"jing"}, "js": {"jiong"},
        "jq": {"jiu"}, "jv": {"ju"}, "jr": {"juan"}, "jt": {"jue"},
        "jy": {"jun"},
        "ka": {"ka"}, "kd": {"kai"}, "kj": {"kan"}, "kh": {"kang"}, "kk": {"kao"},
        "ke": {"ke"}, "kw": {"kei"}, "kf": {"ken"}, "kg": {"keng"},
        "ks": {"kong"}, "kz": {"kou"}, "ku": {"ku"},
        "kc": {"kua"}, "kx": {"kuai"}, "kr": {"kuan"}, "kl": {"kuang"},
        "kv": {"kui"}, "ky": {"kun"}, "ko": {"kuo"},
        "la": {"la"}, "ld": {"lai"}, "lj": {"lan"}, "lh": {"lang"}, "lk": {"lao"},
        "le": {"le"}, "lw": {"lei"}, "lg": {"leng"},
        "li": {"li"}, "lb": {"lia"}, "lm": {"lian"}, "ll": {"liang"}, "lc": {"liao"},
        "lp": {"lie"}, "ll": {"lin"}, "ly": {"ling"}, "ln": {"liu"},
        "ls": {"long"}, "lz": {"lou"}, "lu": {"lu"}, "lv": {"lv"},
        "lr": {"luan"}, "lt": {"lve"}, "ly": {"lun"}, "lo": {"luo"},
        "ma": {"ma"}, "md": {"mai"}, "mj": {"man"}, "mh": {"mang"}, "mk": {"mao"},
        "me": {"me"}, "mw": {"mei"}, "mf": {"men"}, "mg": {"meng"},
        "mi": {"mi"}, "mm": {"mian"}, "mc": {"miao"}, "mp": {"mie"},
        "ml": {"min"}, "my": {"ming"}, "mn": {"miu"},
        "mo": {"mo"}, "mz": {"mou"}, "mu": {"mu"},
        "na": {"na"}, "nd": {"nai"}, "nj": {"nan"}, "nh": {"nang"}, "nk": {"nao"},
        "ne": {"ne"}, "nw": {"nei"}, "nf": {"nen"}, "ng": {"neng"},
        "ni": {"ni"}, "nm": {"nian"}, "nl": {"niang"}, "nc": {"niao"}, "np": {"nie"},
        "nl": {"nin"}, "ny": {"ning"}, "nn": {"niu"},
        "ns": {"nong"}, "nz": {"nou"}, "nu": {"nu"}, "nv": {"nv"},
        "nr": {"nuan"}, "nt": {"nve"}, "no": {"nuo"},
        "oo": {"o"}, "oz": {"ou"},
        "pa": {"pa"}, "pd": {"pai"}, "pj": {"pan"}, "ph": {"pang"}, "pk": {"pao"},
        "pw": {"pei"}, "pf": {"pen"}, "pg": {"peng"},
        "pi": {"pi"}, "pm": {"pian"}, "pc": {"biao"}, "pp": {"pie"},
        "pl": {"pin"}, "py": {"ping"},
        "po": {"po"}, "pz": {"pou"}, "pu": {"pu"},
        "qi": {"qi"}, "qb": {"qia"}, "qm": {"qian"}, "ql": {"qiang"}, "qc": {"qiao"},
        "qp": {"qie"}, "ql": {"qin"}, "qy": {"qing"}, "qs": {"qiong"},
        "qq": {"qiu"}, "qv": {"qu"}, "qr": {"quan"}, "qt": {"que"},
        "qy": {"qun"},
        "rj": {"ran"}, "rh": {"rang"}, "rk": {"rao"}, "re": {"re"},
        "rf": {"ren"}, "rg": {"reng"}, "ri": {"ri"},
        "rs": {"rong"}, "rz": {"rou"}, "ru": {"ru"},
        "rr": {"ruan"}, "rv": {"rui"}, "ry": {"run"}, "ro": {"ruo"},
        "sa": {"sa"}, "sd": {"sai"}, "sj": {"san"}, "sh": {"sang"}, "sk": {"sao"},
        "se": {"se"}, "sf": {"sen"}, "sg": {"seng"},
        "si": {"si"}, "ss": {"song"}, "sz": {"sou"}, "su": {"su"},
        "sr": {"suan"}, "sv": {"sui"}, "sy": {"sun"}, "so": {"suo"},
        "ta": {"ta"}, "td": {"tai"}, "tj": {"tan"}, "th": {"tang"}, "tk": {"tao"},
        "te": {"te"}, "tg": {"teng"},
        "ti": {"ti"}, "tm": {"tian"}, "tc": {"tiao"}, "tp": {"tie"},
        "ty": {"ting"}, "ts": {"tong"}, "tz": {"tou"}, "tu": {"tu"},
        "tr": {"tuan"}, "tv": {"tui"}, "ty": {"tun"}, "to": {"tuo"},
        "wa": {"wa"}, "wd": {"wai"}, "wj": {"wan"}, "wh": {"wang"}, "ww": {"wei"},
        "wf": {"wen"}, "wg": {"weng"}, "wo": {"wo"}, "wu": {"wu"},
        "xi": {"xi"}, "xb": {"xia"}, "xm": {"xian"}, "xl": {"xiang"}, "xc": {"xiao"},
        "xp": {"xie"}, "xl": {"xin"}, "xy": {"xing"}, "xs": {"xiong"},
        "xq": {"xiu"}, "xv": {"xu"}, "xr": {"xuan"}, "xt": {"xue"},
        "xy": {"xun"},
        "ya": {"ya"}, "yj": {"yan"}, "yh": {"yang"}, "yk": {"yao"},
        "ye": {"ye"}, "yi": {"yi"}, "yl": {"yin"}, "yy": {"ying"},
        "yo": {"yo"}, "ys": {"yong"}, "yz": {"you"}, "yv": {"yu"},
        "yr": {"yuan"}, "yt": {"yue"}, "yy": {"yun"},
        "za": {"za"}, "zd": {"zai"}, "zj": {"zan"}, "zh": {"zang"}, "zk": {"zao"},
        "ze": {"ze"}, "zw": {"zei"}, "zf": {"zen"}, "zg": {"zeng"},
        "va": {"zha"}, "vd": {"zhai"}, "vj": {"zhan"}, "vh": {"zhang"}, "vk": {"zhao"},
        "ve": {"zhe"}, "vf": {"zhen"}, "vg": {"zheng"}, "vi": {"zhi"},
        "vs": {"zhong"}, "vz": {"zhou"}, "vu": {"zhu"},
        "vc": {"zhua"}, "vx": {"zhuai"}, "vr": {"zhuan"}, "vl": {"zhuang"},
        "vv": {"zhui"}, "vy": {"zhun"}, "vo": {"zhuo"},
        "zi": {"zi"}, "zs": {"zong"}, "zz": {"zou"}, "zu": {"zu"},
        "zr": {"zuan"}, "zv": {"zui"}, "zy": {"zun"}, "zo": {"zuo"},
    }
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd goime && go test ./internal/engine/ -v -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: implement speller with Xiaohe lookup table"
```

---

### Task 9: Segmentor

**Files:**
- Create: `goime/internal/engine/segmentor.go`
- Create: `goime/internal/engine/segmentor_test.go`

- [ ] **Step 1: Write the failing test**

```go
package engine

import (
    "reflect"
    "testing"
)

func TestSegment(t *testing.T) {
    tests := []struct {
        input string
        want  [][]string
    }{
        {"shuru", [][]string{{"shu", "ru"}}},
        {"nihao", [][]string{{"ni", "hao"}}},
        {"xian", [][]string{{"xi", "an"}, {"xian"}}},
        {"fangan", [][]string{{"fang", "an"}, {"fan", "gan"}}},
        {"a", [][]string{{"a"}}},
        {"", [][]string{}},
        {"xx", [][]string{}},
    }
    for _, tc := range tests {
        got := Segment(tc.input)
        if !reflect.DeepEqual(got, tc.want) {
            t.Errorf("Segment(%q) = %v, want %v", tc.input, got, tc.want)
        }
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd goime && go test ./internal/engine/ -v -count=1 -run TestSegment
```

Expected: FAIL.

- [ ] **Step 3: Implement Segmentor**

```go
package engine

import "github.com/jiazhoulvke/goime/internal/pinyin"

// Segment 将拼音字符串按所有合法音节边界切分
func Segment(s string) [][]string {
    if s == "" {
        return nil
    }
    var result [][]string
    backtrack(s, 0, nil, &result)
    return result
}

func backtrack(s string, start int, cur []string, result *[][]string) {
    if start >= len(s) {
        seg := make([]string, len(cur))
        copy(seg, cur)
        *result = append(*result, seg)
        return
    }
    for end := start + 1; end <= len(s) && end-start <= 6; end++ {
        syl := s[start:end]
        if pinyin.IsValidSyllable(syl) {
            backtrack(s, end, append(cur, syl), result)
        }
    }
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd goime && go test ./internal/engine/ -v -count=1 -run TestSegment
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: implement segmentor with backtracking"
```

---

### Task 10: Translator with user word support

**Files:**
- Create: `goime/internal/engine/translator.go`
- Create: `goime/internal/engine/translator_test.go`

- [ ] **Step 1: Write the failing test**

```go
package engine

import (
    "os"
    "testing"

    "github.com/jiazhoulvke/goime/internal/dict"
)

func TestTranslatorPhrase(t *testing.T) {
    content := "shu1ru4 输入 100\nni3hao3 你好 200\n"
    src, err := os.CreateTemp("", "dict-*.txt")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(src.Name())
    if _, err := src.WriteString(content); err != nil {
        t.Fatal(err)
    }
    src.Close()

    dst, err := os.CreateTemp("", "dict-*.goime")
    if err != nil {
        t.Fatal(err)
    }
    dst.Close()
    defer os.Remove(dst.Name())

    if err := dict.Build(src.Name(), dst.Name()); err != nil {
        t.Fatalf("Build failed: %v", err)
    }
    idx, err := dict.Load(dst.Name())
    if err != nil {
        t.Fatalf("Load failed: %v", err)
    }

    tr := NewTranslator(idx, nil, 8)
    candidates := tr.Query([]string{"shu", "ru"})
    found := false
    for _, c := range candidates {
        if c.Text == "输入" {
            found = true
            break
        }
    }
    if !found {
        t.Errorf("expected '输入' in 'shuru' candidates, got %+v", candidates)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd goime && go test ./internal/engine/ -v -count=1 -run TestTranslator
```

Expected: FAIL.

- [ ] **Step 3: Implement Translator**

```go
package engine

import (
    "sort"

    "github.com/jiazhoulvke/goime/internal/dict"
)

// Translator 候选词生成器
type Translator struct {
    static       *dict.Index
    user         *dict.UserDict
    maxSyllables int
    selections   []selection
}

type selection struct {
    Pinyin string
    Word   string
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

    seen := make(map[string]bool)
    var results []dict.Entry

    addEntry := func(e dict.Entry) {
        if t.user != nil {
            freq, _ := t.user.GetFreq(e.Text, e.Text)
            e.Weight += freq
        }
        if !seen[e.Text] {
            seen[e.Text] = true
            results = append(results, e)
        }
    }

    // 1. User words for full pinyin
    fullPinyin := ""
    for _, s := range syllables {
        fullPinyin += s
    }
    if t.user != nil {
        for _, e := range t.user.GetUserWords(fullPinyin) {
            addEntry(e)
        }
    }

    // 2. Single syllables
    for _, syl := range syllables {
        for _, e := range t.static.Lookup(syl) {
            addEntry(e)
        }
    }

    // 3. Multi-syllable phrases
    for length := 2; length <= len(syllables) && length <= t.maxSyllables; length++ {
        for i := 0; i+length <= len(syllables); i++ {
            pinyin := ""
            for j := i; j < i+length; j++ {
                pinyin += syllables[j]
            }
            for _, e := range t.static.Lookup(pinyin) {
                addEntry(e)
            }
        }
    }

    sort.Slice(results, func(i, j int) bool {
        return results[i].Weight > results[j].Weight
    })
    return results
}

// AppendSelection 追加选词历史（用于自造词）
func (t *Translator) AppendSelection(pinyin, word string) {
    t.selections = append(t.selections, selection{Pinyin: pinyin, Word: word})
}

// ClearSelections 清空选词历史
func (t *Translator) ClearSelections() {
    t.selections = nil
}

// Selections 获取当前选词历史
func (t *Translator) Selections() []selection {
    return t.selections
}

// CommitSelections 合并选词历史写入用户词库（仅包含多个词时生效）
func (t *Translator) CommitSelections(weight int) {
    if len(t.selections) < 2 || t.user == nil {
        t.ClearSelections()
        return
    }
    pinyin := ""
    word := ""
    for _, s := range t.selections {
        pinyin += s.Pinyin
        word += s.Word
    }
    t.user.AddUserWord(pinyin, word, weight)
    t.ClearSelections()
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd goime && go test ./internal/engine/ -v -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: implement translator with user word support"
```

---

### Task 11: Session manager

**Files:**
- Create: `goime/internal/server/session.go`
- Create: `goime/internal/server/session_test.go`

- [ ] **Step 1: Write the failing test**

```go
package server

import "testing"

func TestSessionAppendSelection(t *testing.T) {
    s := NewSession("xiaohe")
    s.AppendSelection("nihao", "你好")
    s.AppendSelection("shijie", "世界")
    if len(s.Selections()) != 2 {
        t.Errorf("expected 2 selections, got %d", len(s.Selections()))
    }
}

func TestSessionReset(t *testing.T) {
    s := NewSession("xiaohe")
    s.Append("a")
    s.Append("b")
    if s.Buffer() != "ab" {
        t.Errorf("buffer = %q, want ab", s.Buffer())
    }
    s.Reset()
    if s.Buffer() != "" {
        t.Errorf("buffer should be empty after reset")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd goime && go test ./internal/server/ -v -count=1
```

Expected: FAIL.

- [ ] **Step 3: Implement Session**

```go
package server

// Selection 选词历史条目
type Selection struct {
    Pinyin string
    Word   string
}

// Session 单个编辑器连接的状态
type Session struct {
    buffer     string
    scheme     string
    selections []Selection
}

// NewSession 创建新 session
func NewSession(scheme string) *Session {
    return &Session{scheme: scheme}
}

func (s *Session) Buffer() string          { return s.buffer }
func (s *Session) Scheme() string          { return s.scheme }
func (s *Session) Selections() []Selection { return s.selections }

func (s *Session) SetScheme(name string) {
    s.scheme = name
    s.Clear()
}

func (s *Session) Append(key string) {
    s.buffer += key
}

func (s *Session) Backspace() {
    if len(s.buffer) > 0 {
        s.buffer = s.buffer[:len(s.buffer)-1]
    }
}

func (s *Session) Clear() {
    s.buffer = ""
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
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd goime && go test ./internal/server/ -v -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: implement session manager"
```

---

### Task 12: Unix socket server

**Files:**
- Create: `goime/internal/server/server.go`
- Create: `goime/internal/server/server_test.go`
- Create: `goime/internal/engine/engine.go`

- [ ] **Step 1: Write the failing test**

```go
package server

import (
    "encoding/json"
    "net"
    "os"
    "testing"
    "time"

    "github.com/jiazhoulvke/goime/internal/config"
    "github.com/jiazhoulvke/goime/internal/dict"
    "github.com/jiazhoulvke/goime/internal/protocol"
)

func TestServerHandshake(t *testing.T) {
    dir, err := os.MkdirTemp("", "goime-test-*")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(dir)

    socketPath := dir + "/goime.sock"
    cfg := config.Default()
    cfg.General.SocketPath = socketPath

    srv, err := New(cfg, nil, nil, []string{"xiaohe", "fullpin"})
    if err != nil {
        t.Fatalf("New failed: %v", err)
    }
    go srv.Listen()
    defer srv.Close()
    time.Sleep(100 * time.Millisecond)

    conn, err := net.Dial("unix", socketPath)
    if err != nil {
        t.Fatalf("Dial failed: %v", err)
    }
    defer conn.Close()

    req := protocol.Request{Method: "hello", Version: 1, Client: "test"}
    if err := json.NewEncoder(conn).Encode(req); err != nil {
        t.Fatalf("Encode failed: %v", err)
    }
    var resp protocol.Response
    if err := json.NewDecoder(conn).Decode(&resp); err != nil {
        t.Fatalf("Decode failed: %v", err)
    }
    if resp.Type != "welcome" {
        t.Errorf("expected welcome, got %s", resp.Type)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd goime && go test ./internal/server/ -v -count=1
```

Expected: FAIL.

- [ ] **Step 3: Implement server**

```go
package server

import (
    "bufio"
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
    static     *dict.Index
    user       *dict.UserDict
    translator *engine.Translator
    ln         net.Listener
    mu         sync.Mutex
    sessions   map[net.Conn]*Session
    lastActive time.Time
}

// New 创建服务器
func New(cfg *config.Config, static *dict.Index, user *dict.UserDict, schemes []string) (*Server, error) {
    tr := engine.NewTranslator(static, user, cfg.Translator.MaxSyllables)
    return &Server{
        cfg:        cfg,
        schemes:    schemes,
        static:     static,
        user:       user,
        translator: tr,
        sessions:   make(map[net.Conn]*Session),
        lastActive: time.Now(),
    }, nil
}

// Listen 监听 Unix Socket
func (s *Server) Listen() error {
    if conn, err := net.Dial("unix", s.cfg.SocketPath()); err == nil {
        conn.Close()
        return fmt.Errorf("socket %s already in use", s.cfg.SocketPath())
    }
    os.Remove(s.cfg.SocketPath())

    var err error
    s.ln, err = net.Listen("unix", s.cfg.SocketPath())
    if err != nil {
        return fmt.Errorf("listen: %w", err)
    }
    if err := os.Chmod(s.cfg.SocketPath(), 0600); err != nil {
        return fmt.Errorf("chmod: %w", err)
    }

    for {
        conn, err := s.ln.Accept()
        if err != nil {
            return err
        }
        s.mu.Lock()
        session := NewSession(s.cfg.Scheme.Active)
        s.sessions[conn] = session
        s.mu.Unlock()
        go s.handleConn(conn, session)
    }
}

// Close 关闭服务器
func (s *Server) Close() {
    if s.ln != nil {
        s.ln.Close()
    }
    os.Remove(s.cfg.SocketPath())
}

func (s *Server) handleConn(conn net.Conn, session *Session) {
    defer func() {
        conn.Close()
        s.mu.Lock()
        delete(s.sessions, conn)
        s.mu.Unlock()
        s.translator.CommitSelections(s.cfg.UserDict.NewWordWeight)
    }()

    scanner := bufio.NewScanner(conn)
    scanner.Buffer(make([]byte, 4096), 8192)

    for scanner.Scan() {
        line := scanner.Text()
        var req protocol.Request
        if err := json.Unmarshal([]byte(line), &req); err != nil {
            s.send(conn, protocol.NewError("invalid JSON: "+err.Error()))
            continue
        }
        s.mu.Lock()
        resp := s.handleRequest(req, session)
        s.mu.Unlock()

        if err := s.send(conn, resp); err != nil {
            return
        }
        if req.Method == "close" {
            return
        }
    }
}

func (s *Server) handleRequest(req protocol.Request, session *Session) protocol.Response {
    switch req.Method {
    case "hello":
        return protocol.NewWelcome(1, s.schemes, session.Scheme(), s.cfg.Candidates.PageSize)
    case "input":
        if len(req.Key) != 1 || req.Key[0] < 'a' || req.Key[0] > 'z' {
            return protocol.NewError("input only accepts a-z")
        }
        session.Append(req.Key)
        return s.buildInputResponse(session)
    case "enter":
        text := session.Buffer()
        session.Reset()
        s.translator.ClearSelections()
        if text == "" {
            return protocol.NewIdle()
        }
        return protocol.NewCommit(text, "")
    case "escape":
        session.Reset()
        s.translator.ClearSelections()
        return protocol.NewIdle()
    case "backspace":
        session.Backspace()
        return s.buildInputResponse(session)
    case "space":
        if session.Buffer() == "" {
            return protocol.NewIdle()
        }
        return s.handleSelect(session, 0)
    case "select":
        return s.handleSelect(session, req.Index)
    case "commit_preedit":
        text := session.Buffer()
        session.Reset()
        s.translator.ClearSelections()
        return protocol.NewCommit(text, "")
    case "set_scheme":
        session.SetScheme(req.Name)
        return protocol.NewWelcome(1, s.schemes, session.Scheme(), s.cfg.Candidates.PageSize)
    default:
        return protocol.NewError("unknown method: " + req.Method)
    }
}

func (s *Server) handleSelect(session *Session, index int) protocol.Response {
    buffer := session.Buffer()
    if buffer == "" {
        return protocol.NewIdle()
    }
    session.Clear()
    session.AppendSelection(buffer, buffer)
    s.translator.AppendSelection(buffer, buffer)
    return protocol.NewCommit(buffer, "")
}

func (s *Server) buildInputResponse(session *Session) protocol.Response {
    buffer := session.Buffer()
    if buffer == "" {
        return protocol.NewIdle()
    }
    return protocol.NewPreedit(buffer, len(buffer))
}

func (s *Server) send(conn net.Conn, resp protocol.Response) error {
    data, err := json.Marshal(resp)
    if err != nil {
        return err
    }
    data = append(data, '\n')
    _, err = conn.Write(data)
    return err
}

func (s *Server) cleanIdle() {
    timeout, _ := time.ParseDuration(s.cfg.General.IdleTimeout)
    if timeout > 0 && time.Since(s.lastActive) > timeout {
        slog.Info("shutting down due to idle timeout")
        s.Close()
    }
}
```

- [ ] **Step 4: Create engine.go**

```go
package engine

// Engine 输入法引擎主循环（MVP 阶段由 Server 直接驱动）
type Engine struct {
    Speller  *Speller
    Transltr *Translator
}

// NewEngine 创建输入法引擎
func NewEngine(speller *Speller, translator *Translator) *Engine {
    return &Engine{
        Speller:  speller,
        Transltr: translator,
    }
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
cd goime && go test ./internal/server/ -v -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "feat: implement Unix socket server"
```

---

### Task 13: goimed entry point

**Files:**
- Modify: `goime/cmd/goimed/main.go`

- [ ] **Step 1: Implement goimed main**

```go
package main

import (
    "flag"
    "fmt"
    "log/slog"
    "os"
    "path/filepath"

    "github.com/jiazhoulvke/goime/internal/config"
    "github.com/jiazhoulvke/goime/internal/dict"
    "github.com/jiazhoulvke/goime/internal/server"
)

func main() {
    configPath := flag.String("config", "", "Path to config (default: ~/.config/goime/goime.toml)")
    flag.Parse()

    var cfg *config.Config
    if *configPath != "" {
        var err error
        cfg, err = config.Load(*configPath)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    } else {
        home, _ := os.UserHomeDir()
        cfgPath := filepath.Join(home, ".config", "goime", "goime.toml")
        var err error
        cfg, err = config.Load(cfgPath)
        if err != nil {
            cfg = config.Default()
        }
    }

    logLevel := &slog.LevelVar{}
    switch cfg.General.LogLevel {
    case "debug":
        logLevel.Set(slog.LevelDebug)
    case "warn":
        logLevel.Set(slog.LevelWarn)
    case "error":
        logLevel.Set(slog.LevelError)
    default:
        logLevel.Set(slog.LevelInfo)
    }
    slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))
    slog.Info("starting goimed", "socket", cfg.SocketPath())

    var idx *dict.Index
    for _, dictPath := range cfg.Dict.Static {
        srcInfo, err1 := os.Stat(dictPath)
        if err1 != nil {
            continue
        }
        dstPath := filepath.Join(cfg.Dict.BuildDir, filepath.Base(dictPath)+".goime")
        if cfg.Dict.AutoBuild {
            dstInfo, err2 := os.Stat(dstPath)
            if err2 != nil || srcInfo.ModTime().After(dstInfo.ModTime()) {
                slog.Info("building dictionary", "src", dictPath, "dst", dstPath)
                if err := dict.Build(dictPath, dstPath); err != nil {
                    slog.Error("build failed", "error", err)
                    os.Exit(1)
                }
            }
        }
        var err error
        idx, err = dict.Load(dstPath)
        if err != nil {
            slog.Error("load failed", "error", err)
            os.Exit(1)
        }
        break
    }

    var userDict *dict.UserDict
    if cfg.UserDict.Enabled {
        var err error
        userDict, err = dict.OpenUserDict(cfg.Dict.User)
        if err != nil {
            slog.Error("open user dict failed", "error", err)
            os.Exit(1)
        }
        defer userDict.Close()
        if cfg.UserDict.FreqDecay {
            userDict.DecayAll(cfg.UserDict.DecayRate)
        }
    }

    schemes := []string{"xiaohe", "fullpin"}
    srv, err := server.New(cfg, idx, userDict, schemes)
    if err != nil {
        slog.Error("create server failed", "error", err)
        os.Exit(1)
    }
    if err := srv.Listen(); err != nil {
        slog.Error("server error", "error", err)
        os.Exit(1)
    }
}
```

- [ ] **Step 2: Verify build**

```bash
cd goime && go build ./cmd/goimed
```

Expected: clean build.

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat: implement goimed entry point"
```

---

### Task 14: Integration test

**Files:**
- Create: `goime/integration_test.go`

- [ ] **Step 1: Write integration test**

```go
package goime

import (
    "encoding/json"
    "net"
    "os"
    "path/filepath"
    "testing"
    "time"

    "github.com/jiazhoulvke/goime/internal/config"
    "github.com/jiazhoulvke/goime/internal/dict"
    "github.com/jiazhoulvke/goime/internal/protocol"
    "github.com/jiazhoulvke/goime/internal/server"
)

func TestIntegrationFullFlow(t *testing.T) {
    dir, err := os.MkdirTemp("", "goime-test-*")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(dir)

    dictContent := "shu1ru4 输入 100\nni3hao3 你好 200\nshi4jie4 世界 150\n"
    dictFile := filepath.Join(dir, "test.dict.txt")
    if err := os.WriteFile(dictFile, []byte(dictContent), 0644); err != nil {
        t.Fatal(err)
    }
    indexFile := filepath.Join(dir, "test.goime")
    if err := dict.Build(dictFile, indexFile); err != nil {
        t.Fatalf("Build: %v", err)
    }
    idx, err := dict.Load(indexFile)
    if err != nil {
        t.Fatalf("Load: %v", err)
    }

    userDB := filepath.Join(dir, "user.db")
    user, err := dict.OpenUserDict(userDB)
    if err != nil {
        t.Fatalf("OpenUserDict: %v", err)
    }
    defer user.Close()

    cfg := config.Default()
    cfg.General.SocketPath = filepath.Join(dir, "goime.sock")

    srv, err := server.New(cfg, idx, user, []string{"xiaohe", "fullpin"})
    if err != nil {
        t.Fatalf("New: %v", err)
    }
    go srv.Listen()
    defer srv.Close()
    time.Sleep(100 * time.Millisecond)

    conn, err := net.Dial("unix", cfg.General.SocketPath)
    if err != nil {
        t.Fatalf("Dial: %v", err)
    }
    defer conn.Close()

    send := func(req protocol.Request) protocol.Response {
        if err := json.NewEncoder(conn).Encode(req); err != nil {
            t.Fatalf("Encode: %v", err)
        }
        var resp protocol.Response
        if err := json.NewDecoder(conn).Decode(&resp); err != nil {
            t.Fatalf("Decode: %v", err)
        }
        return resp
    }

    resp := send(protocol.Request{Method: "hello", Version: 1, Client: "test"})
    if resp.Type != "welcome" {
        t.Fatalf("expected welcome, got %s", resp.Type)
    }

    resp = send(protocol.Request{Method: "input", Key: "a"})
    if resp.Type != "preedit" {
        t.Fatalf("expected preedit, got %s", resp.Type)
    }
    resp = send(protocol.Request{Method: "enter"})
    if resp.Type != "commit" || resp.Text != "a" {
        t.Fatalf("expected commit 'a', got %s %q", resp.Type, resp.Text)
    }
}
```

- [ ] **Step 2: Run integration test**

```bash
cd goime && go test -v -count=1 -run TestIntegrationFullFlow .
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "test: add integration test for full flow"
```

---

### Scope check

**Spec coverage:**
- ✅ Pinyin syllable table (Task 2)
- ✅ Protocol message types (Task 3)
- ✅ Config loading (Task 4)
- ✅ Default config files (Task 4)
- ✅ Dictionary builder (Task 5)
- ✅ Static dict loading (Task 5)
- ✅ User dict with SQLite (Task 6)
- ✅ goime-dict CLI (Task 7)
- ✅ Speller with Xiaohe lookup (Task 8)
- ✅ Segmentor with backtracking (Task 9)
- ✅ Translator with phrase + user words (Task 10)
- ✅ Self-made word via selections (Task 10)
- ✅ Session management (Task 11)
- ✅ Unix socket server (Task 12)
- ✅ goimed entry point (Task 13)
- ✅ Integration test (Task 14)

**Deferred (future):**
- Rime .dict.yaml import
- Complete candidate word lookups via Speller → Segmentor → Translator pipeline (MVP uses direct buffer-based select)
- Dynamic algebra rule parsing (Xiaohe lookup is hardcoded)
- Full vim plugin (separate project `goime.vim`)
- N-gram language model
- SIGHUP hot reload
- Double-array trie optimization
