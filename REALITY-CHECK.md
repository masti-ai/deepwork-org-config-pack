# Reality Check — What's Real vs Planned

This document clarifies which commands and features exist today vs what's planned for future releases.

## ✅ REAL — These Work Today

### GT Commands (Available Now)
```bash
# Molecule system
gt mol attach <bead>      # Attach molecule to hook
gt mol step done          # Complete current step
gt mol squash [message]   # Compress to digest
gt mol burn               # Discard molecule
gt mol dag [molecule]     # Visualize DAG
gt mol status             # Show hook status
gt mol current            # Show current work
gt mol progress           # Show execution progress

# Scheduler
gt scheduler status       # Show scheduler state
gt scheduler list         # List scheduled beads
gt scheduler run          # Manual dispatch trigger
gt scheduler pause        # Pause dispatch
gt scheduler resume       # Resume dispatch
gt scheduler clear        # Remove beads from scheduler

# Refinery (Merge Queue)
gt refinery status        # Show refinery status
gt refinery queue         # Show merge queue
gt refinery ready         # List ready MRs
gt refinery blocked       # List blocked MRs
gt refinery claim <mr>    # Claim MR for processing
gt refinery release <mr>  # Release MR back to queue
gt refinery start         # Start refinery
gt refinery stop          # Stop refinery
gt refinery restart       # Restart refinery

# Hooks
gt hooks sync             # Regenerate all settings.json
gt hooks diff             # Preview sync changes
gt hooks list             # Show managed locations
gt hooks scan             # Scan for existing hooks
gt hooks registry         # List available hooks
gt hooks install <hook>   # Install hook from registry
gt hooks base             # Edit base config
gt hooks override <role>  # Edit role overrides

# Wasteland
gt wl browse              # Browse wanted items
gt wl post                # Post new wanted item
gt wl show <id>           # Show item details
gt wl claim <id>          # Claim an item
gt wl done <id>           # Submit completion
gt wl sync                # Pull upstream changes

# General
gt sling <bead> <agent>   # Dispatch work to agent
gt ready                  # Show work ready across town
gt bead claim <id>        # Claim open bead
gt done                   # Submit work to refinery
```

### GT Monitor API (Available Now)
```bash
# Endpoints
curl http://localhost:9090/v1/capabilities    # List 64 capabilities
curl http://localhost:9090/v1/health          # Provider health
curl http://localhost:9090/v1/execute         # Execute commands (POST)

# Available capabilities:
# - Refinery, Molecules, Orphans, Synthesis, Directives, Hooks
# - Scheduler, Formulas, DeaconHealth, etc.
```

## 🚧 PLANNED — Coming Soon

These commands are documented in the pack as specifications, but not yet implemented:

### GT Mode System
```bash
gt mode set eco|balanced|turbo|maintenance   # NOT YET IMPLEMENTED
gt mode get                                   # NOT YET IMPLEMENTED
```
**Workaround:** Set mode in `pack.yaml` config, or use GT_MONITOR_MODE env var.

### DI Integration
```bash
gt di generate readme|arch|runbook           # NOT YET IMPLEMENTED
gt di pr --create                            # NOT YET IMPLEMENTED
gt di release <version>                      # NOT YET IMPLEMENTED
gt di wasteland --post                       # NOT YET IMPLEMENTED
gt di catalog                                # NOT YET IMPLEMENTED
```
**Workaround:** Use the DI MCP server directly (configured in `~/gt/.mcp.json`).

### Smart Cron
```bash
gt cron create                               # NOT YET IMPLEMENTED
gt cron apply                                # NOT YET IMPLEMENTED
gt cron validate                             # NOT YET IMPLEMENTED
gt cron status                               # NOT YET IMPLEMENTED
```
**Workaround:** Edit `crons/town-crons.yaml` directly and restart scheduler.

### Debug Mine
```bash
gt debug-mine enable                         # NOT YET IMPLEMENTED
gt debug-mine capture                        # NOT YET IMPLEMENTED
gt debug-mine analyze                        # NOT YET IMPLEMENTED
```
**Workaround:** Manual capture via hooks and logging.

## Configuration vs Commands

Many features work via **configuration files** even if CLI commands don't exist:

| Feature | Config File | CLI Status |
|---------|-------------|------------|
| GT Modes | `pack.yaml` modes: section | Planned |
| Smart Cron | `crons/town-crons.yaml` | Planned |
| DI Templates | `templates/*.yaml` | Planned |
| Debug Mine | `debug-mine.yaml` | Planned |
| Hooks | `hooks/*.json` | ✅ Real |
| Molecules | `mol` commands | ✅ Real |

## Version Compatibility

This pack (v4.0.0) documents features planned for:
- **gt-monitor**: v0.2.0+ (current: v0.1.0)
- **Gas Town CLI**: v2.0.0+ (current: v1.x)

Check your versions:
```bash
gt --version           # Gas Town CLI version
gt-monitor --version   # GT Monitor version (if installed)
```

## Migration Path

As commands are implemented:
1. Configuration files will continue to work
2. New CLI commands will wrap the config
3. Both methods will be supported

Example:
```bash
# Today (config only)
echo 'mode: turbo' >> pack.yaml

# Future (both work)
gt mode set turbo      # CLI command
echo 'mode: turbo' >> pack.yaml  # Still works
```
