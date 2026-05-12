package file

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const MaxFileBytes = 50 * 1024 * 1024 // 50 MB

func OpenFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() > MaxFileBytes {
		return "", fmt.Errorf("file too large (%d MB); maximum is 50 MB", info.Size()/(1024*1024))
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer file.Close()

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text() + "\n")
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return content.String(), nil

}

// SaveBytes atomically writes data to path, preserving original file permissions.
func SaveBytes(path string, data []byte) error {
	perm := os.FileMode(0644)
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode()
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".tengo-save-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		os.Remove(tmpName)
		return err
	}

	return os.Rename(tmpName, path)
}

// SaveFile converts a rune buffer to bytes and atomically writes it to path.
func SaveFile(path string, buf [][]rune) error {
	var b strings.Builder
	for i, line := range buf {
		b.WriteString(string(line))
		if i < len(buf)-1 {
			b.WriteByte('\n')
		}
	}
	return SaveBytes(path, []byte(b.String()))
}