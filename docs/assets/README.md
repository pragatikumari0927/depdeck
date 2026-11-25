# Documentation assets

The [README](../../README.md) uses the Pixel Archive hero, current demo, and a
mobile screenshot. The [testing guide](../TESTING.md#pixel-archive-verification)
records the offline fixture and browser checks.

| Asset | Use |
| --- | --- |
| [pixel-archive-hero.svg](pixel-archive-hero.svg) | README branding |
| [deck-demo.gif](deck-demo.gif) | Current Pixel Archive capture from the Go renderer; offline sample Cards and working group filters |
| [pixel-archive-desktop.png](pixel-archive-desktop.png) | Static desktop Pixel Archive capture, retained for test evidence |
| [pixel-archive-mobile.png](pixel-archive-mobile.png) | Mobile Pixel Archive capture |
| [deck-demo-v1.gif](deck-demo-v1.gif) | Historical demonstration of the earlier Card layout, retained for reference |
| [banner.png](banner.png) | Earlier README branding |
| [architecture.svg](architecture.svg), [architecture.png](architecture.png) | Earlier exported architecture diagram |
| [scan-sequence.svg](scan-sequence.svg), [scan-sequence.png](scan-sequence.png) | Earlier exported scan sequence diagram |

The current README uses Mermaid for the architecture and scan sequence. The
current GIF was captured from Go-rendered HTML using the offline sample
assessment described in the testing guide. It predates package identity artwork:
it shows the Pixel Archive layout, but not the curated Vite mark or deterministic
fallback now present on `main`. Keep that visual limitation clear when using the
capture. The older Card design remains in `deck-demo-v1.gif`.
