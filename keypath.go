package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// fileTypeFromPath returns "json", "yaml", "toml", or "" based on extension.
func fileTypeFromPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	default:
		return ""
	}
}

// parseKeyPath splits a dot-notation path into segments.
// e.g. "database.host" → ["database", "host"]
// e.g. "servers.0.port" → ["servers", "0", "port"]
func parseKeyPath(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, ".")
}

// getAtPath navigates a nested structure (maps/slices) by path segments.
func getAtPath(data interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return data, nil
	}
	key := path[0]
	rest := path[1:]

	switch v := data.(type) {
	case map[string]interface{}:
		val, ok := v[key]
		if !ok {
			return nil, fmt.Errorf("key %q not found", key)
		}
		return getAtPath(val, rest)
	case []interface{}:
		idx, err := strconv.Atoi(key)
		if err != nil {
			return nil, fmt.Errorf("expected array index, got %q", key)
		}
		if idx < 0 || idx >= len(v) {
			return nil, fmt.Errorf("index %d out of range (len %d)", idx, len(v))
		}
		return getAtPath(v[idx], rest)
	default:
		return nil, fmt.Errorf("cannot index %T with key %q", data, key)
	}
}

// setAtPath sets a value at path within a nested structure.
// Map mutations are in-place; the (possibly new) root is returned.
func setAtPath(data interface{}, path []string, value interface{}) (interface{}, error) {
	if len(path) == 0 {
		return value, nil
	}
	key := path[0]
	rest := path[1:]

	switch v := data.(type) {
	case map[string]interface{}:
		child := v[key]
		if child == nil && len(rest) > 0 {
			child = map[string]interface{}{}
		}
		newChild, err := setAtPath(child, rest, value)
		if err != nil {
			return nil, err
		}
		v[key] = newChild
		return v, nil
	case []interface{}:
		idx, err := strconv.Atoi(key)
		if err != nil {
			return nil, fmt.Errorf("expected array index, got %q", key)
		}
		if idx < 0 || idx >= len(v) {
			return nil, fmt.Errorf("index %d out of range (len %d)", idx, len(v))
		}
		newChild, err := setAtPath(v[idx], rest, value)
		if err != nil {
			return nil, err
		}
		v[idx] = newChild
		return v, nil
	default:
		if len(rest) == 0 {
			return value, nil
		}
		return nil, fmt.Errorf("cannot descend into %T with key %q", data, key)
	}
}

// parseScalarValue converts a string to the most appropriate Go type.
func parseScalarValue(s string) interface{} {
	switch strings.ToLower(s) {
	case "true":
		return true
	case "false":
		return false
	case "null", "~":
		return nil
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

// printValue formats a value for -key output.
func printValue(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case map[string]interface{}, []interface{}:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// readFileContent reads a file (or stdin when path=="-") as a string.
func readFileContent(path string) (string, error) {
	var b []byte
	var err error
	if path == "-" {
		b, err = io.ReadAll(os.Stdin)
	} else {
		b, err = os.ReadFile(path)
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// unmarshalGeneric decodes content into a generic map/slice tree.
func unmarshalGeneric(typePath, content string) (interface{}, error) {
	ft := fileTypeFromPath(typePath)
	switch ft {
	case "json":
		var data interface{}
		if err := json.Unmarshal([]byte(content), &data); err != nil {
			return nil, err
		}
		return data, nil
	case "yaml":
		var data interface{}
		if err := yaml.Unmarshal([]byte(content), &data); err != nil {
			return nil, err
		}
		return data, nil
	case "toml":
		var data map[string]interface{}
		if _, err := toml.Decode(content, &data); err != nil {
			return nil, err
		}
		return data, nil
	default:
		ext := filepath.Ext(typePath)
		return nil, fmt.Errorf("unsupported file type %q for key navigation (use .json, .yaml, or .toml)", ext)
	}
}

// marshalGeneric encodes data back to the file format inferred from typePath.
// Note: marshaling does not preserve original formatting or comments.
func marshalGeneric(typePath string, data interface{}) (string, error) {
	ft := fileTypeFromPath(typePath)
	switch ft {
	case "json":
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", err
		}
		return string(b) + "\n", nil
	case "yaml":
		b, err := yaml.Marshal(data)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case "toml":
		var buf strings.Builder
		enc := toml.NewEncoder(&buf)
		if err := enc.Encode(data); err != nil {
			return "", err
		}
		return buf.String(), nil
	default:
		ext := filepath.Ext(typePath)
		return "", fmt.Errorf("unsupported file type %q for marshaling (use .json, .yaml, or .toml)", ext)
	}
}
