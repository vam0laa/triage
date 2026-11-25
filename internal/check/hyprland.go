package check

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"strings"

	"github.com/vam0laa/triage/internal/i18n"
)

type hyprMonitor struct {
	Name        string  `json:"name"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	RefreshRate float64 `json:"refreshRate"`
	Scale       float64 `json:"scale"`
}

// fractionalLogical reports whether the scale yields a non-integer logical
// resolution. Hyprland then rounds to the nearest clean scale or looks blurry.
func fractionalLogical(m hyprMonitor) bool {
	if m.Scale <= 0 {
		return false
	}
	w, h := float64(m.Width)/m.Scale, float64(m.Height)/m.Scale
	return math.Abs(w-math.Round(w)) > 0.01 || math.Abs(h-math.Round(h)) > 0.01
}

// Hyprland checks the session, config errors and monitor scale.
func Hyprland(ctx context.Context, _ Options) Result {
	t := i18n.T
	if _, err := exec.LookPath("hyprctl"); err != nil {
		return Result{Name: "hyprland", Status: Skip, Summary: t("hypr.nocmd")}
	}
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
		return Result{Name: "hyprland", Status: Skip, Summary: t("hypr.nosession")}
	}
	r := Result{Name: "hyprland", Status: OK}

	if out, err := run(ctx, "hyprctl", "version", "-j"); err == nil {
		var v struct {
			Tag string `json:"tag"`
		}
		if json.Unmarshal([]byte(out), &v) == nil && v.Tag != "" {
			r.Details = append(r.Details, t("hypr.version", v.Tag))
		}
	}

	if out, _ := run(ctx, "hyprctl", "configerrors"); strings.TrimSpace(out) != "" {
		r.Status = Fail
		ls := strings.Split(out, "\n")
		r.Details = append(r.Details, t("hypr.cfg_errs", len(ls)))
		for i, l := range ls {
			if i >= 5 {
				r.Details = append(r.Details, t("hypr.cfg_more"))
				break
			}
			r.Details = append(r.Details, "  "+strings.TrimSpace(l))
		}
	} else {
		r.Details = append(r.Details, t("hypr.cfg_ok"))
	}

	out, err := run(ctx, "hyprctl", "monitors", "-j")
	if err != nil {
		r.Status = worst(r.Status, Warn)
		r.Details = append(r.Details, t("hypr.mon_fail"))
	} else {
		var ms []hyprMonitor
		if json.Unmarshal([]byte(out), &ms) == nil {
			for _, m := range ms {
				line := t("hypr.mon", m.Name, m.Width, m.Height, m.RefreshRate, m.Scale)
				if fractionalLogical(m) {
					r.Status = worst(r.Status, Warn)
					line += t("hypr.frac")
				}
				r.Details = append(r.Details, line)
			}
		}
	}

	switch r.Status {
	case OK:
		r.Summary = t("hypr.ok")
	case Warn:
		r.Summary = t("hypr.warn")
	default:
		r.Summary = t("hypr.fail")
	}
	return r
}
