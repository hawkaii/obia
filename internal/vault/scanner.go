package vault

import (
	"os"
	"path/filepath"
	"strings"
)

var skipDirs = map[string]bool{
	".obsidian": true,
	".trash":    true,
	".opencode": true,
	"templates": true,
	"extras":    true,
}

// ScanMarkdownFiles returns all .md file paths under root, skipping excluded directories.
func ScanMarkdownFiles(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") && name != "." {
				if skipDirs[name] {
					return filepath.SkipDir
				}
			}
			if skipDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(d.Name(), ".md") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
