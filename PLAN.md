# vamola-doctor

Arch Linux + Hyprland için tek komutluk sistem sağlık kontrolü.

## Kontroller (v0.1)
| Modül      | Ne bakar                                              | Kaynak komut            |
|------------|-------------------------------------------------------|-------------------------|
| network    | ping gecikmesi, DNS, indirme hızı                     | ping, net/http          |
| disk       | doluluk, btrfs hataları                               | df, btrfs device stats  |
| smart      | NVMe/SSD sağlığı (root gerekir, yoksa atlar)          | smartctl                |
| memory     | RAM ve swap                                           | /proc/meminfo           |
| services   | başarısız systemd servisleri                          | systemctl --failed      |
| journal    | bu boot'taki hatalar; bilinen zararsızları ayırır     | journalctl -p 3 -b      |
| hyprland   | monitör ölçeği, çözünürlük, config hataları           | hyprctl                 |
| packages   | yetim paket, eski önbellek, güncelleme bekleyen sayısı| pacman                  |

## Çıktı
- Her modül: ✔ sorun yok / ⚠ dikkat / ✖ hata
- Sonda özet satırı ve çıkış kodu (0 = temiz, 1 = uyarı, 2 = hata)
- `--json` ile makine okunur çıktı, `--only network,disk` ile seçim

## Kurallar
- Varsayılan olarak SALT OKUNUR; hiçbir şeyi değiştirmez
- Root gerektiren kontrol yetki yoksa sessizce atlanır, "atlandı" diye bildirilir
- Düzeltme (`--fix`) v0.2'de, her adım için onay ister

## Bilinen zararsız loglar (journal modülü ayırır)
- `ACPI BIOS Error (bug)` / `ACPI Error: AE_*` — BIOS tablosu kusuru
- `Ignoring duplicate name 'org.freedesktop.Notifications'` — çift dbus servis dosyası
