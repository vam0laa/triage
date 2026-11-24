package check

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// parseMeminfo /proc/meminfo içeriğini kB cinsinden bir haritaya çevirir.
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

// Memory RAM ve swap kullanımını denetler.
func Memory(_ context.Context, _ Options) Result {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return Result{Name: "memory", Status: Skip, Summary: "/proc/meminfo okunamadı"}
	}
	m := parseMeminfo(string(b))
	total, avail := m["MemTotal"], m["MemAvailable"]
	if total == 0 {
		return Result{Name: "memory", Status: Skip, Summary: "bellek bilgisi ayrıştırılamadı"}
	}
	r := Result{Name: "memory", Status: OK}
	pct := avail / total * 100
	r.Details = append(r.Details, fmt.Sprintf("RAM: %s kullanılabilir / %s (%%%.0f boş)", human(avail*1024), human(total*1024), pct))
	switch {
	case pct < 5:
		r.Status = Fail
	case pct < 10:
		r.Status = Warn
	}
	if st := m["SwapTotal"]; st > 0 {
		used := (st - m["SwapFree"]) / st * 100
		r.Details = append(r.Details, fmt.Sprintf("Swap: %%%.0f kullanımda (%s)", used, human(st*1024)))
		if used > 80 {
			r.Status = worst(r.Status, Warn)
		}
	} else {
		r.Details = append(r.Details, "Swap: yok")
	}
	switch r.Status {
	case OK:
		r.Summary = fmt.Sprintf("bellek rahat, %%%.0f boş", pct)
	case Warn:
		r.Summary = "bellek azalıyor"
	default:
		r.Summary = "bellek kritik seviyede"
	}
	return r
}
