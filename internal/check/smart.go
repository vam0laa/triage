package check

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
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

// analyzeSmart smartctl -j çıktısını yorumlar. ok=false: aygıta erişilemedi.
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
		parts = append(parts, "SMART: geçti")
	} else {
		st = Fail
		parts = append(parts, "SMART: BAŞARISIZ")
	}
	if s.Temperature.Current > 0 {
		parts = append(parts, fmt.Sprintf("%d°C", s.Temperature.Current))
		if s.Temperature.Current >= 70 {
			st = worst(st, Warn)
		}
	}
	if s.NVMe.PercentageUsed > 0 || s.NVMe.SpareThreshold > 0 {
		parts = append(parts, fmt.Sprintf("aşınma %%%d", s.NVMe.PercentageUsed))
		if s.NVMe.PercentageUsed >= 90 {
			st = worst(st, Warn)
		}
		if s.NVMe.SpareLeft > 0 && s.NVMe.SpareLeft <= s.NVMe.SpareThreshold {
			st = worst(st, Fail)
			parts = append(parts, "yedek alan eşiğin altında")
		}
		if s.NVMe.MediaErrors > 0 {
			st = worst(st, Warn)
			parts = append(parts, fmt.Sprintf("%d medya hatası", s.NVMe.MediaErrors))
		}
	}
	name := dev
	if s.ModelName != "" {
		name = fmt.Sprintf("%s (%s)", dev, s.ModelName)
	}
	return st, name + ": " + strings.Join(parts, ", "), true
}

// SMART NVMe/SATA disklerin sağlık durumunu okur. Çoğu sistemde root gerekir.
func SMART(ctx context.Context, _ Options) Result {
	if _, err := exec.LookPath("smartctl"); err != nil {
		return Result{Name: "smart", Status: Skip, Summary: "smartctl kurulu değil (smartmontools)"}
	}
	var devs []string
	for _, pat := range []string{"/dev/nvme[0-9]n1", "/dev/sd[a-z]"} {
		m, _ := filepath.Glob(pat)
		devs = append(devs, m...)
	}
	if len(devs) == 0 {
		return Result{Name: "smart", Status: Skip, Summary: "disk bulunamadı"}
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
		return Result{Name: "smart", Status: Skip, Summary: "atlandı: root yetkisi gerekiyor (sudo vamola-doctor --only smart)"}
	}
	switch r.Status {
	case OK:
		r.Summary = "diskler sağlıklı"
	case Warn:
		r.Summary = "disk için dikkat edilecek değer var"
	default:
		r.Summary = "disk SMART testinden geçemedi"
	}
	return r
}
