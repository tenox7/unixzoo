// UNIX ZOO static site generator

package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
)

var (
	contentDir = flag.String("content", "content", "content folder, one subfolder per card")
	outDir     = flag.String("out", "site", "folder to build the site into, empty to only check the content")
	strict     = flag.Bool("strict", false, "fail if anything in the content folder is left out of the site, for pull request checks")
	preview    = flag.Bool("preview", false, "show the site in the browser from a local web server, reload the page to see edits")
	listenAddr = flag.String("addr", "localhost:8080", "web server address for -preview, eg: :8080 to let other computers see it")
	numCols    = flag.Int("cols", 5, "cards per row in old browsers, modern browsers fit the window")
	thumbSize  = flag.Int("thumb", 160, "max thumbnail width and height in pixels for old browsers")
)

// generate reads the content folder and builds the site
func generate() (map[string][]byte, error) {
	cards, err := readContent(*contentDir)
	if err != nil {
		return nil, err
	}
	// an unmounted share or a wrong path would otherwise wipe the live site
	if len(cards) == 0 {
		return nil, fmt.Errorf("no cards in %v, refusing to build an empty site", *contentDir)
	}
	header, err := fragment(filepath.Join(*contentDir, "header.html"))
	if err != nil {
		return nil, err
	}
	footer, err := fragment(filepath.Join(*contentDir, "footer.html"))
	if err != nil {
		return nil, err
	}
	return build(cards, header, footer, *numCols, *thumbSize)
}

func main() {
	flag.Parse()
	if *numCols < 1 || *thumbSize < 1 {
		log.Fatal("-cols and -thumb must be 1 or more")
	}
	if *preview {
		log.Fatal(serve(*listenAddr))
	}

	site, err := generate()
	if err != nil {
		log.Fatal(err)
	}
	if *strict && problems > 0 {
		log.Fatalf("-strict: fix the %v problem(s) above", problems)
	}
	if *outDir == "" {
		return
	}
	if err := save(*outDir, site); err != nil {
		log.Fatal(err)
	}
	log.Printf("saved %v", filepath.Join(*outDir, "index.html"))
}
