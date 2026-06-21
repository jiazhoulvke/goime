# GoIME — 服务端输入法引擎 · 设计文档

## 概述

GoIME 是一个基于 Go 语言实现的服务端中文输入法引擎。通过 Unix Socket 对外提供服务，与 Vim/Neovim 等编辑器深度集成，完全脱离系统输入法框架即可输入中文。支持双拼（小鹤等）和全拼两种输入方式，通过可插拔方案切换。

目标：在保留 Rime 核心设计理念的基础上，提供更简洁、更快的 Go 原生实现。

## 架构

```
┌──────────────────────────────────────────────┐
│                goimed (守护进程)               │
│                                              │
│  ┌────────────┐    ┌──────────────────────┐  │
│  │ Unix Socket│───►│   Session Manager    │  │
│  │  Listener  │    │  (每个连接一个Session) │  │
│  └────────────┘    └──────────┬───────────┘  │
│                               ▼              │
│  ┌────────────────────────────────────────┐  │
│  │          Input Engine                  │  │
│  │  ┌──────────┐ ┌──────────┐ ┌────────┐ │  │
│  │  │ Speller  │►│Segmentor │►│Trans-  │ │  │
│  │  │(algebra) │ │(分词)    │ │lator   │ │  │
│  │  └──────────┘ └──────────┘ └───┬────┘ │  │
│  └────────────────────────────────┼────────┘  │
│                                   ▼           │
│  ┌────────────────────────────────────────┐  │
│  │          Dictionary Engine             │  │
│  │  ┌─────────────────┐ ┌──────────────┐  │  │
│  │  │   Static Dict   │ │  User Dict   │  │  │
│  │  │  (mmap Trie)    │ │  (SQLite)    │  │  │
│  │  └─────────────────┘ └──────────────┘  │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  ┌────────────────────────────────────────┐  │
│  │       Config Manager                   │  │
│  │  (TOML: 方案/词典/双拼映射表)           │  │
│  └────────────────────────────────────────┘  │
└──────────────────────────────────────────────┘
```

### 模块职责

| 模块 | 职责 |
|------|------|
| Session Manager | 管理每个编辑器的连接状态（输入缓冲区、候选词列表、翻页位置、当前方案） |
| Speller | 将用户按键通过 algebra 规则生成的映射表转换为拼音音节序列 |
| Segmentor | 将拼音序列按所有合法音节边界切分，生成全部可能切分 |
| Translator | 根据拼音搜索词库，生成排序后的候选词列表 |
| Static Dict | 只读词库，mmap 加载二进制 trie 索引 |
| User Dict | 可写词库，SQLite WAL 模式存储用户词频数据 |
| Config Manager | 加载双拼方案、词库路径、用户自定义设置 |

## 目录结构

```
goime/
├── cmd/
│   ├── goimed/              # 守护进程入口
│   │   └── main.go
│   └── goime-dict/         # 词库构建工具
│       └── main.go
├── internal/
│   ├── server/             # Unix socket 监听 + 连接管理
│   │   ├── server.go
│   │   └── session.go
│   ├── protocol/           # JSON-Lines 协议定义
│   │   └── message.go
│   ├── pinyin/             # 拼音音节表（~410 个合法音节）
│   │   ├── syllables.go
│   │   └── syllables_test.go
│   ├── engine/             # 输入法核心引擎
│   │   ├── engine.go       # 主循环
│   │   ├── speller.go      # Algebra 规则引擎 + 映射表生成
│   │   ├── segmentor.go    # 拼音音节切分（歧义感知）
│   │   └── translator.go   # 候选项生成（词级，预留历史窗口扩展点）
│   ├── dict/               # 词库引擎
│   │   ├── dict.go         # 接口定义
│   │   ├── static.go       # 静态词库 (mmap double-array trie)
│   │   ├── user.go         # 用户词库 (SQLite WAL)
│   │   └── builder.go      # goime-dict 词库编译
│   └── config/             # 配置管理 (TOML)
│       └── config.go
├── configs/                # 默认配置文件
│   ├── goime.toml          # 主配置
│   └── schemes/
│       ├── xiaohe.toml     # 小鹤双拼方案（MVP）
│       └── fullpin.toml    # 全拼方案（MVP）
├── dicts/                  # 词库文件（用户自行放置）
│   └── README.md
├── go.mod
├── go.sum
└── Makefile
```

Vim/Neovim 插件是独立项目 `goime.vim`：

```
goime.vim/
├── plugin/
│   └── goime.vim           # Vim8 主插件文件（vimscript）
├── autoload/
│   ├── goime.vim           # Vim8 按需加载函数（含状态栏接口）
│   └── airline/
│       └── extensions/
│           └── goime.vim   # airline 自动发现的状态栏扩展
└── lua/
    └── goime/
        ├── init.lua         # Neovim 入口
        ├── client.lua       # Unix socket 客户端
        ├── ui.lua           # 候选窗/popupmenu 渲染
        ├── status.lua       # 状态栏接口
        └── config.lua       # 插件配置
```

## 通信协议（Unix Socket / JSON-Lines）

每行一个完整的 JSON 对象，`\n` 作为分隔符。Go 的 `bufio.Scanner` 原生支持。

### Socket 路径

优先 `$XDG_RUNTIME_DIR/goime.sock`，回退 `/tmp/goime-$UID.sock`，权限 `0600`。

服务端启动时探测已有 socket，连不上则删除残留文件重新创建。客户端同理，连不上则拉起 goimed。

### 握手

连接建立后第一条消息交换版本和方案信息，业务消息不带版本字段：

```json
→ {"method":"hello","version":1,"client":"nvim-goime-0.1"}
← {"type":"welcome","version":1,"schemes":["xiaohe","fullpin"],"active":"xiaohe","page_size":5}
```

版本不兼容时服务端返回 `error`，客户端提示升级。

### 请求方法（客户端 → 服务端）

特殊键走独立方法，`input` 只处理普通字母字符，无歧义：

| method | 参数 | 说明 |
|--------|------|------|
| `hello` | `version`, `client` | 握手，交换版本和方案列表 |
| `input` | `key`: 单个字母字符（a-z） | 输入一个字符 |
| `enter` | — | 上屏原始输入码（不管有无候选词） |
| `escape` | — | 清空缓冲区，返回 idle |
| `backspace` | — | 删除缓冲区最后一个字符 |
| `arrow` | `dir`: "up"/"down"/"left"/"right" | 方向键 |
| `page` | `dir`: "next"/"prev" | 翻页 |
| `space` | — | 选首选词上屏（有候选时）或上屏原始输入码（无候选时） |
| `select` | `index`: 候选序号（0-based） | 选择候选词 |
| `commit_preedit` | — | 直接上屏当前输入码 |
| `set_scheme` | `name`: 方案名 | 切换输入方案 |
| `reset` | — | 重置状态 |

**`input` 方法只接受字母 a-z**。非字母字符（数字、标点、符号等）不走 `input`，由客户端直接处理：客户端先触发当前缓冲区上屏，再插入该字符。空缓冲区时直接插入。

### 响应类型（服务端 → 客户端）

所有改变状态的方法都返回当前状态快照，客户端按 `type` 分发渲染：

| type | 字段 | 说明 |
|------|------|------|
| `welcome` | `version`, `schemes`, `active`, `page_size` | 握手响应 |
| `preedit` | `text`, `pos` | 输入缓冲区实时预览（未上屏） |
| `candidates` | `list`, `page`, `total` | 候选词列表 |
| `commit` | `text`, `pending_key`(可选) | 上屏文字；`pending_key` 为需透传的后续字符 |
| `idle` | — | 无待处理内容（缓冲区空） |
| `error` | `message` | 错误信息 |

### Session 管理

- 每个 Unix Socket 连接对应一个 session，有独立输入缓冲区、当前方案、选词历史
- 连接本身隔离状态，协议无 session 字段
- 异常断连自动清理，15 分钟无活动超时断开（可配置）
- 多编辑器实例互不干扰

Session 结构包含选词历史，用于自造词：

```
Session {
    buffer: "nihaoshijie"              # 当前输入缓冲区
    scheme: "xiaohe"                   # 当前输入方案
    selections: [                      # 本轮选词历史
        {pinyin: "nihao", word: "你好"},
        {pinyin: "shijie", word: "世界"},
    ]
}
```

当缓冲区完全清空时算一轮结束，此时如果 `selections` 包含多个词，组合成新词存入用户词库。连续分段选词（同一次输入过程）才产生自造词，分次输入不产生。

## 词库系统

### 词库源文件格式（纯文本）

```
shu1ru4 输入 100
shu1wu1 书屋 50
ni3hao3 你好 200
hang2 行 80
xing2 行 90
```

格式：`拼音<tab>词语<tab>权重`。
拼音带声调（数字 1-4 表示四声，5 或省略表示轻声）。
多音字按读音分别存储条目，天然支持多音字。

### Rime 词库兼容

goime-dict 提供 `goime-dict import --rime` 子命令，将 Rime 的 `.dict.yaml` 转换为 GoIME 纯文本格式。运行时只支持 GoIME 自有格式。

**多词库合并**：`dict.static` 可配置多个词库文件。同拼音同词不同权重时，取最大值。

### 用户词库

- **运行时**：SQLite WAL 模式，高频写入、事务安全、支持复杂查询
- **同步/备份**：纯文本 `user_dict.txt`，人类可读、可版本控制、可跨设备同步
- goime-dict 提供 `goime-dict export` / `goime-dict import` 在两者间互转

**两张表（职责分离）：**

```sql
-- 词频表：记录静态词库词的使用频率
CREATE TABLE word_freq (
    pinyin     TEXT,      -- 去调拼音，如 "shuru"
    word       TEXT,      -- 词语，如 "输入"
    frequency  INTEGER,   -- 被选次数
    PRIMARY KEY (pinyin, word)
);

-- 自造词表：记录用户新组合的词及其频率
CREATE TABLE user_words (
    pinyin     TEXT,      -- 去调拼音，如 "nihaoshijie"
    word       TEXT,      -- 词语，如 "你好世界"
    frequency  INTEGER,   -- 被选次数
    created_at INTEGER,   -- 首次创建时间
    PRIMARY KEY (pinyin, word)
);
```

**写入时机：**

| 选词来源 | 写入位置 | 说明 |
|---------|---------|------|
| 静态词库的词 | `word_freq` frequency +1 | 追踪用户偏好，提升常用词排序 |
| `user_words` 自造词 | `user_words` frequency +1 | 自造词自己管频率，不重复记录 |

**查询时权重计算：**

- 静态词库词最终权重 = `静态权重 + word_freq.frequency`
- 自造词最终权重 = `user_words.frequency`（新建时初始权重为 `new_word_weight`，可配置）

**新自造词初始权重：**

新建自造词时 frequency 设为 `new_word_weight`（默认 100，可配置），保证用户下次输入时能看到。如果不常用，词频衰减机制会自然降低排序。

**自造词生成流程：**

1. 用户输入 `nihaoshijie`（4 音节）
2. 静态词库无"你好世界"完整匹配
3. 用户分段选词：`nihao`→你好（缓冲区剩 `shijie`），`shijie`→世界（缓冲区清空）
4. Session 记录选词历史：`[{nihao, 你好}, {shijie, 世界}]`
5. 缓冲区清空时触发：合并拼音和词语 `nihaoshijie` → 你好世界
6. 写入 `user_words` 表（如已存在则 frequency +1）

注意：如果用户先打 `nihao` 选"你好"上屏（缓冲区清空），再打 `shijie` 选"世界"上屏——这是两轮独立输入，不产生自造词。只有同一次输入过程中连续分段选词才产生。

**查询流程：**

1. 先查 `user_words` 表，命中的自造词权重上浮，优先展示
2. 再查静态词库分段匹配
3. 合并排序

每次用户选择候选词时，对应条目频率 +1。
查询结果 = 自造词 ∪ 静态词库结果，合并后按最终权重排序。

### 构建流程

```
dict.txt ──► goime-dict ──► dict.goime（二进制 double-array trie）
                                    │
                     goimed 启动时 mmap 加载
                                    ▼
                              查询 O(n)
```

goime-dict 将纯文本词库编译为二进制索引文件。编译时以**去调拼音**为 key 建索引，声调作为元数据附带。运行时查无声调 key 命中所有同音条目，按权重排序。goimed 启动时 mmap 加载，毫秒级完成。

## 输入引擎

### Speller（Algebra 规则引擎）

参考 Rime 的 Speller 设计，支持以下代数变换：

| 类型 | 作用 | 示例 |
|------|------|------|
| `xform` | 无条件替换 | `xform/iu$/Q/` |
| `derive` | 增加变体，保留原输入 | `derive/^([zcs])h/$1/` |
| `abbrev` | 取首字母做简拼 | `abbrev/^([a-z]).+$/$1/` |
| `erase` | 删除匹配项 | `erase/^xx$/` |

**实现方式**：Speller 内部按方案类型分发：

- **双拼方案**（`type = "shuangpin"`）：方案加载时用 algebra 规则生成 `map[输入码][]拼音音节` 的完整映射表存内存。运行时 O(1) 查表，不跑正则。切换方案时重新生成表。
- **全拼方案**（`type = "pinyin"`）：pass-through 模式，用户输入直接就是拼音，不需要转换。algebra 规则为空。

双拼方案配置示例：

```toml
# configs/schemes/xiaohe.toml
[speller]
type = "shuangpin"
alphabet = "abcdefghijklmnopqrstuvwxyz"
delimiter = "'"
algebra = [
  "xform/iu$/I/",
  "xform/(.)ei$/$1E/",
  "xform/^zh/V/",
  # ...
]
```

全拼方案配置示例：

```toml
# configs/schemes/fullpin.toml
[speller]
type = "pinyin"
alphabet = "abcdefghijklmnopqrstuvwxyz"
delimiter = "'"
algebra = []   # 空规则 = pass-through
```

### Segmentor（音节切分）

将拼音字符串按**所有合法音节边界**切分，生成全部可能切分：

```
xian → [[xi, an], [xian]]
fangan → [[fang, an], [fan, gan]]
```

使用预编译的合法拼音集合（约 410 个），枚举所有合法切分。各切分分别查词库，结果合并后按权重排序。拼音歧义切分的组合数很少（最多几个分支），查词库是 mmap 查表，成本极低。

### Translator（候选词生成）

**MVP 阶段：纯词级输入 + 自造词**，每次查询只匹配 1-8 音节的词，用户逐词上屏。用户分段选词后，组合词自动存入用户词库，下次直接作为候选。

流程：
1. 先查用户自造词表（`user_words`），命中的自造词优先展示
2. 逐音节查单字
3. 贪心组合多音节查词组（1-8 音节）
4. 词频加权：已有词的频率影响排序
5. 结果合并排序，自造词 > 高频词 > 低频词

**自造词机制：**

- Session 记录本轮输入的选词历史
- 用户分段选词（如 `nihao`→你好、`shijie`→世界）时，每次 `select` 都追加到 `selections`
- 缓冲区完全清空时，若 `selections` 包含多个词，合并拼音和词语存入 `user_words`
- 已存在的自造词 frequency +1

**预留扩展点**：Translator 接口预留 `RankWithHistory()` 方法，后期可接入：
- **N-gram 语言模型**：参考 Rime Octagram 数据实现整句输入

## Vim/Neovim 集成

### 插件目录结构

Vim/Neovim 插件是独立项目 `goime.vim`，与 Go 服务端分开维护：

```
goime.vim/
├── plugin/
│   └── goime.vim           # Vim8 入口（vimscript）
├── autoload/
│   ├── goime.vim           # Vim8 按需加载函数（含状态栏接口）
│   └── airline/
│       └── extensions/
│           └── goime.vim   # airline 自动发现的状态栏扩展
└── lua/                      # Neovim Lua 模块
    └── goime/
        ├── init.lua           # Neovim 入口
        ├── client.lua         # Unix socket 客户端
        ├── ui.lua             # 候选窗/popupmenu 渲染
        ├── status.lua         # 状态栏接口
        └── config.lua         # 插件配置
```

插件通过 Unix Socket JSON-Lines 协议与 goimed 通信，不依赖 Go 源码，可独立发版。

### 环境适配策略

```vim
if has('nvim')
  " Neovim: Lua + vim.rpcrequest + nvim_put + floating window
  lua require('goime').setup()
elseif has('job')
  " Vim8: job_start + channel_sendraw + popup
  call goime#init()
else
  " 降级: system("nc -U /tmp/goime-$UID.sock") 轮询，无异步候选窗
endif
```

三种情况共用同一个 Unix Socket JSON-Lines 协议，服务端无关。

### 交互流程

1. 用户进入插入模式，插件异步连接 goimed socket
2. 握手交换版本和方案列表
3. 每次字母键通过 `input` 发送到 goimed，非字母字符由客户端直接处理（触发缓冲区上屏 + 插入字符）
4. goimed 返回 `candidates` 或 `commit`
5. `commit` → 编辑器直接插入文本（若有 `pending_key` 则继续插入该字符）
6. `candidates` → popupmenu/浮动窗口展示候选
7. 用户 `<Tab>`/`<Up>`/`<Down>` 选词，`select` 提交

**Escape 处理**：Escape 不走 goimed。客户端拦截 Escape，异步发 `escape` 给 goimed 清空缓冲区，同时正常切 Normal 模式。用户无感知。

### 中英文切换

由客户端处理，服务端无感知：
- 英文模式下按键直接透传给 Vim，不走 goimed，零延迟
- 切换键约定由插件配置，默认 `<C-Space>` 或 `<Shift>`
- 中文模式下字母键走 goimed，英文模式下走 Vim 默认行为

### 标点符号输入

由客户端处理，两个独立开关（参考 Rime 设计）：
- **中/英模式**（`ascii_mode`）：管字母数字走不走 goimed
- **中/英标点**（`ascii_punct`）：管标点用全角还是半角

两个开关独立，都在客户端。中文模式下 `,` 直接替换成 `，` 上屏，零延迟。映射表可配置。

### 自启动与生命周期

- 插件连接 socket 失败时自动 `job_start("goimed")`，等 socket 就绪后连接
- goimed 作为独立进程运行，插件退出不影响 goimed
- 后续 vim 共享同一个 goimed（词频等状态共享）
- 关闭所有 vim 后，goimed 因 idle 超时（默认 15 分钟）自动退出
- 不需要 systemd，零配置

### 状态栏集成

插件暴露获取当前输入法状态的函数，兼容主流状态栏插件（airline、lightline、lualine）。

**设计原则**：三个状态栏插件的通用模式是"函数返回字符串，空串表示隐藏"。GoIME 遵循这一约定，同时维护 Vimscript 全局函数和 Lua 模块两条路径。

**暴露的接口：**

| 接口 | 语言 | 签名 | 说明 |
|------|------|------|------|
| `goime#status()` | Vimscript | `→ string` | Vim8/通用，返回当前状态文本 |
| `require('goime.status').current()` | Lua | `→ string` | Neovim，返回当前状态文本 |

**返回值约定：**

- 中文模式：`中`（或配置的图标，如 `🇨🇳`、`CN`）
- 英文模式：`EN`（或配置的图标，如 `🇬🇧`、`ASCII`）
- 未连接 goimed：空字符串 `""`（状态栏段隐藏）

返回值格式可配置：

```vim
let g:goime_status_cn = '中'    " 中文模式显示
let g:goime_status_en = 'EN'    " 英文模式显示
let g:goime_status_off = ''     " 未连接时显示（空=隐藏）
```

**自动集成：**

- **airline**：插件提供 `autoload/airline/extensions/goime.vim`，airline 自动发现。检测到 GoIME 已连接时自动在 section A 显示状态。无需用户配置。
- **lightline**：无自动发现机制（by design），用户需手动配置一行：
  ```vim
  let g:lightline.component_function = { 'goime': 'goime#status' }
  ```
- **lualine**：用户需手动配置一行：
  ```lua
  sections = { lualine_a = { { 'require"goime.status".current()' } } }
  ```

**状态变更通知：**

插件在模式切换（中/英）时主动触发状态栏刷新：
- Neovim：`vim.cmd('redrawstatus')` 或 `require('lualine').refresh()`
- Vim8：`redrawstatus!`

## 配置

主配置文件 `goime.toml`，所有配置项均有中文注释，解释作用和使用方式：

```toml
# =============================================================================
# GoIME 主配置文件
# 路径：~/.config/goime/goime.toml
# =============================================================================

# -----------------------------------------------------------------------------
# [general] 通用设置
# 控制 goimed 守护进程的基本行为
# -----------------------------------------------------------------------------
[general]

# 日志级别，控制日志输出的详细程度
# 可选值：debug / info / warn / error
# debug：输出所有日志（调试用，生产环境不建议）
# info：输出常规运行信息（默认值，推荐）
# warn：仅输出警告和错误
# error：仅输出错误
log_level = "info"

# Unix Socket 文件路径
# 留空（推荐）则自动推导：
#   优先使用 $XDG_RUNTIME_DIR/goime.sock（systemd 提供，per-user，自动清理）
#   回退到 /tmp/goime-$UID.sock（带 UID 避免多用户冲突）
# 手动指定时需确保路径可写且不与其他用户冲突
# 权限自动设为 0600（仅当前用户可连接）
socket_path = ""

# 空闲超时时间，所有客户端断开后多久自动退出 goimed
# 支持时间单位：s（秒）、m（分钟）、h（小时）
# 设为 0 表示永不退出（需手动管理进程）
# 默认 15 分钟：关闭所有编辑器后 15 分钟自动清理，下次打开编辑器时插件重新拉起
idle_timeout = "15m"

# -----------------------------------------------------------------------------
# [scheme] 输入方案设置
# 控制使用哪个输入法方案（双拼/全拼）及方案文件的位置
# -----------------------------------------------------------------------------
[scheme]

# 默认激活的输入方案名称
# 必须是 dir 目录下存在的 .toml 方案文件名（不含扩展名）
# 可运行时通过协议 set_scheme 方法切换
# MVP 支持的方案：xiaohe（小鹤双拼）、fullpin（全拼）
active = "xiaohe"

# 方案文件目录，存放各输入方案的 TOML 配置
# 支持波浪线 ~ 展开
dir = "~/.config/goime/schemes/"

# -----------------------------------------------------------------------------
# [dict] 词库设置
# 控制静态词库和用户词库的路径及构建行为
# -----------------------------------------------------------------------------
[dict]

# 静态词库文件列表（GoIME 纯文本格式）
# 可配置多个文件，goimed 会全部加载并合并
# 合并规则：同拼音同词不同权重时，取最大值
# 文件格式：拼音<tab>词语<tab>权重，如：shu1ru4<TAB>输入<TAB>100
# 支持 ~ 路径展开
# 使用 goime-dict import --rime 可从 Rime .dict.yaml 转换
static = ["~/.config/goime/dicts/zhonghua.dict.txt"]

# 用户词库 SQLite 数据库路径
# 存储：1) 静态词库词的使用频率  2) 用户自造词及其频率
# 使用 WAL 模式，崩溃不丢数据
user = "~/.config/goime/user_dict.db"

# 用户词库纯文本同步/备份文件路径
# 人类可读、可版本控制、可跨设备同步
# 使用 goime-dict export 导出、goime-dict import 导入
sync_file = "~/.config/goime/user_dict.txt"

# 词库二进制索引构建目录
# goime-dict 将纯文本词库编译为 .goime 二进制索引存放于此
# goimed 启动时 mmap 加载，毫秒级完成
build_dir = "~/.cache/goime/"

# 自动构建开关
# true（推荐）：goimed 启动时比较静态词库源文件和已构建索引的修改时间（mtime）
#   源文件更新则自动调用 goime-dict 重新构建，无需手动操作
# false：需手动运行 goime-dict 构建索引
auto_build = true

# -----------------------------------------------------------------------------
# [candidates] 候选词设置
# 控制候选词列表的展示行为
# -----------------------------------------------------------------------------
[candidates]

# 每页显示的候选词数量
# 用户翻页时每页展示的候选词个数
# 常见值：5（搜狗/Rime 默认）、7、9
# 值越大单页信息越多但选择可能更慢，建议 5-9
page_size = 5

# 单次查询最多返回的候选词总数
# 防止极端情况下返回过多候选词导致性能下降
# 超出部分需翻页查看
max_candidates = 100

# -----------------------------------------------------------------------------
# [translator] 翻译器设置
# 控制拼音到候选词的匹配行为
# -----------------------------------------------------------------------------
[translator]

# 单次查询最大匹配音节数
# 决定能匹配的最长词组音节长度
# 8 可覆盖绝大多数中文词组（成语最多 4-6 字，"中华人民共和国"7 音节）
# 值越大查询范围越广但性能越低，建议 6-8
max_syllables = 8

# -----------------------------------------------------------------------------
# [user_dict] 用户词库设置
# 控制用户词频和自造词的行为
# -----------------------------------------------------------------------------
[user_dict]

# 是否启用用户词库
# true（推荐）：记录词频、支持自造词
# false：纯静态词库模式，不记录任何用户数据
enabled = true

# 是否启用词频衰减
# true：长期不用的词频率会逐渐降低，让新常用词浮上来
# false：词频只增不减
freq_decay = true

# 词频衰减率，每次衰减时乘以此系数
# 0.99（推荐）：每次衰减 1%，温和下降
# 0.95：快速衰减，适合频繁切换话题的用户
# 范围：0.0-1.0，越接近 1.0 衰减越慢
# 衰减触发时机：goimed 启动时对所有词频统一衰减一次
decay_rate = 0.99

# 新自造词的初始权重
# 用户首次组合出自造词时写入的 frequency 值
# 100（推荐）：保证下次输入时能看到该词
# 太低（如 1）可能导致排序靠后用户找不到
# 太高（如 1000）可能干扰已有高频词的排序
new_word_weight = 100
```

### 配置与词库热重载

MVP 不支持热重载，修改配置或词库后重启 goimed 生效（启动 < 100ms）。代码将"加载配置"和"加载词库"封装成独立函数，预留 `SIGHUP` 信号重载扩展点。

## 崩溃恢复

- **数据安全**：用户词库 SQLite WAL 模式，崩溃不丢数据。静态词库只读，无风险。
- **进程恢复**：客户端检测连接断开时自动重新拉起 goimed。
- **崩溃日志**：goimed 崩溃时写入 `$XDG_STATE_HOME/goime/crash.log`（回退 `~/.local/state/goime/crash.log`）。
- **输入状态**：崩溃时输入缓冲区丢失，用户重新打字即可。
- MVP 不做自动重启守护（systemd 的事）。

## 日志

使用 Go 标准库 `slog`，结构化 JSON 日志：

- 路径：`$XDG_STATE_HOME/goime/goimed.log`（回退 `~/.local/state/goime/goimed.log`）
- 级别：跟随配置 `log_level`（debug/info/warn/error）
- 轮转：交给用户（logrotate 或不轮转）

## 测试策略

### 单元测试

每个 `internal/` 包独立测试：
- Pinyin：音节表完整性（~410 个合法音节无遗漏）
- Speller：algebra 规则解析、映射表生成、全拼 pass-through
- Segmentor：歧义切分（重点测 `xian`、`fangan` 等经典 case）
- Translator：候选生成、排序、自造词
- Dict：mmap 加载、查询、多词库合并

### 集成测试

端到端，用测试 Unix socket：
- 启动 goimed → 连接 → 握手 → 输入 `uuru` → 选词 → 验证 `commit`
- 覆盖协议握手、切换方案、翻页、取消等流程

### Benchmark

关键路径性能：
- `BenchmarkSpeller_Lookup` — 双拼码查表
- `BenchmarkTranslator_Query` — 完整查询链路
- `BenchmarkDict_Load` — mmap 加载速度

## 安装方式

GoIME（Go 服务端）和 goime.vim（编辑器插件）是两个独立项目，分别安装。

### GoIME（Go 服务端）

**MVP：**

```bash
go install github.com/jiazhoulvke/goime/cmd/goimed@latest
go install github.com/jiazhoulvke/goime/cmd/goime-dict@latest
```

**Release 阶段：**

GitHub Release 提供预编译二进制（Linux amd64/arm64），Makefile 里有 `make release` target 用 goreleaser 打包。

### goime.vim（编辑器插件）

通过插件管理器安装：

```vim
" vim-plug
Plug 'jiazhoulvke/goime.vim'

" packer.nvim
use { 'jiazhoulvke/goime.vim' }
```

插件启动时检测 `goimed` 是否在 PATH 里，不存在则提示安装。

## 性能目标

| 指标 | 目标 |
|------|------|
| 词库加载（百万条） | < 10ms（mmap） |
| 单次查询延迟（P99） | < 1ms |
| 并发 session | 无上限（goroutine per conn） |
| 内存占用（百万条词库） | < 50MB |
| 首次启动时间 | < 100ms |

## 代码规范

### 注释规范

**所有导出函数和重要内部函数必须有中文注释**，说明：
- 函数的作用（做什么）
- 参数含义
- 返回值含义
- 关键实现逻辑（如有）

示例：

```go
// ToPinyin 将用户输入的双拼码转换为拼音音节序列。
// code: 用户输入的编码（如双拼方案下的 "uuru" 或全拼方案下的 "shuru"）
// 返回: 所有可能对应的拼音音节切分结果
// 双拼方案查预生成的映射表（O(1)），全拼方案直接返回输入本身（pass-through）。
func (s *Speller) ToPinyin(code string) [][]string {
    // ...
}

// Segment 将拼音字符串按所有合法音节边界切分，生成全部可能切分。
// pinyin: 去调后的拼音字符串（如 "xian"）
// 返回: 所有可能的音节切分组合，如 [[xi, an], [xian]]
// 使用预编译的合法拼音集合（~410 个音节），递归枚举所有合法边界。
func Segment(pinyin string) [][]string {
    // ...
}
```

### 配置注释规范

**所有配置文件中的选项必须有中文注释**，说明：
- 选项作用
- 可选值及含义
- 推荐值及理由
- 与其他选项的关联（如有）

已在上方 `goime.toml` 配置示例中体现。

## 未来扩展

- **N-gram 整句输入**：参考 Rime Octagram 数据，用 Viterbi 解码实现整句输入
- **SIGHUP 热重载**：信号触发重新加载配置和词库
- **更多双拼方案**：自然码、微软双拼、智能ABC 等
