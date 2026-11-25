// Package ui renders check results as a colorful boxed terminal report.
package ui

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"syscall"
	"unicode/utf8"
	"unsafe"

	"github.com/vam0laa/triage/internal/check"
	"github.com/vam0laa/triage/internal/i18n"
)

// Style controls how the report is drawn.
type Style struct {
	Color bool
	Width int // total box width in columns
	Cols  int // terminal width; detail lines are cut to fit it (0 = never cut)
}

// 256-color palette.
const (
	cAccent = "38;5;111" // soft blue
	cOK     = "38;5;114" // green
	cWarn   = "38;5;179" // amber
	cFail   = "38;5;203" // red
	cDim    = "38;5;244" // grey
	cBold   = "1"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func (s Style) paint(code, text string) string {
	if !s.Color || text == "" {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func visibleLen(s string) int { return utf8.RuneCountInString(ansiRE.ReplaceAllString(s, "")) }

func padRight(s string, n int) string {
	if d := n - visibleLen(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

// TermWidth returns the terminal width in columns, or 80 when unknown.
func TermWidth() int {
	var ws struct{ Row, Col, X, Y uint16 }
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if e != 0 || ws.Col == 0 {
		return 80
	}
	return int(ws.Col)
}

// BoxWidth picks a comfortable box width for the given terminal width.
func BoxWidth(cols int) int {
	switch {
	case cols < 50:
		return 50
	case cols > 68:
		return 68
	}
	return cols
}

// Score returns the health score (0-100) over all non-skipped modules.
func Score(rs []check.Result) (pct int, ok bool) {
	var sum float64
	n := 0
	for _, r := range rs {
		switch r.Status {
		case check.OK:
			sum++
		case check.Warn:
			sum += 0.6
		case check.Fail:
		default:
			continue
		}
		n++
	}
	if n == 0 {
		return 0, false
	}
	return int(sum/float64(n)*100 + 0.5), true
}

// Bar draws a progress bar of the given width.
func Bar(pct, width int) (filled, empty string) {
	n := pct * width / 100
	if n < 0 {
		n = 0
	}
	if n > width {
		n = width
	}
	return strings.Repeat("█", n), strings.Repeat("░", width-n)
}

// box draws a rounded box around lines; title is optional.
func (s Style) box(title string, ls []string) string {
	inner := s.Width - 2
	var b strings.Builder
	top := "─"
	if title != "" {
		top = "─ " + title + " "
	}
	topLen := visibleLen(top)
	b.WriteString(s.paint(cAccent, "╭"+top+strings.Repeat("─", max(0, inner-topLen))+"╮") + "\n")
	for _, l := range ls {
		b.WriteString(s.paint(cAccent, "│") + padRight(" "+l, inner) + s.paint(cAccent, "│") + "\n")
	}
	b.WriteString(s.paint(cAccent, "╰"+strings.Repeat("─", inner)+"╯") + "\n")
	return b.String()
}

// center pads a line so it sits in the middle of a box.
func (s Style) center(line string) string {
	if d := (s.Width - 3 - visibleLen(line)) / 2; d > 0 {
		return strings.Repeat(" ", d) + line
	}
	return line
}

func (s Style) statusColor(st check.Status) string {
	switch st {
	case check.OK:
		return cOK
	case check.Warn:
		return cWarn
	case check.Fail:
		return cFail
	}
	return cDim
}

func icon(st check.Status) string {
	switch st {
	case check.OK:
		return "✔"
	case check.Warn:
		return "⚠"
	case check.Fail:
		return "✖"
	}
	return "–"
}

// Render writes the full report.
func Render(w io.Writer, version string, cs []check.Check, rs []check.Result, s Style) {
	// Banner
	fmt.Fprint(w, s.box("", []string{
		"",
		s.center(s.paint(cBold+";"+cAccent, "triage") + " " + s.paint(cDim, "v"+version)),
		s.center(s.paint(cDim, i18n.T("cli.subtitle"))),
		"",
	}))
	fmt.Fprintln(w)

	var ok, warn, fail, skip int
	for i, r := range rs {
		col := s.statusColor(r.Status)
		title := padRight(s.paint(cBold, cs[i].Title()), 24)
		fmt.Fprintf(w, " %s  %s%s\n", s.paint(col, icon(r.Status)), title, s.paint(col, r.Summary))
		renderDetails(w, s, r.Details)
		fmt.Fprintln(w)
		switch r.Status {
		case check.OK:
			ok++
		case check.Warn:
			warn++
		case check.Fail:
			fail++
		default:
			skip++
		}
	}

	// Summary box
	lines := []string{
		"",
		s.paint(cDim, i18n.T("cli.counts", ok, warn, fail, skip)),
	}
	if pct, has := Score(rs); has {
		col, verdict := cOK, i18n.T("cli.verdict_ok")
		switch {
		case fail > 0 || pct < 60:
			col, verdict = cFail, i18n.T("cli.verdict_fail")
		case warn > 0 || pct < 90:
			col, verdict = cWarn, i18n.T("cli.verdict_warn")
		}
		f, e := Bar(pct, 24)
		lines = append(lines, "",
			s.paint(cBold, i18n.T("cli.score"))+"  "+s.paint(col, f)+s.paint(cDim, e)+"  "+s.paint(col+";"+cBold, fmt.Sprintf("%d%%", pct)),
			s.paint(col, verdict))
	}
	lines = append(lines, "")
	fmt.Fprint(w, s.box(i18n.T("cli.summary_title"), lines))
}

// clip shortens txt to at most n columns, adding an ellipsis.
func clip(txt string, n int) string {
	if n <= 1 || utf8.RuneCountInString(txt) <= n {
		return txt
	}
	return string([]rune(txt)[:n-1]) + "…"
}

// renderDetails draws the detail lines as a tree. Lines starting with spaces are
// notes that belong to the item above them.
func renderDetails(w io.Writer, s Style, details []string) {
	last := -1
	for i, d := range details {
		if !strings.HasPrefix(d, " ") {
			last = i
		}
	}
	for i, d := range details {
		note := strings.HasPrefix(d, " ")
		txt := strings.TrimSpace(d)
		if s.Cols > 0 {
			txt = clip(txt, s.Cols-9) // 4 indent + glyph + spaces
		}
		if note {
			gutter := "│"
			if i > last {
				gutter = " "
			}
			fmt.Fprintf(w, "    %s   %s\n", s.paint(cDim, gutter), s.paint(cDim, txt))
			continue
		}
		glyph := "├─"
		if i == last {
			glyph = "└─"
		}
		fmt.Fprintf(w, "    %s %s\n", s.paint(cDim, glyph), s.paint(cDim, txt))
	}
}
