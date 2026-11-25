package check

import (
	"context"
	"regexp"
	"strings"

	"github.com/vam0laa/triage/internal/i18n"
)

// harmless lists known, benign journal lines that are separated from real errors.
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

// classify splits lines into (harmless, real); blank lines are dropped.
func classify(in []string) (known, unknown []string) {
next:
	for _, l := range in {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		for _, re := range harmless {
			if re.MatchString(l) {
				known = append(known, l)
				continue next
			}
		}
		unknown = append(unknown, l)
	}
	return
}

// Journal inspects error-level (priority 3 and above) entries of the current boot.
func Journal(ctx context.Context, _ Options) Result {
	t := i18n.T
	out, err := run(ctx, "journalctl", "-p", "3", "-b", "--no-pager", "-o", "cat")
	if err != nil && out == "" {
		return Result{Name: "journal", Status: Skip, Summary: t("jr.denied")}
	}
	known, unknown := classify(strings.Split(out, "\n"))
	r := Result{Name: "journal", Status: OK}
	if len(known) > 0 {
		r.Details = append(r.Details, t("jr.ignored", len(known)))
	}
	if len(unknown) > 0 {
		r.Status = Warn
		shown := unknown
		if len(shown) > 8 {
			shown = shown[:8]
		}
		for _, l := range shown {
			if len([]rune(l)) > 120 {
				l = string([]rune(l)[:120]) + "…"
			}
			r.Details = append(r.Details, "• "+l)
		}
		if len(unknown) > len(shown) {
			r.Details = append(r.Details, t("jr.more", len(unknown)-len(shown)))
		}
		r.Summary = t("jr.warn", len(unknown))
	} else {
		r.Summary = t("jr.ok")
	}
	return r
}
