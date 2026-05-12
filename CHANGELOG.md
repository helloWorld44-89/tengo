# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] - 2025-05-12

Initial release.

### Interactive Editor
- Full-screen terminal editor with raw-mode input
- Multi-file tabs (`tengo file1.yaml file2.json`), switch with `Ctrl+W` or `Alt+1`…`Alt+9`
- Undo / redo stacks (`Ctrl+Z` / `Ctrl+Y`)
- Find (`Ctrl+F`) and find & replace (`Ctrl+R`)
- System clipboard: copy (`Ctrl+C`), cut (`Ctrl+X`), paste (`Ctrl+V`)
- Syntax highlighting for YAML, JSON, TOML, INI keys and values
- Inline syntax validation on save (`Ctrl+S`) and on demand (`Ctrl+E`) with gutter error marker
- Auto-format / pretty-print (`Ctrl+T`) for YAML, JSON, TOML
- Go to line (`Ctrl+G`), toggle line numbers (`Ctrl+N`), toggle word-wrap (`Alt+W`)
- Auto-close brackets and quotes; skip-over on closing character
- Word-level navigation (`Ctrl+Left` / `Ctrl+Right`)
- Fast 4-step navigation (`Alt+Arrow`)
- Duplicate line (`Ctrl+D`), toggle comment (`Ctrl+/`)
- Insert line above / below (`Ctrl+Shift+Enter` / `Shift+Enter`)
- Indent / outdent selection (`Ctrl+]` / `Ctrl+[`)
- Status bar: unsaved indicator, file type, cursor position, selection size
- Current-line highlight; help overlay (`Ctrl+H`)
- Bracketed paste support

### CLI Mode
- `-F <text>` — find all occurrences with line numbers and counts
- `-F <text> -R <new>` — interactive find & replace (per-occurrence prompts)
- `--confirm-all` — apply all replacements without prompting
- `--dry-run` — colored unified diff preview without writing
- `--stdout` — print result to stdout instead of saving in place
- `--backup` — write `.bak` before any destructive write
- `-validate` — validate YAML, JSON, TOML, XML syntax; exit 0/1
- `-format` — auto-format / pretty-print in place
- `-key <path>` — read a dot-notation key value (YAML, JSON, TOML)
- `-set <path>=<value>` — set a dot-notation key value
- `-L <n>` — print line n and exit
- `-type <ext>` — specify file type when reading from stdin
- `--regex` — treat `-F` pattern as a regular expression
- Pipe support: `cat config.yaml | tengo -validate -type yaml`
- Proper exit codes: 0 = success, 1 = invalid/not found, 2 = I/O error

### Shell Completions
- `-completion bash` — bash completion script
- `-completion zsh` — zsh completion script
- `-completion fish` — fish completion script

### Security
- Atomic saves (temp file + rename) with original permissions preserved
- 50 MB file size cap on open (interactive and CLI)
- YAML alias expansion depth limit (100 000-node cap) to prevent billion-laughs DoS
- CRLF normalization on Windows file reads
- `-type` flag restricted to known file types
