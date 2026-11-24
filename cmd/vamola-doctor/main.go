// vamola-doctor: a read-only system health check for Arch Linux + Hyprland.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vam0laa/vamola-doctor/internal/check"
	"github.com/vam0laa/vamola-doctor/internal/i18n"
	"github.com/vam0laa/vamola-doctor/internal/ui"
)

const version = "0.2.0"

func isTTY(f *os.File) bool {
	st, err := f.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

// preScan finds --lang and --choose-lang before flag parsing, so that the flag
// descriptions and usage text can already be shown in the right language.
func preScan(args []string) (lang string, choose, jsonOut bool) {
	for i, a := range args {
		a = strings.TrimLeft(a, "-")
		switch {
		case a == "choose-lang":
			choose = true
		case a == "json":
			jsonOut = true
		case strings.HasPrefix(a, "lang="):
			lang = strings.TrimPrefix(a, "lang=")
		case a == "lang" && i+1 < len(args):
			lang = args[i+1]
		}
	}
	return
}

// askLanguage shows the first-run language menu and returns the choice.
func askLanguage() i18n.Lang {
	def := i18n.Detect()
	fmt.Println()
	fmt.Println("  " + i18n.T("lang.prompt"))
	fmt.Println()
	fmt.Println("    1) English")
	fmt.Println("    2) Türkçe")
	fmt.Println()
	defName := "English"
	if def == i18n.TR {
		defName = "Türkçe"
	}
	i18n.Set(def)
	fmt.Print("  " + i18n.T("lang.choice", defName) + " › ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	if l, ok := i18n.Parse(line); ok {
		def = l
	}
	return def
}

// resolveLanguage picks the UI language: --lang, saved preference, first-run
// prompt (interactive only), or the system locale.
func resolveLanguage(flagLang string, choose, jsonOut bool) {
	if flagLang != "" {
		l, ok := i18n.Parse(flagLang)
		if !ok {
			fmt.Fprintln(os.Stderr, i18n.T("lang.bad", flagLang))
			os.Exit(64)
		}
		i18n.Set(l)
		return
	}
	saved, has := i18n.Saved()
	if has && !choose {
		i18n.Set(saved)
		return
	}
	if (choose || !has) && !jsonOut && isTTY(os.Stdin) && isTTY(os.Stdout) {
		l := askLanguage()
		i18n.Set(l)
		if err := i18n.Save(l); err == nil {
			fmt.Println("  " + i18n.T("lang.saved"))
			fmt.Println()
		}
		return
	}
	i18n.Set(i18n.Detect())
}

func main() {
	langFlag, choose, jsonOut := preScan(os.Args[1:])
	resolveLanguage(langFlag, choose, jsonOut)
	t := i18n.T

	var (
		asJSON  = flag.Bool("json", false, t("flag.json"))
		only    = flag.String("only", "", t("flag.only"))
		list    = flag.Bool("list", false, t("flag.list"))
		speed   = flag.Bool("speed", false, t("flag.speed"))
		noColor = flag.Bool("no-color", false, t("flag.nocolor"))
		showVer = flag.Bool("version", false, t("flag.version"))
	)
	flag.String("lang", "", t("flag.lang"))
	flag.Bool("choose-lang", false, t("flag.chooselang"))
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, t("cli.usage", version))
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, t("cli.exitcodes"))
	}
	flag.Parse()

	if *showVer {
		fmt.Println("vamola-doctor", version)
		return
	}
	all := check.All()
	if *list {
		for _, c := range all {
			fmt.Printf("%-10s %s\n", c.Name, c.Title())
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
			fmt.Fprintln(os.Stderr, t("cli.unknown_module", n))
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
		enc.Encode(map[string]any{"version": version, "lang": string(i18n.Current()), "results": results, "exit": exit})
	} else {
		color := !*noColor && os.Getenv("NO_COLOR") == "" && isTTY(os.Stdout)
		ui.Render(os.Stdout, version, selected, results, ui.Style{Color: color, Width: ui.BoxWidth(ui.TermWidth()), Cols: ui.TermWidth()})
	}
	os.Exit(exit)
}
