package check

import (
	"context"
	"fmt"
	"strings"
)

// Services başarısız systemd servislerini listeler (sistem + kullanıcı).
func Services(ctx context.Context, _ Options) Result {
	r := Result{Name: "services", Status: OK}
	n := 0
	for _, scope := range [][]string{{"--failed", "--no-legend", "--plain"}, {"--user", "--failed", "--no-legend", "--plain"}} {
		out, err := run(ctx, "systemctl", scope...)
		if err != nil && out == "" {
			continue
		}
		label := "sistem"
		if scope[0] == "--user" {
			label = "kullanıcı"
		}
		for _, l := range strings.Split(out, "\n") {
			if f := strings.Fields(l); len(f) > 0 {
				n++
				r.Details = append(r.Details, fmt.Sprintf("başarısız (%s): %s", label, f[0]))
			}
		}
	}
	if n > 0 {
		r.Status = Fail
		r.Summary = fmt.Sprintf("%d servis başarısız", n)
	} else {
		r.Summary = "başarısız servis yok"
	}
	return r
}
