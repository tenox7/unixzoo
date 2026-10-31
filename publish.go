package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

var mimeTypes = map[string]string{
	".html": "text/html; charset=utf-8",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
}

// managed objects belong to this tool, anything else in the bucket
// such as 404.html is left alone
func managed(n string) bool {
	return n == "index.html" || strings.HasPrefix(n, "img/")
}

// save writes a copy of the site to a local folder
func save(dir string, site map[string][]byte) error {
	for n, b := range site {
		p := filepath.Join(dir, filepath.FromSlash(n))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(p, b, 0644); err != nil {
			return err
		}
	}
	return nil
}

// publish makes the bucket match the site: uploads new and changed files,
// then deletes images nothing uses any more
func publish(ctx context.Context, bucket string, site map[string][]byte, dry bool) error {
	c, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()
	b := c.Bucket(bucket)

	old := map[string][]byte{} // name: md5
	q := &storage.Query{}
	if err := q.SetAttrSelection([]string{"Name", "MD5"}); err != nil {
		return err
	}
	it := b.Objects(ctx, q)
	for {
		a, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("list gs://%v: %w", bucket, err)
		}
		if managed(a.Name) {
			old[a.Name] = a.MD5
		}
	}

	// images first and the page last, so the page never shows a missing image
	names := slices.DeleteFunc(slices.Sorted(maps.Keys(site)), func(n string) bool { return n == "index.html" })
	names = append(names, "index.html")

	var up, del int
	for _, n := range names {
		sum := md5.Sum(site[n])
		if bytes.Equal(old[n], sum[:]) {
			continue
		}
		log.Printf("upload %v", n)
		up++
		if dry {
			continue
		}
		w := b.Object(n).NewWriter(ctx)
		w.ContentType = mimeTypes[path.Ext(n)]
		if n == "index.html" {
			w.CacheControl = "no-cache" // changes show up right away
		}
		w.MD5 = sum[:] // the bucket checks it
		_, err := w.Write(site[n])
		if cerr := w.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			return fmt.Errorf("upload %v: %w", n, err)
		}
	}

	for _, n := range slices.Sorted(maps.Keys(old)) {
		if _, ok := site[n]; ok {
			continue
		}
		log.Printf("delete %v", n)
		del++
		if dry {
			continue
		}
		if err := b.Object(n).Delete(ctx); err != nil {
			return fmt.Errorf("delete %v: %w", n, err)
		}
	}

	if dry {
		log.Printf("dry run: %v files, would upload %v, would delete %v", len(site), up, del)
		return nil
	}
	log.Printf("gs://%v: %v files, %v uploaded, %v deleted", bucket, len(site), up, del)
	return nil
}
