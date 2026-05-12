package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"tengo/editor"
	"tengo/file"
)

const (
	clrRed   = "\x1b[31;1m"
	clrGreen = "\x1b[32;1m"
	clrBold  = "\x1b[1m"
	clrReset = "\x1b[0m"
)

// wrapOccurrences wraps every occurrence of term in line with the given ANSI color.
func wrapOccurrences(line, term, color string) string {
	parts := strings.Split(line, term)
	return strings.Join(parts, color+term+clrReset)
}

// wrapReplacement rebuilds line with replacement shown in color at every position
// where term appeared, without affecting pre-existing text.
func wrapReplacement(line, term, replacement, color string) string {
	parts := strings.Split(line, term)
	return strings.Join(parts, color+replacement+clrReset)
}

func readFileLines(path string) (lines []string, original []byte, err error) {
	original, err = os.ReadFile(path)
	if err != nil {
		return
	}
	lines = strings.Split(string(original), "\n")
	// Drop the phantom empty element produced by a trailing \n.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return
}

// runFind prints every line containing term, with line numbers and counts.
//
//	Exit 0 — matches found
//	Exit 1 — no matches
//	Exit 2 — I/O error
func runFind(path, term string, quiet bool) int {
	lines, _, err := readFileLines(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	type hit struct {
		lineNum int
		content string
		count   int
	}

	var hits []hit
	totalCount := 0
	for i, line := range lines {
		n := strings.Count(line, term)
		if n > 0 {
			hits = append(hits, hit{i + 1, line, n})
			totalCount += n
		}
	}

	if len(hits) == 0 {
		if !quiet {
			fmt.Printf("No occurrences of %q found in %s\n", term, path)
		}
		return 1
	}

	if !quiet {
		fmt.Printf("%sFound %d occurrence(s) of %q across %d line(s) in %s:%s\n\n",
			clrBold, totalCount, term, len(hits), path, clrReset)
		for _, h := range hits {
			label := fmt.Sprintf("  Line %d", h.lineNum)
			if h.count > 1 {
				label += fmt.Sprintf(" (%d×)", h.count)
			}
			fmt.Printf("%s:  %s\n", label, wrapOccurrences(h.content, term, clrRed))
		}
		fmt.Println()
	}

	return 0
}

// runReplace finds every occurrence of term and replaces with replacement.
//
// Without --confirm-all the user is prompted (y/n/q) for each occurrence.
// With    --confirm-all all changes are applied immediately; the diff is still printed.
//
//	Exit 0 — success or clean abort
//	Exit 1 — no matches
//	Exit 2 — I/O error
func runReplace(path, term, replacement string, confirmAll, dryRun, backup, toStdout, quiet bool) int {
	lines, original, err := readFileLines(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	type occurrence struct {
		lineIdx int
		oldLine string
		newLine string
		count   int
	}

	var occs []occurrence
	totalCount := 0
	for i, line := range lines {
		n := strings.Count(line, term)
		if n > 0 {
			occs = append(occs, occurrence{
				lineIdx: i,
				oldLine: line,
				newLine: strings.ReplaceAll(line, term, replacement),
				count:   n,
			})
			totalCount += n
		}
	}

	if len(occs) == 0 {
		if !quiet {
			fmt.Printf("No occurrences of %q found in %s\n", term, path)
		}
		return 1
	}

	if !quiet {
		fmt.Printf("%sFound %d occurrence(s) of %q in %s%s\n\n",
			clrBold, totalCount, term, path, clrReset)
	}

	newLines := make([]string, len(lines))
	copy(newLines, lines)

	replacedLines := 0
	reader := bufio.NewReader(os.Stdin)

	for i, occ := range occs {
		if !quiet {
			fmt.Printf("%s[%d/%d] Line %d:%s\n", clrBold, i+1, len(occs), occ.lineIdx+1, clrReset)
			fmt.Printf("  %s-%s %s\n",
				clrRed, clrReset,
				wrapOccurrences(occ.oldLine, term, clrRed))
			fmt.Printf("  %s+%s %s\n",
				clrGreen, clrReset,
				wrapReplacement(occ.oldLine, term, replacement, clrGreen))
		}

		if confirmAll {
			newLines[occ.lineIdx] = occ.newLine
			replacedLines++
			if !quiet {
				fmt.Println()
			}
			continue
		}

		fmt.Print("  Replace? (y/n/q): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		fmt.Println()

		switch answer {
		case "y":
			newLines[occ.lineIdx] = occ.newLine
			replacedLines++
		case "q":
			if !quiet {
				fmt.Println("Aborted.")
			}
			return 0
		// "n" or anything else: skip this occurrence
		}
	}

	if replacedLines == 0 {
		if !quiet {
			fmt.Println("No replacements made.")
		}
		return 0
	}

	result := strings.Join(newLines, "\n") + "\n"

	if toStdout {
		fmt.Print(result)
		return 0
	}

	if dryRun {
		if !quiet {
			fmt.Printf("[dry-run] Would modify %d line(s). No file written.\n", replacedLines)
		}
		return 0
	}

	if backup {
		backupPath := path + ".bak"
		if err := os.WriteFile(backupPath, original, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
			return 2
		}
		if !quiet {
			fmt.Printf("Backup written to %s\n", backupPath)
		}
	}

	if err := file.SaveBytes(path, []byte(result)); err != nil {
		fmt.Fprintf(os.Stderr, "Save failed: %v\n", err)
		return 2
	}

	if !quiet {
		fmt.Printf("Replaced %d line(s). Saved %s.\n", replacedLines, path)
	}

	return 0
}

// runValidate validates file syntax and exits.
//
//	Exit 0 — valid
//	Exit 1 — invalid (syntax error)
//	Exit 2 — I/O error
func runValidate(path string, quiet bool) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	result := editor.ValidateContent(path, string(content))
	if result.Valid {
		if !quiet {
			fmt.Printf("%s✓ %s%s\n", clrGreen, result.Message, clrReset)
		}
		return 0
	}

	if !quiet {
		if result.Line > 0 {
			fmt.Printf("%s✗ Line %d: %s%s\n", clrRed, result.Line, result.Message, clrReset)
		} else {
			fmt.Printf("%s✗ %s%s\n", clrRed, result.Message, clrReset)
		}
	}
	return 1
}

// runFormat auto-formats the file in place (or to stdout with --stdout).
//
//	Exit 0 — success
//	Exit 1 — parse/format error
//	Exit 2 — I/O error
func runFormat(path string, dryRun, toStdout, backup, quiet bool) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	formatted, err := editor.FormatContent(path, string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Format error: %v\n", err)
		return 1
	}

	if toStdout {
		fmt.Print(formatted)
		return 0
	}

	if dryRun {
		if !quiet {
			fmt.Printf("[dry-run] Would format %s\n", path)
			fmt.Println(formatted)
		}
		return 0
	}

	if backup {
		backupPath := path + ".bak"
		if err := os.WriteFile(backupPath, content, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
			return 2
		}
		if !quiet {
			fmt.Printf("Backup written to %s\n", backupPath)
		}
	}

	if err := file.SaveBytes(path, []byte(formatted)); err != nil {
		fmt.Fprintf(os.Stderr, "Save failed: %v\n", err)
		return 2
	}

	if !quiet {
		fmt.Printf("Formatted %s\n", path)
	}
	return 0
}
