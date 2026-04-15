# How to Work with Smart Cron

Step-by-step guide for creating and managing cron jobs in Gas Town.

## Quick Start

### 1. Create a Simple Cron Job

```bash
# Edit the town crons configuration
vim ~/gt/.beads/config/crons/my-job.yaml
```

```yaml
[[order]]
id = "my-daily-task"
description = "Run my script every day at 6 AM"
exec = "/path/to/my-script.sh"
gate = "cron"
schedule = "0 6 * * *"
```

### 2. Apply the Configuration

```bash
# Validate the cron configuration
gt cron validate

# Apply to crontab
gt cron apply

# Check status
gt cron status
```

## Understanding Cron Expressions

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6, Sunday = 0)
│ │ │ │ │
* * * * *
```

### Common Patterns

| Schedule | Expression | Description |
|----------|------------|-------------|
| Every minute | `* * * * *` | Run every minute |
| Every 5 minutes | `*/5 * * * *` | Run every 5 minutes |
| Every hour | `0 * * * *` | Run at the top of each hour |
| Every 6 hours | `0 */6 * * *` | Run at 00:00, 06:00, 12:00, 18:00 |
| Daily at 6 AM | `0 6 * * *` | Run once daily at 6:00 AM |
| Weekly on Sunday | `0 4 * * 0` | Run Sundays at 4:00 AM |
| Monthly | `0 3 1 * *` | Run 1st of month at 3:00 AM |

## Smart Cron Features

### Retry Strategies

When a cron job fails, Smart Cron can automatically retry:

```yaml
[[order]]
id = "sync-data"
exec = "sync.sh"
gate = "cron"
schedule = "*/10 * * * *"

# Retry configuration
retry = "exponential"      # Strategy: fixed, exponential, linear
retry_max = 3              # Max retry attempts
```

**Retry Strategies:**

- `fixed` — Wait the same time between retries (e.g., 5min, 5min, 5min)
- `exponential` — Double wait time each retry (e.g., 1min, 2min, 4min)
- `linear` — Add constant increment (e.g., 1min, 2min, 3min)

### Mode-Aware Scheduling

Cron jobs automatically adjust based on GT mode:

```yaml
[[order]]
id = "expensive-task"
exec = "process-large-dataset.sh"
gate = "cron"
schedule = "*/30 * * * *"
mode_aware = true          # Intervals adjust to GT mode
```

**Mode Adjustments:**

| Mode | Multiplier | Example: 30min becomes... |
|------|------------|---------------------------|
| eco | 2.0x | 60 minutes |
| balanced | 1.0x | 30 minutes |
| turbo | 0.5x | 15 minutes |
| maintenance | paused | Disabled (unless allow_manual) |

### Dependency Management

Wait for other jobs to complete:

```yaml
# Job A: Fetch data
[[order]]
id = "fetch-data"
exec = "fetch.sh"
schedule = "0 * * * *"

# Job B: Process data (depends on A)
[[order]]
id = "process-data"
exec = "process.sh"
schedule = "5 * * * *"     # 5 minutes after fetch
depends_on = ["fetch-data"]  # Wait for fetch-data to complete
```

### Cooldown Gates

Prevent jobs from running too frequently:

```yaml
[[order]]
id = "api-call"
exec = "call-api.sh"
gate = "cooldown"
cooldown = "15m"           # Minimum 15 minutes between runs
```

## Practical Examples

### Example 1: Daily Backup with Retry

```yaml
[[order]]
id = "daily-backup"
description = "Backup Dolt databases daily"
exec = "gt dolt backup --all"
gate = "cron"
schedule = "0 2 * * *"     # 2 AM daily
retry = "exponential"
retry_max = 3
mode_aware = true          # Run less frequently in eco mode
```

### Example 2: Wasteland Sync with Dependencies

```yaml
# First: Scan for changes
[[order]]
id = "scan-beads"
exec = "scan-for-wasteland.sh"
schedule = "0 */4 * * *"   # Every 4 hours

# Second: Sync to wasteland (depends on scan)
[[order]]
id = "sync-wasteland"
exec = "wasteland-sync.sh"
schedule = "5 */4 * * *"   # 5 minutes after scan
depends_on = ["scan-beads"]
retry = "fixed"
retry_max = 2
```

### Example 3: Health Check with Alert

```yaml
[[order]]
id = "health-check"
description = "Check system health and alert if issues"
exec = "health-check.sh --alert"
gate = "cron"
schedule = "*/5 * * * *"   # Every 5 minutes
retry = "fixed"
retry_max = 1              # Don't retry alerts
```

### Example 4: Maintenance Task (Maintenance Mode Only)

```yaml
[[order]]
id = "deep-cleanup"
description = "Run only during maintenance windows"
exec = "deep-cleanup.sh"
gate = "cron"
schedule = "0 3 * * 0"     # Sundays at 3 AM
mode_aware = true
# In maintenance mode, this runs
# In other modes, it's skipped
```

## Managing Cron Jobs

### List All Jobs

```bash
# List active cron jobs
gt cron list

# List with status
gt cron list --status

# Show next run times
gt cron list --schedule
```

### Check Job Status

```bash
# Check specific job
gt cron status my-daily-task

# Check job history
gt cron history my-daily-task

# View logs
gt cron logs my-daily-task --tail 50
```

### Disable/Enable Jobs

```bash
# Disable a job temporarily
gt cron disable my-daily-task

# Re-enable
gt cron enable my-daily-task

# Disable all jobs (maintenance)
gt cron disable --all
```

### Run Jobs Manually

```bash
# Trigger a job immediately
gt cron run my-daily-task

# Run with dry-run (don't execute, just show what would happen)
gt cron run my-daily-task --dry-run

# Run with debug output
gt cron run my-daily-task --debug
```

## Troubleshooting

### Job Not Running

```bash
# Check if job is enabled
gt cron status my-job

# Check GT mode (might be paused)
gt mode get

# Check dependencies
gt cron deps my-job

# Verify schedule syntax
gt cron validate my-job
```

### Job Failing Repeatedly

```bash
# View recent failures
gt cron logs my-job --failures

# Check retry settings
gt cron config my-job --show

# Temporarily increase retries
gt cron config my-job --retry-max 5
```

### Mode-Aware Not Working

```bash
# Verify mode_aware is set
gt cron config my-job | grep mode_aware

# Check current GT mode
gt mode get

# Force run regardless of mode
gt cron run my-job --force
```

## Best Practices

### 1. Use Mode-Aware for Resource-Intensive Jobs

```yaml
# Good: Respects eco mode
[[order]]
id = "heavy-processing"
exec = "process.sh"
schedule = "*/30 * * * *"
mode_aware = true

# Bad: Runs too frequently in eco mode
[[order]]
id = "heavy-processing"
exec = "process.sh"
schedule = "*/30 * * * *"
```

### 2. Set Appropriate Retry Limits

```yaml
# Good: Limited retries with backoff
retry = "exponential"
retry_max = 3

# Bad: Infinite retries could overwhelm system
retry = "fixed"
# (missing retry_max = infinite)
```

### 3. Use Dependencies for Related Jobs

```yaml
# Good: Clear dependency chain
job-a → job-b → job-c

# Bad: Race conditions
job-a runs every 10min
job-b runs every 10min (might run before a finishes)
```

### 4. Include Descriptions

```yaml
# Good: Clear description
[[order]]
id = "cleanup-logs"
description = "Remove logs older than 30 days to prevent disk fill"

# Bad: Unclear purpose
[[order]]
id = "cleanup-logs"
```

### 5. Test Before Scheduling

```bash
# Always test your script first
./my-script.sh

# Then test via cron system
gt cron run my-job --dry-run
gt cron run my-job

# Finally, enable the schedule
gt cron enable my-job
```

## Advanced: Custom Gates

Create custom execution gates:

```yaml
[[order]]
id = "conditional-task"
exec = "process.sh"
gate = "custom"
gate_condition = "beads_ready > 5"  # Only run if 5+ beads ready
```

## See Also

- [Patterns: Smart Cron](../../knowledge/patterns.md)
- [Anti-Patterns: Cron Mistakes](../../knowledge/anti-patterns.md)
- [Town Crons Reference](../../crons/town-crons.yaml)
