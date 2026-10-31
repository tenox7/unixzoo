package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"golang.org/x/image/draw"
)

//go:embed index.tmpl
var indexTmpl string

var page = template.Must(template.New("index").Parse(indexTmpl))

// bigThumb is the thumbnail size for modern browsers, their cards are about
// 220 css pixels wide so this stays sharp on hidpi screens
const bigThumb = 400

// tile is a card the way the page shows it
type tile struct {
	Title            string
	Links            []link
	Full, Thumb, Big string // screenshot urls, empty if none
	W, H             int    // small thumbnail size
}

// build renders the page and the screenshots, the site is a map of
// bucket object names to their contents
func build(cards []card, header, footer template.HTML, cols, size int) (map[string][]byte, error) {
	site := map[string][]byte{}
	used := map[string]bool{}
	var tiles []tile
	for _, c := range cards {
		t := tile{Title: c.Title, Links: c.Links}
		if c.Shot != "" {
			if err := addShot(site, &t, c.Shot, slug(c.Title, used), size); err != nil {
				log.Printf("ignoring %v: %v", c.Shot, err)
			}
		}
		tiles = append(tiles, t)
	}
	var rows [][]tile
	for i := 0; i < len(tiles); i += cols {
		rows = append(rows, tiles[i:min(i+cols, len(tiles))])
	}
	var b bytes.Buffer
	err := page.Execute(&b, struct {
		Header, Footer template.HTML
		Rows           [][]tile
		Width          int
	}{header, footer, rows, 100 / cols})
	site["index.html"] = ascii(b.Bytes())
	return site, err
}

// addShot puts a screenshot and its thumbnails on the site, a small one
// for old browsers and a big one for modern browsers
func addShot(site map[string][]byte, t *tile, file, name string, size int) error {
	full, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	src, _, err := image.Decode(bytes.NewReader(full))
	if err != nil {
		return err
	}
	thumb, w, h, err := thumbnail(src, size)
	if err != nil {
		return err
	}
	big, _, _, err := thumbnail(src, bigThumb)
	if err != nil {
		return err
	}
	t.Full = "img/" + name + strings.ToLower(filepath.Ext(file))
	t.Thumb = "img/" + name + "_t.jpg"
	t.Big = "img/" + name + "_m.jpg"
	t.W, t.H = w, h
	site[t.Full], site[t.Thumb], site[t.Big] = full, thumb, big
	return nil
}

// thumbnail scales an image down to fit a size x size box, never up,
// as jpeg because old browsers that can not show png can show jpeg
func thumbnail(src image.Image, size int) ([]byte, int, int, error) {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	if w > size || h > size {
		if w >= h {
			w, h = size, max(1, h*size/w)
		} else {
			w, h = max(1, w*size/h), size
		}
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), image.White, image.Point{}, draw.Src) // white behind transparent pixels
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	var o bytes.Buffer
	err := jpeg.Encode(&o, dst, &jpeg.Options{Quality: 85})
	return o.Bytes(), w, h, err
}

// slug makes a file name from a title, AT&T 3b1 / 7300 Unix becomes
// at-t-3b1-7300-unix, used keeps the names unique
func slug(t string, used map[string]bool) string {
	var b strings.Builder
	for _, r := range strings.ToLower(t) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "-"):
			b.WriteByte('-')
		}
	}
	s := strings.TrimSuffix(b.String(), "-")
	if s == "" {
		s = "card"
	}
	n := s
	for i := 2; used[n]; i++ {
		n = fmt.Sprintf("%v-%v", s, i)
	}
	used[n] = true
	return n
}

// ascii turns everything outside ascii into references like &#8482;
// so the page does not depend on the browser knowing utf-8
func ascii(b []byte) []byte {
	var o bytes.Buffer
	for _, r := range string(b) {
		if r < utf8.RuneSelf {
			o.WriteByte(byte(r))
			continue
		}
		fmt.Fprintf(&o, "&#%d;", r)
	}
	return o.Bytes()
}
