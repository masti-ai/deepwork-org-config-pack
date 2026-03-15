# Deepwork Base Pack

**The portable configuration pack for Gas Town mesh networks.**

This replaces `gtconfig` — instead of a monolithic config file, Gas Town instances sync knowledge, roles, rules, and skills through this pack.

---

## 📚 Table of Contents

- [What's Inside](#whats-inside)
- [Getting Started](#getting-started)
- [How to Use](#how-to-use)
- [Knowledge Files](#knowledge-files)
- [Roles](#roles)
- [Rules](#rules)
- [Skills](#skills)
- [Templates](#templates)
- [Contributing](#contributing)
- [Version History](#version-history)

---

## What's Inside

```
deepwork-base/
├── pack.yaml              # Pack manifest (version, contents list)
├── knowledge/             # Shared learnings, patterns, anti-patterns, bug fixes
│   ├── shared-knowledge.md
│   ├── conventions.md
│   ├── bug-fixes.md
│   ├── anti-patterns.md
│   ├── mail-routing.md
│   └── worker-sla.md
├── roles/                 # Agent role definitions (planner, worker, reviewer)
│   ├── coordinator/
│   ├── worker/
│   └── reviewer/
├── rules/                 # Governance rules (branch format, PR policy, etc.)
│   ├── git-workflow.md
│   ├── pr-policy.md
│   └── merge-requirements.md
├── skills/                # Claude Code skills for mesh operations
│   └── mesh/
└── templates/             # PR body template, mesh.yaml template
    ├── pr-template.md
    └── mesh.yaml.template
```

---

## Getting Started

### Installation

#### Option 1: As a Pack (Recommended)

```bash
# From your Gas Town root directory
cd /path/to/your/gt

# Clone this repo into the packs directory
git clone https://github.com/Deepwork-AI/deepwork-base.git .packs/deepwork-base

# Install the pack
gt mesh packs install deepwork-base

# Verify installation
gt mesh packs list
```

#### Option 2: Direct Clone

```bash
# Clone anywhere and reference manually
git clone https://github.com/Deepwork-AI/deepwork-base.git
cd deepwork-base

# Copy files to your GT's .mesh-config/
cp -r knowledge/ roles/ rules/ /path/to/your/gt/.mesh-config/
```

### Initial Setup

After installation, your GT will have access to:

1. **Shared Knowledge** — Best practices, patterns, anti-patterns
2. **Role Definitions** — What coordinators, workers, and reviewers do
3. **Governance Rules** — Branch naming, PR requirements, merge policies
4. **Claude Skills** — Reusable skills for common mesh operations

Verify the pack is active:

```bash
# Check pack status
gt mesh packs status deepwork-base

# View loaded knowledge
gt mesh knowledge list

# View active rules
gt mesh rules list
```

---

## How to Use

### For a New Gas Town Instance

1. **Clone this repo** into your GT workspace:
   ```bash
   cd /path/to/your/gt
   git clone git@github.com:Deepwork-AI/deepwork-base.git .packs/deepwork-base
   ```

2. **Install the pack** (copies knowledge to `.mesh-config/`):
   ```bash
   gt mesh packs install deepwork-base
   ```

3. **The knowledge, roles, and rules are now available** to all agents in your GT.

### Syncing Updates

Pull the latest from upstream:

```bash
# Update the pack
cd .packs/deepwork-base && git pull origin main

# Re-install to apply changes
gt mesh packs install deepwork-base
```

### Contributing Learnings Back

If your GT discovers a new pattern, bug fix, or anti-pattern:

1. **Edit the relevant file** in `knowledge/`
2. **Commit and push** to a branch
3. **Create a PR** to `main`
4. **The coordinator (gt-local) reviews and merges**

This way, one GT's learning becomes every GT's prevention.

---

## Knowledge Files

| File | Contains | When to Read |
|------|----------|--------------|
| `shared-knowledge.md` | GitHub labels, branch naming, PR conventions | Onboarding, before first PR |
| `conventions.md` | Coding standards, commit format, review process | Daily reference |
| `bug-fixes.md` | Known issues and their fixes | When debugging |
| `anti-patterns.md` | What NOT to do (learned from mistakes) | Code review, planning |
| `mail-routing.md` | How mesh mail routing works | Setting up cross-GT mail |
| `worker-sla.md` | Worker SLA expectations and enforcement | Role assignment |
| `rules.md` | Governance rules reference | Dispute resolution |

### Example: Using Knowledge

```bash
# Query knowledge base
gt mesh knowledge search "branch naming"

# Read specific file
gt mesh knowledge read conventions.md

# Add new learning
gt mesh knowledge add "Always use strict TypeScript mode"
```

---

## Roles

Deepwork Base defines three primary behavioral roles:

### Coordinator
- **Creates Tasks** — Identifies work, creates beads, assigns priorities
- **Assigns Work** — Routes beads to appropriate workers
- **Reviews PRs** — Final quality gate before merge
- **Merges** — Can merge to protected branches
- **Writes Code** — ❌ Never writes code directly

### Worker
- **Writes Code** — Implements features, fixes bugs
- **Creates PRs** — Submits work for review
- **Creates Tasks** — ❌ Does not create beads
- **Assigns Work** — ❌ Cannot assign work
- **Merges** — ❌ Never merges own PRs

### Reviewer
- **Reviews PRs** — Quality gate, approves/rejects
- **Merges** — Can merge approved PRs
- **Creates Tasks** — ❌ Does not create work
- **Writes Code** — ❌ Does not implement
- **Assigns Work** — ❌ Does not assign

### Role Configuration

```yaml
# mesh.yaml
behavioral_role:
  this_gt: "coordinator"
  behavior:
    writes_code: false          # HARD BLOCK — config-enforced
    delegates_always: true      # Must send work to workers
  peer_roles:
    gt-worker-1:
      role: "worker"
      specialties: ["backend", "infrastructure"]
    gt-reviewer-1:
      role: "reviewer"
      specialties: ["security", "performance"]
```

---

## Rules

### Git Workflow Rules

| Rule | Description | Enforcement |
|------|-------------|-------------|
| **Branch Naming** | `gt/{node-id}/{issue}-{description}` | Pre-commit hook |
| **Commit Format** | Conventional commits required | CI check |
| **Co-Authored-By** | Required for all AI-generated commits | Pre-commit hook |
| **PR Target** | Must target `dev` branch | Branch protection |
| **No Direct Main** | Cannot push directly to main | Branch protection |

### PR Policy

| Requirement | Details |
|-------------|---------|
| **Review Required** | At least 1 approving review |
| **CI Passing** | All checks must pass |
| **No Conflicts** | Branch must be up-to-date |
| **Linked Issue** | PR should reference a bead |
| **Description** | Must explain what and why |

### Merge Requirements

```yaml
merge_requirements:
  min_approvals: 1
  required_checks:
    - ci/build
    - ci/test
    - ci/lint
  branch_protection:
    main:
      - require_reviews
      - require_up_to_date
      - no_force_push
    dev:
      - require_reviews
```

---

## Skills

Claude Code skills included in this pack:

| Skill | Purpose | Usage |
|-------|---------|-------|
| `mesh-init` | Initialize GT mesh node | `/mesh-init --role worker` |
| `mesh-send` | Send cross-GT message | `/mesh-send gt-1 "Subject" "Body"` |
| `mesh-sync` | Force DoltHub sync | `/mesh-sync` |
| `knowledge-search` | Query knowledge base | `/knowledge-search "pattern"` |

---

## Templates

### PR Template

```markdown
## Summary
Brief description of changes

## Related Bead
Fixes #123

## Type of Change
- [ ] Bug fix
- [ ] Feature
- [ ] Refactor
- [ ] Documentation

## Testing
- [ ] Unit tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows conventions
- [ ] Commit messages are clear
- [ ] Co-Authored-By included
```

### mesh.yaml Template

```yaml
# mesh.yaml - Gas Town Mesh Configuration
mesh:
  name: "your-gt-name"
  role: "worker"  # coordinator | worker | reviewer

behavioral_role:
  this_gt: "worker"
  behavior:
    writes_code: true
    delegates_always: false

  peer_roles:
    gt-coordinator:
      role: "coordinator"

packs:
  installed:
    - deepwork-base

knowledge:
  auto_sync: true
  sync_interval: 2m
```

---

## How This Replaces gtconfig

Instead of a single YAML config that tries to define everything about a GT instance, this pack provides:

| gtconfig (Old) | deepwork-base (New) |
|----------------|---------------------|
| Static config values | Living knowledge that agents read and learn |
| Hardcoded settings | Skills that agents invoke |
| Documentation-only rules | Enforced governance |
| Single file | Organized directory structure |
| Manual updates | Git-based sync and versioning |

A new GT clones this repo, installs the pack, and immediately has all the shared knowledge of the network. Updates flow through git:

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Upstream    │────►│  Your .packs/ │────►│  .mesh-config/│
│  (main)      │ pull │  deepwork-base│install│  (active)    │
└──────────────┘     └──────────────┘     └──────────────┘
       ▲                                          │
       └──────────────────────────────────────────┘
                    (contributions back)
```

---

## Contributing

### Adding Knowledge

1. **Identify the gap** — What did your GT learn that others should know?
2. **Choose the right file** — bug-fixes.md, anti-patterns.md, conventions.md
3. **Write clearly** — Include context, solution, and when to apply
4. **Submit PR** — Target `main`, explain the learning

### Adding Rules

1. **Propose in issues first** — Rules affect all GTs
2. **Get consensus** — Coordinator reviews
3. **Update enforcement** — Add hooks/CI checks
4. **Document** — Update rules.md

### Version Bumps

When making changes:

1. Update `pack.yaml` version
2. Update this README
3. Tag release: `git tag v2.x.x`
4. Push tags: `git push origin --tags`

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| **v2.1.0** | 2026-03-12 | Added coordinator/reviewer blueprints, bash-only scalar agents |
| **v2.0.0** | 2026-03-11 | Renamed packs to blueprints, added org-as-code system |
| **v1.5.0** | 2026-03-10 | Added persistent memory protocol, learning pipeline |
| **v1.0.0** | 2026-03-07 | Initial release with 3-role system |

Current: **v2.1.0**

See `pack.yaml` for the full manifest.

---

<p align="center">
  <sub>Part of <a href="https://github.com/Deepwork-AI">Deepwork AI</a> • Built for <a href="https://github.com/steveyegge/gastown">Gas Town</a></sub>
</p>
