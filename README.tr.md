# vamola-doctor

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Arch Linux](https://img.shields.io/badge/Arch-Linux-1793D1?logo=archlinux&logoColor=white)](https://archlinux.org)

**Arch Linux + Hyprland için tek komutluk, salt okunur sistem sağlık kontrolü.**

Yeni kurduğun sistemde "her şey yolunda mı?" sorusuna saniyeler içinde cevap verir ve
zararsız log gürültüsünü (ACPI, dbus...) gerçek hatalardan ayırır.

🇬🇧 [English README](README.md)

```
╭──────────────────────────────────────────────────────────────────╮
│                                                                  │
│                       vamola-doctor v0.2.0                       │
│           Arch Linux + Hyprland sistem sağlık kontrolü           │
│                                                                  │
╰──────────────────────────────────────────────────────────────────╯

 ✔  Ağ                      bağlantı sağlıklı, gecikme 10 ms
    ├─ gecikme: 10 ms (3/3 başarılı)
    └─ DNS: çalışıyor (1 ms)

 ✔  Disk                    tüm bölümlerde yeterli alan var
    ├─ /              btrfs    1.6% dolu  (937.8 GB boş / 952.9 GB)
    │   btrfs hata sayaçları: temiz
    └─ /boot          vfat    18.3% dolu  (834.6 MB boş / 1022.0 MB)

 ✔  Bellek                  bellek rahat, %47 boş
    ├─ RAM: 6.9 GB kullanılabilir / 14.9 GB (%47 boş)
    └─ Swap: %0 kullanımda (4.0 GB)

 ✔  Hyprland                Hyprland sağlıklı, config temiz
    ├─ sürüm: v0.57.2
    ├─ config: hata yok
    └─ eDP-1: 2880x1800@120Hz, ölçek 1.67

 ⚠  Paketler                paket temizliği önerilir
    ├─ 6 yetim paket: cli11, extra-cmake-modules, html2markdown-debug, python-pillow, …
    │   inceleyip kaldırmak için: pacman -Rns $(pacman -Qdtq)
    ├─ bekleyen güncelleme: 0 (son senkronizasyona göre)
    └─ paket önbelleği: 2.7 GB

╭─ Özet ───────────────────────────────────────────────────────────╮
│                                                                  │
│ sorunsuz 4 · uyarı 1 · hata 0 · atlandı 0                        │
│                                                                  │
│ Sağlık puanı  ██████████████████████░░  92%                      │
│ Genel olarak iyi, birkaç şeye bakmakta fayda var.                │
│                                                                  │
╰──────────────────────────────────────────────────────────────────╯
```

## Kurulum

```bash
git clone https://github.com/vam0laa/vamola-doctor
cd vamola-doctor
go build -o vamola-doctor ./cmd/vamola-doctor
sudo install -m755 vamola-doctor /usr/local/bin/
```

Gereken: `go`. İsteğe bağlı: SMART modülü için `smartmontools`.

## Kullanım

```bash
vamola-doctor                    # tüm kontroller
vamola-doctor --only disk,smart  # seçili modüller
vamola-doctor --speed            # indirme hızı testi de yap (~10 MB)
vamola-doctor --json             # makine okunur çıktı
vamola-doctor --list             # modülleri listele
vamola-doctor --lang en          # bu çalıştırmada İngilizce
vamola-doctor --choose-lang      # dili yeniden seç ve kaydet
sudo vamola-doctor --only smart  # SMART root ister
```

İlk etkileşimli çalıştırmada dilini (English / Türkçe) sorar ve `~/.config/vamola-doctor/lang`
dosyasına kaydeder. Terminal yoksa (pipe, cron) sistem diline bakar.

**Çıkış kodu:** `0` temiz · `1` uyarı var · `2` hata var · `64` hatalı kullanım.

## Modüller

| Modül | Ne kontrol eder |
|---|---|
| `network` | gecikme, DNS, (isteğe bağlı) indirme hızı |
| `disk` | bölüm doluluğu (%85 uyarı, %95 hata), btrfs hata sayaçları |
| `smart` | NVMe/SATA SMART durumu, sıcaklık, aşınma (root gerekir) |
| `memory` | RAM ve swap |
| `services` | başarısız systemd servisleri (sistem + kullanıcı) |
| `journal` | bu boot'taki hata kayıtları, bilinen zararsızlar ayrılır |
| `hyprland` | `hyprctl configerrors`, monitör ölçeği (kesirli çözünürlük uyarısı) |
| `packages` | yetim paketler, bekleyen güncelleme, önbellek, `.pacnew` dosyaları |

**Sağlık puanı:** sorunsuz modül %100, uyarı %60, hata %0 sayılır; atlananlar hesaba katılmaz.

## Güvenlik

- Hiçbir şeyi **değiştirmez**, silmez, kurmaz. Sadece okur.
- Root gerektiren kontrol, yetki yoksa atlanır ve "atlandı" diye bildirilir.
- Hiçbir veri gönderilmez. Tek ağ trafiği: gecikme ölçümü, bir DNS sorgusu ve isteğe bağlı `--speed`.

## Yeni modül eklemek

`internal/check/` altına `func(ctx, Options) Result` yaz, `check.go` içindeki `All()` listesine ekle,
metinlerini `internal/i18n/catalog.go` içinde **iki dilde de** ekle (bir test iki kataloğun senkron
kalmasını sağlar) ve ayrıştırma mantığını birim testle doğrula.

## Yol haritası

- [ ] `--fix`: düzeltilebilen sorunlar için, her adımda onay isteyerek
- [ ] AUR paketi
- [ ] Sıcaklık ve fan (`sensors`) modülü
- [ ] Wi-Fi sinyal kalitesi

## Lisans

[MIT](LICENSE) © 2026 vam0laa
