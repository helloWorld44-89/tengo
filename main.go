package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"github.com/helloWorld44-89/tengo/editor"
)

var version = "0.1.0"

func printSplash() {
	fmt.Print(`
  ╔══════════════════════════════════════════════════╗
  ║                                                  ║
  ║    t e n g o   —   quick CLI text editor         ║
  ║    YAML · JSON · TOML · INI · XML · Go           ║
  ║                                                  ║
  ╚══════════════════════════════════════════════════╝

  Usage:
    tengo <file>                             Open file in interactive editor
    tengo <file> -F <text>                   Find occurrences (line numbers + count)
    tengo <file> -F <text> -R <new>          Interactive find & replace
    tengo <file> -F <text> -R <new> --confirm-all   Replace all without prompts
    tengo <file> -validate                   Validate file syntax
    tengo <file> -format                     Pretty-print file in place
    tengo <file> -L <n>                      Print line n
    cat file.yaml | tengo -validate -type yaml       Validate from stdin

  Keyboard shortcuts (inside editor):
    Ctrl+S   Save / validate   Ctrl+Z / Ctrl+Y   Undo / Redo
    Ctrl+E   Validate syntax   Ctrl+T            Format file
    Ctrl+F   Find              Ctrl+R            Find & Replace
    Ctrl+G   Go to line        Ctrl+N            Toggle line numbers
    Ctrl+H   Help              Ctrl+Q / Esc      Quit

`)
}

func printUsage() {
	fmt.Print(`tengo ` + version + ` — quick CLI editor for structured config files

Usage:
  tengo <file> [file2 ...]          Open file(s) in interactive editor
  tengo <file> -F <text>            Find all occurrences with line numbers
  tengo <file> -F <text> -R <new>   Find & replace (interactive per-occurrence)
  tengo <file> -validate            Validate file syntax
  tengo <file> -format              Pretty-print file in place
  tengo <file> -key <path>          Read value at dot-notation key path
  tengo <file> -set <path>=<value>  Set value at key path
  tengo <file> -L <n>               Print line n and exit
  cat file.yaml | tengo -validate -type yaml

Flags:
`)
	flag.PrintDefaults()
	fmt.Print(`
Interactive editor shortcuts:
  Ctrl+S / Ctrl+Q   Save / Quit          Ctrl+Z / Ctrl+Y   Undo / Redo
  Ctrl+F / Ctrl+R   Find / Find&Replace  Ctrl+H            Help overlay
  Ctrl+E / Ctrl+T   Validate / Format    Ctrl+N            Toggle line numbers
  Ctrl+G            Go to line           Alt+W             Toggle word-wrap
  Ctrl+[ / Ctrl+]   Outdent / Indent     Ctrl+W            Next tab
  Ctrl+C / Ctrl+X   Copy / Cut           Ctrl+V            Paste
`)
}

func main() {
	versionFlag := flag.Bool("version", false, "Print version and exit")
	updateFlag := flag.Bool("update", false, "Download and install the latest release")
	completionFlag := flag.String("completion", "", "Print shell completion script: bash, zsh, or fish")
	findFlag := flag.String("F", "", "Find occurrences of `text`")
	replaceFlag := flag.String("R", "", "Replacement `text` (requires -F)")
	confirmAll := flag.Bool("confirm-all", false, "Apply all replacements without per-occurrence confirmation")
	dryRun := flag.Bool("dry-run", false, "Show unified diff of changes without writing to disk")
	toStdout := flag.Bool("stdout", false, "Print result to stdout instead of saving in-place")
	backup := flag.Bool("backup", false, "Write a .bak backup before modifying the file")
	quiet := flag.Bool("q", false, "Suppress output; rely on exit codes only")
	useRegex := flag.Bool("regex", false, "Treat -F pattern as a regular expression")
	validateFlag := flag.Bool("validate", false, "Validate file syntax and exit (0=valid, 1=invalid, 2=error)")
	formatFlag := flag.Bool("format", false, "Auto-format/pretty-print the file in place")
	lineFlag := flag.Int("L", 0, "Print line `n` (1-based) and exit")
	typeFlag := flag.String("type", "", "File `type` for stdin: yaml, json, toml, xml")
	keyFlag := flag.String("key", "", "Read value at dot-notation `path` (e.g. database.host)")
	setFlag := flag.String("set", "", "Set value at dot-notation path (e.g. database.port=5432)")

	flag.Usage = printUsage

	// --version / --help before any filename extraction.
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-version") {
		fmt.Println("tengo " + version)
		os.Exit(0)
	}

	// If the first argument is a filename (not a flag), extract it before
	// flag.Parse so that flags appearing after the filename are still parsed.
	var filePath string
	if len(os.Args) >= 2 && !strings.HasPrefix(os.Args[1], "-") {
		filePath = os.Args[1]
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	flag.Parse()

	if *completionFlag != "" {
		switch *completionFlag {
		case "bash":
			fmt.Print(bashCompletion)
		case "zsh":
			fmt.Print(zshCompletion)
		case "fish":
			fmt.Print(fishCompletion)
		default:
			fmt.Fprintf(os.Stderr, "Unknown shell %q; use: bash, zsh, or fish\n", *completionFlag)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Filename may also appear as a positional arg after all flags.
	if filePath == "" {
		if flag.NArg() > 0 {
			filePath = flag.Arg(0)
		} else if stdinIsPiped() {
			filePath = "-" // read from stdin
		} else {
			printSplash()
			os.Exit(0)
		}
	}

	// typePath is used for file-type detection; it may be a synthetic name
	// when the user overrides via -type or when reading from stdin.
	typePath := filePath
	if *typeFlag != "" {
		switch strings.ToLower(*typeFlag) {
		case "yaml", "yml", "json", "toml", "xml", "ini":
			typePath = "stdin." + strings.ToLower(*typeFlag)
		default:
			fmt.Fprintf(os.Stderr, "Unknown -type %q; valid types: yaml, json, toml, xml, ini\n", *typeFlag)
			os.Exit(1)
		}
	}

	if *versionFlag {
		fmt.Println("tengo " + version)
		os.Exit(0)
	}

	if *updateFlag {
		os.Exit(runUpdate(*quiet))
	}

	// CLI (non-interactive) mode.
	if *keyFlag != "" {
		os.Exit(runKeyGet(filePath, *keyFlag, typePath, *quiet))
	}
	if *setFlag != "" {
		os.Exit(runKeySet(filePath, *setFlag, typePath, *dryRun, *backup, *toStdout, *quiet))
	}
	if *lineFlag > 0 {
		os.Exit(runLine(filePath, *lineFlag, *quiet))
	}
	if *validateFlag {
		os.Exit(runValidate(filePath, typePath, *quiet))
	}
	if *formatFlag {
		os.Exit(runFormat(filePath, typePath, *dryRun, *toStdout, *backup, *quiet))
	}
	if *findFlag != "" {
		if *replaceFlag == "" {
			os.Exit(runFind(filePath, *findFlag, *useRegex, *quiet))
		}
		os.Exit(runReplace(filePath, *findFlag, *replaceFlag,
			*useRegex, *confirmAll, *dryRun, *backup, *toStdout, *quiet))
	}

	// Refuse to open interactive editor on stdin.
	if filePath == "-" {
		fmt.Fprintln(os.Stderr, "Specify a CLI flag (-validate, -format, -F, -L) when reading from stdin.")
		os.Exit(1)
	}

	// Collect all file paths: the first one (pre-extracted) plus any positional args.
	filePaths := []string{filePath}
	filePaths = append(filePaths, flag.Args()...)

	// Interactive editor mode.

	// Check for updates in the background; notify via channel when found.
	updateCh := make(chan string, 1)
	go func() {
		latest, err := fetchLatestTag()
		if err != nil {
			return
		}
		if strings.TrimPrefix(latest, "v") != strings.TrimPrefix(version, "v") {
			updateCh <- fmt.Sprintf("Update available %s — run: tengo -update", latest)
		}
	}()

	restore := editor.EnableRaw()
	defer restore()

	fmt.Print("\x1b[?2004h")
	defer fmt.Print("\x1b[?2004l")

	fmt.Print("\x1b[?1049h")
	defer fmt.Print("\x1b[?1049l")

	editor.RunQuickEditor(filePaths, updateCh)
}
