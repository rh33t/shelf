package fs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	reNonAlphaNum = regexp.MustCompile(`[^a-z0-9\s-]`)
	reMultiSep    = regexp.MustCompile(`[-\s]+`)
)

// Slugify converts text to kebab-case ASCII slug, matching the Python tool's behavior.
func Slugify(text string) string {
	text = strings.ToLower(text)
	// Strip non-ASCII characters
	var b strings.Builder
	for _, r := range text {
		if r < 128 {
			b.WriteRune(r)
		}
	}
	text = b.String()
	text = strings.ReplaceAll(text, "_", "-")
	text = reNonAlphaNum.ReplaceAllString(text, "")
	text = reMultiSep.ReplaceAllString(text, "-")
	text = strings.Trim(text, "-")
	return text
}

// WalkAllDirs returns all subdirectory paths under root, relative to root.
func WalkAllDirs(root string) ([]string, error) {
	var results []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || path == root || !d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		results = append(results, rel)
		return nil
	})
	return results, err
}

// ListDirs returns the names of subdirectories in dir.
func ListDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs, nil
}

// MkdirAll creates dir and all parents with 0755 permissions.
func MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

// SpawnTmux creates a tmux session named sessionName with dir as the working
// directory. If the session already exists it is reused. When called from
// inside an existing tmux session the client is switched to the new session
// automatically; otherwise the session is left detached for manual attachment.
func SpawnTmux(sessionName, dir string) error {
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("tmux not found in PATH")
	}

	// Create only if the session does not already exist.
	if err := exec.Command("tmux", "has-session", "-t", sessionName).Run(); err != nil {
		if err := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", dir).Run(); err != nil {
			return fmt.Errorf("creating tmux session %q: %w", sessionName, err)
		}
	}

	// Switch to it if we are already inside tmux.
	if os.Getenv("TMUX") != "" {
		if err := exec.Command("tmux", "switch-client", "-t", sessionName).Run(); err != nil {
			return fmt.Errorf("switching to tmux session %q: %w", sessionName, err)
		}
	}

	return nil
}

// DeleteDir removes dir and all its contents.
func DeleteDir(path string) error {
	return os.RemoveAll(path)
}

// RenameDir moves oldPath to newPath.
func RenameDir(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// WriteNotesTemplate creates notes.md for a new box.
func WriteNotesTemplate(dir, platform, box string) error {
	content := fmt.Sprintf(
		"# Notes for %s\n\n"+
			"## Box Information\n"+
			"- Platform: %s\n"+
			"- Box Name: %s\n"+
			"- Difficulty:\n"+
			"- IP Address: \n"+
			"- Target OS: \n\n"+
			"## Initial Reconnaissance\n"+
			"```bash\n"+
			"# Nmap full scan\n"+
			"fastmap %s\n"+
			"```\n\n"+
			"## Enumeration\n\n"+
			"### Commands & tools used\n"+
			"```bash\n"+
			"# something great\n"+
			"```\n\n"+
			"## Exploitation\n\n"+
			"### Initial Access\n"+
			"```bash\n"+
			"# Exploit commands\n"+
			"```\n\n"+
			"## Privilege Escalation\n\n"+
			"### Local Enumeration\n"+
			"```bash\n"+
			"# Enumeration commands\n"+
			"```\n\n"+
			"### Escalation\n"+
			"```bash\n"+
			"# Privilege escalation commands\n"+
			"```\n\n"+
			"### Loot\n"+
			"- User Flag: \n"+
			"- Root Flag: \n\n"+
			"## Errors Made\n"+
			"- \n\n"+
			"## Lessons Learned\n"+
			"- \n\n"+
			"## References\n"+
			"- \n",
		box, platform, box, box,
	)
	return os.WriteFile(filepath.Join(dir, "notes.md"), []byte(content), 0644)
}
