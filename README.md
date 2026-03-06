# Deepwork Base Pack

**The portable configuration pack for Gas Town mesh networks.**

This replaces `gtconfig` — instead of a monolithic config file, Gas Town instances sync knowledge, roles, rules, and skills through this pack.

## What's Inside

```
deepwork-base/
  pack.yaml              # Pack manifest (version, contents list)
  knowledge/             # Shared learnings, patterns, anti-patterns, bug fixes
  roles/                 # Agent role definitions (planner, worker, reviewer)
  rules/                 # Governance rules (branch format, PR policy, etc.)
  skills/                # Claude Code skills for mesh operations
  templates/             # PR body template, mesh.yaml template
```

## How to Use

### For a new Gas Town instance

1. Clone this repo into your GT workspace:
   ```bash
   cd /path/to/your/gt
   git clone git@github.com:Deepwork-AI/deepwork-base.git .packs/deepwork-base
   ```

2. Install the pack (copies knowledge to `.mesh-config/`):
   ```bash
   gt mesh packs install deepwork-base
   ```

3. The knowledge, roles, and rules are now available to all agents in your GT.

### Syncing updates

Pull the latest from upstream:
```bash
cd .packs/deepwork-base && git pull origin main
gt mesh packs install deepwork-base  # Re-install to apply changes
```

### Contributing learnings back

If your GT discovers a new pattern, bug fix, or anti-pattern:

1. Edit the relevant file in `knowledge/`
2. Commit and push to a branch
3. Create a PR to `main`
4. The coordinator (gt-local) reviews and merges

This way, one GT's learning becomes every GT's prevention.

## Knowledge Files

| File | Contains |
|------|----------|
| `shared-knowledge.md` | GitHub labels, branch naming, PR conventions |
| `conventions.md` | Coding standards, commit format, review process |
| `bug-fixes.md` | Known issues and their fixes |
| `anti-patterns.md` | What NOT to do (learned from mistakes) |
| `mail-routing.md` | How mesh mail routing works |
| `rules.md` | Governance rules reference |
| `worker-sla.md` | Worker SLA expectations and enforcement |

## Roles

- **Planner** — Delegates work, reviews PRs, manages roadmap
- **Worker** — Executes code, creates PRs, never merges own PRs
- **Reviewer** — Quality gate, approves/rejects, merges

## How This Replaces gtconfig

Instead of a single YAML config that tries to define everything about a GT instance, this pack provides:

- **Knowledge** that agents can read and learn from (not just config values)
- **Skills** that agents can invoke (not just settings)
- **Rules** that are enforced (not just documented)
- **Templates** that standardize output (PRs, mesh config)

A new GT clones this repo, installs the pack, and immediately has all the shared knowledge of the network. Updates flow through git — pull to get new learnings, PR to contribute back.

## Version

Current: **v2.1.0**

See `pack.yaml` for the full manifest.
