# Adding to UNIX ZOO

Each folder in [content](content) is one card on http://www.unixzoo.com.
When your pull request is merged, the site updates within a few minutes.

Looking for ideas? See the wanted list in [TODO.md](TODO.md).

## A card

```
content/
  Dell Unix 2.2/              folder name = card title
    screenshot.png            .png .jpg or .gif, one per card
    86Box Net HiRes.url       one link file = one link, file name = link text
```

Look at the cards already in [content](content) and do the same.

## Link files

The easy way: drag the link from your web browser's address bar into the
card folder.

- Windows makes a `.url` file.
- Mac makes a `.webloc` file.
- Linux makes a `.desktop` file.

Then rename the file to the link text you want, eg: `86Box Net HiRes.url`.

Or write a `.url` file in any text editor:

```
[InternetShortcut]
URL=https://example.com/unix.zip
```

Only http, https and ftp links.

## Names

Names must work on Windows, Mac and Linux.

- For a `/` use `∕` (division slash). Copy it from here: ∕
  - It shows as a normal `/` on the site.
  - Example: folder `Apple A∕UX 3.x` is the card "Apple A/UX 3.x".
  - Mac: do not type `/` in Finder. Finder saves it as `:`, and Windows can
    not use `:`.
- Not allowed: `< > : " \ | ? *`
- No space or dot at the end of a name.
- A name starting with `_` is hidden from the site.

## Send it

### On the GitHub web site, no tools needed

1. Open [content](content) on GitHub.
2. Press **Add file**, then **Upload files**.
3. Drag your card folder from your computer onto the page.
4. Press **Propose changes**, then **Create pull request**.

To fix a link: open the link file on GitHub, press the pencil, edit it, then
press **Propose changes**.

### With git

1. Fork this repo and clone your fork.
2. Add or change a folder in `content/`.
3. See the result: `go run . -preview` (needs [Go](https://go.dev/dl/)).
   It opens the site in your browser and uploads nothing. Reload the page
   after each edit.
4. Commit, push, and open a pull request.

## The check

GitHub checks every pull request. A red X means something is wrong. Press
**Details** to see what. Common problems:

- `no url in link file`: a `.url` file needs the `[InternetShortcut]` line.
- `not a screenshot or a link file`: use `.png`, `.jpg` or `.gif`.
- `Windows can not have ...`: rename it, see Names.

The first time you send a pull request, the check may wait until a
maintainer starts it.
