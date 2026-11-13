# Review checklist

Review the change against the ticket and [SPEC.md](../SPEC.md).

- Check behavior, error handling, bounds, and compatibility.
- Treat files, registry responses, and tool input as untrusted.
- Check that HTML output escapes user and registry data.
- Check that failures do not hide partial assessment data.
- Check tests for offline operation and clear failure cases.
- Check that docs and tickets agree with the implementation.
- Report findings with severity, file, evidence, and impact. Do not mark a finding fixed without checking the change.

For release review, also follow [SECURITY.md](../SECURITY.md), [THREAT_MODEL.md](../THREAT_MODEL.md), and [TESTING.md](../TESTING.md). Record unavailable scans as unavailable.
