package main

import (
	"log"
	"net"
	"net/http"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"
)

// previewHandler serves the site from memory, every load of the page
// rebuilds it so edits in the content folder show up on reload
func previewHandler() (http.Handler, error) {
	site, err := generate()
	if err != nil {
		return nil, err
	}
	var mu sync.Mutex
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := strings.TrimPrefix(r.URL.Path, "/")
		if n == "" {
			n = "index.html"
		}
		mu.Lock()
		if n == "index.html" {
			s, err := generate()
			if err != nil {
				mu.Unlock()
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			site = s
		}
		b, ok := site[n]
		mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", mimeTypes[path.Ext(n)])
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(b)
	}), nil
}

// serve runs the preview web server, nothing is uploaded
func serve(addr string) error {
	h, err := previewHandler()
	if err != nil {
		return err
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	if host == "" {
		host = "localhost"
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	u := "http://" + net.JoinHostPort(host, port) + "/"
	log.Printf("preview on %v reload the page to see edits, ctrl+c to stop", u)
	go openBrowser(u)
	return http.Serve(l, h)
}

// openBrowser shows a url in the default browser, best effort
func openBrowser(u string) {
	c := exec.Command("xdg-open", u)
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", u)
	case "windows":
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	}
	if err := c.Run(); err != nil {
		log.Printf("can not open a browser, go to %v: %v", u, err)
	}
}
