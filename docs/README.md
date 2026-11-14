# Project documentation

This directory holds the canonical project guidance. Read `AGENTS.md` first. Then read this map and the documents that match your task.

## Authority order

1. [SPEC.md](SPEC.md) defines the product and technical contract.
2. [DECISIONS.md](DECISIONS.md) indexes accepted decisions. Read the linked ADR for the reason and scope of a decision.
3. [CONTEXT.md](CONTEXT.md) defines domain terms.
4. [Design](DESIGN.md), [development](DEVELOPMENT.md), [rules](RULES.md), [phases](PHASES.md), and [security](SECURITY.md) explain how to build and maintain the product.
5. Tickets in `.scratch/` plan work. They do not override the contract.

## By task

- Security: [SECURITY.md](SECURITY.md) and [THREAT_MODEL.md](THREAT_MODEL.md).
- Tests: [TESTING.md](TESTING.md).
- Release: [RELEASING.md](RELEASING.md).
- Agent work: [agents/README.md](agents/README.md).
- Decisions: [adr/](adr/).

Root Markdown files with moved names are short compatibility pointers. Edit the document in `docs/`. Tool rules, hooks, and skills stay at their discovery paths.

Update this map when the document layout changes.
