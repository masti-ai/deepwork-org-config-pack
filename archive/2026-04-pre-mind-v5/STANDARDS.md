# Deepwork Standards

Single source of truth for all formats across the org. Every agent, script, and formula MUST follow these exactly.

## Commit Format

```
<type>(<scope>): <description>

[optional body]

Co-Authored-By: <model> <noreply@anthropic.com>
```

| Field | Rules |
|-------|-------|
| **type** | `feat` `fix` `refactor` `chore` `test` `docs` `perf` `ci` |
| **scope** | Bead ID (e.g. `vap-b65l`) or area (e.g. `dashboard`, `api`, `mobile`) |
| **description** | Lowercase, imperative, no period. Max 72 chars. |
| **body** | Optional. Wrap at 72 chars. Explain WHY, not what. |
| **Co-Authored-By** | Required on all AI commits. Model name + noreply. |

**Examples:**
```
feat(vap-b65l): add shelf detection confidence score to API response
fix(dashboard): prevent infinite recompile in Docker production build
refactor(api): extract auth middleware into shared module
docs(vap-ci): update deploy pipelines for live AWS infra
```

**Bad:**
```
Fixed stuff                          # no type, no scope, vague
feat: Add new feature.               # period, uppercase, no scope
update                               # meaningless
bd: backup 2026-04-01                # noise — should be .gitignored
```

## Branch Format

```
gt/<agent-or-user>/<bead-id>-<short-description>
```

| Field | Rules |
|-------|-------|
| **prefix** | Always `gt/` |
| **agent-or-user** | Agent name (`polecat-nitro`, `crew-mel`) or human username |
| **bead-id** | The bead being worked on |
| **short-description** | Lowercase, hyphenated, 3-5 words max |

**Examples:**
```
gt/polecat-nitro/vap-b65l-shelf-confidence
gt/pratham/ds-dws-a59-vercel-deploy
gt/crew-mel/vaa-dd1-whatsapp-gateway
```

## PR Format

**Title:** Same as the primary commit (type + scope + description).

**Body:**
```markdown
## Summary
<1-3 bullet points: what changed and why>

## Bead
<bead-id> — <bead title>

## Changes
- <file>: <what changed>
- <file>: <what changed>

## Testing
- [ ] <how you verified this works>
- [ ] <edge cases tested>

## Screenshots (if UI change)
<before/after if applicable>
```

**Target branch:** Always `dev`. Never push directly to `main`.

## Release Format

**Tag:** Semantic versioning `vMAJOR.MINOR.PATCH`

| Commit type | Bump |
|-------------|------|
| `feat` | MINOR (v1.2.0 → v1.3.0) |
| `fix`, `refactor`, `perf` | PATCH (v1.2.0 → v1.2.1) |
| `BREAKING CHANGE` in body | MAJOR (v1.2.0 → v2.0.0) |
| `chore`, `docs`, `test`, `ci` | No release |

**Release title:** `vX.Y.Z`

**Release notes:**
```markdown
## Highlights
<1-3 sentences: most important changes for users>

## Features
- **<human-readable title>** — what it does, why it matters

## Bug Fixes
- **<human-readable title>** — what was broken, how it's fixed

## Improvements
- <refactors, performance, DX>

---
**X commits** since vPREV
```

**Rules:**
- NO bead IDs in release notes (translate to human language)
- NO agent jargon (polecat, witness, sling, molecule)
- NO "bd: backup" or merge commits
- Write for product users, not agent developers

## Issue Format (Beads)

Created via `bd create --rig <rig> "<title>"`:

**Title:** `<category>: <clear description>`

Categories: `Security`, `UI`, `API`, `Infra`, `Mobile`, `Dashboard`, `Docs`, `Performance`

**Description:**
```
## Context
<What this is about, why it matters>

## Current Behavior
<What happens now>

## Desired Behavior
<What should happen>

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
```

## Wasteland Item Format

Posted via `gt wl post`. Written by witness during patrol (LLM-enriched, not raw bead copy).

**Title:** `<category>: <clear, outsider-readable description>`

Must be understandable by someone with zero context about Gas Town internals.

**Description:**
```
Bead: <bead-id>

## Context
<1-2 sentences: what this project is, what this task is about>

## Repo
<GitHub URL> (masti-ai org)
Branch: dev (create feature branch from dev)

## Current Behavior
<What exists / what's broken / what's missing>

## Desired Behavior
<What it should do when complete>

## Key Files
- `path/to/file.py` — what this file does
- `path/to/component.tsx` — what this component is

## How to Test
<Step-by-step verification>

## Acceptance Criteria
- [ ] Concrete, checkable condition
- [ ] Another condition

## Setup
git clone <repo-url>
cd <repo>
<install commands>
```

**Filter rules** (witness skips these):
- Agent infrastructure (deacon, witness, refinery, patrol, hooks, formulas)
- Town housekeeping (Dolt health, process cleanup, session handoff)
- Titles starting with: "Pre-existing:", "Merge:", "Boot:"

## Wasteland Completion Format

When polecat finishes work:
```bash
gt wl claim <wl-id>
gt wl done <wl-id> --evidence "<PR URL or commit hash>"
```

Evidence must be a link — PR URL, commit URL, or deploy URL. Not "closed locally."

## Changelog Entry Format

```markdown
## YYYY-MM-DD — <short title>

**Type:** decision | deploy | fix | incident | milestone | infra
**Rigs:** <comma-separated rig names> or "town"

<What happened, why, and what changed. 2-5 lines.>
```

## Knowledge Entry Format

```markdown
### <Title> (YYYY-MM-DD)
<What we learned, why it matters, what to do about it.>
Source: <bead ID or incident reference>
```

Types: `pattern` (what works), `anti-pattern` (what breaks), `decision`, `operations`, `product`

## README Format (per-repo)

```markdown
# Project Name

<1-2 line description>

## Quick Start
<Clone, install, run — 3 commands max>

## Architecture
<Brief tech stack overview>

## Development
<Local dev setup, run tests>

## Deployment
<Where deployed, CI/CD workflow>

## API Reference (if applicable)
<Key endpoints>

## Contributing
<Link to wasteland board, PR workflow, branch naming>

## License
```

## Who Enforces What

| Standard | Enforced By | How |
|----------|------------|-----|
| Commit format | Pre-commit hook (bd hooks) | Rejects non-conforming messages |
| Branch format | PreToolUse hook (gt tap guard) | Blocks wrong branch names |
| PR format | Template (.gitea/PULL_REQUEST_TEMPLATE.md) | Pre-filled on PR creation |
| Release format | gitea-to-github.sh + mol-dog-release-notes | Semver auto-tag + LLM enrichment |
| Issue format | AGENTS.md instructions | Agents follow on bd create |
| Wasteland format | Witness patrol | LLM writes rich descriptions |
| Changelog/Knowledge | capture.sh / append.sh | Scripts enforce structure |
