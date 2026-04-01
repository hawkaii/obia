package vault

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hawkaii/obia/internal/caldav"
	"github.com/hawkaii/obia/internal/task"
)

var (
	taskPattern       = regexp.MustCompile(`^\s*[-*+] \[([ xX])\] (.+)$`)
	wikiLinkPattern   = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	tagPattern        = regexp.MustCompile(`(?:^|\s)#(\S+)`)
	linkedTaskPattern = regexp.MustCompile(`\[\[([^| \]]+)\|([^\]]+)\]\]`)
)

// ParseTasks reads a markdown file and extracts all checkbox tasks.
func ParseTasks(filePath string) ([]task.Task, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	fileMod := info.ModTime()

	frontmatter := parseFrontmatter(f)
	f.Seek(0, 0)

	var tasks []task.Task
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		matches := taskPattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		marker := matches[1]
		description := strings.TrimSpace(matches[2])

		status := task.Todo
		if marker == "x" || marker == "X" {
			status = task.Done
		}

		wikiLinks := extractWikiLinks(description)
		tags := extractTags(description)

		t := task.Task{
			Description: description,
			Status:      status,
			Tags:        tags,
			WikiLinks:   wikiLinks,
			Source: task.Source{
				FilePath: filePath,
				Line:     lineNum,
				FileMod:  fileMod,
			},
		}

		// Attach frontmatter metadata if present (single-task files)
		if frontmatter.calDAVUID != "" {
			t.CalDAVUID = frontmatter.calDAVUID
		}
		if frontmatter.due != nil {
			t.Due = frontmatter.due
		}

		tasks = append(tasks, t)
	}

	return tasks, scanner.Err()
}

// ParseAllTasks scans the vault and parses tasks from every markdown file.
// taskFilesFolder is the vault-relative folder where task files live (e.g. "tasks").
func ParseAllTasks(vaultPath, taskFilesFolder string) ([]task.Task, error) {
	files, err := ScanMarkdownFiles(vaultPath)
	if err != nil {
		return nil, err
	}

	var allTasks []task.Task
	for _, f := range files {
		tasks, err := ParseTasks(f)
		if err != nil {
			continue // skip unreadable files
		}
		allTasks = append(allTasks, tasks...)
	}

	// Sort by file modification time, newest first.
	// Within the same file, preserve line order.
	sort.SliceStable(allTasks, func(i, j int) bool {
		if allTasks[i].Source.FileMod.Equal(allTasks[j].Source.FileMod) {
			return allTasks[i].Source.Line < allTasks[j].Source.Line
		}
		return allTasks[i].Source.FileMod.After(allTasks[j].Source.FileMod)
	})

	// Hydrate CalDAVUID from sync.json for plain pushed tasks.
	if uidMap, err := caldav.LoadUIDMap(); err == nil {
		for i := range allTasks {
			key := allTasks[i].Source.FilePath + ":" + strconv.Itoa(allTasks[i].Source.Line)
			if uid, ok := uidMap[key]; ok {
				allTasks[i].CalDAVUID = uid
			}
		}
	}

	// Hydrate linked tasks: resolve [[uid|alias]] wikilinks to task files.
	taskFilesDir := filepath.Join(vaultPath, taskFilesFolder)
	for i := range allTasks {
		m := linkedTaskPattern.FindStringSubmatch(strings.TrimSpace(allTasks[i].Description))
		if m == nil {
			continue
		}
		uid := m[1]
		alias := m[2]
		taskFile := filepath.Join(taskFilesDir, uid+".md")

		f, err := os.Open(taskFile)
		if err != nil {
			// task file missing — keep raw description
			continue
		}
		fm := parseFrontmatter(f)
		f.Close()

		allTasks[i].Description = alias
		allTasks[i].LinkedTaskFile = taskFile
		if fm.calDAVUID != "" {
			allTasks[i].CalDAVUID = fm.calDAVUID
		} else {
			allTasks[i].CalDAVUID = uid
		}
		if fm.due != nil {
			allTasks[i].Due = fm.due
		}
		allTasks[i].Priority = fm.priority
		allTasks[i].CalDAVStatus = fm.status
		allTasks[i].Body = readTaskFileBody(taskFile)
	}

	return allTasks, nil
}

func extractWikiLinks(s string) []string {
	matches := wikiLinkPattern.FindAllStringSubmatch(s, -1)
	links := make([]string, 0, len(matches))
	for _, m := range matches {
		links = append(links, m[1])
	}
	return links
}

func extractTags(s string) []string {
	matches := tagPattern.FindAllStringSubmatch(s, -1)
	tags := make([]string, 0, len(matches))
	for _, m := range matches {
		tags = append(tags, m[1])
	}
	return tags
}

type frontmatterData struct {
	calDAVUID  string
	due        *time.Time
	isTaskFile bool // true if frontmatter contains "type: task"
	priority   int
	status     string
}

func parseFrontmatter(f *os.File) frontmatterData {
	var data frontmatterData
	scanner := bufio.NewScanner(f)

	// First line must be ---
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return data
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "---" {
			break
		}

		key, value, ok := parseYAMLLine(line)
		if !ok {
			continue
		}

		switch key {
		case "caldav-uid":
			data.calDAVUID = value
		case "due":
			if t, err := time.Parse("2006-01-02", value); err == nil {
				data.due = &t
			} else if t, err := time.Parse(time.RFC3339, value); err == nil {
				data.due = &t
			}
		case "type":
			if value == "task" {
				data.isTaskFile = true
			}
		case "priority":
			if p, err := strconv.Atoi(value); err == nil {
				data.priority = p
			}
		case "status":
			data.status = value
		}
	}

	return data
}

// readTaskFileBody reads the body text from a task file (content after the # Title heading).
func readTaskFileBody(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	inFrontmatter := false
	pastFrontmatter := false
	pastTitle := false
	var body []string
	for _, line := range lines {
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
		if !pastTitle {
			if strings.HasPrefix(line, "# ") {
				pastTitle = true
			}
			continue
		}
		body = append(body, line)
	}
	return strings.TrimSpace(strings.Join(body, "\n"))
}

func parseYAMLLine(line string) (key, value string, ok bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	value = strings.TrimSpace(line[idx+1:])
	value = strings.Trim(value, `"'`)
	return key, value, true
}
