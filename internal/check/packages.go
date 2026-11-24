package check

import (
	"context"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vam0laa/vamola-doctor/internal/i18n"
)

// dirSize returns the total size of a directory in bytes, skipping unreadable parts.
func dirSize(root string) float64 {
	var n int64
	filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			if info, e := d.Info(); e == nil {
				n += info.Size()
			}
		}
		return nil
	})
	return float64(n)
}

// Packages checks orphans, pending updates, the cache and unmerged .pacnew files.
func Packages(ctx context.Context, _ Options) Result {
	t := i18n.T
	if _, err := exec.LookPath("pacman"); err != nil {
		return Result{Name: "packages", Status: Skip, Summary: t("pkg.nopacman")}
	}
	r := Result{Name: "packages", Status: OK}

	orphans, _ := run(ctx, "pacman", "-Qdtq")
	if o := lines(orphans); len(o) > 0 {
		r.Status = worst(r.Status, Warn)
		show := o
		if len(show) > 4 {
			show = show[:4]
		}
		names := strings.Join(show, ", ")
		if len(o) > len(show) {
			names += ", …"
		}
		r.Details = append(r.Details, t("pkg.orphans", len(o), names))
		r.Details = append(r.Details, t("pkg.orphan_hint"))
	} else {
		r.Details = append(r.Details, t("pkg.noorphans"))
	}

	// -Qu reads the local database only; it does not touch the network.
	upd, _ := run(ctx, "pacman", "-Qu")
	nu := len(lines(upd))
	r.Details = append(r.Details, t("pkg.updates", nu))
	if nu > 100 {
		r.Status = worst(r.Status, Warn)
	}

	if c := dirSize("/var/cache/pacman/pkg"); c > 0 {
		r.Details = append(r.Details, t("pkg.cache", human(c)))
		if c > 10*1024*1024*1024 {
			r.Status = worst(r.Status, Warn)
			r.Details = append(r.Details, t("pkg.cache_hint"))
		}
	}

	var pacnew []string
	filepath.WalkDir("/etc", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && (strings.HasSuffix(p, ".pacnew") || strings.HasSuffix(p, ".pacsave")) {
			pacnew = append(pacnew, p)
		}
		return nil
	})
	if len(pacnew) > 0 {
		r.Status = worst(r.Status, Warn)
		r.Details = append(r.Details, t("pkg.pacnew", len(pacnew), strings.Join(pacnew, ", ")))
	}

	if r.Status == OK {
		r.Summary = t("pkg.ok")
	} else {
		r.Summary = t("pkg.warn")
	}
	return r
}
