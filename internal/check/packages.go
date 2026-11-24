package check

import (
	"context"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
)

func lines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// dirSize bir klasörün toplam boyutunu bayt olarak verir; okunamayan yerleri atlar.
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

// Packages yetim paketleri, bekleyen güncellemeleri, önbelleği ve .pacnew dosyalarını denetler.
func Packages(ctx context.Context, _ Options) Result {
	if _, err := exec.LookPath("pacman"); err != nil {
		return Result{Name: "packages", Status: Skip, Summary: "pacman bulunamadı (Arch tabanlı değil)"}
	}
	r := Result{Name: "packages", Status: OK}

	orphans, _ := run(ctx, "pacman", "-Qdtq")
	if o := lines(orphans); len(o) > 0 {
		r.Status = worst(r.Status, Warn)
		show := o
		if len(show) > 6 {
			show = show[:6]
		}
		r.Details = append(r.Details, fmt.Sprintf("%d yetim paket: %s", len(o), strings.Join(show, ", ")))
		r.Details = append(r.Details, "  inceleyip kaldırmak için: pacman -Rns $(pacman -Qdtq)")
	} else {
		r.Details = append(r.Details, "yetim paket yok")
	}

	// -Qu yerel veritabanına bakar; ağa gitmez, son senkronizasyona göre sayar.
	upd, _ := run(ctx, "pacman", "-Qu")
	nu := len(lines(upd))
	r.Details = append(r.Details, fmt.Sprintf("bekleyen güncelleme: %d (son senkronizasyona göre)", nu))
	if nu > 100 {
		r.Status = worst(r.Status, Warn)
	}

	if c := dirSize("/var/cache/pacman/pkg"); c > 0 {
		r.Details = append(r.Details, "paket önbelleği: "+human(c))
		if c > 10*1024*1024*1024 {
			r.Status = worst(r.Status, Warn)
			r.Details = append(r.Details, "  temizlemek için: paccache -r")
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
		r.Details = append(r.Details, fmt.Sprintf("%d .pacnew/.pacsave dosyası birleştirilmeyi bekliyor: %s", len(pacnew), strings.Join(pacnew, ", ")))
	}

	switch r.Status {
	case OK:
		r.Summary = "paketler düzenli"
	default:
		r.Summary = "paket temizliği önerilir"
	}
	return r
}
