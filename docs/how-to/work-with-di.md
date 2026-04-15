# How to Work with Deepwork Intelligence (DI)

Step-by-step guide for using DI to generate structured content.

## What is DI?

Deepwork Intelligence (DI) is an MCP (Model Context Protocol) server that provides structured content generation for Gas Town operations. It uses LLMs (Claude via MiniMax M2.5 on local H100 GPUs) to generate:

- Documentation (READMEs, arch docs, runbooks)
- PR descriptions and release notes
- Wasteland work packages
- Knowledge base entries

## Quick Start

### 1. Check DI Status

```bash
# Check if DI server is running
gt di status

# View recent invocations
gt di list --recent

# Check catalog of available templates
gt di catalog
```

### 2. Generate Documentation

```bash
# Generate README for a rig
gt di generate readme --rig my-rig

# Generate architecture doc
gt di generate arch --rig my-rig --title "System Architecture"

# Generate runbook
gt di generate runbook --title "Incident Response"
```

### 3. Create PR Description

```bash
# Auto-generate PR description from changes
gt di pr --branch feature-branch --base main

# Include specific context
gt di pr --branch feature-branch --context "Fixes memory leak in worker pool"
```

## DI Commands

### Core Commands

```bash
# Generate documentation
gt di generate <type> [options]

# Create PR description  
gt di pr [options]

# Create release notes
gt di release <version> [options]

# Sync with wasteland
gt di wasteland [options]

# Query catalog
gt di catalog [filter]

# View invocation history
gt di list [options]
```

## Practical Examples

### Example 1: Generate README

```bash
# Basic README generation
gt di generate readme --rig gt-monitor

# With specific focus
gt di generate readme --rig gt-monitor --focus "API endpoints and capabilities"

# Include usage examples
gt di generate readme --rig gt-monitor --include-examples
```

**Output:** Structured README with:
- Overview section
- Installation instructions
- Usage examples
- API reference
- Configuration options

### Example 2: Create PR Description

```bash
# From current branch
gt di pr

# Specify branches
gt di pr --branch feature/auth --base develop

# Include testing notes
gt di pr --include-tests

# Add breaking change notice
gt di pr --breaking "API v1 endpoints removed"
```

**Output:**
```markdown
## Summary
Implements authentication middleware for API endpoints.

## Changes
- Add JWT token validation
- Implement auth middleware
- Add user context to requests
- Update error responses

## Testing
- Unit tests for JWT validation
- Integration tests for protected endpoints
- Manual testing with curl

## Breaking Changes
API v1 endpoints removed. Migrate to v2.

## Checklist
- [x] Tests pass
- [x] Documentation updated
- [x] Breaking changes noted
```

### Example 3: Generate Release Notes

```bash
# Generate from commits since last tag
gt di release v2.1.0

# Include specific commits
gt di release v2.1.0 --since v2.0.0

# Add highlights
gt di release v2.1.0 --highlights "New DI integration, Smart Cron"
```

**Output:**
```markdown
# Release v2.1.0

## Highlights
- New DI integration for structured content
- Smart Cron with retry strategies
- GT Modes for resource management

## New Features
- DI MCP server integration
- GT modes: eco, balanced, turbo, maintenance
- Smart cron with exponential backoff
- Molecule workflow system

## Improvements
- Reduced memory usage in polecat workers
- Faster startup times
- Better error messages

## Bug Fixes
- Fixed race condition in scheduler
- Resolved memory leak in witness

## Migration Guide
See docs/MIGRATION-v2.1.md for upgrade instructions.
```

### Example 4: Wasteland Work Package

```bash
# Create rich wasteland item from bead
gt di wasteland --bead GTM-123

# Include code context
gt di wasteland --bead GTM-123 --include-code

# Auto-post to wasteland
gt di wasteland --bead GTM-123 --post
```

**Output:**
```markdown
## Context
The refinery merge queue processor needs status monitoring.

## Repo
https://github.com/masti-ai/gt-monitor

## Current Behavior
No visibility into refinery processing state.

## Desired Behavior
API endpoint to view queue status, blocked MRs, ready MRs.

## Acceptance Criteria
- [ ] GET /v1/refinery/queue returns queue state
- [ ] Shows blocked MRs with reason
- [ ] Shows ready MRs available for claiming
- [ ] Includes MR metadata (author, title, age)

## Files to Modify
- gt-monitor-server/src/main.rs
- gt-monitor-executor/src/lib.rs
```

### Example 5: Generate Architecture Doc

```bash
# System architecture
gt di generate arch --rig gt-monitor --scope system

# Component architecture
gt di generate arch --component executor

# Data flow
gt di generate arch --type data-flow
```

### Example 6: Create Runbook

```bash
# Incident response runbook
gt di generate runbook --title "Dolt Database Recovery"

# Operational procedure
gt di generate runbook --title "Adding New Rig"

# Troubleshooting guide
gt di generate runbook --title "Agent Stuck Recovery"
```

## Using DI in Formulas

### Formula with DI Step

```toml
[[steps]]
id = "generate-docs"
title = "Generate README with DI"
description = """
Use DI to generate updated README based on changes.
"""
exec = """
gt di generate readme \\
  --rig {{ .Rig }} \\
  --output {{ .OutputDir }}/README.md
gt di generate arch \\
  --rig {{ .Rig }} \\
  --output {{ .OutputDir }}/ARCHITECTURE.md
"""
```

### Formula with PR Description

```toml
[[steps]]
id = "create-pr"
title = "Create PR with DI-generated description"
exec = """
# Generate PR description
gt di pr \\
  --branch {{ .Branch }} \\
  --base main \\
  --output /tmp/pr-body.md

# Create PR using generated description
gh pr create \\
  --title "{{ .Title }}" \\
  --body-file /tmp/pr-body.md
"""
```

## Using DI in Hooks

### SessionStop Hook with DI

```json
{
  "hooks": {
    "SessionStop": {
      "actions": [
        {
          "type": "di_capture",
          "template": "session-summary",
          "output": "~/.gt/sessions/{{ .SessionId }}.md"
        }
      ]
    }
  }
}
```

### PreCompact Hook

```json
{
  "hooks": {
    "PreCompact": {
      "actions": [
        {
          "type": "di_generate",
          "template": "work-summary",
          "context": "{{ .MoleculeSummary }}"
        }
      ]
    }
  }
}
```

## DI Templates

### Available Templates

```bash
# List all templates
gt di catalog

# Filter by type
gt di catalog --type docs
gt di catalog --type pr
gt di catalog --type release

# Show template details
gt di catalog readme --show
```

### Template Types

| Type | Purpose | Output |
|------|---------|--------|
| `readme` | Project README | Markdown |
| `arch` | Architecture docs | Markdown |
| `runbook` | Operational procedures | Markdown |
| `pr` | PR descriptions | Markdown |
| `release` | Release notes | Markdown |
| `wasteland` | Work packages | Markdown |
| `knowledge` | Knowledge base entries | Markdown |

### Custom Templates

Create custom templates in `~/gt/deepwork_intelligence/templates/`:

```yaml
# my-template.yaml
template:
  name: custom-readme
  base: readme
  
sections:
  - overview
  - installation
  - usage
  - api
  - custom_section: |
      ## Custom Section
      
      {{ .CustomContent }}
```

## DI Context Variables

### Available in All Templates

| Variable | Description | Example |
|----------|-------------|---------|
| `{{ .Rig }}` | Current rig name | `gt-monitor` |
| `{{ .TownId }}` | Town identifier | `gt-pratham2` |
| `{{ .Mode }}` | Current GT mode | `balanced` |
| `{{ .DiContext }}` | DI session context | Session metadata |
| `{{ .DiCatalog }}` | Available templates | Template list |
| `{{ .DiHistory }}` | Recent invocations | Last 10 calls |

### Template-Specific Variables

**PR Template:**
- `{{ .Branch }}` — Source branch
- `{{ .Base }}` — Target branch
- `{{ .Commits }}` — Commit list
- `{{ .Files }}` — Changed files
- `{{ .Diff }}` — PR diff (summary)

**Release Template:**
- `{{ .Version }}` — Release version
- `{{ .Since }}` — Previous version
- `{{ .Commits }}` — Commits since last release
- `{{ .Highlights }}` — User-provided highlights

**Wasteland Template:**
- `{{ .BeadId }}` — Bead identifier
- `{{ .BeadTitle }}` — Bead title
- `{{ .BeadDescription }}` — Bead description
- `{{ .Priority }}` — Priority level
- `{{ .AcceptanceCriteria }}` — AC from bead

## DI Configuration

### Global Config

```yaml
# ~/.gt/config.yaml
di:
  enabled: true
  server:
    host: "localhost"
    port: 3000
  defaults:
    model: "anthropic/claude-sonnet-4"
    timeout: "60s"
  templates:
    dir: "~/gt/deepwork_intelligence/templates"
  output:
    format: "markdown"
    save_history: true
```

### Per-Rig Config

```yaml
# ~/gt/rigs/my-rig/.claude/di.yaml
di:
  overrides:
    model: "anthropic/claude-opus-4"
    timeout: "120s"
  templates:
    custom:
      - my-custom-template
  context:
    project_type: "rust"
    team: "platform"
```

## Troubleshooting

### DI Server Not Running

```bash
# Check status
gt di status

# Start DI server
gt di start

# Or check systemd status
systemctl status di-server
```

### Generation Fails

```bash
# Check DI logs
gt di logs --tail 50

# Try with debug
gt di generate readme --debug

# Check context size
gt di generate readme --show-context
```

### Template Not Found

```bash
# List available templates
gt di catalog

# Check template path
gt di config --show-templates-dir

# Verify custom template
gt di catalog my-template --validate
```

### Slow Generation

```bash
# Use faster model
gt di generate readme --model claude-haiku

# Reduce context
gt di generate readme --max-context 4000

# Enable streaming
gt di generate readme --stream
```

## Best Practices

### 1. Always Review DI Output

```bash
# Generate and review before using
gt di pr --output /tmp/pr.md
cat /tmp/pr.md
# Edit if needed
vim /tmp/pr.md
gh pr create --body-file /tmp/pr.md
```

### 2. Provide Context

```bash
# Good: Specific context
gt di pr --context "Fixes critical memory leak in production"

# Bad: No context (generic output)
gt di pr
```

### 3. Use Templates Consistently

```bash
# Create team template
gt di template create --name team-pr --base pr
# Customize for team needs
# Use consistently
gt di pr --template team-pr
```

### 4. Cache Frequently Used Output

```bash
# Generate once, reuse
gt di generate arch --output docs/ARCHITECTURE.md
# Update only when architecture changes
```

### 5. Include in Automation

```toml
# In formulas
[[steps]]
id = "docs"
exec = "gt di generate readme --output README.md"
```

## See Also

- [Patterns: DI Integration](../../knowledge/patterns.md)
- [Pack Config: DI Section](../../pack.yaml)
- [Templates: DI Templates](../../templates/)
