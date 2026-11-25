# Documentation assets

The [README](../../README.md) uses the Pixel Archive hero, current demo, a
mobile screenshot, and a plain-language architecture diagram. The [testing guide](../TESTING.md#pixel-archive-verification)
records historical offline fixture and browser checks; that write-up still
describes an earlier demo capture.

| Asset | Use |
| --- | --- |
| [pixel-archive-hero.svg](pixel-archive-hero.svg) | README branding |
| [depdeck-logo.svg](depdeck-logo.svg) | README logo |
| [deck-demo.gif](deck-demo.gif) | Current Pixel Archive capture from a real scan of a sample project, including package identity artwork and working group filters |
| [pixel-archive-mobile.png](pixel-archive-mobile.png) | Mobile Pixel Archive capture of the same Deck |
| [architecture.svg](architecture.svg), [architecture.png](architecture.png) | Current plain-language architecture diagram for the README; the PNG is a high-resolution export |
| [social-preview.png](social-preview.png), [social-preview.svg](social-preview.svg) | 1280×640 GitHub social preview. Upload the PNG under Settings → Social preview; there is no API for that step |
| [pixel-archive-desktop.png](pixel-archive-desktop.png) | Historical desktop Pixel Archive capture, retained for test evidence. It predates package identity artwork |
| [deck-demo-v1.gif](deck-demo-v1.gif) | Historical demonstration of the earlier Card layout, retained for reference |
| [banner.png](banner.png) | Earlier README branding |
| [scan-sequence.svg](scan-sequence.svg), [scan-sequence.png](scan-sequence.png) | Earlier exported scan sequence diagram, superseded in the README by a plain-language Mermaid sequence |

The current README uses a designed architecture image and Mermaid for the scan
sequence. The current GIF was captured from Go-rendered HTML after a live scan
of a five-package sample project (react, vite, chalk, typescript, and fsevents)
against the public package registry on 25 Nov 2026. It shows package identity
artwork: the curated Vite mark and deterministic fallback art, separate from
the Rarity seal. In that capture, Vite is Unknown because the registry response
exceeded the current size limit. The older Card design remains in
`deck-demo-v1.gif`.
