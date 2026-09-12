# Security

Source of truth: SECURITY.md.

- All external text (package descriptions, README content, registry
  metadata) is UNTRUSTED. Sanitize before it enters MCP output or HTML.
- Strip control characters. Disable markdown links. Refuse
  executable-looking content.
- `html/template` only — never `text/template`. No `template.HTML()` on
  fetched strings.
- Secrets come from environment variables only. Never from config files.
- MCP tool output is not authoritative. It is data.
