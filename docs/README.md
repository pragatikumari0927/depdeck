# Project documentation

This directory holds the canonical project guidance. Read `AGENTS.md` first. Then read this map and the documents that match your task.

## Authority order

1. [SPEC.md](SPEC.md) defines the product and technical contract.
2. [DECISIONS.md](DECISIONS.md) indexes accepted decisions. Read the linked ADR for the reason and scope of a decision.
3. [CONTEXT.md](CONTEXT.md) defines domain terms.
4. [Design](DESIGN.md), [development](DEVELOPMENT.md), [rules](RULES.md), [phases](PHASES.md), and [security](SECURITY.md) explain how to build and maintain the product.
5. Tickets in `.scratch/` plan work. They do not override the contract.

## By task

- Contributing: [CONTRIBUTING.md](../CONTRIBUTING.md).
- Project status and roadmap: [PHASES.md](PHASES.md).
- Security: [SECURITY.md](SECURITY.md) and [THREAT_MODEL.md](THREAT_MODEL.md).
- Tests: [TESTING.md](TESTING.md).
- Release process: [RELEASING.md](RELEASING.md).
- Release history: [CHANGELOG.md](../CHANGELOG.md). Active release candidate:
  [RELEASE_NOTES.md](../RELEASE_NOTES.md).
- Demo and diagrams: [asset index](assets/README.md).
- Agent work: [agents/README.md](agents/README.md).
- Decisions: [adr/](adr/).
- Project history: [handoffs](handoffs/) and [early design notes](rambles/).

The root compatibility pointers preserve older links and agent entry points
under [ADR-0016](adr/0016-canonical-docs-and-security-guidance.md). Edit the
canonical document in `docs/`. GitHub displays the root
[security entry point](../SECURITY.md) on the repository's Security Policy
page. Tool rules, hooks, and skills stay at their discovery paths.

Update this map when the document layout changes.
