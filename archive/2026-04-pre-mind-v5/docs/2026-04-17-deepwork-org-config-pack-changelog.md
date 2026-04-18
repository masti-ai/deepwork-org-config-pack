# deepwork-org-config-pack Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.1.0] - 2026-04-14

### Added

- **Deepwork Mind MCP Tools Reference** — New knowledge resource added as mandatory reference for all agents. Provides comprehensive MCP tool documentation for agent workflows.

### Changed

- **Pack refresh** — Formula pack updated with latest configurations across all formula files.

## [2.0.0] - 2026-04-08

### Added

- **New formula files** — Expanded formula library with the following additions:
  - `mol-polecat-work-monorepo.formula.toml` — Monorepo workflow patterns
  - `mol-polecat-code-review.formula.toml` — Code review workflow
  - `mol-polecat-commit.formula.toml` — Commit workflow patterns
  - `mol-polecat-conflict-resolve.formula.toml` — Conflict resolution workflows
  - `mol-deacon-patrol.formula.toml` — Patrol workflow formula
  - `mol-refinery-patrol.formula.toml` — Refinery patrol patterns
  - `mol-dep-propagate.formula.toml` — Dependency propagation
  - `mol-graceful-handoff.formula.toml` — Handoff workflow patterns
  - `mol-shutdown-dance.formula.toml` — Shutdown workflow sequences
  - `mol-plan-review.formula.toml` — Plan review workflows
  - `mol-dog-checkpoint.formula.toml` — Checkpoint patterns
  - `mol-convoy-feed.formula.toml` — Convoy feed workflows
  - `towers-of-hanoi-10.formula.toml` — Task sequencing formula
  - `beads-workflow.formula.toml` — Bead-based workflow patterns
  - `beads-creation-expansion.formula.toml` — Creation expansion workflows
  - `spec-questions-interview-expansion.formula.toml` — Specification interview patterns
  - `plan-writing-expansion.formula.toml` — Plan writing workflows

### Changed

- **Pack restructuring** — Significant reorganization of formula files with 479 insertions and 565 deletions across 6 files.

## [1.0.0] - 2026-04-03

### Added

- **Initial formula pack** — Core configuration pack with foundational formula files:
  - `design.formula.toml` — Design workflow patterns
  - `spec-workflow.formula.toml` — Specification workflow
  - `code-review.formula.toml` — Code review patterns
  - `shiny.formula.toml` — General workflow patterns
  - `shiny-secure.formula.toml` — Secure workflow patterns
  - `gastown-release.formula.toml` — Release workflow patterns

---

## Automated Updates

This project uses scheduled automated updates that run every 6 hours. These updates include:

- Formula content refinements
- Configuration parameter adjustments
- Documentation alignment
- Pattern optimization based on usage feedback

Automated update commits are not individually listed in the changelog but are included in the release notes when they contribute to a version change.

---

## Versioning

This project follows [Semantic Versioning](https://semver.org/). Versions are tagged using the format `vMAJOR.MINOR.PATCH`.

- **MAJOR** — Incompatible API changes or major restructuring
- **MINOR** — New formulas or significant workflow additions
- **PATCH** — Bug fixes, formula refinements, or documentation updates

## Release Schedule

Automated pack updates occur on the following schedule (UTC):

| Time      | Frequency |
|-----------|-----------|
| 00:30     | Daily     |
| 06:30     | Daily     |
| 12:30     | Daily     |
| 18:30     | Daily     |

---

*Last updated: 2026-04-14*