package check

import "testing"

func TestParseMeminfo(t *testing.T) {
	m := parseMeminfo("MemTotal:       16000000 kB\nMemAvailable:    8000000 kB\nbozuk satir\nSwapTotal: 0 kB\n")
	if m["MemTotal"] != 16000000 || m["MemAvailable"] != 8000000 {
		t.Fatalf("beklenmeyen: %v", m)
	}
	if _, ok := m["bozuk satir"]; ok {
		t.Fatal("bozuk satır eklenmemeliydi")
	}
}

func TestClassify(t *testing.T) {
	ok, real := classify([]string{
		"ACPI BIOS Error (bug): Failure creating named object",
		"ACPI Error: AE_NOT_FOUND, During name lookup/catalog",
		"virt/tdx: TDX not supported by the host platform",
		"",
		"nvme nvme0: I/O 12 QID 3 timeout",
	})
	if len(ok) != 3 || len(real) != 1 {
		t.Fatalf("ok=%d real=%d", len(ok), len(real))
	}
}

func TestBtrfsErrors(t *testing.T) {
	clean := "[/dev/nvme0n1p2].write_io_errs    0\n[/dev/nvme0n1p2].read_io_errs     0\n"
	bad := clean + "[/dev/nvme0n1p2].corruption_errs  3\n"
	if btrfsErrors(clean) != 0 || btrfsErrors(bad) != 1 {
		t.Fatal("btrfs sayacı yanlış")
	}
}

func TestWorst(t *testing.T) {
	if worst(OK, Warn) != Warn || worst(Fail, Skip) != Fail || worst(Skip, OK) != OK {
		t.Fatal("worst yanlış")
	}
}

func TestFractionalLogical(t *testing.T) {
	if fractionalLogical(hyprMonitor{Width: 2880, Height: 1800, Scale: 1.6666666}) {
		t.Fatal("1.666 temiz olmalı (1728x1080)")
	}
	if !fractionalLogical(hyprMonitor{Width: 2880, Height: 1800, Scale: 1.55}) {
		t.Fatal("1.55 kesirli olmalı")
	}
	if fractionalLogical(hyprMonitor{Width: 1920, Height: 1080, Scale: 1}) {
		t.Fatal("1.0 temiz olmalı")
	}
}

func TestAnalyzeSmart(t *testing.T) {
	good := `{"model_name":"X","smart_status":{"passed":true},"temperature":{"current":41},"nvme_smart_health_information_log":{"percentage_used":3,"available_spare":100,"available_spare_threshold":10,"media_errors":0}}`
	if st, _, ok := analyzeSmart("/dev/nvme0n1", []byte(good)); !ok || st != OK {
		t.Fatalf("sağlıklı disk: st=%v ok=%v", st, ok)
	}
	bad := `{"smart_status":{"passed":false}}`
	if st, _, _ := analyzeSmart("/dev/sda", []byte(bad)); st != Fail {
		t.Fatal("başarısız SMART Fail olmalı")
	}
	hot := `{"smart_status":{"passed":true},"temperature":{"current":75}}`
	if st, _, _ := analyzeSmart("/dev/sda", []byte(hot)); st != Warn {
		t.Fatal("sıcak disk Warn olmalı")
	}
	denied := `{"smartctl":{"messages":[{"string":"Smartctl open device: /dev/nvme0n1 failed: Permission denied","severity":"error"}]}}`
	if _, _, ok := analyzeSmart("/dev/nvme0n1", []byte(denied)); ok {
		t.Fatal("izin hatası ok=false olmalı")
	}
	if _, _, ok := analyzeSmart("/dev/x", []byte("json değil")); ok {
		t.Fatal("bozuk girdi ok=false olmalı")
	}
}

func TestLines(t *testing.T) {
	if got := lines("a\n\n  b \n"); len(got) != 2 || got[1] != "b" {
		t.Fatalf("lines: %v", got)
	}
}
