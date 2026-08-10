package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// maxWalkDepth bounds the ctrl+f search to the deepest structural level
// (platform/category/challenge) so it never descends into a target's own
// working files.
const maxWalkDepth = 3

var (
	reNonAlphaNum = regexp.MustCompile(`[^a-z0-9\s-]`)
	reMultiSep    = regexp.MustCompile(`[-\s]+`)
)

// Slugify converts text to a kebab-case ASCII slug.
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

// WalkAllDirs returns subdirectory paths under root, relative to root, most
// recently modified first. Dot directories and anything past maxWalkDepth are
// skipped.
func WalkAllDirs(root string) ([]string, error) {
	var found []dirInfo
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || path == root {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		rel, _ := filepath.Rel(root, path)
		found = append(found, dirInfo{rel, modTime(d)})
		if strings.Count(rel, string(filepath.Separator))+1 >= maxWalkDepth {
			return filepath.SkipDir
		}
		return nil
	})
	return byRecency(found), err
}

// ListDirs returns the names of subdirectories in dir, most recently modified
// first.
func ListDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var dirs []dirInfo
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, dirInfo{e.Name(), modTime(e)})
		}
	}
	return byRecency(dirs), nil
}

// MkdirAll creates dir and all parents with 0755 permissions.
func MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

// DeleteDir removes dir and all its contents.
func DeleteDir(path string) error {
	return os.RemoveAll(path)
}

// RenameDir moves oldPath to newPath.
func RenameDir(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// WriteBoxNotes creates notes.md for a new box.
func WriteBoxNotes(dir, platform, box string) error {
	return writeNotes(dir, fmt.Sprintf(
		"# %s\n\n- Platform: %s\n- IP:\n\n## Recon\n\n## Foothold\n\n## Privesc\n\n## Flag\n\n- user:\n- root:\n",
		box, platform))
}

// WriteChallengeNotes creates notes.md for a new CTF challenge.
func WriteChallengeNotes(dir, source, category, challenge string) error {
	return writeNotes(dir, fmt.Sprintf(
		"# %s\n\n- Source: %s\n- Category: %s\n\n## Analysis\n\n## Solution\n\n## Flag\n\n-\n",
		challenge, source, category))
}

func writeNotes(dir, content string) error {
	return os.WriteFile(filepath.Join(dir, "notes.md"), []byte(content), 0644)
}

type dirInfo struct {
	name string
	mod  time.Time
}

// byRecency sorts newest first so the last target worked on stays at the top.
func byRecency(dirs []dirInfo) []string {
	sort.SliceStable(dirs, func(i, j int) bool { return dirs[i].mod.After(dirs[j].mod) })
	names := make([]string, len(dirs))
	for i, d := range dirs {
		names[i] = d.name
	}
	return names
}

func modTime(d os.DirEntry) time.Time {
	info, err := d.Info()
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
