package main

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"gopkg.in/ini.v1"
	"howett.net/plist"
)

type link struct {
	Name string
	URL  template.URL // checked by readLink: http, https or ftp only
}

type card struct {
	Title string
	Shot  string // screenshot file, empty if none
	Links []link
}

var (
	shotExt = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true}
	linkExt = map[string]bool{".url": true, ".desktop": true, ".webloc": true}
	// files Windows and macOS leave in folders by themselves
	junk = map[string]bool{"thumbs.db": true, "desktop.ini": true, "icon\r": true}
	// device names Windows does not allow, with or without an extension
	winReserved = regexp.MustCompile(`(?i)^(con|prn|aux|nul|com[1-9]|lpt[1-9])(\.|$)`)
)

// slash is ∕ (U+2215 division slash), names use it for a / since no file
// system allows a real one, it shows as / on the site
const slash = "∕"

// problems counts what warn reported, -strict fails the build on any
var problems int

// warn logs something in the content folder that the site leaves out
func warn(format string, v ...any) {
	problems++
	log.Printf(format, v...)
}

// skip hidden and junk files, names starting with _ are drafts
func skip(n string) bool {
	return strings.HasPrefix(n, ".") || strings.HasPrefix(n, "_") || junk[strings.ToLower(n)]
}

// portable fails a name that Windows can not have, git for Windows does not
// check out such a path, so one bad name breaks the repo for every Windows user
func portable(n string) error {
	if i := strings.IndexFunc(n, func(r rune) bool { return r < ' ' || strings.ContainsRune(`<>:"/\|?*`, r) }); i >= 0 {
		if n[i] == ':' {
			return fmt.Errorf("Windows can not have ':' in a name, for a / use %v (division slash)", slash)
		}
		return fmt.Errorf("Windows can not have %q in a name", n[i])
	}
	if strings.HasSuffix(n, " ") || strings.HasSuffix(n, ".") {
		return errors.New("Windows can not have a name ending with a space or a dot")
	}
	if winReserved.MatchString(n) {
		return errors.New("Windows does not allow this name")
	}
	return nil
}

// title turns a file or folder name into display text
func title(n string) string {
	return strings.ReplaceAll(n, slash, "/")
}

// readContent reads one card per folder, sorted the way file managers sort
func readContent(dir string) ([]card, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var cards []card
	for _, e := range ents {
		n := e.Name()
		p := filepath.Join(dir, n)
		if skip(n) {
			continue
		}
		if err := portable(n); err != nil {
			return nil, fmt.Errorf("%v: %w", p, err)
		}
		fi, err := os.Stat(p) // follows symlinks, e.IsDir() does not
		if err != nil {
			warn("ignoring %v: %v", p, err)
			continue
		}
		if !fi.IsDir() {
			if n != "header.html" && n != "footer.html" {
				warn("ignoring %v: not a folder", p)
			}
			continue
		}
		c, err := readCard(p)
		if err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	slices.SortFunc(cards, func(a, b card) int { return natCmp(a.Title, b.Title) })
	return cards, nil
}

// readCard reads a card folder: the folder name is the title, inside is
// a screenshot and any number of link files
func readCard(dir string) (card, error) {
	c := card{Title: title(filepath.Base(dir))}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return c, err
	}
	var shots []string
	for _, e := range ents {
		n := e.Name()
		p := filepath.Join(dir, n)
		if skip(n) {
			continue
		}
		if err := portable(n); err != nil {
			return c, fmt.Errorf("%v: %w", p, err)
		}
		ext := strings.ToLower(filepath.Ext(n))
		switch {
		case e.IsDir():
			warn("ignoring %v: folder inside a card", p)
		case shotExt[ext]:
			shots = append(shots, n)
		case linkExt[ext]:
			u, err := readLink(p)
			if err != nil {
				warn("ignoring %v: %v", p, err)
				continue
			}
			c.Links = append(c.Links, link{Name: title(strings.TrimSuffix(n, filepath.Ext(n))), URL: template.URL(u)})
		default:
			warn("ignoring %v: not a screenshot or a link file", p)
		}
	}
	slices.SortFunc(c.Links, func(a, b link) int { return natCmp(a.Name, b.Name) })
	slices.SortFunc(shots, natCmp)
	if len(shots) > 0 {
		c.Shot = filepath.Join(dir, shots[0])
	}
	if len(shots) > 1 {
		warn("%v: more than one screenshot, using %v", dir, shots[0])
	}
	return c, nil
}

// readLink returns the url from a link file: .url (Windows), .desktop (Linux)
// or .webloc (macOS), the same formats wfm understands
func readLink(p string) (string, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	var u string
	switch ext := strings.ToLower(filepath.Ext(p)); ext {
	case ".url", ".desktop":
		// # and ; are common in urls, they do not start a comment here
		i, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, b)
		if err != nil {
			return "", err
		}
		s := i.Section("InternetShortcut")
		if ext == ".desktop" {
			s = i.Section("Desktop Entry")
		}
		u = s.Key("URL").String()
		if u == "" {
			u = s.Key("URL[$e]").String() // KDE writes this
		}
	case ".webloc":
		var w struct{ URL string }
		if _, err := plist.Unmarshal(b, &w); err != nil {
			return "", err
		}
		u = w.URL
	}
	u = strings.TrimSpace(u)
	pu, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	switch pu.Scheme {
	case "http", "https", "ftp":
		return u, nil
	case "":
		return "", errors.New("no url in link file")
	}
	return "", fmt.Errorf("unsupported url %q", u)
}

// fragment reads optional html from the content folder, eg: header.html
func fragment(p string) (template.HTML, error) {
	b, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	return template.HTML(bytes.TrimPrefix(b, []byte("\xef\xbb\xbf"))), err // Notepad likes to add a BOM
}

// natCmp compares the way Finder and Explorer do: ignoring case, with
// numbers in number order, so QNX 4.25 comes before QNX 10
func natCmp(a, b string) int {
	x, y := strings.ToLower(a), strings.ToLower(b)
	for x != "" && y != "" {
		if nx, ny := digits(x), digits(y); nx > 0 && ny > 0 {
			dx, dy := strings.TrimLeft(x[:nx], "0"), strings.TrimLeft(y[:ny], "0")
			if c := cmp.Compare(len(dx), len(dy)); c != 0 {
				return c
			}
			if c := strings.Compare(dx, dy); c != 0 {
				return c
			}
			x, y = x[nx:], y[ny:]
			continue
		}
		rx, sx := utf8.DecodeRuneInString(x)
		ry, sy := utf8.DecodeRuneInString(y)
		if rx != ry {
			return cmp.Compare(rx, ry)
		}
		x, y = x[sx:], y[sy:]
	}
	if c := cmp.Compare(len(x), len(y)); c != 0 {
		return c
	}
	return strings.Compare(a, b) // same apart from case or leading zeros
}

func digits(s string) int {
	n := 0
	for n < len(s) && s[n] >= '0' && s[n] <= '9' {
		n++
	}
	return n
}
