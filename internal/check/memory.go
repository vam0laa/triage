package check

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/vam0laa/vamola-doctor/internal/i18n"
)

// parseMeminfo turns /proc/meminfo into a map of values in kB.
func parseMeminfo(s string) map[string]float64 {
	m := map[string]float64{}
	for _, l := range strings.Split(s, "\n") {
		k, v, ok := strings.Cut(l, ":")
		if !ok {
			continue
		}
		f := strings.Fields(v)
		if len(f) == 0 {
			continue
		}
		if n, err := strconv.ParseFloat(f[0], 64); err == nil {
			m[k] = n
		}
	}
	return m
}

// Memory checks RAM and swap usage.
func Memory(_ context.Context, _ Options) Result {
	t := i18n.T
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return Result{Name: "memory", Status: Skip, Summary: t("mem.err")}
	}
	m := parseMeminfo(string(b))
	total, avail := m["MemTotal"], m["MemAvailable"]
	if total == 0 {
		return Result{Name: "memory", Status: Skip, Summary: t("mem.parse")}
	}
	r := Result{Name: "memory", Status: OK}
	pct := avail / total * 100
	r.Details = append(r.Details, t("mem.ram", human(avail*1024), human(total*1024), pct))
	switch {
	case pct < 5:
		r.Status = Fail
	case pct < 10:
		r.Status = Warn
	}
	if st := m["SwapTotal"]; st > 0 {
		used := (st - m["SwapFree"]) / st * 100
		r.Details = append(r.Details, t("mem.swap", used, human(st*1024)))
		if used > 80 {
			r.Status = worst(r.Status, Warn)
		}
	} else {
		r.Details = append(r.Details, t("mem.noswap"))
	}
	switch r.Status {
	case OK:
		r.Summary = t("mem.ok", pct)
	case Warn:
		r.Summary = t("mem.warn")
	default:
		r.Summary = t("mem.fail")
	}
	return r
}
