# Patterns — What Works

Proven approaches for Gas Town operations.

## Molecule Workflow Pattern

**Use for:** Structured agent work with checkpoints

```bash
# 1. Attach molecule to hook
gt mol attach <molecule-id>

# 2. Complete setup step
gt mol step done

# 3. Do the work...

# 4. Complete implementation
gt mol step done

# 5. Squash to digest
gt mol squash "Implement feature X"

# 6. Submit
gt done
```

**Why it works:**
- Clear progress tracking
- Recovery from interruptions
- Audit trail

## Smart Cron Pattern

**Use for:** Reliable scheduled tasks

```yaml
[[order]]
id = "my-task"
schedule = "*/5 * * * *"
retry = "exponential"
retry_max = 3
mode_aware = true
```

**Why it works:**
- Auto-recovery from transient failures
- Respects GT mode settings
- Dependencies respected

## GT Mode Pattern

**Use for:** Resource management

| Mode | Use Case |
|------|----------|
| eco | Low activity periods, save costs |
| balanced | Normal operations |
| turbo | Sprint mode, deadline crunch |
| maintenance | Downtime, upgrades |

```bash
# Switch modes
gt mode set turbo
gt mode set eco
```

## Hook Overlay Pattern

**Use for:** Role-specific agent behavior

```
.claude/
├── settings.json           # Base config
├── settings.json.d/
│   ├── mayor.json          # Mayor overlay
│   ├── deacon.json         # Deacon overlay
│   └── polecat.json        # Polecat overlay
└── agents/
    └── alice.json          # Agent-specific
```

## Wasteland Sync Pattern

**Use for:** Public work board publishing

```bash
# Auto-sync via config
wasteland:
  auto_sync: true
  min_priority: 1
  
# Manual sync
gt wl post --title "..." --type feature --priority 1
```

## DI Integration Pattern

**Use for:** Structured content generation

```bash
# In formulas
exec = "di-generate --type=readme --context={{ .Input }}"

# Or via MCP
docs_generate(rig, 'readme')
```

## Debug Mine Pattern

**Use for:** Deep troubleshooting

```bash
# Enable debug mine
gt debug-mine enable

# Auto-capture on issues
gt debug-mine capture --trigger estop

# Analyze
gt debug-mine analyze <capture-id>
```
