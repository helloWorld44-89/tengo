package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"tengo/editor"
	"tengo/file"
)

const (
	clrRed   = "\x1b[31;1m"
	clrGreen = "\x1b[32;1m"
	clrCyan  = "\x1b[36m"
	clrBold  = "\x1b[1m"
	clrDim   = "\x1b[2m"
	clrReset = "\x1b[0m"
)

// readFileLines reads content from a file or stdin ("-"), splitting into lines.
// The phantom empty element produced by a trailing \n is dropped.
func readFileLines(path string) (lines []string, original []byte, err error) {
	if path == "-" {
		original, err = io.ReadAll(os.Stdin)
	} else {
		original, err = os.ReadFile(path)
	}
	if err != nil {
		return
	}
	lines = strings.Split(string(original), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return
}

// stdinIsPiped reports whether stdin is connected to a pipe rather than a terminal.
func stdinIsPiped() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

// ── Plain-text highlight helpers ─────────────────────────────────────────────

func wrapOccurrences(line, term, color string) string {
	parts := strings.Split(line, term)
	return strings.Join(parts, color+term+clrReset)
}

func wrapReplacement(line, term, replacement, color string) string {
	parts := strings.Split(line, term)
	return strings.Join(parts, color+replacement+clrReset)
}

// ── Regex highlight helpers ───────────────────────────────────────────────────

// wrapMatchesRegex highlights every regex match in line using color.
func wrapMatchesRegex(line string, re *regexp.Regexp, color string) string {
	locs := re.FindAllStringIndex(line, -1)
	if len(locs) == 0 {
		return line
	}
	var b strings.Builder
	prev := 0
	for _, loc := range locs {
		b.WriteString(line[prev:loc[0]])
		b.WriteString(color)
		b.WriteString(line[loc[0]:loc[1]])
		b.WriteString(clrReset)
		prev = loc[1]
	}
	b.WriteString(line[prev:])
	return b.String()
}

// wrapReplacementRegex shows replacement text in color at each match position.
func wrapReplacementRegex(line, replacement string, re *regexp.Regexp, color string) string {
	return re.ReplaceAllString(line, color+replacement+clrReset)
}

// ── Unified diff ─────────────────────────────────────────────────────────────

// unifiedDiff produces a colored unified diff between original and modified line slices.
func unifiedDiff(path string, original, modified []string) string {
	const ctx = 3

	// Collect 0-based indices of changed lines.
	n := len(original)
	if len(modified) > n {
		n = len(modified)
	}
	var changed []int
	for i := 0; i < n; i++ {
		var a, b string
		if i < len(original) {
			a = original[i]
		}
		if i < len(modified) {
			b = modified[i]
		}
		if a != b {
			changed = append(changed, i)
		}
	}
	if len(changed) == 0 {
		return ""
	}

	// Group nearby changes into hunks.
	type hunk struct{ start, end int }
	var hunks []hunk
	hs, he := changed[0], changed[0]
	for _, c := range changed[1:] {
		if c-he <= 2*ctx {
			he = c
		} else {
			hunks = append(hunks, hunk{hs, he})
			hs, he = c, c
		}
	}
	hunks = append(hunks, hunk{hs, he})

	var sb strings.Builder
	sb.WriteString(clrBold + "--- a/" + path + clrReset + "\n")
	sb.WriteString(clrBold + "+++ b/" + path + clrReset + "\n")

	for _, h := range hunks {
		lo := h.start - ctx
		if lo < 0 {
			lo = 0
		}
		hi := h.end + ctx + 1
		if hi > len(original) {
			hi = len(original)
		}

		sb.WriteString(clrCyan + fmt.Sprintf("@@ -%d,%d +%d,%d @@", lo+1, hi-lo, lo+1, hi-lo) + clrReset + "\n")

		for i := lo; i < hi; i++ {
			var a, b string
			if i < len(original) {
				a = original[i]
			}
			if i < len(modified) {
				b = modified[i]
			}
			if a != b {
				sb.WriteString(clrRed + "-" + a + clrReset + "\n")
				sb.WriteString(clrGreen + "+" + b + clrReset + "\n")
			} else {
				sb.WriteString(clrDim + " " + a + clrReset + "\n")
			}
		}
	}
	return sb.String()
}

// ── runLine ───────────────────────────────────────────────────────────────────

// runLine prints a single line from the file (1-based).
//
//	Exit 0 — success
//	Exit 1 — out of range
//	Exit 2 — I/O error
func runLine(path string, n int, quiet bool) int {
	lines, _, err := readFileLines(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}
	if n < 1 || n > len(lines) {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Line %d out of range (file has %d lines)\n", n, len(lines))
		}
		return 1
	}
	fmt.Println(lines[n-1])
	return 0
}

// ── runFind ───────────────────────────────────────────────────────────────────

// runFind prints every line containing term, with line numbers and counts.
//
//	Exit 0 — matches found
//	Exit 1 — no matches
//	Exit 2 — I/O error or bad regex
func runFind(path, term string, useRegex, quiet bool) int {
	lines, _, err := readFileLines(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	// Build matcher / highlighter.
	var re *regexp.Regexp
	if useRegex {
		re, err = regexp.Compile(term)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid regex: %v\n", err)
			return 2
		}
	}

	type hit struct {
		lineNum int
		content string
		count   int
	}

	var hits []hit
	totalCount := 0
	for i, line := range lines {
		var n int
		if useRegex {
			n = len(re.FindAllString(line, -1))
		} else {
			n = strings.Count(line, term)
		}
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
			var highlighted string
			if useRegex {
				highlighted = wrapMatchesRegex(h.content, re, clrRed)
			} else {
				highlighted = wrapOccurrences(h.content, term, clrRed)
			}
			fmt.Printf("%s:  %s\n", label, highlighted)
		}
		fmt.Println()
	}

	return 0
}

// ── runReplace ────────────────────────────────────────────────────────────────

// runReplace finds every occurrence of term and replaces with replacement.
// --dry-run shows a unified diff without writing.
// --confirm-all applies all without prompting; per-occurrence prompts otherwise.
//
//	Exit 0 — success or clean abort
//	Exit 1 — no matches
//	Exit 2 — I/O error or bad regex
func runReplace(path, term, replacement string, useRegex, confirmAll, dryRun, backup, toStdout, quiet bool) int {
	lines, original, err := readFileLines(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	var re *regexp.Regexp
	if useRegex {
		re, err = regexp.Compile(term)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid regex: %v\n", err)
			return 2
		}
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
		var n int
		var newLine string
		if useRegex {
			n = len(re.FindAllString(line, -1))
			newLine = re.ReplaceAllString(line, replacement)
		} else {
			n = strings.Count(line, term)
			newLine = strings.ReplaceAll(line, term, replacement)
		}
		if n > 0 {
			occs = append(occs, occurrence{i, line, newLine, n})
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
	// dry-run implicitly applies all so we can compute the diff.
	autoApply := confirmAll || dryRun
	reader := bufio.NewReader(os.Stdin)

	for i, occ := range occs {
		if !quiet {
			fmt.Printf("%s[%d/%d] Line %d:%s\n", clrBold, i+1, len(occs), occ.lineIdx+1, clrReset)
			var oldDisplay, newDisplay string
			if useRegex {
				oldDisplay = wrapMatchesRegex(occ.oldLine, re, clrRed)
				newDisplay = wrapReplacementRegex(occ.oldLine, replacement, re, clrGreen)
			} else {
				oldDisplay = wrapOccurrences(occ.oldLine, term, clrRed)
				newDisplay = wrapReplacement(occ.oldLine, term, replacement, clrGreen)
			}
			fmt.Printf("  %s-%s %s\n", clrRed, clrReset, oldDisplay)
			fmt.Printf("  %s+%s %s\n", clrGreen, clrReset, newDisplay)
		}

		if autoApply {
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
			fmt.Print(unifiedDiff(path, lines, newLines))
			fmt.Printf("\n%s[dry-run]%s %d line(s) would be modified. No file written.\n",
				clrBold, clrReset, replacedLines)
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

// ── runValidate ───────────────────────────────────────────────────────────────

// runValidate validates file syntax and exits.
// typePath is used for type detection (may differ from path when reading stdin).
//
//	Exit 0 — valid
//	Exit 1 — invalid
//	Exit 2 — I/O error
func runValidate(path, typePath string, quiet bool) int {
	var content []byte
	var err error
	if path == "-" {
		content, err = io.ReadAll(os.Stdin)
	} else {
		content, err = os.ReadFile(path)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	result := editor.ValidateContent(typePath, string(content))
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

// ── runKeyGet ─────────────────────────────────────────────────────────────────

// runKeyGet reads a value at a dot-notation key path and prints it.
//
//	Exit 0 — value found and printed
//	Exit 1 — key not found or parse error
//	Exit 2 — I/O error
func runKeyGet(path, keyPath, typePath string, quiet bool) int {
	content, err := readFileContent(path)
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return 2
	}
	data, err := unmarshalGeneric(typePath, content)
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		}
		return 1
	}
	segments := parseKeyPath(keyPath)
	val, err := getAtPath(data, segments)
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Key error: %v\n", err)
		}
		return 1
	}
	if !quiet {
		fmt.Println(printValue(val))
	}
	return 0
}

// ── runKeySet ─────────────────────────────────────────────────────────────────

// runKeySet sets a value at a dot-notation key path and saves the file.
// assignment must be in the form "key.path=value".
// Note: the file is re-marshaled from the parsed structure, so original
// formatting and comments are not preserved.
//
//	Exit 0 — value set and saved
//	Exit 1 — parse/key error
//	Exit 2 — I/O error
func runKeySet(path, assignment, typePath string, dryRun, backup, toStdout, quiet bool) int {
	eqIdx := strings.Index(assignment, "=")
	if eqIdx < 0 {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Invalid -set value %q: expected key.path=value\n", assignment)
		}
		return 1
	}
	keyPath := assignment[:eqIdx]
	valueStr := assignment[eqIdx+1:]

	content, err := readFileContent(path)
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return 2
	}
	original := content

	data, err := unmarshalGeneric(typePath, content)
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		}
		return 1
	}

	segments := parseKeyPath(keyPath)
	newData, err := setAtPath(data, segments, parseScalarValue(valueStr))
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Key error: %v\n", err)
		}
		return 1
	}

	result, err := marshalGeneric(typePath, newData)
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Marshal error: %v\n", err)
		}
		return 1
	}

	if toStdout || path == "-" {
		fmt.Print(result)
		return 0
	}

	if dryRun {
		if !quiet {
			origLines := strings.Split(strings.TrimSuffix(original, "\n"), "\n")
			newLines := strings.Split(strings.TrimSuffix(result, "\n"), "\n")
			fmt.Print(unifiedDiff(path, origLines, newLines))
			fmt.Printf("\n%s[dry-run]%s No file written.\n", clrBold, clrReset)
		}
		return 0
	}

	if backup {
		backupPath := path + ".bak"
		if err := os.WriteFile(backupPath, []byte(original), 0644); err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
			}
			return 2
		}
		if !quiet {
			fmt.Printf("Backup written to %s\n", backupPath)
		}
	}

	if err := os.WriteFile(path, []byte(result), 0644); err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Save failed: %v\n", err)
		}
		return 2
	}

	if !quiet {
		fmt.Printf("Set %s = %s in %s\n", keyPath, valueStr, path)
	}
	return 0
}

// ── runFormat ─────────────────────────────────────────────────────────────────

// runFormat auto-formats the file in place (or to stdout / dry-run).
// typePath is used for type detection (may differ from path when reading stdin).
//
//	Exit 0 — success
//	Exit 1 — parse/format error
//	Exit 2 — I/O error
func runFormat(path, typePath string, dryRun, toStdout, backup, quiet bool) int {
	var content []byte
	var err error
	if path == "-" {
		content, err = io.ReadAll(os.Stdin)
	} else {
		content, err = os.ReadFile(path)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	formatted, err := editor.FormatContent(typePath, string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Format error: %v\n", err)
		return 1
	}

	if toStdout || path == "-" {
		fmt.Print(formatted)
		return 0
	}

	if dryRun {
		if !quiet {
			origLines := strings.Split(strings.TrimSuffix(string(content), "\n"), "\n")
			fmtLines := strings.Split(strings.TrimSuffix(formatted, "\n"), "\n")
			fmt.Print(unifiedDiff(path, origLines, fmtLines))
			fmt.Printf("\n%s[dry-run]%s No file written.\n", clrBold, clrReset)
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
