package fs

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A directory name carrying shell metacharacters must reach the command as
// data, never as script.
func TestCommandDoesNotInterpolate(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "pwned")
	hostile := "box; touch " + marker

	cfg := &Config{Cmd: "echo $path"}
	argv := cfg.Command("session", hostile)

	out, err := exec.Command(argv[0], argv[1:]...).Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(out)); got != hostile {
		t.Errorf("cmd output = %q, want %q", got, hostile)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Error("injected command executed")
	}
}

func TestCommandSubstitutesBothVars(t *testing.T) {
	cfg := &Config{Cmd: "echo $session $path"}
	argv := cfg.Command("mybox", "/work/mybox")

	out, err := exec.Command(argv[0], argv[1:]...).Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(out)); got != "mybox /work/mybox" {
		t.Errorf("cmd output = %q", got)
	}
}

func TestMergeKeepsDefaultsForUnsetFields(t *testing.T) {
	cfg := defaults()
	merge(&cfg, Config{Cmd: "custom"})

	if cfg.Cmd != "custom" {
		t.Errorf("Cmd = %q, want %q", cfg.Cmd, "custom")
	}
	if cfg.BaseDir != defaults().BaseDir {
		t.Errorf("BaseDir = %q, want default", cfg.BaseDir)
	}
	if len(cfg.CTFCategories) != len(defaults().CTFCategories) {
		t.Errorf("CTFCategories = %v, want default", cfg.CTFCategories)
	}
}

func TestResolveBaseDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory")
	}

	cfg := Config{BaseDir: "~/work"}
	t.Setenv("SHELF_BASE_DIR", "")
	cfg.resolveBaseDir()
	if want := filepath.Join(home, "work"); cfg.BaseDir != want {
		t.Errorf("BaseDir = %q, want %q", cfg.BaseDir, want)
	}

	cfg = Config{BaseDir: "~/work"}
	t.Setenv("SHELF_BASE_DIR", "/lab")
	cfg.resolveBaseDir()
	if cfg.BaseDir != "/lab" {
		t.Errorf("env override ignored, BaseDir = %q", cfg.BaseDir)
	}
}
