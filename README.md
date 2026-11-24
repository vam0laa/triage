# vamola-doctor

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Arch Linux](https://img.shields.io/badge/Arch-Linux-1793D1?logo=archlinux&logoColor=white)](https://archlinux.org)

**A read-only, one-command health check for Arch Linux + Hyprland.**

Right after a fresh install you want to know: *is everything OK?* `vamola-doctor` answers in a
couple of seconds, and it separates harmless log noise (ACPI, dbus...) from real problems.

🇹🇷 [Türkçe README](README.tr.md)

```
╭──────────────────────────────────────────────────────────────────╮
│                                                                  │
│                       vamola-doctor v0.2.0                       │
│                Arch Linux + Hyprland health check                │
│                                                                  │
╰──────────────────────────────────────────────────────────────────╯

 ✔  Network                 connection healthy, latency 12 ms
    ├─ latency: 12 ms (3/3 succeeded)
    └─ DNS: working (2 ms)

 ✔  Disk                    all filesystems have enough free space
    ├─ /              btrfs    1.6% used  (937.8 GB free of 952.9 GB)
    │   btrfs error counters: clean
    └─ /boot          vfat    18.3% used  (834.6 MB free of 1022.0 MB)

 ✔  Memory                  memory is fine, 47% free
    ├─ RAM: 7.0 GB available of 14.9 GB (47% free)
    └─ Swap: 0% in use (4.0 GB)

 ✔  Hyprland                Hyprland is healthy, config is clean
    ├─ version: v0.57.2
    ├─ config: no errors
    └─ eDP-1: 2880x1800@120Hz, scale 1.67

 ⚠  Packages                package cleanup recommended
    ├─ 6 orphan packages: cli11, extra-cmake-modules, html2markdown-debug, python-pillow, …
    │   review, then remove with: pacman -Rns $(pacman -Qdtq)
    ├─ pending updates: 0 (as of the last sync)
    └─ package cache: 2.7 GB

╭─ Summary ────────────────────────────────────────────────────────╮
│                                                                  │
│ healthy 4 · warnings 1 · errors 0 · skipped 0                    │
│                                                                  │
│ Health score  ██████████████████████░░  92%                      │
│ Mostly good, a few things are worth a look.                      │
│                                                                  │
╰──────────────────────────────────────────────────────────────────╯
```

## Install

```bash
git clone https://github.com/vam0laa/vamola-doctor
cd vamola-doctor
go build -o vamola-doctor ./cmd/vamola-doctor
sudo install -m755 vamola-doctor /usr/local/bin/
```

Requires `go`. Optional: `smartmontools` for the SMART module.

## Usage

```bash
vamola-doctor                    # run every check
vamola-doctor --only disk,smart  # only some modules
vamola-doctor --speed            # also run a download speed test (~10 MB)
vamola-doctor --json             # machine-readable output
vamola-doctor --list             # list modules
vamola-doctor --lang tr          # use Turkish for this run
vamola-doctor --choose-lang      # pick the language again and save it
sudo vamola-doctor --only smart  # SMART needs root
```

On the first interactive run it asks for your language (English / Türkçe) and remembers it in
`~/.config/vamola-doctor/lang`. Without a terminal (pipes, cron) it follows your system locale.

**Exit code:** `0` clean · `1` warnings · `2` errors · `64` bad usage — handy for scripts and timers.

## Modules

| Module | What it checks |
|---|---|
| `network` | latency, DNS, (optional) download speed |
| `disk` | filesystem usage (warn 85%, fail 95%), btrfs error counters |
| `smart` | NVMe/SATA SMART status, temperature, wear (needs root) |
| `memory` | RAM and swap |
| `services` | failed systemd units (system + user) |
| `journal` | errors of the current boot; known harmless lines are filtered out |
| `hyprland` | `hyprctl configerrors`, monitor scale (warns on fractional logical resolution) |
| `packages` | orphans, pending updates, cache size, `.pacnew` files |

The **health score** counts a healthy module as 100%, a warning as 60% and an error as 0%;
skipped modules are ignored.

## Safety

- It never changes, deletes or installs anything. Read-only.
- A check that needs root is skipped and reported as such when you don't have it.
- Nothing is uploaded. The only network traffic is the latency probe, a DNS lookup,
  and the optional `--speed` download.

## Adding a module

Write a `func(ctx, Options) Result` in `internal/check/`, add it to `All()` in `check.go`,
add its texts to **both** languages in `internal/i18n/catalog.go` (a test makes sure the two
catalogs stay in sync), and unit-test the parsing logic.

## Roadmap

- [ ] `--fix` for fixable problems, asking before every step
- [ ] AUR package
- [ ] Temperature and fan (`sensors`) module
- [ ] Wi-Fi signal quality

## License

[MIT](LICENSE) © 2026 vam0laa
