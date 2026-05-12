package editor

import (
	"bytes"
	"encoding/json"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// FormatBuffer re-formats a buffer for the file type detected from filename.
// Returns the new buffer (caller should push undo first) or an error.
func FormatBuffer(filename string, buf [][]rune) ([][]rune, error) {
	content := bufToString(buf)
	formatted, err := FormatContent(filename, content)
	if err != nil {
		return nil, err
	}
	return toBuffer(formatted), nil
}

// FormatContent pretty-prints the given content for the detected file type.
func FormatContent(filename, content string) (string, error) {
	switch detectFileType(filename) {
	case "JSON":
		return formatJSON(content)
	case "YAML":
		return formatYAML(content)
	case "TOML":
		return formatTOML(content)
	default:
		return content, nil
	}
}

func formatJSON(content string) (string, error) {
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return "", err
	}
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out) + "\n", nil
}

func formatYAML(content string) (string, error) {
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(content), &node); err != nil {
		return "", err
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&node); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func formatTOML(content string) (string, error) {
	var v interface{}
	if _, err := toml.Decode(content, &v); err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(v); err != nil {
		return "", err
	}
	return buf.String(), nil
}
