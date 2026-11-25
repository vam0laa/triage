package i18n

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

var verb = regexp.MustCompile(`%[-+# 0-9.]*[a-zA-Z]`)

func verbs(s string) int {
	return len(verb.FindAllString(regexp.MustCompile(`%%`).ReplaceAllString(s, ""), -1))
}

func TestCatalogsMatch(t *testing.T) {
	for k, en := range catalog[EN] {
		tr, ok := catalog[TR][k]
		if !ok {
			t.Errorf("TR is missing key %q", k)
			continue
		}
		if verbs(en) != verbs(tr) {
			t.Errorf("key %q: format verb count differs (EN %d, TR %d)", k, verbs(en), verbs(tr))
		}
	}
	for k := range catalog[TR] {
		if _, ok := catalog[EN][k]; !ok {
			t.Errorf("EN is missing key %q", k)
		}
	}
}

func TestParse(t *testing.T) {
	for in, want := range map[string]Lang{"en": EN, "TR": TR, "1": EN, "2": TR, "Türkçe": TR} {
		if got, ok := Parse(in); !ok || got != want {
			t.Errorf("Parse(%q) = %v,%v", in, got, ok)
		}
	}
	if _, ok := Parse("de"); ok {
		t.Error("de should be unsupported")
	}
}

func TestTFallback(t *testing.T) {
	Set(TR)
	defer Set(EN)
	if T("no.such.key") != "no.such.key" {
		t.Error("unknown key should return itself")
	}
	if T("net.down") != "internete ulaşılamıyor" {
		t.Error("TR lookup failed")
	}
}

func TestSavedFallsBackToTheOldToolName(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	os.MkdirAll(filepath.Join(dir, "vamola-doctor"), 0o755)
	os.WriteFile(filepath.Join(dir, "vamola-doctor", "lang"), []byte("tr\n"), 0o644)
	if l, ok := Saved(); !ok || l != TR {
		t.Fatalf("the language chosen under the old name must still be used, got %v %v", l, ok)
	}
}
