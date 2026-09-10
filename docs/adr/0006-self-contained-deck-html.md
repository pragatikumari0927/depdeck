# Human Deck is one self-contained `deck.html`

The CLI writes a single HTML file (inline CSS/JS, no CDN, no `depdeck serve`). Default path is `deck.html` **next to the scanned `package.json`**, not CWD. `--out -` writes HTML to stdout. First-run default is a file plus a one-line stderr confirmation, not a terminal dump of HTML.

**Status:** accepted

## Modes

`--json` writes JSON to stdout and **does not** create HTML. No `--both` in v1; compose two invocations. JSON may include `generated_at`; HTML must not (byte-identical for the same pipeline input — no timestamps, random IDs, or `Date.now()` in filter JS).

## XSS

Package descriptions are untrusted. Use `html/template` only. Never `template.HTML` / `template.JS` on fetched strings. Filter `<script>` is a static template string.

## Size

Target under 150 KB uncompressed for ~47 Cards. No per-card identicons/logos in v1.
