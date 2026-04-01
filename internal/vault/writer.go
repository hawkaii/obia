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
// target: "daily" → today's daily note, "default"/"" → defaultFile, anything else → vault-relative path.
func ResolveTaskFile(vaultPath, dailyFolder, dailyFormat, defaultFile, target string) string {
	defaultPath := filepath.Join(vaultPath, defaultFile)

	switch target {
	case "default", "":
		return defaultPath
	case "daily":
		// handled below
	default:
		return filepath.Join(vaultPath, target)
	}

	// Check if the daily notes folder exists
	dailyDir := filepath.Join(vaultPath, dailyFolder)
	if info, err := os.Stat(dailyDir); err != nil || !info.IsDir() {
		return defaultPath
	}

	// Build today's daily note path
	today := time.Now().Format(dailyFormat)
	dailyPath := filepath.Join(dailyDir, today+".md")

	// If it already exists, use it
	if _, err := os.Stat(dailyPath); err == nil {
		return dailyPath
	}

	// Try to create from template
	templatePath := filepath.Join(vaultPath, "templates", "diary template.md")
	templateData, err := os.ReadFile(templatePath)
	if err == nil {
		// Replace template variables
		content := strings.ReplaceAll(string(templateData), "{{date}}", today)
		content = strings.ReplaceAll(content, "{{time}}", time.Now().Format("15:04"))
		if err := os.WriteFile(dailyPath, []byte(content), 0o644); err != nil {
			return defaultPath
		}
		return dailyPath
	}

	// No template — create a bare file
	bare := fmt.Sprintf("# %s\n\n", today)
	if err := os.WriteFile(dailyPath, []byte(bare), 0o644); err != nil {
		return defaultPath
	}
	return dailyPath
}

// WriteFrontmatterUID writes caldav-uid to the YAML frontmatter of filePath,
// but only if the file has frontmatter containing "type: task".
// Multi-task files (daily notes, todo.md) are skipped — returns nil.
func WriteFrontmatterUID(filePath, uid string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Must start with ---
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil
	}

	// Find the closing --- and check for "type: task"
	closingIdx := -1
	isTaskFile := false
	for i := 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "---" {
			closingIdx = i
			break
		}
		key, value, ok := parseYAMLLine(trimmed)
		if ok && key == "type" && value == "task" {
			isTaskFile = true
		}
	}

	// No closing --- found or not a single-task file
	if closingIdx < 0 || !isTaskFile {
		return nil
	}

	// Check if caldav-uid already exists in the frontmatter
	for i := 1; i < closingIdx; i++ {
		key, _, ok := parseYAMLLine(strings.TrimSpace(lines[i]))
		if ok && key == "caldav-uid" {
			// Replace existing value
			lines[i] = "caldav-uid: " + uid
			return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0o644)
		}
	}

	// Insert before closing ---
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:closingIdx]...)
	newLines = append(newLines, "caldav-uid: "+uid)
	newLines = append(newLines, lines[closingIdx:]...)
	return os.WriteFile(filePath, []byte(strings.Join(newLines, "\n")), 0o644)
}

// AppendTask adds a new task line to the given file.
func AppendTask(filePath string, description string) error {
	_, err := AppendTaskAt(filePath, description)
	return err
}

// AppendTaskAt adds a new task line to the given file and returns the 1-indexed line number
// where the task was written.
func AppendTaskAt(filePath string, description string) (int, error) {
	// Read existing content to count lines.
	lineCount := 0
	existing, err := os.ReadFile(filePath)
	if err == nil {
		lineCount = strings.Count(string(existing), "\n")
	}
	// lineCount is now the number of existing lines (0 if file doesn't exist).

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	taskLine := fmt.Sprintf("\n- [ ] %s\n", description)
	if _, err = f.WriteString(taskLine); err != nil {
		return 0, err
	}

	// The blank line adds 1, then the task is on the next line.
	return lineCount + 2, nil
}
