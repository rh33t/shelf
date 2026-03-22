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
	PrimaryColor   string `yaml:"primary_color"`
	SecondaryColor string `yaml:"secondary_color"`
	Cmd            string `yaml:"cmd"`
}

// defaults returns the baseline config.
// Every field must have a value here.
func defaults() Config {
	return Config{
		PrimaryColor:   "#7aa2f7",
		SecondaryColor: "#1a1b26",
		Cmd:            "xdg-open $path",
	}
}

// LoadConfig reads ~/.config/shelf/config.yaml and merges it over the defaults.
// If no config file exists, the defaults are returned as-is.
func LoadConfig() (*Config, error) {
	cfg := defaults()

	path, err := configPath()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		// Fall back to .yml extension.
		alt := path[:len(path)-len(".yaml")] + ".yml"
		data, err = os.ReadFile(alt)
		if errors.Is(err, os.ErrNotExist) {
			if err := writeDefaultConfig(path, cfg); err != nil {
				return nil, fmt.Errorf("config: init %s: %w", path, err)
			}
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

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	return &cfg, nil
}

// Path returns the resolved config file path.
// Useful for a --show-config flag.
func Path() (string, error) {
	return configPath()
}

// ExpandCmd substitutes $session and $path in Cmd.
func (c *Config) ExpandCmd(session, path string) string {
	return strings.NewReplacer("$session", session, "$path", path).Replace(c.Cmd)
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

func (c *Config) validate() error {
	if !isValidHex(c.PrimaryColor) {
		return fmt.Errorf("primary_color %q is not a valid hex color", c.PrimaryColor)
	}
	if !isValidHex(c.SecondaryColor) {
		return fmt.Errorf("secondary_color %q is not a valid hex color", c.SecondaryColor)
	}
	return nil
}

func isValidHex(s string) bool {
	if len(s) == 0 || s[0] != '#' {
		return false
	}
	rest := s[1:]
	if len(rest) != 3 && len(rest) != 6 {
		return false
	}
	for _, c := range rest {
		if !('0' <= c && c <= '9') && !('a' <= c && c <= 'f') && !('A' <= c && c <= 'F') {
			return false
		}
	}
	return true
}

func writeDefaultConfig(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content := fmt.Sprintf(`# shelf configuration

# Accent color (hex)
primary_color: "%s"

# Background / secondary color (hex)
secondary_color: "%s"

# Command to run after selecting a directory.
# Available variables: $session (directory name), $path (full path)
cmd: "%s"
`, cfg.PrimaryColor, cfg.SecondaryColor, cfg.Cmd)
	return os.WriteFile(path, []byte(content), 0o644)
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
