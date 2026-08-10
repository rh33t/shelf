package fs

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Hello World":      "hello-world",
		"Under_score":      "under-score",
		"  spaced  out  ":  "spaced-out",
		"Crème Brûlée":     "crme-brle",
		"a//b??c":          "abc",
		"---":              "",
		"already-a-slug":   "already-a-slug",
		"Box #12 (easy!!)": "box-12-easy",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestListDirsRecencyOrder(t *testing.T) {
	root := t.TempDir()
	now := time.Now()
	for i, name := range []string{"oldest", "middle", "newest"} {
		path := filepath.Join(root, name)
		if err := os.Mkdir(path, 0755); err != nil {
			t.Fatal(err)
		}
		stamp := now.Add(time.Duration(i) * time.Hour)
		if err := os.Chtimes(path, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "notes.md"), nil, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := ListDirs(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"newest", "middle", "oldest"}
	if len(got) != len(want) {
		t.Fatalf("ListDirs = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ListDirs = %v, want %v", got, want)
		}
	}
}

func TestWalkAllDirsSkipsDotDirsAndDepth(t *testing.T) {
	root := t.TempDir()
	mkdirs := []string{
		"platform/category/challenge",
		"platform/category/challenge/toolchain", // past maxWalkDepth
		"platform/.git/objects",
	}
	for _, d := range mkdirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	got, err := WalkAllDirs(root)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, g := range got {
		seen[g] = true
	}
	for _, want := range []string{"platform", "platform/category", "platform/category/challenge"} {
		if !seen[filepath.FromSlash(want)] {
			t.Errorf("missing %q in %v", want, got)
		}
	}
	for _, unwanted := range []string{"platform/category/challenge/toolchain", "platform/.git", "platform/.git/objects"} {
		if seen[filepath.FromSlash(unwanted)] {
			t.Errorf("unexpected %q in %v", unwanted, got)
		}
	}
}
