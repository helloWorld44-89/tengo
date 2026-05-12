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

## Phase A: Quick Editor Stabilization & UI Polish ✓

- [x] Refactor error handling — no panics, user-friendly messages
- [x] Add core unit tests (36 tests)
- [x] Undo/redo stacks (`ctrl-z` / `ctrl-y`)
- [x] Clipboard (copy/cut/paste) with cross-platform support
- [x] Scrolling
- [x] Help popup overlay (`ctrl-h`)
- [x] Find (`ctrl-f`) and Find & Replace (`ctrl-r`)
- [x] Finalize and audit all keyboard shortcuts (fixed ctrl-key byte mapping)
- [x] Status bar shows unsaved changes indicator (`●`)
- [x] Status bar shows file type detected (YAML / JSON / TOML / etc.)
- [x] Line numbers (toggleable with `ctrl-n`)
- [x] Improved info bar: show selection size (chars/lines selected)
- [x] Highlight current line
- [x] ASCII art welcome/splash on launch (no file arg)
- [x] Word-wrap toggle (`alt-w`) — soft wrap, no syntax highlight on wrapped segments
- [ ] Cross-platform testing (Windows, macOS)

---

## Phase B: Structured File Support ✓

### Syntax Validation
- [x] Detect file type from extension (`.yaml`, `.yml`, `.json`, `.toml`, `.ini`, `.xml`)
- [x] Validate on save (`ctrl-s`) — status bar shows error with line number
- [x] Validate on demand (`ctrl-e`) — jumps cursor to error line, shows gutter `!`
- [x] Inline gutter error indicator (`!` in red on offending line)
- [x] Exit code 1 from `-validate` flag when invalid

### Syntax Highlighting (stretch)
- [x] Keyword/key coloring for YAML, JSON, TOML keys vs values
- [x] String, number, boolean, null type coloring

### Auto-formatting
- [x] `ctrl-t` — format the whole file in place (pretty-print); JSON, YAML, TOML
- [x] CLI flag `-format` (with `--dry-run`, `--stdout`, `--backup`)

### Key-Path Navigation (YAML/JSON/TOML)
- [x] `ctrl-g` — go to line number
- [x] CLI flag `-key <path>` to read a value
- [x] CLI flag `-set <path>=<value>` to set a value (note: reformats file on save)

---

## Phase C: CLI Non-Interactive Mode ✓

- [x] Parse CLI flags (`-F`, `-R`, `-L`, `-validate`, `-format`, `--dry-run`, `--stdout`, `--backup`, `-q`, `--regex`, `-type`)
- [x] `main.go` routes to interactive editor or non-interactive pipeline
- [x] Non-interactive operations return proper exit codes (0 = success, 1 = error/invalid, 2 = I/O error)
- [x] `-F` / `-R` supports regex patterns (`--regex` flag)
- [x] `--dry-run` prints a colored unified diff
- [x] Pipe support: `cat config.yaml | tengo -validate -type yaml`
- [x] `-key <path>` — read a dot-notation key value
- [x] `-set <path>=<value>` — write a key value

---

## Phase D: Editor Polish

Focused improvements to the Quick Editor experience. Full IDE-style features
(split panes, multiple cursors, LSP, etc.) are tracked separately in
`fulleditor_plan.md` for a potential future v2.

- [x] Multi-file tabs (`tengo file1.yaml file2.json`) — `ctrl-w` / `alt-1`…`alt-9`
- [x] Word-level navigation — `ctrl-left` / `ctrl-right`
- [x] Auto-close brackets and quotes (`{`, `[`, `(`, `"`, `'`, `` ` ``) — skip-over on closing char
- [x] Syntax highlighting — key/value coloring for YAML, JSON, TOML, INI (foreground-only, composes with current-line highlight)
- [x] Status bar: show cursor column as visual column (tab-aware)

---

## Phase E: Polish & Release

- [x] Cross-platform binary builds (Linux x64, macOS arm64/x64, Windows x64) via GitHub Actions
- [x] Single-binary, zero-dependency install (static linking where possible)
- [x] `tengo --version` and `tengo --help`
- [x] Man page / shell completion (bash, zsh, fish)
- [x] README with install instructions and usage examples
- [x] Release assets on GitHub Releases
- [x] Install scripts (Linux/macOS bash, Windows PowerShell)
- [x] `tengo -update` self-update command
- [x] Update available notification in editor status bar

## Phase F: Distribution

- [ ] **Homebrew tap** — create a separate GitHub repo named `homebrew-tengo`, copy `Formula/tengo.rb` into it, push
  - After each release: update `url` to new tag, run `curl -sL <tarball-url> | sha256sum`, update `sha256` in formula, push
  - Users install with: `brew tap helloWorld44-89/tengo && brew install tengo`
- [ ] **AUR (Arch Linux)** — write a `PKGBUILD` file, create an account at aur.archlinux.org, publish package
  - Low effort, no approval process, covers all Arch/Manjaro users
  - Users install with: `yay -S tengo`
- [ ] **Snapcraft** — write `snap/snapcraft.yaml`, publish to snapcraft.io
  - Works on any Linux distro with snapd (Ubuntu pre-installed)
  - Users install with: `snap install tengo`

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
