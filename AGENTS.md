# GoIME

Server-side Chinese input method engine. Unix Socket daemon + dict tool + CLI test client. Works with Vim/Neovim via plugins.

## Project

- **What**: Server-side Chinese IME — supports Xiaohe Shuangpin & Full Pinyin
- **Stack**: Go 1.25, `modernc.org/sqlite` (CGo-free SQLite), `github.com/spf13/cobra` (dict/cli tools), `github.com/BurntSushi/toml` (config), mmap for dict index
- **Module**: `github.com/jiazhoulvke/goime`
- **Entry points**:
  - `cmd/goimed/main.go` — daemon, std `flag` package
  - `cmd/goime-dict/main.go` — dictionary builder, cobra
  - `cmd/goimec/main.go` — CLI test client, cobra

## Commands

| Command | What |
|---------|------|
| `go build ./cmd/goimed` | Build daemon binary |
| `go build ./cmd/goime-dict` | Build dict tool binary |
| `go build ./cmd/goimec` | Build CLI test binary |
| `make build` | Build all three binaries to `./` (appends `.exe` on Windows) |
| `make test` | `go test ./...` |
| `make clean` | Remove local binaries (handles `.exe` on Windows) |
| `make vet` | `go vet ./...` |
| `make release` | `goreleaser release --clean` (tagged releases) |

- E2E test: `go test -v -run TestE2E ./...` (in root package)
- Integration test: `go test -v -run TestIntegration ./...`
- CI: `.github/workflows/ci.yml` — `go test ./...` on ubuntu-latest, stable Go
- Releases: `.goreleaser.yaml` — builds for linux/darwin/freebsd/openbsd (no Windows builds in release)

## Architecture

```
cmd/
  goimed/        — Daemon entry point (flag-based, no cobra)
  goime-dict/    — Dictionary tool (cobra CLI)
  goimec/        — Interactive test client (cobra CLI)
internal/
  config/        — TOML config loading & defaults (General, Scheme, Dict, Candidates, Translator, UserDict)
  server/        — IPC listener, session state machine, idle-loop
  engine/        — Speller interface (xiaohe/fullpin), Segmentor, Translator
  dict/          — Dict Index (mmap lazy-load / in-memory merge), UserDict (SQLite), Rime import
  protocol/      — JSON request/response types (Request, Response, Candidate)
  pinyin/        — Valid-pinyin syllable table and validation
  transport/     — *(planned)* Platform-agnostic IPC abstraction
```

Flow: IPC (currently direct `net.Listen`) → Server → Session → Speller → Segmentor → Translator → Dict (static + user)

## Platform Support

- **Linux / macOS / FreeBSD / OpenBSD**: Unix Domain Socket (default `/tmp/goime-$UID.sock` or `$XDG_RUNTIME_DIR/goime.sock`)
- **Dict mmap**: Unix uses `mmap` syscall (`mmap_unix.go`); Windows fallback uses `os.ReadFile` (`mmap_other.go`)
- **Makefile**: Auto-detects `$(OS)` to append `.exe` extension on Windows

## Conventions

- **Language**: Chinese comments, English code and identifiers
- **Errors**: `fmt.Errorf("context: %w", err)` — always wrap with `%w`
- **Logging**: `log/slog` (standard library) — never `log.Print` or third-party loggers
- **Constructors**: `New*()` functions; return `(*Type, error)` when fallible, plain `*Type` otherwise
- **Tests**: table-driven with anonymous structs; `t.Errorf` / `t.Fatal` for failures; e2e tests in root package (`e2e_test.go`, `integration_test.go`)
- **Config**: TOML via `BurntSushi/toml`; use `toml:` tags on struct fields; `config.Default()` provides defaults; `config.ExpandPath()` for `~` expansion
- **Imports**: group stdlib first, then third-party, no blank line between them
- **Config paths**: `~/.config/goime/goime.toml` default; `~/.config/goime/dicts/` for static dicts; `~/.cache/goime/` for built indices
- **Protocol**: JSON-lines over Unix socket; message structs in `protocol/` package

## Notes

<!-- Quick-add section for future discoveries -->
