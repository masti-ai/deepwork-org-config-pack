# How-To Guides

Step-by-step guides for common Gas Town operations.

## Available Guides

### Core Operations

- **[Work with Smart Cron](work-with-cron.md)** — Create, manage, and troubleshoot cron jobs with retry strategies and mode-aware scheduling

- **[Work with Deepwork Intelligence (DI)](work-with-di.md)** — Generate structured content (docs, PRs, releases) using the DI MCP server

### Coming Soon

- Work with Molecules — Structured agent workflows with checkpoints
- Work with GT Modes — Switching and configuring runtime modes
- Work with Hooks — Managing Claude Code settings
- Work with Wasteland — Publishing to the public work board
- Work with GT Monitor — Using the API control plane

## Quick Reference

### Smart Cron Quick Start

```bash
# Create a cron job
gt cron create --id my-job --schedule "0 6 * * *" --exec "backup.sh"

# Enable mode-aware scheduling
gt cron config my-job --mode-aware true

# Run manually
gt cron run my-job
```

### DI Quick Start

```bash
# Generate README
gt di generate readme --rig my-rig

# Create PR description
gt di pr --branch feature-branch

# Check available templates
gt di catalog
```

### Molecule Quick Start

```bash
# Attach work
gt mol attach my-work

# Complete step
gt mol step done

# Squash to digest
gt mol squash "Summary of work"
```

### GT Mode Quick Start

```bash
# Check current mode
gt mode get

# Switch to eco mode
gt mode set eco

# Turbo mode for sprints
gt mode set turbo
```

## Need Help?

- Check the [Glossary](../GLOSSARY.md) for terminology
- Review [Patterns](../../knowledge/patterns.md) for proven approaches
- See [Anti-Patterns](../../knowledge/anti-patterns.md) for common mistakes
