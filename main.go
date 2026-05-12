package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"tengo/editor"
)

func printSplash() {
	fmt.Print(`
  ╔══════════════════════════════════════════════════╗
  ║                                                  ║
  ║    t e n g o   —   quick CLI text editor         ║
  ║    YAML · JSON · TOML · INI · XML · Go           ║
  ║                                                  ║
  ╚══════════════════════════════════════════════════╝

  Usage:
    tengo <file>                         Open file in interactive editor
    tengo <file> -F <text>               Find all occurrences with line numbers
    tengo <file> -F <text> -R <new>      Interactive find & replace
    tengo <file> -F <text> -R <new> --confirm-all   Replace all without prompts

  Keyboard shortcuts (inside editor):
    Ctrl+S   Save       Ctrl+Z / Ctrl+Y   Undo / Redo
    Ctrl+F   Find       Ctrl+R            Find & Replace
    Ctrl+A   Select all Ctrl+N            Toggle line numbers
    Ctrl+H   Help       Ctrl+Q / Esc      Quit

`)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `tengo — quick CLI editor for structured config files

Usage:
  tengo <file>                                     Open file in interactive editor
  tengo <file> -F <text>                           Find all occurrences with line numbers
  tengo <file> -F <text> -R <new>                  Interactive find & replace (confirm each)
  tengo <file> -F <text> -R <new> --confirm-all    Replace all, show diff, no prompt

Flags:
`)
	flag.PrintDefaults()
}

func main() {
	findFlag := flag.String("F", "", "Find occurrences of `text`")
	replaceFlag := flag.String("R", "", "Replacement `text` (requires -F)")
	confirmAll := flag.Bool("confirm-all", false, "Apply all replacements without per-occurrence confirmation")
	dryRun := flag.Bool("dry-run", false, "Show changes without writing to disk")
	toStdout := flag.Bool("stdout", false, "Print result to stdout instead of saving in-place")
	backup := flag.Bool("backup", false, "Write a .bak backup before modifying the file")
	quiet := flag.Bool("q", false, "Suppress output; rely on exit codes only")
	validateFlag := flag.Bool("validate", false, "Validate file syntax and exit (0=valid, 1=invalid, 2=error)")
	formatFlag := flag.Bool("format", false, "Auto-format/pretty-print the file in place")

	flag.Usage = printUsage

	// If the first argument is a filename (not a flag), extract it before
	// flag.Parse so that flags appearing after the filename are still parsed.
	var filePath string
	if len(os.Args) >= 2 && !strings.HasPrefix(os.Args[1], "-") {
		filePath = os.Args[1]
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	flag.Parse()

	// Filename may also appear as a positional arg after all flags.
	if filePath == "" {
		if flag.NArg() == 0 {
			printSplash()
			os.Exit(0)
		}
		filePath = flag.Arg(0)
	}

	// CLI (non-interactive) mode.
	if *validateFlag {
		os.Exit(runValidate(filePath, *quiet))
	}
	if *formatFlag {
		os.Exit(runFormat(filePath, *dryRun, *toStdout, *backup, *quiet))
	}
	if *findFlag != "" {
		if *replaceFlag == "" {
			os.Exit(runFind(filePath, *findFlag, *quiet))
		}
		os.Exit(runReplace(filePath, *findFlag, *replaceFlag,
			*confirmAll, *dryRun, *backup, *toStdout, *quiet))
	}

	// Interactive editor mode.
	restore := editor.EnableRaw()
	defer restore()

	fmt.Print("\x1b[?2004h")
	defer fmt.Print("\x1b[?2004l")

	fmt.Print("\x1b[?1049h")
	defer fmt.Print("\x1b[?1049l")

	editor.RunQuickEditor(filePath)
}
