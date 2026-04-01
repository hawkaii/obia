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

// EnsureTaskFolder creates the task files folder under vaultPath.
// If the name already exists as a file (not a dir), tries name+"_1", etc.
// Returns the resolved folder path and a notification string (empty if no conflict).
func EnsureTaskFolder(vaultPath, folder string) (string, string) {
	base := filepath.Join(vaultPath, folder)
	candidate := base
	suffix := 1
	for {
		info, err := os.Stat(candidate)
		if os.IsNotExist(err) {
			_ = os.MkdirAll(candidate, 0o755)
			if candidate != base {
				return candidate, fmt.Sprintf("'%s' exists as file, using '%s' instead", folder, filepath.Base(candidate))
			}
			return candidate, ""
		}
		if info.IsDir() {
			return candidate, ""
		}
		candidate = fmt.Sprintf("%s_%d", base, suffix)
		suffix++
	}
}

// CreateTaskFile writes a task file at folderPath/<uid>.md with YAML frontmatter,
// a title header, and an optional description body.
func CreateTaskFile(folderPath, uid, title, description string, due *time.Time, priority int, status string) error {
	if status == "" {
		status = "NEEDS-ACTION"
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("type: task\n")
	fmt.Fprintf(&b, "caldav-uid: %s\n", uid)
	if due != nil {
		b.WriteString("due: " + due.Format(time.RFC3339) + "\n")
	}
	if priority > 0 {
		fmt.Fprintf(&b, "priority: %d\n", priority)
	}
	fmt.Fprintf(&b, "status: %s\n", status)
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "# %s\n", title)
	if description != "" {
		b.WriteString("\n")
		b.WriteString(description)
		b.WriteString("\n")
	}

	filePath := filepath.Join(folderPath, uid+".md")
	return os.WriteFile(filePath, []byte(b.String()), 0o644)
}

// RewriteTaskLine rewrites the task at lineNum in filePath so the description
// becomes a wikilink alias: [[uid|alias]], preserving the checkbox state.
func RewriteTaskLine(filePath string, lineNum int, uid, alias string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	lines := strings.Split(string(data), "\n")
	idx := lineNum - 1
	if idx < 0 || idx >= len(lines) {
		return fmt.Errorf("line %d out of range in %s", lineNum, filePath)
	}

	matches := checkboxPattern.FindStringSubmatch(lines[idx])
	if matches == nil {
		return fmt.Errorf("line %d is not a task checkbox", lineNum)
	}

	lines[idx] = matches[1] + matches[2] + "] [[" + uid + "|" + alias + "]]"
	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0o644)
}

// UpdateTaskFileStatus updates the status: field in a task file's YAML frontmatter.
func UpdateTaskFileStatus(filePath, status string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil
	}

	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			break
		}
		key, _, ok := parseYAMLLine(strings.TrimSpace(lines[i]))
		if ok && key == "status" {
			lines[i] = "status: " + status
			return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0o644)
		}
	}
	return nil
}

// UpdateTaskFileFrontmatter overwrites due/status/priority fields in a task file's frontmatter.
func UpdateTaskFileFrontmatter(filePath string, due *time.Time, status string, priority int) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil
	}

	closingIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closingIdx = i
			break
		}
	}
	if closingIdx < 0 {
		return nil
	}

	updatedDue, updatedStatus, updatedPriority := false, false, false

	// First pass: update or remove existing fields.
	// Build filtered lines to handle removals (due=nil clears the field, priority=0 clears it).
	filtered := make([]string, 0, len(lines))
	filtered = append(filtered, lines[0]) // opening ---
	for i := 1; i < closingIdx; i++ {
		key, _, ok := parseYAMLLine(strings.TrimSpace(lines[i]))
		if !ok {
			filtered = append(filtered, lines[i])
			continue
		}
		switch key {
		case "due":
			updatedDue = true
			if due != nil {
				filtered = append(filtered, "due: "+due.Format(time.RFC3339))
			}
			// due==nil: omit the line (clear due date)
		case "status":
			updatedStatus = true
			if status != "" {
				filtered = append(filtered, "status: "+status)
			} else {
				filtered = append(filtered, lines[i]) // keep existing
			}
		case "priority":
			updatedPriority = true
			if priority > 0 {
				filtered = append(filtered, fmt.Sprintf("priority: %d", priority))
			}
			// priority==0: omit the line (clear priority)
		default:
			filtered = append(filtered, lines[i])
		}
	}

	// Insert missing fields before closing ---
	if !updatedDue && due != nil {
		filtered = append(filtered, "due: "+due.Format(time.RFC3339))
	}
	if !updatedStatus && status != "" {
		filtered = append(filtered, "status: "+status)
	}
	if !updatedPriority && priority > 0 {
		filtered = append(filtered, fmt.Sprintf("priority: %d", priority))
	}

	// Re-append closing --- and everything after
	filtered = append(filtered, lines[closingIdx:]...)

	return os.WriteFile(filePath, []byte(strings.Join(filtered, "\n")), 0o644)
}

// UpdateTaskFileContent updates title, body, and frontmatter fields (due/status/priority)
// in a single read-modify-write pass.
func UpdateTaskFileContent(filePath, title, body string, due *time.Time, status string, priority int) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil
	}

	// Locate closing frontmatter ---
	closingIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closingIdx = i
			break
		}
	}
	if closingIdx < 0 {
		return nil
	}

	// Rewrite frontmatter fields
	updatedDue, updatedStatus, updatedPriority := false, false, false
	filtered := make([]string, 0, len(lines))
	filtered = append(filtered, lines[0])
	for i := 1; i < closingIdx; i++ {
		key, _, ok := parseYAMLLine(strings.TrimSpace(lines[i]))
		if !ok {
			filtered = append(filtered, lines[i])
			continue
		}
		switch key {
		case "due":
			updatedDue = true
			if due != nil {
				filtered = append(filtered, "due: "+due.Format(time.RFC3339))
			}
		case "status":
			updatedStatus = true
			if status != "" {
				filtered = append(filtered, "status: "+status)
			} else {
				filtered = append(filtered, lines[i])
			}
		case "priority":
			updatedPriority = true
			if priority > 0 {
				filtered = append(filtered, fmt.Sprintf("priority: %d", priority))
			}
		default:
			filtered = append(filtered, lines[i])
		}
	}
	if !updatedDue && due != nil {
		filtered = append(filtered, "due: "+due.Format(time.RFC3339))
	}
	if !updatedStatus && status != "" {
		filtered = append(filtered, "status: "+status)
	}
	if !updatedPriority && priority > 0 {
		filtered = append(filtered, fmt.Sprintf("priority: %d", priority))
	}
	filtered = append(filtered, lines[closingIdx]) // closing ---

	// Find title line and rewrite title + body
	titleIdx := -1
	for i := closingIdx + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "# ") {
			titleIdx = i
			break
		}
	}

	if titleIdx >= 0 {
		// Keep lines between closing --- and title (blank lines)
		filtered = append(filtered, lines[closingIdx+1:titleIdx]...)
		filtered = append(filtered, "# "+title)
		filtered = append(filtered, "")
		if body != "" {
			filtered = append(filtered, body)
		}
	} else {
		// No title found — append remaining lines unchanged
		filtered = append(filtered, lines[closingIdx+1:]...)
	}

	return os.WriteFile(filePath, []byte(strings.Join(filtered, "\n")), 0o644)
}

// AppendToFile appends a line to a file, creating it if needed.
func AppendToFile(filePath, line string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, line)
	return err
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

// UpdatePlainTaskDescription rewrites the description text of a plain task checkbox line.
func UpdatePlainTaskDescription(filePath string, lineNum int, newDesc string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	lines := strings.Split(string(data), "\n")
	idx := lineNum - 1
	if idx < 0 || idx >= len(lines) {
		return fmt.Errorf("line %d out of range in %s", lineNum, filePath)
	}

	matches := checkboxPattern.FindStringSubmatch(lines[idx])
	if matches == nil {
		return fmt.Errorf("line %d is not a task checkbox", lineNum)
	}

	lines[idx] = matches[1] + matches[2] + "] " + newDesc
	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0o644)
}

// UpdateTaskFileTitle replaces the # Title heading in a task file.
func UpdateTaskFileTitle(filePath, newTitle string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	lines := strings.Split(string(data), "\n")
	inFrontmatter := false
	pastFrontmatter := false
	for i, line := range lines {
		if !pastFrontmatter {
			if strings.TrimSpace(line) == "---" {
				if !inFrontmatter {
					inFrontmatter = true
				} else {
					pastFrontmatter = true
				}
			}
			continue
		}
		if strings.HasPrefix(line, "# ") {
			lines[i] = "# " + newTitle
			return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0o644)
		}
	}
	return nil
}

// UpdateTaskFileBody replaces the body content (below the # Title heading) in a task file.
func UpdateTaskFileBody(filePath, body string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	lines := strings.Split(string(data), "\n")
	inFrontmatter := false
	pastFrontmatter := false
	titleIdx := -1
	for i, line := range lines {
		if !pastFrontmatter {
			if strings.TrimSpace(line) == "---" {
				if !inFrontmatter {
					inFrontmatter = true
				} else {
					pastFrontmatter = true
				}
			}
			continue
		}
		if strings.HasPrefix(line, "# ") {
			titleIdx = i
			break
		}
	}

	if titleIdx < 0 {
		return nil
	}

	// Keep everything up to and including the title line
	newLines := append(lines[:titleIdx+1], "")
	if body != "" {
		newLines = append(newLines, body)
	}

	return os.WriteFile(filePath, []byte(strings.Join(newLines, "\n")), 0o644)
}
