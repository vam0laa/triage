package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vam0laa/vamola-doctor/internal/check"
)

func TestVisibleLenIgnoresANSI(t *testing.T) {
	if visibleLen("\x1b[1;38;5;111mabç\x1b[0m") != 3 {
		t.Fatal("ANSI sequences must not count")
	}
}

func TestScore(t *testing.T) {
	if p, ok := Score([]check.Result{{Status: check.OK}, {Status: check.Warn}, {Status: check.Skip}}); !ok || p != 80 {
		t.Fatalf("got %d,%v want 80,true", p, ok)
	}
	if _, ok := Score([]check.Result{{Status: check.Skip}}); ok {
		t.Fatal("all-skipped must have no score")
	}
	if p, _ := Score([]check.Result{{Status: check.Fail}}); p != 0 {
		t.Fatal("a failure is 0")
	}
}

func TestBar(t *testing.T) {
	f, e := Bar(50, 10)
	if strings.Count(f, "█") != 5 || strings.Count(e, "░") != 5 {
		t.Fatalf("bar: %q %q", f, e)
	}
	f, e = Bar(150, 4)
	if strings.Count(f, "█") != 4 || e != "" {
		t.Fatal("bar must clamp")
	}
}

func TestBoxLinesAreEqualWidth(t *testing.T) {
	s := Style{Color: true, Width: 40}
	for _, l := range strings.Split(strings.TrimRight(s.box("Title", []string{"a", s.paint(cOK, "çok renkli")}), "\n"), "\n") {
		if visibleLen(l) != 40 {
			t.Fatalf("line %q is %d wide, want 40", l, visibleLen(l))
		}
	}
}

func TestRenderNoColorHasNoEscapes(t *testing.T) {
	var buf bytes.Buffer
	cs := []check.Check{{Name: "x", TitleKey: "title.disk"}}
	rs := []check.Result{{Status: check.Warn, Summary: "s", Details: []string{"a", "  note", "b"}}}
	Render(&buf, "0.0.0", cs, rs, Style{Color: false, Width: 60})
	out := buf.String()
	if strings.Contains(out, "\x1b[") {
		t.Fatal("no-color output must not contain escapes")
	}
	if !strings.Contains(out, "├─ a") || !strings.Contains(out, "└─ b") {
		t.Fatalf("tree glyphs missing:\n%s", out)
	}
}

func TestClip(t *testing.T) {
	if clip("abcdef", 4) != "abc…" {
		t.Fatal("clip should add an ellipsis")
	}
	if clip("abc", 4) != "abc" || clip("çğüş", 4) != "çğüş" {
		t.Fatal("short text must stay intact")
	}
}
