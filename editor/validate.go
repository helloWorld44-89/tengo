package editor

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// ValidationResult holds the outcome of a syntax check.
type ValidationResult struct {
	Valid   bool
	Line   int    // 1-based line number of the error; 0 = unknown
	Message string
}

var lineNumRe = regexp.MustCompile(`\bline (\d+)\b`)

const yamlMaxNodes = 100_000

// countYAMLNodes walks a yaml.Node tree simulating alias expansion and counts
// effective nodes. Aliases are followed each time they appear, so exponentially
// expanded structures (billion-laughs) are caught before the limit is reached.
func countYAMLNodes(n *yaml.Node, count *int) error {
	if n == nil {
		return nil
	}
	if n.Kind == yaml.AliasNode {
		return countYAMLNodes(n.Alias, count)
	}
	*count++
	if *count > yamlMaxNodes {
		return fmt.Errorf("alias expansion exceeds %d nodes; file rejected to prevent memory exhaustion", yamlMaxNodes)
	}
	for _, child := range n.Content {
		if err := countYAMLNodes(child, count); err != nil {
			return err
		}
	}
	return nil
}

// CheckYAMLSafe parses content as YAML and returns an error if alias expansion
// would produce more than yamlMaxNodes effective nodes.
func CheckYAMLSafe(content string) error {
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(content), &node); err != nil {
		return err
	}
	count := 0
	return countYAMLNodes(&node, &count)
}

// ValidateBuffer validates the in-memory buffer for the given filename's type.
func ValidateBuffer(filename string, buf [][]rune) ValidationResult {
	return ValidateContent(filename, bufToString(buf))
}

// ValidateContent validates a string for the file type inferred from filename.
func ValidateContent(filename, content string) ValidationResult {
	switch detectFileType(filename) {
	case "JSON":
		return validateJSON(content)
	case "YAML":
		return validateYAML(content)
	case "TOML":
		return validateTOML(content)
	case "XML":
		return validateXML(content)
	default:
		return ValidationResult{Valid: true, Message: "no validator for this file type"}
	}
}

func validateJSON(content string) ValidationResult {
	var v interface{}
	if err := json.Unmarshal([]byte(content), &v); err != nil {
		line := 0
		if se, ok := err.(*json.SyntaxError); ok {
			line = offsetToLine(content, int(se.Offset))
		} else if ute, ok := err.(*json.UnmarshalTypeError); ok {
			line = offsetToLine(content, int(ute.Offset))
		}
		return ValidationResult{Valid: false, Line: line, Message: fmt.Sprintf("JSON: %v", err)}
	}
	return ValidationResult{Valid: true, Message: "Valid JSON"}
}

func validateYAML(content string) ValidationResult {
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(content), &node); err != nil {
		line := 0
		if m := lineNumRe.FindStringSubmatch(err.Error()); len(m) > 1 {
			line, _ = strconv.Atoi(m[1])
		}
		return ValidationResult{Valid: false, Line: line, Message: fmt.Sprintf("YAML: %v", err)}
	}
	count := 0
	if err := countYAMLNodes(&node, &count); err != nil {
		return ValidationResult{Valid: false, Message: fmt.Sprintf("YAML: %v", err)}
	}
	return ValidationResult{Valid: true, Message: "Valid YAML"}
}

func validateTOML(content string) ValidationResult {
	var v interface{}
	if _, err := toml.Decode(content, &v); err != nil {
		line := 0
		if pe, ok := err.(toml.ParseError); ok {
			line = pe.Position.Line
		}
		return ValidationResult{Valid: false, Line: line, Message: fmt.Sprintf("TOML: %v", err)}
	}
	return ValidationResult{Valid: true, Message: "Valid TOML"}
}

func validateXML(content string) ValidationResult {
	dec := xml.NewDecoder(strings.NewReader(content))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			line := 0
			if se, ok := err.(*xml.SyntaxError); ok {
				line = int(se.Line)
			}
			return ValidationResult{Valid: false, Line: line, Message: fmt.Sprintf("XML: %v", err)}
		}
	}
	return ValidationResult{Valid: true, Message: "Valid XML"}
}

// offsetToLine converts a byte offset in content to a 1-based line number.
func offsetToLine(content string, offset int) int {
	if offset > len(content) {
		offset = len(content)
	}
	return strings.Count(content[:offset], "\n") + 1
}
