package editor

import (
	"strings"
	"unicode"
)

// Foreground-only ANSI codes. We intentionally avoid \x1b[0m (full reset)
// inside highlighted lines so that background colors set by the caller
// (e.g. the current-line gray) are preserved across the line.
const (
	hlKey     = "\x1b[96m"    // bright cyan   — keys / identifiers
	hlString  = "\x1b[32m"    // green         — string values
	hlNumber  = "\x1b[33m"    // yellow        — numbers
	hlBool    = "\x1b[35m"    // magenta       — booleans
	hlNull    = "\x1b[2;37m"  // dim white     — null / nil / ~
	hlComment = "\x1b[2;37m"  // dim white     — comments
	hlSection = "\x1b[33;1m"  // bold yellow   — section / doc markers
	hlPunct   = "\x1b[2m"     // dim           — punctuation / dashes
	hlReset   = "\x1b[39;22m" // reset FG + bold only (preserves background)
)

// highlightLine returns line with ANSI foreground color codes injected for
// the given file type. Returns the original line unchanged for unsupported types.
func highlightLine(fileType, line string) string {
	switch fileType {
	case "YAML":
		return highlightYAML(line)
	case "JSON":
		return highlightJSON(line)
	case "TOML":
		return highlightTOML(line)
	case "INI":
		return highlightINI(line)
	default:
		return line
	}
}

// ── Value coloring ────────────────────────────────────────────────────────────

// colorValue wraps a scalar value string with the appropriate ANSI color.
func colorValue(v string) string {
	if v == "" {
		return ""
	}
	t := strings.TrimSpace(v)

	// Quoted strings
	if len(t) >= 2 &&
		((t[0] == '"' && t[len(t)-1] == '"') ||
			(t[0] == '\'' && t[len(t)-1] == '\'')) {
		return hlString + v + hlReset
	}

	// Booleans (YAML accepts several spellings)
	switch strings.ToLower(t) {
	case "true", "false", "yes", "no", "on", "off":
		return hlBool + v + hlReset
	}

	// Null
	switch strings.ToLower(t) {
	case "null", "nil", "~", "none":
		return hlNull + v + hlReset
	}

	// Numbers
	if isNumericLiteral(t) {
		return hlNumber + v + hlReset
	}

	return v
}

func isNumericLiteral(s string) bool {
	if s == "" {
		return false
	}
	i := 0
	if s[i] == '-' || s[i] == '+' {
		i++
	}
	if i == len(s) {
		return false
	}
	// Hex / octal / binary prefixes (YAML / TOML)
	if strings.HasPrefix(s[i:], "0x") || strings.HasPrefix(s[i:], "0o") || strings.HasPrefix(s[i:], "0b") {
		return len(s[i:]) > 2
	}
	dotSeen, eSeen := false, false
	for _, r := range s[i:] {
		switch {
		case unicode.IsDigit(r):
			// ok
		case r == '.' && !dotSeen && !eSeen:
			dotSeen = true
		case (r == 'e' || r == 'E') && !eSeen:
			eSeen = true
		case (r == '+' || r == '-') && eSeen:
			// exponent sign — ok
		case r == '_': // numeric separator in TOML
			// ok
		default:
			return false
		}
	}
	return true
}

// ── YAML ──────────────────────────────────────────────────────────────────────

func highlightYAML(line string) string {
	if line == "" {
		return line
	}
	stripped := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(stripped)]

	// Comment
	if strings.HasPrefix(stripped, "#") {
		return indent + hlComment + stripped + hlReset
	}

	// Document markers
	if stripped == "---" || stripped == "..." {
		return hlSection + line + hlReset
	}

	// List item  "- value"  or bare "-"
	if stripped == "-" {
		return indent + hlPunct + "-" + hlReset
	}
	if strings.HasPrefix(stripped, "- ") {
		rest := stripped[2:]
		// The value might itself be a key:value — recurse for inline maps
		if colonIdx := findYAMLColon(rest); colonIdx >= 0 {
			return indent + hlPunct + "- " + hlReset + highlightYAMLKeyVal(rest)
		}
		return indent + hlPunct + "- " + hlReset + colorValue(rest)
	}

	// Key: value
	if colonIdx := findYAMLColon(stripped); colonIdx >= 0 {
		return indent + highlightYAMLKeyVal(stripped)
	}

	return line
}

// highlightYAMLKeyVal colors a "key: value" fragment (no leading indent).
func highlightYAMLKeyVal(s string) string {
	colonIdx := findYAMLColon(s)
	if colonIdx < 0 {
		return s
	}
	key := s[:colonIdx]
	after := s[colonIdx+1:]
	sep, val := ":", ""
	if len(after) > 0 && after[0] == ' ' {
		sep = ": "
		val = after[1:]
	}
	valPart, comment := splitYAMLComment(val)
	result := hlKey + key + hlReset + sep + colorValue(valPart)
	if comment != "" {
		result += " " + hlComment + comment + hlReset
	}
	return result
}

// findYAMLColon returns the index of the first unquoted colon followed by
// a space or end-of-string, or -1 if not found.
func findYAMLColon(line string) int {
	inSingle, inDouble := false, false
	for i, r := range line {
		switch {
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case r == ':' && !inSingle && !inDouble:
			if i+1 == len(line) || line[i+1] == ' ' || line[i+1] == '\t' {
				return i
			}
		}
	}
	return -1
}

// splitYAMLComment splits a value string into (value, "# comment") parts.
func splitYAMLComment(val string) (string, string) {
	inSingle, inDouble := false, false
	for i, r := range val {
		switch {
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case r == '#' && !inSingle && !inDouble && i > 0 && val[i-1] == ' ':
			return strings.TrimRight(val[:i-1], " "), val[i:]
		}
	}
	return val, ""
}

// ── JSON ──────────────────────────────────────────────────────────────────────

func highlightJSON(line string) string {
	stripped := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(stripped)]

	// Strip trailing comma and whitespace before matching
	bare := stripped
	trailingComma := ""
	if strings.HasSuffix(bare, ",") {
		trailingComma = ","
		bare = strings.TrimRight(bare[:len(bare)-1], " ")
	}

	// Structural-only lines
	switch bare {
	case "{", "}", "[", "]", "{}", "[]":
		return indent + hlPunct + stripped + hlReset
	}

	// "key": value
	if strings.HasPrefix(bare, `"`) {
		closeQ := strings.Index(bare[1:], `"`)
		if closeQ >= 0 {
			closeQ += 2 // position after closing quote
			if closeQ < len(bare) && bare[closeQ] == ':' {
				key := bare[:closeQ]
				rest := strings.TrimLeft(bare[closeQ+1:], " ")
				return indent + hlKey + key + hlReset + ": " + colorJSONValue(rest) + trailingComma
			}
		}
	}

	// Array value or bare value
	if bare != "" {
		return indent + colorJSONValue(bare) + trailingComma
	}

	return line
}

func colorJSONValue(v string) string {
	switch {
	case strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`):
		return hlString + v + hlReset
	case v == "true" || v == "false":
		return hlBool + v + hlReset
	case v == "null":
		return hlNull + v + hlReset
	case v == "{" || v == "}" || v == "[" || v == "]" || v == "{}" || v == "[]":
		return hlPunct + v + hlReset
	default:
		if isNumericLiteral(v) {
			return hlNumber + v + hlReset
		}
	}
	return v
}

// ── TOML ──────────────────────────────────────────────────────────────────────

func highlightTOML(line string) string {
	stripped := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(stripped)]

	// Comment
	if strings.HasPrefix(stripped, "#") {
		return indent + hlComment + stripped + hlReset
	}

	// Section headers: [section] or [[array-of-tables]]
	if strings.HasPrefix(stripped, "[") {
		return indent + hlSection + stripped + hlReset
	}

	// key = value
	eqIdx := findUnquotedChar(stripped, '=')
	if eqIdx > 0 {
		key := strings.TrimRight(stripped[:eqIdx], " ")
		rest := strings.TrimLeft(stripped[eqIdx+1:], " ")
		valPart, comment := splitTOMLComment(rest)
		result := indent + hlKey + key + hlReset + " = " + colorValue(valPart)
		if comment != "" {
			result += " " + hlComment + comment + hlReset
		}
		return result
	}

	return line
}

func splitTOMLComment(val string) (string, string) {
	inSingle, inDouble := false, false
	for i, r := range val {
		switch {
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case r == '#' && !inSingle && !inDouble:
			return strings.TrimRight(val[:i], " "), val[i:]
		}
	}
	return val, ""
}

// findUnquotedChar returns the index of the first occurrence of ch that is
// not inside single or double quotes, or -1.
func findUnquotedChar(s string, ch rune) int {
	inSingle, inDouble := false, false
	for i, r := range s {
		switch {
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case r == ch && !inSingle && !inDouble:
			return i
		}
	}
	return -1
}

// ── INI ───────────────────────────────────────────────────────────────────────

func highlightINI(line string) string {
	stripped := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(stripped)]

	// Comments
	if strings.HasPrefix(stripped, "#") || strings.HasPrefix(stripped, ";") {
		return indent + hlComment + stripped + hlReset
	}

	// Section header [name]
	if strings.HasPrefix(stripped, "[") && strings.HasSuffix(stripped, "]") {
		return indent + hlSection + stripped + hlReset
	}

	// key = value  or  key: value
	for _, sep := range []string{"=", ":"} {
		idx := strings.Index(stripped, sep)
		if idx > 0 {
			key := strings.TrimRight(stripped[:idx], " ")
			val := strings.TrimLeft(stripped[idx+1:], " ")
			return indent + hlKey + key + hlReset + " " + sep + " " + val
		}
	}

	return line
}
