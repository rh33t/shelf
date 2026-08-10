package model

import (
	"path/filepath"
	"testing"

	"shelf/internal/fs"
)

func TestCurrentListState(t *testing.T) {
	base := filepath.Join("/lab", "challenges")
	cases := []struct {
		name    string
		mode    string
		current string
		want    appState
	}{
		{"no mode", "", base, stateModeSelect},
		{"ctf root", "ctf", base, stateLevel0},
		{"ctf source", "ctf", filepath.Join(base, "hackthebox"), stateLevel1},
		{"ctf category", "ctf", filepath.Join(base, "hackthebox", "web-exploitation"), stateLeaf},
		{"box root", "box", base, stateLevel0},
		{"box platform", "box", filepath.Join(base, "hackthebox"), stateLeaf},
	}
	for _, c := range cases {
		m := Model{mode: c.mode, baseDir: base, currentDir: c.current}
		if got := m.currentListState(); got != c.want {
			t.Errorf("%s: currentListState() = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestLevelLabel(t *testing.T) {
	cases := []struct {
		mode  string
		state appState
		want  string
	}{
		{"ctf", stateLevel0, "source"},
		{"ctf", stateLevel1, "category"},
		{"ctf", stateLeaf, "challenge"},
		{"box", stateLevel0, "platform"},
		{"box", stateLeaf, "box"},
	}
	for _, c := range cases {
		m := Model{mode: c.mode, state: c.state}
		if got := m.levelLabel(); got != c.want {
			t.Errorf("levelLabel(%s, %d) = %q, want %q", c.mode, c.state, got, c.want)
		}
	}
}

// Configured entries are offered at exactly the three levels that have them.
func TestDefaultsForLevel(t *testing.T) {
	cfg := &fs.Config{
		CTFSources:    []string{"hackthebox"},
		CTFCategories: []string{"web-exploitation"},
		BoxPlatforms:  []string{"tryhackme"},
	}
	cases := []struct {
		mode  string
		state appState
		want  string
	}{
		{"ctf", stateLevel0, "hackthebox"},
		{"ctf", stateLevel1, "web-exploitation"},
		{"box", stateLevel0, "tryhackme"},
	}
	for _, c := range cases {
		m := Model{mode: c.mode, state: c.state, cfg: cfg}
		got := m.defaultsForLevel()
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("defaultsForLevel(%s, %d) = %v, want [%s]", c.mode, c.state, got, c.want)
		}
	}

	for _, c := range []struct {
		mode  string
		state appState
	}{{"ctf", stateLeaf}, {"box", stateLeaf}, {"box", stateLevel1}} {
		m := Model{mode: c.mode, state: c.state, cfg: cfg}
		if got := m.defaultsForLevel(); got != nil {
			t.Errorf("defaultsForLevel(%s, %d) = %v, want nil", c.mode, c.state, got)
		}
	}
}

func TestBreadcrumb(t *testing.T) {
	base := filepath.Join("/lab", "boxes")
	cases := []struct {
		mode    string
		current string
		want    string
	}{
		{"", base, ""},
		{"box", base, "box"},
		{"box", filepath.Join(base, "hackthebox"), "box › hackthebox"},
		{"box", filepath.Join(base, "hackthebox", "codify"), "box › hackthebox › codify"},
	}
	for _, c := range cases {
		m := Model{mode: c.mode, baseDir: base, currentDir: c.current}
		if got := m.breadcrumb(); got != c.want {
			t.Errorf("breadcrumb(%q) = %q, want %q", c.current, got, c.want)
		}
	}
}

func TestBaseDirForMode(t *testing.T) {
	m := Model{cfg: &fs.Config{BaseDir: "/lab"}}
	if got, want := m.baseDirForMode("ctf"), "/lab/challenges"; got != want {
		t.Errorf("ctf base = %q, want %q", got, want)
	}
	if got, want := m.baseDirForMode("box"), "/lab/boxes"; got != want {
		t.Errorf("box base = %q, want %q", got, want)
	}
}

// ctrl+f lists every level, but only a challenge or a box is worth opening a
// session on. Anything shallower browses into it instead.
func TestIsTarget(t *testing.T) {
	ctfBase := filepath.Join("/lab", "challenges")
	boxBase := filepath.Join("/lab", "boxes")
	cases := []struct {
		mode string
		base string
		path string
		want bool
	}{
		{"ctf", ctfBase, ctfBase, false},
		{"ctf", ctfBase, filepath.Join(ctfBase, "picoctf"), false},
		{"ctf", ctfBase, filepath.Join(ctfBase, "picoctf", "forensics"), false},
		{"ctf", ctfBase, filepath.Join(ctfBase, "picoctf", "forensics", "tape-deck"), true},
		{"box", boxBase, boxBase, false},
		{"box", boxBase, filepath.Join(boxBase, "hackthebox"), false},
		{"box", boxBase, filepath.Join(boxBase, "hackthebox", "codify"), true},
	}
	for _, c := range cases {
		m := Model{mode: c.mode, baseDir: c.base}
		if got := m.isTarget(c.path); got != c.want {
			t.Errorf("isTarget(%s, %q) = %v, want %v", c.mode, c.path, got, c.want)
		}
	}
}
