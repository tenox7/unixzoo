# UNIX ZOO

Ready to run Unix virtual machines: http://www.unixzoo.com

This repo is the web site. Each folder in [content](content) is one card.
When a pull request is merged, GitHub builds the site and publishes it.

- **To add or fix a card, see [CONTRIBUTING.md](CONTRIBUTING.md).**
- Wanted cards and the submission queue: [TODO.md](TODO.md).

## The site

`unixzoo` is a Go program. It builds the site from the content folder into
the `site` folder. GitHub Pages serves it.

The page is HTML 4.01 Transitional, no JavaScript, plain ASCII.

- Modern browsers: a gallery of cards that fits the window, phones too.
- Old browsers: a grid of white boxes made from a plain table, with small JPEG
  thumbnails. They never load the modern styles or the big thumbnails.

## Content folder

```
content/
  header.html               optional, html above the grid
  footer.html               optional, html below the grid
  Dell Unix 2.2/            one folder = one card, folder name = title
    screenshot.png          the screenshot: .png .jpg or .gif
    86Box Net HiRes.url     one link file = one link, file name = link text
    Blog Post.webloc
    Wiki.desktop
```

Edit it with wfm, Finder, Explorer or any file manager.

- Link files: `.url` (Windows, wfm), `.webloc` (macOS), `.desktop` (Linux).
  Only http, https and ftp links.
- Cards and links are sorted the way Finder and Explorer sort them.
- No name can have a real `/`. Use `∕` (division slash) instead, it shows as
  `/` on the site. Folder `Apple A∕UX 3.x` is the card "Apple A/UX 3.x".
- Names must work on Windows, or git for Windows can not check out the repo:
  no `< > : " \ | ? *`, no space or dot at the end. The tool stops on such a
  name. Finder saves a typed `/` as `:`, so type `∕` instead.
- Names starting with `.` or `_` are skipped. Rename a folder to `_Foo` to hide
  a card.
- One screenshot per card. If there are more, the first one is used.
- Anything else is ignored with a warning. With `-strict` it is an error.

## Publish

GitHub Actions publishes on every push to master, eg: a merged pull request.
It takes about a minute.

- [publish](.github/workflows/publish.yml): tests, builds and puts the site on
  GitHub Pages.
  - Run it by hand: Actions, publish, Run workflow.
- [check](.github/workflows/check.yml): runs on every pull request. Tests, and
  builds with `-strict`. Publishes nothing.

### Hosting

GitHub Pages, with Source set to "GitHub Actions" in Settings, Pages. No key
or password is stored anywhere.

- Only master can deploy. The `github-pages` environment allows no other
  branch.
- DNS is at Namecheap. `www` is a CNAME to `tenox7.github.io`. The bare
  `unixzoo.com` has GitHub's A and AAAA records, and GitHub sends it on to www.
- "Enforce HTTPS" is off on purpose. Old browsers can only do http, so http
  must never redirect to https. Modern browsers try https by themselves.
- `404.html` is the page for a missing file.

## Run it yourself

```
go build
./unixzoo -preview          # opens http://localhost:8080 in your browser
./unixzoo                   # builds the site into the site folder
```

- Preview: leave it running, edit the content folder, reload the page in the
  browser to see the edit. Ctrl+C stops it.
- To see the preview from other computers, like old machines with old
  browsers, use `-preview -addr :8080`.
- Only GitHub publishes the site, from master. See Publish.
- It refuses to build when there are no cards, for example when the content
  folder is on a share that is not mounted.
- All flags: `./unixzoo -h`

## Test

```
go test ./...
```
