package vault

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hawkaii/obia/internal/caldav"
	"github.com/hawkaii/obia/internal/task"
)

var (
	taskPattern     = regexp.MustCompile(`^\s*[-*+] \[([ xX])\] (.+)$`)
	wikiLinkPattern = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	tagPattern      = regexp.MustCompile(`(?:^|\s)#(\S+)`)
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

		// Attach frontmatter metadata if present
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
func ParseAllTasks(vaultPath string) ([]task.Task, error) {
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

	// Hydrate CalDAVUID from sync.json for tasks that have been pushed.
	if uidMap, err := caldav.LoadUIDMap(); err == nil {
		for i := range allTasks {
			key := allTasks[i].Source.FilePath + ":" + strconv.Itoa(allTasks[i].Source.Line)
			if uid, ok := uidMap[key]; ok {
				allTasks[i].CalDAVUID = uid
			}
		}
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
		}
	}

	return data
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
