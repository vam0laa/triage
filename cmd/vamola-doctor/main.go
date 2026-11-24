// vamola-doctor: Arch Linux + Hyprland için salt okunur sistem sağlık kontrolü.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vam0laa/vamola-doctor/internal/check"
)

const version = "0.1.0"

func main() {
	var (
		asJSON  = flag.Bool("json", false, "makine okunur JSON çıktısı")
		only    = flag.String("only", "", "yalnızca bu modüller (virgülle): network,disk,...")
		list    = flag.Bool("list", false, "modülleri listele")
		speed   = flag.Bool("speed", false, "indirme hızı testi yap (≈10 MB indirir)")
		noColor = flag.Bool("no-color", false, "renkleri kapat")
		showVer = flag.Bool("version", false, "sürümü göster")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "vamola-doctor %s: sistem sağlık kontrolü (salt okunur)\n\nKullanım: vamola-doctor [seçenekler]\n\n", version)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nÇıkış kodu: 0 temiz, 1 uyarı var, 2 hata var")
	}
	flag.Parse()

	if *showVer {
		fmt.Println("vamola-doctor", version)
		return
	}
	all := check.All()
	if *list {
		for _, c := range all {
			fmt.Printf("%-10s %s\n", c.Name, c.Title)
		}
		return
	}

	selected := all
	if *only != "" {
		want := map[string]bool{}
		for _, n := range strings.Split(*only, ",") {
			want[strings.TrimSpace(n)] = true
		}
		selected = nil
		for _, c := range all {
			if want[c.Name] {
				selected = append(selected, c)
				delete(want, c.Name)
			}
		}
		for n := range want {
			fmt.Fprintf(os.Stderr, "bilinmeyen modül: %q (liste için --list)\n", n)
			os.Exit(64)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	opts := check.Options{Speed: *speed}
	var results []check.Result
	for _, c := range selected {
		r := c.Run(ctx, opts)
		r.Name = c.Name
		results = append(results, r)
	}

	exit := 0
	for _, r := range results {
		if r.Status == check.Fail {
			exit = 2
		} else if r.Status == check.Warn && exit < 1 {
			exit = 1
		}
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(map[string]any{"version": version, "results": results, "exit": exit})
	} else {
		render(selected, results, useColor(*noColor))
	}
	os.Exit(exit)
}

func useColor(off bool) bool {
	if off || os.Getenv("NO_COLOR") != "" {
		return false
	}
	st, err := os.Stdout.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

func render(cs []check.Check, rs []check.Result, color bool) {
	paint := func(code, s string) string {
		if !color {
			return s
		}
		return "\033[" + code + "m" + s + "\033[0m"
	}
	icon := map[check.Status]string{
		check.OK:   paint("32", "✔"),
		check.Warn: paint("33", "⚠"),
		check.Fail: paint("31", "✖"),
		check.Skip: paint("90", "–"),
	}
	fmt.Println(paint("1", "vamola-doctor "+version))
	fmt.Println()
	var ok, warn, fail, skip int
	for i, r := range rs {
		fmt.Printf("%s %s  %s\n", icon[r.Status], paint("1", cs[i].Title), r.Summary)
		for _, d := range r.Details {
			fmt.Println("    " + paint("90", d))
		}
		switch r.Status {
		case check.OK:
			ok++
		case check.Warn:
			warn++
		case check.Fail:
			fail++
		default:
			skip++
		}
	}
	fmt.Printf("\n%d sorunsuz · %d uyarı · %d hata · %d atlandı\n", ok, warn, fail, skip)
}
