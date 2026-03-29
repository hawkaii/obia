package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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

// ResolveTaskFile determines where to add a new task based on the target setting.
// target: "daily" tries today's daily note, "default" uses the default task file.
func ResolveTaskFile(vaultPath, dailyFolder, dailyFormat, defaultFile, target string) (string, error) {
	defaultPath := filepath.Join(vaultPath, defaultFile)

	if target != "daily" {
		return defaultPath, nil
	}

	// Check if the daily notes folder exists
	dailyDir := filepath.Join(vaultPath, dailyFolder)
	if info, err := os.Stat(dailyDir); err != nil || !info.IsDir() {
		return defaultPath, nil
	}

	// Build today's daily note path
	today := time.Now().Format(dailyFormat)
	dailyPath := filepath.Join(dailyDir, today+".md")

	// If it already exists, use it
	if _, err := os.Stat(dailyPath); err == nil {
		return dailyPath, nil
	}

	// Try to create from template
	templatePath := filepath.Join(vaultPath, "templates", "diary template.md")
	templateData, err := os.ReadFile(templatePath)
	if err == nil {
		// Replace template variables
		content := strings.ReplaceAll(string(templateData), "{{date}}", today)
		content = strings.ReplaceAll(content, "{{time}}", time.Now().Format("15:04"))
		if err := os.WriteFile(dailyPath, []byte(content), 0o644); err != nil {
			return defaultPath, nil
		}
		return dailyPath, nil
	}

	// No template — create a bare file
	bare := fmt.Sprintf("# %s\n\n", today)
	if err := os.WriteFile(dailyPath, []byte(bare), 0o644); err != nil {
		return defaultPath, nil
	}
	return dailyPath, nil
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
