package main

import (
	"bytes"
	"image/jpeg"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestSlug(t *testing.T) {
	used := map[string]bool{}
	cases := []struct{ in, want string }{
		{"AT&T 3b1 / 7300 Unix", "at-t-3b1-7300-unix"},
		{"SCO UnixWare 7.1.4+", "sco-unixware-7-1-4"},
		{"  Quantum QNX 4.25  ", "quantum-qnx-4-25"},
		{"日本語", "card"},
		{"日本語!", "card-2"},
		{"sco unixware 7 1 4", "sco-unixware-7-1-4-2"},
	}
	for _, c := range cases {
		if got := slug(c.in, used); got != c.want {
			t.Errorf("slug(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestThumbnail(t *testing.T) {
	cases := []struct{ w, h, sw, sh, bw, bh int }{ // screenshot, small and big thumbnail
		{1024, 768, 160, 120, 400, 300},
		{1920, 1080, 160, 90, 400, 225},
		{600, 1200, 80, 160, 200, 400},
		{100, 50, 100, 50, 100, 50}, // never scaled up
		{2000, 1, 160, 1, 400, 1},
	}
	d := t.TempDir()
	p := filepath.Join(d, "s.png")
	size := func(b []byte) (int, int) {
		img, err := jpeg.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		return img.Bounds().Dx(), img.Bounds().Dy()
	}
	for _, c := range cases {
		mkPng(t, p, c.w, c.h)
		var site = map[string][]byte{}
		var tl tile
		if err := addShot(site, &tl, p, "x", 160); err != nil {
			t.Fatal(err)
		}
		if w, h := size(site["img/x_t.jpg"]); w != c.sw || h != c.sh || tl.W != c.sw || tl.H != c.sh {
			t.Errorf("%vx%v: small thumbnail %vx%v (tile %vx%v), want %vx%v", c.w, c.h, w, h, tl.W, tl.H, c.sw, c.sh)
		}
		if w, h := size(site["img/x_m.jpg"]); w != c.bw || h != c.bh || tl.Big != "img/x_m.jpg" {
			t.Errorf("%vx%v: big thumbnail %vx%v %q, want %vx%v", c.w, c.h, w, h, tl.Big, c.bw, c.bh)
		}
		if tl.Full != "img/x.png" || site["img/x.png"] == nil {
			t.Errorf("full screenshot %q missing", tl.Full)
		}
	}
	mkFile(t, p, []byte("not an image"))
	if err := addShot(map[string][]byte{}, &tile{}, p, "x", 160); err == nil {
		t.Error("bad image must fail")
	}
}

func TestBuild(t *testing.T) {
	d := t.TempDir()
	shot := filepath.Join(d, "shot.png")
	mkPng(t, shot, 640, 480)
	cards := []card{
		{Title: "AT&T 3b1 / 7300 Unix", Shot: shot, Links: []link{{"FreeBee", "http://tenox.pdp-11.ru/att3b1/freebee/freebee-win32.zip"}}},
		{Title: "Coherent 3.0", Links: []link{{"PCem", "http://example.com/a b?x=1&y=2"}}},
		{Title: "Ünix™"},
	}
	site, err := build(cards, "<p>top", "<p>bottom", 2, 160)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := slices.Sorted(maps.Keys(site)), []string{"img/at-t-3b1-7300-unix.png", "img/at-t-3b1-7300-unix_m.jpg", "img/at-t-3b1-7300-unix_t.jpg", "index.html"}; !slices.Equal(got, want) {
		t.Errorf("site files %q, want %q", got, want)
	}
	h := string(site["index.html"])
	for _, s := range []string{
		`<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN"`,
		"<p>top", "<p>bottom",
		"<b>AT&amp;T 3b1 / 7300 Unix</b>",
		`<a class="shot" href="img/at-t-3b1-7300-unix.png" style="--big: url(img/at-t-3b1-7300-unix_m.jpg)"><img src="img/at-t-3b1-7300-unix_t.jpg" width="160" height="120" alt="AT&amp;T 3b1 / 7300 Unix" border="0"></a>`,
		`<a href="http://example.com/a%20b?x=1&amp;y=2">PCem</a>`,
		`<div class="noshot" title="Coherent 3.0"></div>`,
		"<b>&#220;nix&#8482;</b>",
		`<td class="card" width="50%" bgcolor="#FFFFFF">`,
		`@supports (display: contents) and (aspect-ratio: 1)`,
	} {
		if !strings.Contains(h, s) {
			t.Errorf("index.html is missing %q", s)
		}
	}
	if n := strings.Count(h, "<tr"); n != 2 {
		t.Errorf("%v rows, want 2", n)
	}
	for i, r := range h {
		if r > 127 {
			t.Fatalf("non-ascii %q at %v", r, i)
		}
	}
	if bare, _ := build(cards, "", "", 2, 160); strings.Contains(string(bare["index.html"]), `class="bottom"`) {
		t.Error("no footer.html must leave no bottom section")
	}
	again, _ := build(cards, "<p>top", "<p>bottom", 2, 160)
	for k, v := range site {
		if !bytes.Equal(v, again[k]) {
			t.Errorf("%v differs between runs, every cron run would upload it", k)
		}
	}
}
