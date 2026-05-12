# tengo

A fast, zero-config terminal editor for structured config files.

Open any YAML, JSON, TOML, INI, or XML file for interactive editing, or run non-interactive transforms directly from the command line — no setup required.

---

## Install

### Download a binary

Pre-built binaries for Linux, macOS, and Windows are available on the [Releases](https://github.com/helloWorld44-89/tengo/releases) page.

```bash
# Linux x64 example
curl -Lo tengo https://github.com/helloWorld44-89/tengo/releases/latest/download/tengo-linux-amd64
chmod +x tengo
sudo mv tengo /usr/local/bin/
```

### Build from source

Requires Go 1.23+.

```bash
git clone https://github.com/helloWorld44-89/tengo.git
cd tengo
go build -o tengo .
```

---

## Interactive Editor

```bash
tengo config.yaml
tengo file1.yaml file2.json   # open multiple files in tabs
```

### Keyboard Shortcuts

| Action | Shortcut |
|---|---|
| Save | `Ctrl+S` |
| Quit | `Ctrl+Q` or `Esc` |
| Undo / Redo | `Ctrl+Z` / `Ctrl+Y` |
| Find | `Ctrl+F` |
| Find & Replace | `Ctrl+R` |
| Validate syntax | `Ctrl+E` |
| Format / pretty-print | `Ctrl+T` |
| Go to line | `Ctrl+G` |
| Toggle line numbers | `Ctrl+N` |
| Toggle word-wrap | `Alt+W` |
| Select all | `Ctrl+A` |
| Copy / Cut / Paste | `Ctrl+C` / `Ctrl+X` / `Ctrl+V` |
| Indent / Outdent | `Ctrl+]` / `Ctrl+[` |
| Duplicate line | `Ctrl+D` |
| Toggle comment | `Ctrl+/` |
| Insert line below | `Shift+Enter` |
| Insert line above | `Ctrl+Shift+Enter` |
| Word jump + select | `Ctrl+Arrow` |
| Fast navigation (4 steps) | `Alt+Arrow` |
| Next tab | `Ctrl+W` |
| Switch to tab N | `Alt+1`…`Alt+9` |
| Help overlay | `Ctrl+H` |

---

## CLI Mode

Non-interactive flags operate on the file and exit immediately — no editor is opened.

### Find

```bash
tengo config.yaml -F "localhost"
tengo config.yaml -F "debug: true" --regex
```

### Find & Replace

```bash
# Interactive (confirm each match)
tengo config.yaml -F "localhost" -R "0.0.0.0"

# Apply all without prompts
tengo config.yaml -F "localhost" -R "0.0.0.0" --confirm-all

# Preview changes without writing
tengo config.yaml -F "localhost" -R "0.0.0.0" --dry-run

# Print result to stdout instead of saving
tengo config.yaml -F "localhost" -R "0.0.0.0" --stdout
```

### Validate

```bash
tengo config.yaml -validate           # exit 0 = valid, 1 = invalid
cat config.yaml | tengo -validate -type yaml
```

### Format / Pretty-print

```bash
tengo config.json -format             # in place
tengo config.json -format --dry-run   # preview diff
tengo config.json -format --stdout    # print to stdout
tengo config.json -format --backup    # save .bak before writing
```

### Key-path navigation (YAML / JSON / TOML)

```bash
# Read a value
tengo config.yaml -key database.host
tengo config.yaml -key servers.0.port

# Set a value (note: reformats the file on save)
tengo config.yaml -set database.port=5432
tengo config.yaml -set database.host=0.0.0.0 --dry-run
tengo config.yaml -set debug=false --backup
```

### Other flags

```bash
tengo config.yaml -L 42              # print line 42 and exit
tengo --version                      # print version
tengo --help                         # print usage
```

---

## Supported File Types

| Extension | Syntax highlight | Validate | Format | -key / -set |
|---|:---:|:---:|:---:|:---:|
| `.yaml` / `.yml` | ✓ | ✓ | ✓ | ✓ |
| `.json` | ✓ | ✓ | ✓ | ✓ |
| `.toml` | ✓ | ✓ | ✓ | ✓ |
| `.ini` / `.cfg` | ✓ | ✓ | — | — |
| `.xml` | — | ✓ | — | — |
| `.go` / `.md` / `.txt` | — | — | — | — |

---

## Pipe support

```bash
cat config.yaml | tengo -validate -type yaml
cat config.json | tengo -format -type json --stdout
cat config.yaml | tengo -F "host" -type yaml
```

---

## License

MIT
