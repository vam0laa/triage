package check

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
)

var realFS = map[string]bool{"ext4": true, "btrfs": true, "xfs": true, "f2fs": true, "vfat": true, "ntfs3": true, "exfat": true}

// Disk bağlı bölümlerin doluluğunu ve btrfs hata sayaçlarını denetler.
func Disk(ctx context.Context, _ Options) Result {
	r := Result{Name: "disk", Status: OK}
	f, err := os.Open("/proc/self/mounts")
	if err != nil {
		return Result{Name: "disk", Status: Skip, Summary: "bağlı bölümler okunamadı"}
	}
	defer f.Close()

	seen := map[string]bool{} // aynı cihaz birden çok kez bağlanabilir (btrfs alt birimleri)
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
		r.Details = append(r.Details, fmt.Sprintf("%-14s %-6s %5.1f%% dolu  (%s boş / %s)", p[1], p[2], pct, human(free), human(total)))
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
				r.Details = append(r.Details, "  btrfs hata sayaçları: atlandı (root gerekir)")
			case btrfsErrors(out) > 0:
				r.Status = worst(r.Status, Fail)
				r.Details = append(r.Details, "  btrfs: sıfır olmayan hata sayacı var → `btrfs device stats "+p[1]+"`")
			default:
				r.Details = append(r.Details, "  btrfs hata sayaçları: temiz")
			}
		}
	}
	if len(seen) == 0 {
		return Result{Name: "disk", Status: Skip, Summary: "denetlenecek bölüm bulunamadı"}
	}
	switch r.Status {
	case OK:
		r.Summary = "tüm bölümlerde yeterli alan var"
	case Warn:
		r.Summary = "bir bölüm dolmaya yakın (%85+)"
	default:
		r.Summary = "disk sorunu: bölüm dolu (%95+) ya da btrfs hatası"
	}
	return r
}

// btrfsErrors `btrfs device stats` çıktısındaki sıfır olmayan sayaçları sayar.
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
