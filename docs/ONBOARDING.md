# Onboarding Guide — New Team Member Setup

Welcome! This guide will get you productive with Gas Town, Deepwork Intelligence, GitHub, and Wasteland in under 30 minutes.

## What You Get

The **deepwork-org-config-pack** gives you:

- ✅ Pre-configured GT Modes (eco/balanced/turbo/maintenance)
- ✅ Smart Cron templates with retry strategies
- ✅ DI integration for automated docs/PRs
- ✅ Standardized hooks and molecules
- ✅ Wasteland publishing workflows
- ✅ Role-based configurations (mayor, deacon, polecat, crew)

## Prerequisites

Before starting, ensure you have:

1. **GitHub Access** — Member of `masti-ai` organization
2. **Wasteland Access** — Account on wasteland board
3. **Gas Town Installed** — `gt` CLI on your machine
4. **DI Access** — MCP server configured (for content generation)

## 5-Minute Quick Start

### Step 1: Clone the Config Pack

```bash
# Clone the pack
git clone https://github.com/masti-ai/deepwork-org-config-pack.git ~/deepwork-org-config-pack

# Link to your GT config
ln -s ~/deepwork-org-config-pack ~/gt/.config/packs/deepwork-org
```

### Step 2: Configure Your Environment

```bash
# Copy the mesh template
cp ~/deepwork-org-config-pack/templates/mesh.yaml.template ~/.config/gt/mesh.yaml

# Edit with your details
vim ~/.config/gt/mesh.yaml
```

Fill in:
- `TownId` — Your town identifier
- `Rigs` — List of rigs you'll work with
- `GithubRepo` — Your GitHub repository URLs
- `Mode` — Default GT mode (start with `balanced`)

### Step 3: Initialize GT

```bash
# Start Gas Town
gt start

# Set your mode
gt mode set balanced

# Verify everything works
gt status
gt di status
```

### Step 4: Configure Hooks

```bash
# Install standard hooks
gt hooks install deepwork-base

# Sync to all your rigs
gt hooks sync

# Verify
gt hooks list
```

### Step 5: Test Wasteland Connection

```bash
# Check wasteland status
gt wl status

# Test posting (dry run)
gt wl post --title "Test: Onboarding Complete" --type task --priority 3 --dry-run
```

## Your First Hour

### 1. Pick Your First Bead (15 min)

```bash
# See what's available
gt ready

# Or check the wasteland
gt wl browse

# Claim a bead
gt bead claim GTM-xxx
```

### 2. Start Working with Molecules (20 min)

```bash
# Attach the molecule
gt mol attach GTM-xxx

# Complete setup step
gt mol step done

# Do your work...
# (edit code, write tests, etc.)

# Complete implementation
gt mol step done

# Squash your work
gt mol squash "Implement feature: xxx"

# Submit
gt done
```

### 3. Create Your First PR with DI (15 min)

```bash
# Push your branch
git push origin feature-branch

# Generate PR description with DI
gt di pr --branch feature-branch --base main

# Or create directly
gt di pr --branch feature-branch --base main --create
```

### 4. Publish to Wasteland (10 min)

```bash
# If your work is suitable for external contributors
gt di wasteland --bead GTM-xxx --post

# Or manual post with rich context
gt wl post \
  --title "Feature: Add user authentication" \
  --type feature \
  --priority 1 \
  --description "## Context..." \
  --acceptance "- [ ] Criteria 1"
```

## Standardizing Templates

### Create Team Standards

As a team, agree on:

1. **GT Mode Schedule**
   - Default: `balanced`
   - After hours: `eco`
   - Sprints: `turbo`
   - Maintenance: `maintenance`

2. **DI Templates**
   - PR template style
   - README sections
   - Release note format

3. **Wasteland Criteria**
   - What goes to wasteland (P0/P1 only?)
   - Required acceptance criteria
   - Review process

### Customize Templates

```bash
# Copy templates to your rig
cp ~/deepwork-org-config-pack/templates/*.template ~/gt/my-rig/.templates/

# Customize for your team
vim ~/gt/my-rig/.templates/pr-body.md
```

Example team PR template:

```markdown
## {{ .Title }}

### Changes
{{ range .Changes }}- {{ . }}
{{ end }}

### Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

### Screenshots
{{ .Screenshots }}

### Related
- Bead: {{ .BeadId }}
- Wasteland: {{ .WastelandId }}

---
**Team:** {{ .Team }} | **Sprint:** {{ .Sprint }}
```

### Share Templates

```bash
# Commit to your rig
git add .templates/
git commit -m "chore: add team templates"
git push

# Others can pull
git pull origin main
```

## Wasteland Workflow

### Automatic Publishing

Configure auto-sync for your beads:

```yaml
# In your rig config
wasteland:
  auto_sync: true
  min_priority: 1      # Only P0/P1
  include_acceptance: true
  github_repo: "https://github.com/masti-ai/your-repo"
```

### Manual Publishing

For beads that need review first:

```bash
# Create rich wasteland item
gt di wasteland --bead GTM-xxx --output /tmp/wl-item.md

# Review and edit
vim /tmp/wl-item.md

# Post when ready
gt wl post --from-file /tmp/wl-item.md
```

### Sync Status

```bash
# Check what's synced
gt wl status

# See unsynced beads
gt wl pending

# Force sync
gt wl sync --force
```

## GitHub Integration

### Repository Setup

Ensure your repos are configured:

```bash
# Check GitHub remotes
gt rig list --remotes

# Add GitHub remote if missing
gt rig add-remote --github masti-ai/your-repo

# Verify
git remote -v
```

### PR Automation

```bash
# Generate and create PR in one command
gt di pr \
  --branch feature-branch \
  --base main \
  --title "feat: new feature" \
  --create

# With custom context
gt di pr \
  --branch feature-branch \
  --context "This fixes the memory leak reported in #123" \
  --create
```

### Release Management

```bash
# Generate release notes
gt di release v1.2.0 --since v1.1.0

# Create GitHub release
gt di release v1.2.0 --create --draft
```

## Role-Based Setup

### If You're a Developer (Polecat/Crew)

```bash
# Install polecat hooks
gt hooks install polecat-base

# Use molecule workflow for all work
gt mol attach <bead>
# ... do work ...
gt mol squash "Summary"
gt done
```

### If You're a Team Lead (Mayor)

```bash
# Install mayor hooks
gt hooks install mayor-base

# Use governance overlays
gt hooks overlay governance

# Review wasteland items before publishing
gt wl review --pending

# Approve for publishing
gt wl approve <item-id>
```

### If You're DevOps (Deacon)

```bash
# Install deacon hooks
gt hooks install deacon-base

# Configure patrol tasks
gt deacon config --from ~/deepwork-org-config-pack/roles/deacon.yaml

# Monitor health
gt deacon status
gt deacon patrol
```

## Common Workflows

### Workflow 1: New Feature

```bash
# 1. Get bead
gt ready
gt bead claim GTM-123

# 2. Work
gt mol attach GTM-123
gt mol step done  # Setup
# ... code ...
gt mol step done  # Implementation
gt mol squash "Add feature X"

# 3. Submit
gt done

# 4. Create PR
gt di pr --create

# 5. Optionally publish to wasteland
gt di wasteland --post
```

### Workflow 2: Bug Fix

```bash
# 1. Claim bug bead
gt bead claim GTM-456

# 2. Work
gt mol attach GTM-456
gt mol step done
# ... fix ...
gt mol squash "Fix: issue description"

# 3. Submit
gt done

# 4. PR with context
gt di pr --context "Fixes crash in production"
```

### Workflow 3: Documentation

```bash
# 1. Generate with DI
gt di generate readme --rig my-rig

# 2. Review and edit
vim README.md

# 3. Commit
git add README.md
git commit -m "docs: update README"

# 4. PR
gt di pr --create
```

### Workflow 4: Wasteland Item

```bash
# 1. Find bead suitable for external
gt bead show GTM-789

# 2. Generate rich description
gt di wasteland --bead GTM-789 --output /tmp/wl.md

# 3. Review
cat /tmp/wl.md

# 4. Post
gt wl post --from-file /tmp/wl.md

# 5. Track
gt wl status
```

## Troubleshooting

### "gt command not found"

```bash
# Add to your shell profile
echo 'export PATH="$HOME/gt/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### "DI server not running"

```bash
# Start DI server
gt di start

# Or check if it's configured
cat ~/.config/gt/mesh.yaml | grep di
```

### "Wasteland auth failed"

```bash
# Check wasteland config
gt config get wasteland.token

# Or re-authenticate
gt wl login
```

### "Hooks not syncing"

```bash
# Force sync
gt hooks sync --force

# Check permissions
ls -la ~/gt/.claude/
```

## Next Steps

### Week 1: Get Comfortable

- [ ] Complete 3 beads using molecule workflow
- [ ] Create 2 PRs using DI
- [ ] Post 1 item to wasteland
- [ ] Configure your preferred GT mode

### Week 2: Optimize

- [ ] Customize PR template for your team
- [ ] Set up personal cron jobs
- [ ] Create a runbook for your service
- [ ] Help onboard another team member

### Week 3: Contribute

- [ ] Add a pattern to the knowledge base
- [ ] Create a new formula
- [ ] Improve a template
- [ ] Submit pack update

## Getting Help

- **Quick questions:** `gt help <command>`
- **Knowledge base:** `~/gt/mayor/knowledge/`
- **This pack:** `~/deepwork-org-config-pack/`
- **Team chat:** #gastown-help
- **Issues:** GitHub issues on your rig repo

## Checklist

Before you start working:

- [ ] Cloned deepwork-org-config-pack
- [ ] Configured mesh.yaml
- [ ] GT status shows healthy
- [ ] Hooks installed and synced
- [ ] Wasteland connection tested
- [ ] DI status shows ready
- [ ] First bead claimed and completed

You're ready! Start with `gt ready` to see what's available.
