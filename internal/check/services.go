package check

import (
	"context"
	"strings"

	"github.com/vam0laa/triage/internal/i18n"
)

// Services lists failed systemd units (system and user).
func Services(ctx context.Context, _ Options) Result {
	t := i18n.T
	r := Result{Name: "services", Status: OK}
	n := 0
	for _, scope := range [][]string{{"--failed", "--no-legend", "--plain"}, {"--user", "--failed", "--no-legend", "--plain"}} {
		out, err := run(ctx, "systemctl", scope...)
		if err != nil && out == "" {
			continue
		}
		label := t("svc.system")
		if scope[0] == "--user" {
			label = t("svc.user")
		}
		for _, l := range strings.Split(out, "\n") {
			if f := strings.Fields(l); len(f) > 0 {
				n++
				r.Details = append(r.Details, t("svc.line", label, f[0]))
			}
		}
	}
	if n > 0 {
		r.Status = Fail
		r.Summary = t("svc.fail", n)
	} else {
		r.Summary = t("svc.ok")
	}
	return r
}
