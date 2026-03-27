package file

import (
	"bufio"
	"os"
	"strings"
)

func OpenFile(path string) (string, error) {
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

func SaveFile(path string, buf [][]rune) error {
	var b strings.Builder

	for i, line := range buf {
		b.WriteString(string(line))
		if i < len(buf)-1 {
			b.WriteByte('\n')
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}