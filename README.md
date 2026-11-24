# vamola-doctor

Arch Linux + Hyprland için tek komutluk, **salt okunur** sistem sağlık kontrolü.
*A single-command, read-only health check for Arch Linux + Hyprland.*

Yeni kurduğun sistemde "her şey yolunda mı?" sorusuna saniyeler içinde cevap verir ve
zararsız log gürültüsünü (ACPI, dbus...) gerçek hatalardan ayırır.

```
✔ Ağ  bağlantı sağlıklı, gecikme 11 ms
✔ Disk  tüm bölümlerde yeterli alan var
– Disk sağlığı (SMART)  atlandı: root yetkisi gerekiyor
✔ Bellek  bellek rahat, %49 boş
✔ Servisler  başarısız servis yok
✔ Sistem günlüğü  bilinmeyen hata kaydı yok
    16 zararsız bilinen uyarı yok sayıldı (ACPI, dbus...)
✔ Hyprland  Hyprland sağlıklı, config temiz
⚠ Paketler  paket temizliği önerilir

6 sorunsuz · 1 uyarı · 0 hata · 1 atlandı
```

## Kurulum

```bash
git clone https://github.com/vam0laa/vamola-doctor
cd vamola-doctor
go build -o vamola-doctor ./cmd/vamola-doctor
sudo install -m755 vamola-doctor /usr/local/bin/
```

Gereken: `go`. İsteğe bağlı: `smartmontools` (SMART modülü için).

## Kullanım

```bash
vamola-doctor                    # tüm kontroller
vamola-doctor --only disk,smart  # seçili modüller
vamola-doctor --speed            # indirme hızı testi de yap (~10 MB indirir)
vamola-doctor --json             # makine okunur çıktı
vamola-doctor --list             # modülleri listele
sudo vamola-doctor --only smart  # SMART root ister
```

**Çıkış kodu:** `0` temiz · `1` uyarı var · `2` hata var · `64` hatalı kullanım.
Betiklerde ve cron/systemd zamanlayıcılarında kullanılabilir.

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

## Güvenlik

- Hiçbir şeyi **değiştirmez**, silmez, kurmaz. Sadece okur.
- Root gerektiren kontrol yetki yoksa sessizce atlanır ve "atlandı" diye bildirilir.
- `--speed` dışında ağa dosya indirmez; veri göndermez.

## Yeni modül eklemek

`internal/check/` altına `func(ctx, Options) Result` imzalı bir fonksiyon yaz ve
`check.go` içindeki `All()` listesine ekle. Ayrıştırma mantığını ayrı fonksiyona koyup
`check_test.go` içinde test et.

## Yol haritası

- [ ] `--fix`: düzeltilebilen sorunlar için, her adımda onay isteyerek
- [ ] AUR paketi
- [ ] Sıcaklık ve fan (`sensors`) modülü
- [ ] Wi-Fi sinyal kalitesi

## Lisans

[MIT](LICENSE) © 2026 vam0laa
