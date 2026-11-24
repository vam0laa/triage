// Package i18n holds the message catalog and the saved language preference.
package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Lang is a supported UI language.
type Lang string

const (
	EN Lang = "en"
	TR Lang = "tr"
)

var cur = EN

// Parse understands "en", "tr", "english", "türkçe", and the menu choices "1"/"2".
func Parse(s string) (Lang, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "en", "eng", "english", "1":
		return EN, true
	case "tr", "tur", "turkish", "türkçe", "turkce", "2":
		return TR, true
	}
	return EN, false
}

// Set switches the active language.
func Set(l Lang) { cur = l }

// Current returns the active language.
func Current() Lang { return cur }

// Detect guesses the language from the locale environment variables.
func Detect() Lang {
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANGUAGE", "LANG"} {
		if v := os.Getenv(k); v != "" {
			if strings.HasPrefix(strings.ToLower(v), "tr") {
				return TR
			}
			return EN
		}
	}
	return EN
}

// T returns the message for key in the active language, formatted with args.
// It falls back to English, then to the key itself.
func T(key string, args ...any) string {
	s, ok := catalog[cur][key]
	if !ok {
		s, ok = catalog[EN][key]
	}
	if !ok {
		return key
	}
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

func prefPath() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "vamola-doctor", "lang"), nil
}

// Saved returns the language stored by a previous run, if any.
func Saved() (Lang, bool) {
	p, err := prefPath()
	if err != nil {
		return EN, false
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return EN, false
	}
	return Parse(string(b))
}

// Save stores the language preference under the user config directory.
func Save(l Lang) error {
	p, err := prefPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(string(l)+"\n"), 0o644)
}
