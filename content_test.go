package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"howett.net/plist"
)

func mkFile(t *testing.T, p string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0644); err != nil {
		t.Fatal(err)
	}
}

func mkPng(t *testing.T, p string, w, h int) {
	t.Helper()
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
}

func TestReadLink(t *testing.T) {
	bin, err := plist.Marshal(struct{ URL string }{"https://example.com/bin"}, plist.BinaryFormat)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		file, data, want string // want empty: expect an error
	}{
		{"wfm.url", "[InternetShortcut]\r\nURL=http://tenox.pdp-11.ru/dellunix/DellUnix.ova\r\n", "http://tenox.pdp-11.ru/dellunix/DellUnix.ova"},
		{"hash.url", "[InternetShortcut]\r\nURL=https://loomcom.com/3b2/emulator.html#orgff8ead5\r\n", "https://loomcom.com/3b2/emulator.html#orgff8ead5"},
		{"semi.url", "[InternetShortcut]\nURL=http://example.com/a;b?c=1&d=2\n", "http://example.com/a;b?c=1&d=2"},
		{"bom.url", "\uFEFF[InternetShortcut]\r\nURL=http://example.com/bom\r\n", "http://example.com/bom"},
		{"windows.url", "[{000214A0-0000-0000-C000-000000000046}]\r\nProp3=19,11\r\n[InternetShortcut]\r\nIDList=\r\nURL=https://example.com/win\r\nIconIndex=0\r\n", "https://example.com/win"},
		{"upper.URL", "[InternetShortcut]\r\nURL=ftp://example.com/pub/\r\n", "ftp://example.com/pub/"},
		{"gnome.desktop", "[Desktop Entry]\nType=Link\nName=Foo\nURL=https://example.com/gnome\n", "https://example.com/gnome"},
		{"kde.desktop", "[Desktop Entry]\nIcon=text-html\nType=Link\nURL[$e]=https://example.com/kde\n", "https://example.com/kde"},
		{"xml.webloc", `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict><key>URL</key><string>https://example.com/xml</string></dict></plist>`, "https://example.com/xml"},
		{"bin.webloc", string(bin), "https://example.com/bin"},
		{"js.url", "[InternetShortcut]\r\nURL=javascript:alert(1)\r\n", ""},
		{"file.url", "[InternetShortcut]\r\nURL=file:///C:/foo.txt\r\n", ""},
		{"nourl.url", "[InternetShortcut]\r\nIconIndex=0\r\n", ""},
		{"wrongsection.desktop", "[InternetShortcut]\nURL=https://example.com/\n", ""},
		{"empty.webloc", "", ""},
	}
	d := t.TempDir()
	for _, c := range cases {
		p := filepath.Join(d, c.file)
		mkFile(t, p, []byte(c.data))
		got, err := readLink(p)
		if c.want == "" {
			if err == nil {
				t.Errorf("%v: want error, got %q", c.file, got)
			}
			continue
		}
		if err != nil || got != c.want {
			t.Errorf("%v: got %q, %v, want %q", c.file, got, err, c.want)
		}
	}
}

func TestNatCmp(t *testing.T) {
	want := []string{
		"Apple A/UX 0.7",
		"Apple A/UX 3.x",
		"QNX 1.2",
		"qnx 2.15",
		"QNX 4.25",
		"QNX 10",
		"SCO Xenix 86 2.1.3",
		"SCO Xenix 386 2.3.4",
		"Sun Solaris 11",
		"Sun SunOS 2.0",
		"SunSoft Interactive Unix 4.1",
	}
	got := slices.Clone(want)
	slices.Reverse(got)
	slices.SortFunc(got, natCmp)
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
	if natCmp("Foo", "foo") == 0 || natCmp("v01", "v1") == 0 {
		t.Error("different names must not compare equal")
	}
}

func TestReadContent(t *testing.T) {
	d := t.TempDir()
	url := []byte("[InternetShortcut]\r\nURL=http://example.com/\r\n")
	mkFile(t, filepath.Join(d, "QNX 10", "Wiki.url"), url)
	mkFile(t, filepath.Join(d, "QNX 4.25", "VMware.url"), url)
	mkFile(t, filepath.Join(d, "QNX 4.25", "86Box.url"), url)
	mkFile(t, filepath.Join(d, "QNX 4.25", "Bochs∕Qemu.url"), url)
	mkFile(t, filepath.Join(d, "QNX 4.25", "notes.txt"), nil)
	mkFile(t, filepath.Join(d, "QNX 4.25", ".DS_Store"), nil)
	mkFile(t, filepath.Join(d, "QNX 4.25", "Thumbs.db"), nil)
	mkFile(t, filepath.Join(d, "QNX 4.25", "_old.url"), url)
	mkPng(t, filepath.Join(d, "QNX 4.25", "shot10.png"), 8, 8)
	mkPng(t, filepath.Join(d, "QNX 4.25", "shot2.png"), 8, 8)
	mkFile(t, filepath.Join(d, "Apple A∕UX 3.x", "Shoebill.webloc"), []byte(`<plist version="1.0"><dict><key>URL</key><string>https://example.com/</string></dict></plist>`))
	mkFile(t, filepath.Join(d, "_Draft OS", "Wiki.url"), url)
	mkFile(t, filepath.Join(d, ".git", "config"), nil)
	mkFile(t, filepath.Join(d, "header.html"), []byte("<p>hi"))
	mkFile(t, filepath.Join(d, "stray.url"), url)

	before := problems
	cards, err := readContent(d)
	if err != nil {
		t.Fatal(err)
	}
	if n := problems - before; n != 3 { // notes.txt, the second screenshot, stray.url
		t.Errorf("%v problems, want 3", n)
	}
	var titles []string
	for _, c := range cards {
		titles = append(titles, c.Title)
	}
	if want := []string{"Apple A/UX 3.x", "QNX 4.25", "QNX 10"}; !slices.Equal(titles, want) {
		t.Fatalf("titles %q, want %q", titles, want)
	}
	q := cards[1]
	var names []string
	for _, l := range q.Links {
		names = append(names, l.Name)
	}
	if want := []string{"86Box", "Bochs/Qemu", "VMware"}; !slices.Equal(names, want) {
		t.Errorf("links %q, want %q", names, want)
	}
	if filepath.Base(q.Shot) != "shot2.png" {
		t.Errorf("screenshot %q, want shot2.png", q.Shot)
	}
	if cards[0].Shot != "" || len(cards[0].Links) != 1 {
		t.Errorf("A/UX card %+v", cards[0])
	}

	if _, err := readContent(filepath.Join(d, "missing")); err == nil {
		t.Error("missing content folder must fail")
	}
}

func TestPortable(t *testing.T) {
	for _, n := range []string{"Apple A∕UX 3.x", "AT&T 3b1 ∕ 7300 Unix", "SCO UnixWare 7.1.4+", "86Box Net HiRes.url", "Auxiliary.url", "con2.png"} {
		if err := portable(n); err != nil {
			t.Errorf("%q: %v", n, err)
		}
	}
	for _, n := range []string{"Apple A:UX 3.x", "What?.url", `a\b`, "x*", `"q"`, "a<b>", "a|b", "Icon\r", "Foo.", "Foo ", "AUX", "aux.url", "COM1.png", "lpt9"} {
		if err := portable(n); err == nil {
			t.Errorf("%q: want error", n)
		}
	}

	if runtime.GOOS == "windows" {
		return // can not make these files to begin with
	}
	url := []byte("[InternetShortcut]\r\nURL=http://example.com/\r\n")
	for _, p := range []string{"Apple A:UX 3.x/Wiki.url", "QNX 4.25/What?.url"} {
		d := t.TempDir()
		mkFile(t, filepath.Join(d, p), url)
		if _, err := readContent(d); err == nil || !strings.Contains(err.Error(), "Windows") {
			t.Errorf("%v: got %v, want a Windows error", p, err)
		}
	}
}

func TestFragment(t *testing.T) {
	d := t.TempDir()
	mkFile(t, filepath.Join(d, "header.html"), []byte("\uFEFF<p>hi"))
	h, err := fragment(filepath.Join(d, "header.html"))
	if err != nil || h != "<p>hi" {
		t.Errorf("got %q, %v", h, err)
	}
	if h, err := fragment(filepath.Join(d, "footer.html")); err != nil || h != "" {
		t.Errorf("missing fragment: got %q, %v", h, err)
	}
}
