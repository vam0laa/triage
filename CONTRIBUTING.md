# Contributing to triage

triage is a read-only health check for Arch Linux + Hyprland. Rules that keep it useful:

- **Read-only.** It never changes, deletes or installs anything. A check that needs root is skipped and says so.
- **Plain Go, standard library only.**
- **A finding is worth its noise.** Known harmless log lines are filtered; a new check must separate real problems from normal noise.
- **English and Turkish.** Every user-visible string lives in `internal/i18n/catalog.go` with both languages (a test checks the formats match).
- **Each check has a test** in `internal/check` that runs without the real system.

```bash
make test     # go vet + tests
make build    # ./triage
```
