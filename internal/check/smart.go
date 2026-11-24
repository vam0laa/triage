package check

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vam0laa/vamola-doctor/internal/i18n"
)

type smartJSON struct {
	ModelName string `json:"model_name"`
	Smartctl  struct {
		Messages []struct {
			String   string `json:"string"`
			Severity string `json:"severity"`
		} `json:"messages"`
	} `json:"smartctl"`
	SmartStatus *struct {
		Passed bool `json:"passed"`
	} `json:"smart_status"`
	Temperature struct {
		Current int `json:"current"`
	} `json:"temperature"`
	NVMe struct {
		PercentageUsed int `json:"percentage_used"`
		SpareLeft      int `json:"available_spare"`
		SpareThreshold int `json:"available_spare_threshold"`
		MediaErrors    int `json:"media_errors"`
	} `json:"nvme_smart_health_information_log"`
}

// analyzeSmart interprets `smartctl -j` output. ok=false means the device was not readable.
func analyzeSmart(dev string, raw []byte) (st Status, line string, ok bool) {
	var s smartJSON
	if err := json.Unmarshal(raw, &s); err != nil {
		return Skip, "", false
	}
	for _, m := range s.Smartctl.Messages {
		if m.Severity == "error" && strings.Contains(strings.ToLower(m.String), "permission denied") {
			return Skip, "", false
		}
	}
	if s.SmartStatus == nil {
		return Skip, "", false
	}
	st = OK
	parts := []string{}
	if s.SmartStatus.Passed {
		parts = append(parts, i18n.T("smart.passed"))
	} else {
		st = Fail
		parts = append(parts, i18n.T("smart.failed"))
	}
	if s.Temperature.Current > 0 {
		parts = append(parts, fmt.Sprintf("%d°C", s.Temperature.Current))
		if s.Temperature.Current >= 70 {
			st = worst(st, Warn)
		}
	}
	if s.NVMe.PercentageUsed > 0 || s.NVMe.SpareThreshold > 0 {
		parts = append(parts, i18n.T("smart.wear", s.NVMe.PercentageUsed))
		if s.NVMe.PercentageUsed >= 90 {
			st = worst(st, Warn)
		}
		if s.NVMe.SpareLeft > 0 && s.NVMe.SpareLeft <= s.NVMe.SpareThreshold {
			st = worst(st, Fail)
			parts = append(parts, i18n.T("smart.spare_low"))
		}
		if s.NVMe.MediaErrors > 0 {
			st = worst(st, Warn)
			parts = append(parts, i18n.T("smart.media_err", s.NVMe.MediaErrors))
		}
	}
	name := dev
	if s.ModelName != "" {
		name = fmt.Sprintf("%s (%s)", dev, s.ModelName)
	}
	return st, name + ": " + strings.Join(parts, ", "), true
}

// SMART reads the health of NVMe/SATA drives. Usually needs root.
func SMART(ctx context.Context, _ Options) Result {
	if _, err := exec.LookPath("smartctl"); err != nil {
		return Result{Name: "smart", Status: Skip, Summary: i18n.T("smart.nocmd")}
	}
	var devs []string
	for _, pat := range []string{"/dev/nvme[0-9]n1", "/dev/sd[a-z]"} {
		m, _ := filepath.Glob(pat)
		devs = append(devs, m...)
	}
	if len(devs) == 0 {
		return Result{Name: "smart", Status: Skip, Summary: i18n.T("smart.nodisk")}
	}
	r := Result{Name: "smart", Status: OK}
	denied := 0
	for _, d := range devs {
		out, _ := run(ctx, "smartctl", "-j", "-H", "-A", d)
		st, line, ok := analyzeSmart(d, []byte(out))
		if !ok {
			denied++
			continue
		}
		r.Status = worst(r.Status, st)
		r.Details = append(r.Details, line)
	}
	if denied == len(devs) {
		return Result{Name: "smart", Status: Skip, Summary: i18n.T("smart.denied")}
	}
	switch r.Status {
	case OK:
		r.Summary = i18n.T("smart.ok")
	case Warn:
		r.Summary = i18n.T("smart.warn")
	default:
		r.Summary = i18n.T("smart.fail")
	}
	return r
}
