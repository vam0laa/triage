package check

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
)

type hyprMonitor struct {
	Name        string  `json:"name"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	RefreshRate float64 `json:"refreshRate"`
	Scale       float64 `json:"scale"`
}

// fractionalLogical ölçek, mantıksal çözünürlüğü tam sayıya bölmüyorsa true döner.
// Böyle ölçeklerde Hyprland değeri en yakın temiz ölçeğe yuvarlar ya da görüntü bulanıklaşır.
func fractionalLogical(m hyprMonitor) bool {
	if m.Scale <= 0 {
		return false
	}
	w, h := float64(m.Width)/m.Scale, float64(m.Height)/m.Scale
	return math.Abs(w-math.Round(w)) > 0.01 || math.Abs(h-math.Round(h)) > 0.01
}

// Hyprland oturum, config hataları ve monitör ölçeğini denetler.
func Hyprland(ctx context.Context, _ Options) Result {
	if _, err := exec.LookPath("hyprctl"); err != nil {
		return Result{Name: "hyprland", Status: Skip, Summary: "hyprctl kurulu değil"}
	}
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
		return Result{Name: "hyprland", Status: Skip, Summary: "Hyprland oturumunda değilsin"}
	}
	r := Result{Name: "hyprland", Status: OK}

	if out, err := run(ctx, "hyprctl", "version", "-j"); err == nil {
		var v struct {
			Tag string `json:"tag"`
		}
		if json.Unmarshal([]byte(out), &v) == nil && v.Tag != "" {
			r.Details = append(r.Details, "sürüm: "+v.Tag)
		}
	}

	if out, _ := run(ctx, "hyprctl", "configerrors"); strings.TrimSpace(out) != "" {
		r.Status = Fail
		lines := strings.Split(out, "\n")
		r.Details = append(r.Details, fmt.Sprintf("config hataları (%d):", len(lines)))
		for i, l := range lines {
			if i >= 5 {
				r.Details = append(r.Details, "  … `hyprctl configerrors` ile tamamını gör")
				break
			}
			r.Details = append(r.Details, "  "+strings.TrimSpace(l))
		}
	} else {
		r.Details = append(r.Details, "config: hata yok")
	}

	out, err := run(ctx, "hyprctl", "monitors", "-j")
	if err != nil {
		r.Status = worst(r.Status, Warn)
		r.Details = append(r.Details, "monitör bilgisi okunamadı")
	} else {
		var ms []hyprMonitor
		if json.Unmarshal([]byte(out), &ms) == nil {
			for _, m := range ms {
				line := fmt.Sprintf("%s: %dx%d@%.0fHz, ölçek %.2f", m.Name, m.Width, m.Height, m.RefreshRate, m.Scale)
				if fractionalLogical(m) {
					r.Status = worst(r.Status, Warn)
					line += "  ⚠ kesirli mantıksal çözünürlük (bulanık olabilir)"
				}
				r.Details = append(r.Details, line)
			}
		}
	}

	switch r.Status {
	case OK:
		r.Summary = "Hyprland sağlıklı, config temiz"
	case Warn:
		r.Summary = "monitör ölçeği gözden geçirilmeli"
	default:
		r.Summary = "Hyprland config hataları var"
	}
	return r
}
