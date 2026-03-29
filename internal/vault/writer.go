package vault

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hawkaii/obia/internal/task"
)

var checkboxPattern = regexp.MustCompile(`^(\s*[-*+] \[)([ xX])(\] .+)$`)

// ToggleTask reads the source file, flips the checkbox on the task's line, and writes it back.
func ToggleTask(t *task.Task) error {
	data, err := os.ReadFile(t.Source.FilePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", t.Source.FilePath, err)
	}

	lines := strings.Split(string(data), "\n")
	lineIdx := t.Source.Line - 1
	if lineIdx < 0 || lineIdx >= len(lines) {
		return fmt.Errorf("line %d out of range in %s", t.Source.Line, t.Source.FilePath)
	}

	matches := checkboxPattern.FindStringSubmatch(lines[lineIdx])
	if matches == nil {
		return fmt.Errorf("line %d is not a task checkbox", t.Source.Line)
	}

	// Flip the marker
	newMarker := "x"
	if matches[2] == "x" || matches[2] == "X" {
		newMarker = " "
	}

	lines[lineIdx] = matches[1] + newMarker + matches[3]

	return os.WriteFile(t.Source.FilePath, []byte(strings.Join(lines, "\n")), 0o644)
}

// AppendTask adds a new task line to the given file.
func AppendTask(filePath string, description string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("\n- [ ] %s\n", description)
	_, err = f.WriteString(line)
	return err
}
