package check

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// harmless bilinen, zararsız günlük satırlarıdır. Gerçek hatalardan ayrılır.
var harmless = []*regexp.Regexp{
	regexp.MustCompile(`ACPI BIOS Error \(bug\)`),
	regexp.MustCompile(`ACPI Error: AE_`),
	regexp.MustCompile(`ACPI Error: Aborting method`),
	regexp.MustCompile(`Ignoring duplicate name 'org\.freedesktop\.Notifications'`),
	regexp.MustCompile(`gkr-pam: unable to locate daemon control file`),
	regexp.MustCompile(`virt/tdx: TDX not supported by the host platform`),
	regexp.MustCompile(`^\S+ : a password is required ;`),
	regexp.MustCompile(`Couldn't start kglobalaccel from org\.kde\.kglobalaccel\.service`),
}

// classify satırları (zararsız, gerçek) olarak ayırır; boş satırları yok sayar.
func classify(lines []string) (ok, real []string) {
next:
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		for _, re := range harmless {
			if re.MatchString(l) {
				ok = append(ok, l)
				continue next
			}
		}
		real = append(real, l)
	}
	return
}

// Journal bu boot'taki hata (öncelik 3 ve üstü) kayıtlarını inceler.
func Journal(ctx context.Context, _ Options) Result {
	out, err := run(ctx, "journalctl", "-p", "3", "-b", "--no-pager", "-o", "cat")
	if err != nil && out == "" {
		return Result{Name: "journal", Status: Skip, Summary: "günlük okunamadı (systemd-journal grubunda mısın?)"}
	}
	known, unknown := classify(strings.Split(out, "\n"))
	r := Result{Name: "journal", Status: OK}
	if len(known) > 0 {
		r.Details = append(r.Details, fmt.Sprintf("%d zararsız bilinen uyarı yok sayıldı (ACPI, dbus...)", len(known)))
	}
	if len(unknown) > 0 {
		r.Status = Warn
		shown := unknown
		if len(shown) > 8 {
			shown = shown[:8]
		}
		for _, l := range shown {
			if len(l) > 140 {
				l = l[:140] + "…"
			}
			r.Details = append(r.Details, "• "+l)
		}
		if len(unknown) > len(shown) {
			r.Details = append(r.Details, fmt.Sprintf("… ve %d satır daha", len(unknown)-len(shown)))
		}
		r.Summary = fmt.Sprintf("%d incelenmesi gereken hata kaydı", len(unknown))
	} else {
		r.Summary = "bilinmeyen hata kaydı yok"
	}
	return r
}
