// Package check contains the health checks of vamola-doctor.
// Every check is read-only: it never changes anything on the system.
package check

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/vam0laa/vamola-doctor/internal/i18n"
)

// Status is the outcome of a check. A higher value is worse (except Skip).
type Status int

const (
	OK Status = iota
	Warn
	Fail
	Skip
)

func (s Status) String() string {
	switch s {
	case OK:
		return "ok"
	case Warn:
		return "warn"
	case Fail:
		return "fail"
	default:
		return "skip"
	}
}

// MarshalText makes JSON output use "ok", "warn", ... instead of numbers.
func (s Status) MarshalText() ([]byte, error) { return []byte(s.String()), nil }

// Result is the output of a single module.
type Result struct {
	Name    string   `json:"name"`
	Status  Status   `json:"status"`
	Summary string   `json:"summary"`
	Details []string `json:"details,omitempty"`
}

// Options carries command line settings into the checks.
type Options struct {
	Speed bool // run a download speed test (~10 MB); off by default
}

// Check is one module.
type Check struct {
	Name     string
	TitleKey string
	Run      func(ctx context.Context, o Options) Result
}

// Title returns the module title in the active language.
func (c Check) Title() string { return i18n.T(c.TitleKey) }

// All returns every module in execution order.
func All() []Check {
	return []Check{
		{"network", "title.network", Network},
		{"disk", "title.disk", Disk},
		{"smart", "title.smart", SMART},
		{"memory", "title.memory", Memory},
		{"services", "title.services", Services},
		{"journal", "title.journal", Journal},
		{"hyprland", "title.hyprland", Hyprland},
		{"packages", "title.packages", Packages},
	}
}

// worst returns the more severe of two statuses (Skip is ignored).
func worst(a, b Status) Status {
	if b == Skip {
		return a
	}
	if a == Skip || b > a {
		return b
	}
	return a
}

// run executes an external command with a 10 second timeout.
func run(ctx context.Context, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// lines splits s into trimmed, non-empty lines.
func lines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}
