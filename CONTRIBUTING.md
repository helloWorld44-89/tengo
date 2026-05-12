# Contributing to tengo

## Prerequisites

- Go 1.23 or later
- A terminal that supports ANSI escape sequences

## Build

```bash
git clone https://github.com/helloWorld44-89/tengo.git
cd tengo
go build -o tengo .
```

## Run

```bash
go run . <filename>          # interactive editor
go run . <filename> -validate  # CLI mode
```

## Test

```bash
go test ./...                  # all packages
go test ./editor/... -v        # editor package, verbose
go test ./editor/... -run TestToBuffer  # single test
```

## Project layout

```
main.go          Entry point; flag parsing; routes to editor or CLI mode
cli.go           Non-interactive CLI commands (runFind, runReplace, etc.)
keypath.go       Dot-notation key-path navigation for YAML/JSON/TOML
completion.go    Embedded shell completion scripts

editor/
  quickEditor.go  Main event loop; undo/redo; keystroke dispatch
  buffer.go       Core buffer primitives; cursor; selection helpers
  input.go        Raw key reading; ANSI escape parsing; moveCursor
  render.go       Full-screen redraw; syntax highlighting dispatch
  clipboard.go    System clipboard integration with internal fallback
  validate.go     Syntax validation for YAML/JSON/TOML/XML
  format.go       Auto-formatting for YAML/JSON/TOML
  highlight.go    Syntax highlighting (ANSI colors)
  terminal.go     Raw mode; terminal size
  help.go         Help overlay popup

file/
  file.go         OpenFile / SaveFile / SaveBytes (atomic writes)
```

## Submitting changes

1. Fork the repo and create a branch from `main`.
2. Make your changes and ensure `go test ./...` and `go vet ./...` pass.
3. Keep commits focused; one logical change per commit.
4. Open a pull request against `main` with a clear description of what and why.

## Code style

- Standard `gofmt` formatting (enforced by `go vet` in CI).
- No comments that restate what the code does — only add one when the *why* is non-obvious.
- No panics in user-facing paths; return errors instead.
- All file writes go through `file.SaveBytes` (atomic temp-file + rename).

## Reporting bugs

Open an issue on GitHub with:
- The tengo version (`tengo --version`)
- OS and terminal emulator
- Steps to reproduce
- Expected vs. actual behaviour
