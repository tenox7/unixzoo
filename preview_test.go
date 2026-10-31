package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreview(t *testing.T) {
	d := t.TempDir()
	old := *contentDir
	*contentDir = d
	t.Cleanup(func() { *contentDir = old })
	url := []byte("[InternetShortcut]\r\nURL=http://example.com/\r\n")
	mkFile(t, filepath.Join(d, "Coherent 3.0", "PCem.url"), url)
	mkPng(t, filepath.Join(d, "Coherent 3.0", "shot.png"), 640, 480)

	h, err := previewHandler()
	if err != nil {
		t.Fatal(err)
	}
	get := func(p string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", p, nil))
		return w
	}
	cases := []struct {
		path  string
		code  int
		ctype string
	}{
		{"/", 200, "text/html; charset=utf-8"},
		{"/index.html", 200, "text/html; charset=utf-8"},
		{"/img/coherent-3-0_t.jpg", 200, "image/jpeg"},
		{"/img/coherent-3-0_m.jpg", 200, "image/jpeg"},
		{"/img/coherent-3-0.png", 200, "image/png"},
		{"/nope.html", 404, ""},
		{"/../content.go", 404, ""}, // served from memory, never from disk
	}
	for _, c := range cases {
		w := get(c.path)
		if w.Code != c.code || (c.ctype != "" && w.Header().Get("Content-Type") != c.ctype) {
			t.Errorf("%v: got %v %q, want %v %q", c.path, w.Code, w.Header().Get("Content-Type"), c.code, c.ctype)
		}
	}

	// edits show up on reload
	mkFile(t, filepath.Join(d, "Dell Unix 2.2", "OVA.url"), url)
	if w := get("/"); !strings.Contains(w.Body.String(), "<b>Dell Unix 2.2</b>") {
		t.Error("new card missing after reload")
	}

	// a broken content folder shows the error, not an old page
	for _, c := range []string{"Coherent 3.0", "Dell Unix 2.2"} {
		if err := os.RemoveAll(filepath.Join(d, c)); err != nil {
			t.Fatal(err)
		}
	}
	if w := get("/"); w.Code != 500 || !strings.Contains(w.Body.String(), "no cards") {
		t.Errorf("empty content: got %v %q", w.Code, w.Body.String())
	}
}
