package check

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/vam0laa/triage/internal/i18n"
)

var realFS = map[string]bool{"ext4": true, "btrfs": true, "xfs": true, "f2fs": true, "vfat": true, "ntfs3": true, "exfat": true}

// Disk checks filesystem usage and btrfs error counters.
func Disk(ctx context.Context, _ Options) Result {
	t := i18n.T
	r := Result{Name: "disk", Status: OK}
	f, err := os.Open("/proc/self/mounts")
	if err != nil {
		return Result{Name: "disk", Status: Skip, Summary: t("disk.mounts_err")}
	}
	defer f.Close()

	seen := map[string]bool{} // one device can be mounted many times (btrfs subvolumes)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		p := strings.Fields(sc.Text())
		if len(p) < 3 || !realFS[p[2]] || seen[p[0]] {
			continue
		}
		seen[p[0]] = true
		var st syscall.Statfs_t
		if err := syscall.Statfs(p[1], &st); err != nil || st.Blocks == 0 {
			continue
		}
		total := float64(st.Blocks) * float64(st.Bsize)
		free := float64(st.Bavail) * float64(st.Bsize)
		pct := (total - free) / total * 100
		r.Details = append(r.Details, t("disk.row", p[1], p[2], pct, human(free), human(total)))
		switch {
		case pct >= 95:
			r.Status = worst(r.Status, Fail)
		case pct >= 85:
			r.Status = worst(r.Status, Warn)
		}
		if p[2] == "btrfs" {
			out, err := run(ctx, "btrfs", "device", "stats", p[1])
			switch {
			case err != nil:
				r.Details = append(r.Details, "  "+t("disk.btrfs_skip"))
			case btrfsErrors(out) > 0:
				r.Status = worst(r.Status, Fail)
				r.Details = append(r.Details, "  "+t("disk.btrfs_err", p[1]))
			default:
				r.Details = append(r.Details, "  "+t("disk.btrfs_ok"))
			}
		}
	}
	if len(seen) == 0 {
		return Result{Name: "disk", Status: Skip, Summary: t("disk.none")}
	}
	switch r.Status {
	case OK:
		r.Summary = t("disk.ok")
	case Warn:
		r.Summary = t("disk.warn")
	default:
		r.Summary = t("disk.fail")
	}
	return r
}

// btrfsErrors counts non-zero counters in `btrfs device stats` output.
func btrfsErrors(out string) int {
	n := 0
	for _, l := range strings.Split(out, "\n") {
		p := strings.Fields(l)
		if len(p) == 2 && p[1] != "0" {
			n++
		}
	}
	return n
}

func human(b float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for b >= 1024 && i < len(units)-1 {
		b /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", b, units[i])
}
