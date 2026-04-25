package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Vault struct {
	Path             string   `toml:"path"`
	DailyNotesFolder string   `toml:"daily_notes_folder"` // kept for backward compat
	Folders          []string `toml:"folders"`             // generic folder list; takes precedence over daily_notes_folder
	DailyNotesFormat string   `toml:"daily_notes_format"`
	DefaultTaskFile  string   `toml:"default_task_file"`
	AddTaskTarget    string   `toml:"add_task_target"`   // "daily" | "default" | vault-relative path
	ExtraTargets     []string `toml:"extra_targets"`     // additional vault-relative paths
	TaskFilesFolder  string   `toml:"task_files_folder"` // folder for task files (default "tasks")
	InboxFile        string   `toml:"inbox_file"`        // landing zone for remote-only pulled tasks
}

type CalDAV struct {
	URL      string `toml:"url"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	AutoPush bool   `toml:"auto_push"`
}

type UI struct {
	DefaultTab string `toml:"default_tab"`
	Grouped    bool   `toml:"grouped"`
}

type Config struct {
	Vault  Vault  `toml:"vault"`
	CalDAV CalDAV `toml:"caldav"`
	UI     UI     `toml:"ui"`
}

func DefaultConfig() Config {
	return Config{
		Vault: Vault{
			DailyNotesFolder: "diary",
			DailyNotesFormat: "2006-01-02",
			DefaultTaskFile:  "todo.md",
			AddTaskTarget:    "daily",
			TaskFilesFolder:  "tasks",
			InboxFile:        "tasks/inbox.md",
		},
		UI: UI{
			DefaultTab: "tasks",
		},
	}
}

func configDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "obia"), nil
}

func ConfigPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

func CachePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "cache.json"), nil
}

func Load() (Config, error) {
	cfg := DefaultConfig()

	path, err := ConfigPath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

func Save(cfg Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	path := filepath.Join(dir, "config.toml")
	return os.WriteFile(path, data, 0o644)
}
