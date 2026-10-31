# UNIX ZOO

Ready to run Unix virtual machines: http://www.unixzoo.com

This repo is the web site. Each folder in [content](content) is one card.
When a pull request is merged, GitHub builds the site and uploads it.

- **To add or fix a card, see [CONTRIBUTING.md](CONTRIBUTING.md).**
- Wanted cards and the submission queue: [TODO.md](TODO.md).

## The site

`unixzoo` is a Go program. It builds the site from the content folder and
uploads it to the Google Cloud Storage bucket `www.unixzoo.com`.

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

- [publish](.github/workflows/publish.yml): tests, builds and uploads.
  - Run it by hand: Actions, publish, Run workflow.
  - Tick "dry run" to only see what would be uploaded or deleted.
- [check](.github/workflows/check.yml): runs on every pull request. Tests, and
  builds with `-strict`. Uploads nothing.

### Bucket login

No key or password is stored anywhere. It uses Google Workload Identity
Federation:

1. Each publish run gets a signed token from GitHub that says which repo,
   branch and event it is.
2. Google checks it, and gives back a login that works for one hour.

- Google lets in only this repo, only the master branch, only pushes and manual
  runs. It checks the repo by its id, so a new repo with the same name does
  not get in.
- The login can only list, upload, replace and delete files in the
  `www.unixzoo.com` bucket. It can not change the bucket's settings or touch
  anything else.
- In project `tenox7`: pool `github`, provider `unixzoo`, role
  `unixzooPublisher`. The provider name in publish.yml is not a secret.

## Run it yourself

```
go build
./unixzoo -preview          # opens http://localhost:8080 in your browser, uploads nothing
./unixzoo -n                # show what would be uploaded or deleted, uploads nothing
./unixzoo                   # build and upload
```

- Preview: leave it running, edit the content folder, reload the page in the
  browser to see the edit. Ctrl+C stops it.
- To see the preview from other computers, like old machines with old
  browsers, use `-preview -addr :8080`.
- Upload: normally GitHub does it, see Publish. An upload from your computer
  puts your local folder on the site, even what is not pushed yet. The next
  GitHub run puts master back.
- Only changed files are uploaded.
- The tool owns `index.html` and `img/` in the bucket. It deletes images no card
  uses. Everything else in the bucket, like `404.html`, is left alone.
- It refuses to publish when there are no cards, for example when the content
  folder is on a share that is not mounted.
- All flags: `./unixzoo -h`

## Login

Only needed to upload from your own computer. Uses Google Application Default
Credentials: run `gcloud auth application-default login` once.

## Test

```
go test ./...
```
