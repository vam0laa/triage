package check

import (
	"context"
	"io"
	"net"
	"net/http"
	"sort"
	"time"

	"github.com/vam0laa/triage/internal/i18n"
)

// Network measures latency, DNS and (optionally) download speed.
func Network(ctx context.Context, o Options) Result {
	t := i18n.T
	r := Result{Name: "network", Status: OK}

	// Latency: three TCP connects to 1.1.1.1:443, take the median.
	var times []time.Duration
	for i := 0; i < 3; i++ {
		start := time.Now()
		d := net.Dialer{Timeout: 3 * time.Second}
		c, err := d.DialContext(ctx, "tcp", "1.1.1.1:443")
		if err == nil {
			times = append(times, time.Since(start))
			c.Close()
		}
	}
	if len(times) == 0 {
		r.Status = Fail
		r.Summary = t("net.down")
		return r
	}
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	med := times[len(times)/2]
	r.Details = append(r.Details, t("net.latency", med.Milliseconds(), len(times)))
	if med > 150*time.Millisecond || len(times) < 3 {
		r.Status = worst(r.Status, Warn)
	}

	// DNS
	rc, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	start := time.Now()
	if _, err := net.DefaultResolver.LookupHost(rc, "archlinux.org"); err != nil {
		r.Status = worst(r.Status, Fail)
		r.Details = append(r.Details, t("net.dns_fail", err.Error()))
	} else {
		r.Details = append(r.Details, t("net.dns_ok", time.Since(start).Milliseconds()))
	}

	if o.Speed {
		if mbit, err := speed(ctx); err != nil {
			r.Status = worst(r.Status, Warn)
			r.Details = append(r.Details, t("net.speed_fail", err.Error()))
		} else {
			r.Details = append(r.Details, t("net.speed", mbit))
		}
	}

	switch r.Status {
	case OK:
		r.Summary = t("net.ok", med.Milliseconds())
	case Warn:
		r.Summary = t("net.warn")
	default:
		r.Summary = t("net.fail")
	}
	return r
}

func speed(ctx context.Context) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://speed.cloudflare.com/__down?bytes=10000000", nil)
	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	n, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return 0, err
	}
	return float64(n) * 8 / 1e6 / time.Since(start).Seconds(), nil
}
