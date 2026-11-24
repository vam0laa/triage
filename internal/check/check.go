// Package check, vamola-doctor'un sağlık kontrollerini içerir.
// Her kontrol salt okunurdur: sistemi hiçbir şekilde değiştirmez.
package check

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// Status bir kontrolün sonucunu belirtir. Büyük değer = daha kötü (Skip hariç).
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

// MarshalText JSON çıktısında "ok", "warn" gibi okunur metin verir.
func (s Status) MarshalText() ([]byte, error) { return []byte(s.String()), nil }

// Result tek bir modülün çıktısıdır.
type Result struct {
	Name    string   `json:"name"`
	Status  Status   `json:"status"`
	Summary string   `json:"summary"`
	Details []string `json:"details,omitempty"`
}

// Options komut satırından gelen ayarlardır.
type Options struct {
	Speed bool // indirme hızı testi (≈10 MB indirir), varsayılan kapalı
}

// Check bir kontrol modülüdür.
type Check struct {
	Name  string
	Title string
	Run   func(ctx context.Context, o Options) Result
}

// All tüm modülleri çalıştırma sırasıyla döndürür.
func All() []Check {
	return []Check{
		{"network", "Ağ", Network},
		{"disk", "Disk", Disk},
		{"smart", "Disk sağlığı (SMART)", SMART},
		{"memory", "Bellek", Memory},
		{"services", "Servisler", Services},
		{"journal", "Sistem günlüğü", Journal},
		{"hyprland", "Hyprland", Hyprland},
		{"packages", "Paketler", Packages},
	}
}

// worst sonuçların en kötüsünü döndürür (Skip sayılmaz).
func worst(a, b Status) Status {
	if b == Skip {
		return a
	}
	if a == Skip || b > a {
		return b
	}
	return a
}

// run harici bir komutu 10 sn zaman aşımıyla çalıştırır.
func run(ctx context.Context, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
