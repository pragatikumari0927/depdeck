# ADR-0016: Canonical guidance lives in docs

**Date:** 2026-11-02
**Status:** accepted
**Supersedes:** ADR-0012's threat-catalog location only. ADR-0015's tool discovery and rule-source decisions remain in force.

## Decision

Keep full project guidance in `docs/`. Keep short root pointers for the moved documents. Keep root `AGENTS.md` as the discoverable startup entry. Agents must follow its links to the canonical instructions.

Separate [security policy](../SECURITY.md), [threat model](../THREAT_MODEL.md), and [test procedures](../TESTING.md). Preserve the five threats from ADR-0012. Record added input and resource threats in the threat model. Keep the root security reporting pointer so old links still work.

Keep configs, rules, hooks, and skill bundles at their tool discovery paths. Cursor rules remain the rule source. Generate Grok rules with the existing sync script. Keep tickets in `.scratch/`.

## Reason

One document map reduces duplicate guidance. Root pointers preserve agent discovery and historical links. Moving tool configuration into docs would prevent tools from loading it. Accepted ADRs and historical handoffs remain unchanged.

## Evidence and maintenance

Check local links, pointer targets, rule frontmatter, and rule-sync parity. Update active references when a canonical path changes. Update the docs index with each new document. The approved [completion plan](../../.scratch/depdeck-v1/COMPLETION_PLAN.md) records the scope.
