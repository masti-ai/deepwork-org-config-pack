# Anti-Patterns — What Breaks

Common mistakes and their consequences.

## Not Using Molecules for Complex Work

**Problem:** Agents lose progress on interruption

**Symptom:** Half-finished beads, repeated work

**Fix:** Always use molecule workflow for multi-step work

```bash
# BAD: Direct work
gt bead show 123
# ... work ...
# (interrupted, no record)

# GOOD: Molecule workflow
gt mol attach 123
gt mol step done  # Setup
gt mol step done  # Implementation
gt mol squash "Done"
```

## Ignoring GT Modes

**Problem:** Resource waste during low-activity periods

**Symptom:** High costs, idle agents

**Fix:** Set mode based on activity level

```bash
# End of day
gt mode set eco

# Sprint start
gt mode set turbo

# Maintenance window
gt mode set maintenance
```

## Hardcoded Schedules Without Mode Awareness

**Problem:** Cron jobs run too frequently in eco mode

**Symptom:** Resource waste, unnecessary work

**Fix:** Use mode_aware: true

```yaml
# BAD: Fixed schedule
schedule = "*/1 * * * *"

# GOOD: Mode-aware
schedule = "*/5 * * * *"
mode_aware = true  # Becomes */10 in eco
```

## Not Using Retry Strategies

**Problem:** Transient failures kill jobs permanently

**Symptom:** Failed crons, manual intervention needed

**Fix:** Configure retry with exponential backoff

```yaml
# BAD: No retry
[[order]]
id = "sync"
exec = "sync.sh"

# GOOD: Retry configured
[[order]]
id = "sync"
exec = "sync.sh"
retry = "exponential"
retry_max = 3
```

## Direct Wasteland Posts Without Review

**Problem:** Poor quality items on public board

**Symptom:** External contributors confused

**Fix:** Use auto_sync with review or manual posting

```yaml
# BAD: Always auto-post
wasteland:
  auto_sync: true
  
# GOOD: Review for quality
wasteland:
  auto_sync: false  # Mayor reviews
  min_priority: 1   # Only P0/P1
```

## Not Handling Hook Failures

**Problem:** Agents run with stale config

**Symptom:** Wrong behavior, errors

**Fix:** Regular hook sync

```bash
# In cron
gt hooks sync

# Or check before work
gt hooks diff
```

## Debug Mine Disabled During Issues

**Problem:** No context when things break

**Symptom:** Difficult troubleshooting

**Fix:** Enable debug mine by default

```yaml
debug_mine:
  enabled: true
  triggers:
    - estop
    - agent_stuck
    - molecule_burn
```
