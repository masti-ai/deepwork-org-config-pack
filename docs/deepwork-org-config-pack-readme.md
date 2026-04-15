# deepwork-org-config-pack

A production configuration pack for Deepwork Labs organizations. This pack provides standardized templates, automation rules, and tooling configurations for consistent team workflows.

## Overview

The `deepwork-org-config-pack` centralizes organization-wide configuration management, enabling:

- **Standardized workflows** across teams via GT Modes
- **Modular configuration** through Molecules
- **Automated scheduling** with Smart Cron
- **Dependency injection** patterns via DI
- **Debugging capabilities** with Debug Mine

## Version

**v4.0.0** — Current stable release

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/deepwork-labs/deepwork-org-config-pack.git

# Navigate to the pack directory
cd deepwork-org-config-pack

# Review the current configuration state
cat config/pack.yaml
```

### Initial Setup

1. Copy the desired configuration files to your project root
2. Update `pack.yaml` with your organization-specific values
3. Run the validation script:

```bash
./scripts/validate-config.sh
```

## Core Features

### GT Modes

Governance and Task modes for project lifecycle management.

| Mode | Purpose |
|------|---------|
| `governance` | Policy enforcement and compliance rules |
| `task` | Work item tracking and execution |
| `debug` | Development and troubleshooting |

### Molecules

Reusable configuration modules that can be composed together.

```
molecules/
├── base/           # Core configuration
├── auth/           # Authentication patterns
├── storage/        # Data persistence
└── network/        # Network and API configuration
```

### Smart Cron

Intelligent scheduling with automatic retry and failure recovery.

```yaml
cron:
  schedule: "0 */6 * * *"  # Every 6 hours
  retry:
    max_attempts: 3
    backoff: exponential
  alerts:
    on_failure: true
    channels: ["#ops-alerts"]
```

### DI (Dependency Injection)

Configuration-driven dependency management for services.

```yaml
services:
  api:
    inject:
      - database
      - cache
      - logger
    config:
      timeout: 5000
      retries: 3
```

### Debug Mine

Integrated debugging toolkit for troubleshooting production issues.

```bash
# Activate debug mode
./scripts/debug-mine.sh --activate

# Run diagnostic suite
./scripts/debug-mine.sh --diagnose --scope=full

# Export debug artifacts
./scripts/debug-mine.sh --export --format=json
```

## Project Structure

```
deepwork-org-config-pack/
├── config/                 # Configuration files
│   ├── pack.yaml          # Main pack manifest
│   └── molecules/         # Modular configurations
├── docs/                   # Documentation
│   ├── guides/            # How-to guides
│   ├── onboarding/        # New member guides
│   └── research/          # Technical research
├── scripts/               # Automation scripts
│   ├── validate-config.sh
│   ├── debug-mine.sh
│   └── sync.sh
├── templates/             # Project templates
└── README.md
```

## Documentation

| Document | Description |
|----------|-------------|
| [Onboarding Guide](docs/onboarding/) | Getting started for new team members |
| [How-To: Smart Cron](docs/guides/smart-cron.md) | Configuring scheduled tasks |
| [How-To: DI Patterns](docs/guides/di-patterns.md) | Dependency injection best practices |
| [Reality Check](docs/reality-check.md) | Distinguishing implemented vs planned features |

## Configuration Reference

### pack.yaml

```yaml
version: "4.0.0"
organization: deepwork-labs

defaults:
  timezone: UTC
  logging:
    level: info
    format: json

modes:
  - gt-modes
  - molecules
  - smart-cron
  - di
  - debug-mine
```

## Maintenance

### Auto-Update Schedule

The pack receives automated updates every 6 hours via scheduled sync jobs.

```bash
# Manual sync
./scripts/sync.sh

# Check sync status
git log --oneline -5
```

### Validation

Before deploying configuration changes:

```bash
# Full validation suite
make validate

# Quick syntax check
make lint
```

## Contributing

1. Create a feature branch from `main`
2. Make changes with corresponding test updates
3. Submit a pull request with documentation updates
4. Ensure validation passes before merge

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for detailed version history.

### v4.0.0 (Latest)

- GT Modes for governance and task workflows
- Molecules for modular configuration
- Smart Cron with intelligent scheduling
- DI patterns for service configuration
- Debug Mine debugging toolkit

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: GitHub Issues
- **Channels**: `#ops-alerts`, `#config-pack`