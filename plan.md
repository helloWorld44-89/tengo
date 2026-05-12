# Tengo CLI Text Editor — Product Plan

## Vision

Tengo is the go-to CLI editor for structured config files (YAML, JSON, TOML, INI, XML).
It is fast to open, fast to navigate, and works entirely in the terminal without setup.
It also doubles as a scriptable CLI tool for non-interactive file manipulation.

---

## Two Modes of Operation

### 1. Interactive Editor (`tengo <file>`)
Opens the file in a full-screen terminal UI for manual editing.

### 2. CLI / Non-Interactive (`tengo <file> [flags]`)
Applies transformations directly from the command line, prints results, and exits.
Ideal for scripting and CI pipelines.

---

## CLI Commands (Non-Interactive Flags)

These flags operate on the file and exit immediately — no UI is opened.

| Flag | Description |
|---|---|
| `-F <text>` | Find and print all lines containing `<text>` |
| `-F <text> -R <new>` | Find all occurrences of `<text>` and replace with `<new>`, save in-place |
| `-F <text> -R <new> --dry-run` | Print what would change without writing |
| `-L <n>` | Print line `n` |
| `-validate` | Validate file syntax (YAML/JSON/TOML/INI/XML) and exit 0 (valid) or 1 (invalid) |
| `-format` | Auto-format/pretty-print the file in place |
| `-key <key.path>` | Read a value at a dot-notation key path (e.g. `server.port`) |
| `-set <key.path>=<value>` | Set a value at a key path and save |
| `--stdout` | Print result to stdout instead of writing in-place (combine with `-F -R`, `-format`, `-set`) |
| `--backup` | Create a `.bak` copy before any write |
| `-q` / `--quiet` | Suppress output, only use exit codes |

**Example usage:**
```bash
tengo config.yaml -F "localhost" -R "0.0.0.0"
tengo config.yaml -validate
tengo config.yaml -key database.host
tengo config.yaml -set database.port=5432 --backup
tengo config.yaml -format --stdout
tengo config.yaml -F "debug: true" -R "debug: false" --dry-run
```

---

## Phase A: Quick Editor Stabilization & UI Polish (Current)

### Done
- [x] Refactor error handling — no panics, user-friendly messages
- [x] Add core unit tests
- [x] Undo/redo stacks
- [x] Clipboard (copy/cut/paste) with cross-platform support
- [x] Scrolling
- [x] Help popup overlay
- [x] Find (`ctrl-f`) and Find & Replace (`ctrl-r`)

### Remaining
- [ ] Finalize and audit all keyboard shortcuts for consistency
- [ ] Status bar shows unsaved changes indicator (`●`)
- [ ] Status bar shows file type detected (YAML / JSON / TOML / etc.)
- [ ] Line numbers (toggleable with `ctrl-n`)
- [ ] Improved info bar: show selection size (chars/lines selected)
- [ ] Highlight current line
- [ ] Word-wrap toggle
- [ ] ASCII art welcome/splash on launch (no file arg)
- [ ] Cross-platform validation (Windows, Linux, macOS)

---

## Phase B: Structured File Support (YAML / JSON / TOML / INI / XML)

This is the core differentiator for tengo.

### Syntax Validation
- [ ] Detect file type from extension (`.yaml`, `.yml`, `.json`, `.toml`, `.ini`, `.xml`)
- [ ] Validate on save (`ctrl-s`) — show a popup or status bar message with error details
- [ ] Validate on demand (`ctrl-e`) — explicit lint/check shortcut
- [ ] Show inline error indicator on the offending line (e.g., a `!` glyph in the gutter)
- [ ] Exit code 1 from `-validate` flag when invalid

### Syntax Highlighting (stretch)
- [ ] Keyword/key coloring for YAML, JSON, TOML keys vs values
- [ ] String, number, boolean, null type coloring
- [ ] Highlight mismatched braces/brackets

### Auto-formatting
- [ ] `ctrl-shift-f` — format the whole file in place (pretty-print)
- [ ] Format on save (opt-in setting)
- [ ] CLI flag `-format`

### Key-Path Navigation (YAML/JSON/TOML)
- [ ] `ctrl-g` — jump to key path (e.g., type `server.port`, cursor moves to that line)
- [ ] CLI flag `-key <path>` to read a value
- [ ] CLI flag `-set <path>=<value>` to set a value

---

## Phase C: CLI Non-Interactive Mode

- [ ] Parse CLI flags (`-F`, `-R`, `-L`, `-validate`, `-format`, `-key`, `-set`, `--dry-run`, `--stdout`, `--backup`, `-q`)
- [ ] `main.go` routes to interactive editor if no transformation flags are present, or runs non-interactive pipeline if flags are given
- [ ] Non-interactive operations return proper exit codes (0 = success, 1 = error/invalid)
- [ ] `-F` / `-R` supports regex patterns (opt-in with `--regex` flag)
- [ ] `--dry-run` prints a unified diff of changes to stdout
- [ ] Pipe support: `cat config.yaml | tengo -validate` reads from stdin

---

## Phase D: Full Editor Mode (Advanced)

A richer editing experience beyond the Quick Editor.

- [ ] Multi-file tabs (`tengo file1.yaml file2.json`)
- [ ] Split-pane view (horizontal or vertical)
- [ ] Configurable keybindings via `~/.config/tengo/keys.toml`
- [ ] Auto-close brackets and quotes (`"`, `'`, `{`, `[`)
- [ ] Word-level navigation (`ctrl-left`/`ctrl-right` skip whole words, not just characters)
- [ ] Column/block selection mode
- [ ] Multiple cursors (stretch goal)
- [ ] Search across all open files
- [ ] Persistent undo history across sessions (stored in `.tengo/` alongside the file)

---

## Phase E: Polish & Release

- [ ] Cross-platform binary builds (Linux x64, macOS arm64/x64, Windows x64) via GitHub Actions
- [ ] Single-binary, zero-dependency install (static linking where possible)
- [ ] `tengo --version` and `tengo --help`
- [ ] Man page / shell completion (bash, zsh, fish)
- [ ] README with install instructions and usage examples
- [ ] Release assets on GitHub Releases

---

## Security Considerations

These apply to both interactive and CLI modes.

### Input Validation
- [ ] Validate and sanitize all CLI flag values before use — reject shell metacharacters in `-F`/`-R`/`-key`/`-set` when passed to external processes
- [ ] Clamp all cursor/selection indices before buffer access to prevent out-of-bounds panics
- [ ] Cap file size on open (e.g., warn/refuse files > 50 MB) to prevent memory exhaustion

### File Operations
- [ ] Atomic saves: write to a temp file, then rename — prevents data loss if the process is killed mid-write
- [ ] Preserve original file permissions and ownership on save
- [ ] `--backup` creates a `.bak` before any destructive in-place operation
- [ ] Refuse to follow symlinks to sensitive paths without explicit confirmation (stretch)
- [ ] Validate that the file path argument does not escape the working directory via `../` traversal (relevant for future server/plugin modes)

### Clipboard
- [ ] Shell-out commands for clipboard (`xclip`, `xsel`, `pbpaste`, PowerShell) must not be constructed from user-controlled strings — hardcode the command, pass content via stdin or flags only

### Structured File Parsing
- [ ] Use safe, well-audited Go libraries for YAML/JSON/TOML parsing — do not eval or exec file content
- [ ] YAML parser must disable alias expansion beyond a reasonable depth to prevent billion-laughs-style DoS
- [ ] JSON/TOML parsers should have a max-depth or max-size limit

### Distribution
- [ ] Pin dependency versions in `go.sum` and verify checksums in CI
- [ ] Run `govulncheck` in CI to catch known CVEs in dependencies
- [ ] Do not bundle credentials, tokens, or environment-specific paths in the binary (fix the current Windows paths in `go.mod` before release)

---

## Dependency Plan

| Purpose | Library |
|---|---|
| Terminal raw mode & size | `golang.org/x/term` |
| Clipboard (copy) | `github.com/tiagomelo/go-clipboard` |
| YAML validation/formatting | `gopkg.in/yaml.v3` |
| JSON validation/formatting | `encoding/json` (stdlib) |
| TOML validation/formatting | `github.com/BurntSushi/toml` |
| INI parsing | `gopkg.in/ini.v1` |
| XML validation | `encoding/xml` (stdlib) |
| CLI flag parsing | `flag` (stdlib) or `github.com/spf13/cobra` |
| Vulnerability scanning | `golang.org/x/vuln/cmd/govulncheck` (CI only) |

---

## Notes

- Prioritize stability and speed over features — the editor must open instantly
- All file writes must be atomic (temp file + rename)
- Keep zero required config — tengo works out of the box with no dotfiles
- The Quick Editor is the primary mode; Full Editor is additive
