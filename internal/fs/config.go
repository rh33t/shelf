package fs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

const appName = "shelf"

// Config holds all user-configurable values.
// To add a new field: add it here + set its default in defaults().
type Config struct {
	BaseDir       string   `yaml:"base_dir"`
	Cmd           string   `yaml:"cmd"`
	CTFSources    []string `yaml:"ctf_sources"`
	CTFCategories []string `yaml:"ctf_categories"`
	BoxPlatforms  []string `yaml:"box_platforms"`
}

// defaults returns the baseline config.
// Every field must have a value here.
func defaults() Config {
	return Config{
		BaseDir:    "~/work",
		Cmd:        "tmux new-session -ds $session -c $path 2>/dev/null; tmux switch-client -t $session 2>/dev/null || tmux attach -t $session",
		CTFSources: []string{"hackthebox", "picoctf", "root-me", "tryhackme"},
		CTFCategories: []string{
			"web",
			"reverse",
			"binary",
			"crypto",
			"forensics",
			"osint",
			"steganography",
			"mobile",
			"blockchain",
			"misc",
		},
		BoxPlatforms: []string{"hackthebox", "tryhackme", "hackmyvm"},
	}
}

// LoadConfig reads ~/.config/shelf/config.yaml and merges it over the defaults.
// If no config file exists, one is written with the defaults.
func LoadConfig() (*Config, error) {
	cfg := defaults()

	path, err := configPath()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		// Fall back to .yml extension.
		alt := strings.TrimSuffix(path, ".yaml") + ".yml"
		data, err = os.ReadFile(alt)
		if errors.Is(err, os.ErrNotExist) {
			if err := writeDefaultConfig(path, cfg); err != nil {
				return nil, fmt.Errorf("config: init %s: %w", path, err)
			}
			cfg.resolveBaseDir()
			return &cfg, nil
		}
		if err != nil {
			return nil, fmt.Errorf("config: read %s: %w", alt, err)
		}
		path = alt
	} else if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	var fromFile Config
	if err := yaml.Unmarshal(data, &fromFile); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}

	// Only override fields that are explicitly set in the file.
	// Empty fields keep their default value.
	merge(&cfg, fromFile)
	cfg.resolveBaseDir()

	return &cfg, nil
}

// Path returns the resolved config file path.
func Path() (string, error) {
	return configPath()
}

// Command returns the argv running the configured cmd. $session and $path are
// passed as positional parameters rather than interpolated into the script, so
// directory names containing shell metacharacters cannot execute.
func (c *Config) Command(session, path string) []string {
	script := strings.NewReplacer("$session", `"$1"`, "$path", `"$2"`).Replace(c.Cmd)
	return []string{"sh", "-c", script, appName, session, path}
}

// resolveBaseDir applies the SHELF_BASE_DIR override and expands a leading ~.
func (c *Config) resolveBaseDir() {
	if v := os.Getenv("SHELF_BASE_DIR"); v != "" {
		c.BaseDir = v
	}
	c.BaseDir = ExpandHome(c.BaseDir)
}

// merge overwrites non-zero fields of dst with values from src.
// Works automatically for any new field added to Config.
func merge(dst *Config, src Config) {
	d := reflect.ValueOf(dst).Elem()
	s := reflect.ValueOf(src)
	for i := range d.NumField() {
		if !s.Field(i).IsZero() {
			d.Field(i).Set(s.Field(i))
		}
	}
}

// ExpandHome resolves a leading ~ to the user's home directory.
func ExpandHome(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}

func writeDefaultConfig(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	header := `# shelf configuration
#
# base_dir       workspace root, overridden by $SHELF_BASE_DIR
# cmd            run after selecting a directory
#                variables: $session (directory name), $path (full path)
#
# The three lists below are offered in the picker whether or not they exist
# on disk. Nothing is created until you select it.
#
# ctf_sources    permanent challenge sources; one-off events are typed in
# ctf_categories challenge categories, under every source
# box_platforms  box platforms

`
	body, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, append([]byte(header), body...), 0o644)
}

func configPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, appName, "config.yaml"), nil
}
