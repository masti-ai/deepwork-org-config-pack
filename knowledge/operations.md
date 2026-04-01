# Operations — Debug Shortcuts and Tricks

## Dolt

### Quick health check without gt
```bash
dolt --host 127.0.0.1 --port 3307 --user root --password "" --no-tls sql -q "SELECT 1"
```

### List all databases
```bash
dolt --host 127.0.0.1 --port 3307 --user root --password "" --no-tls sql -q "SHOW DATABASES"
```

### Schema is `status` not `state`
The issues table uses `status` column (open/closed), not `state`. Common mistake.

## Process Management

### Check process budget
```bash
ulimit -u                                    # Current limit
ps -u pratham2 --no-headers | wc -l          # Current usage
ps -u pratham2 --no-headers -o pid,nlwp,args --sort=-nlwp | head -10  # Thread-heavy
ps -u pratham2 --no-headers -o pid,ppid,args | awk '$2==1'            # Orphans
```

### Emergency agent cleanup order
1. Kill all witnesses (auto-respawn, not critical)
2. Kill orphaned bd/gt (ppid=1)
3. Kill Go compile if running in /tmp/gastown-patch
4. Check ulimit -u and raise if possible
5. Retry the operation

## tmux

### Session naming convention
- `hq-*` — HQ agents (mayor, deacon, boot, dog)
- `<rig>-witness` — Per-rig witness
- `<rig>-refinery` — Per-rig refinery
- `<rig>-polecat-<name>` — Named polecats
- `<rig>-crew-<name>` — Crew members

### Peek at an agent
```bash
tmux capture-pane -t <session> -p -S -30
```

## Knowledge System

### Capture a lesson
```bash
bash ~/gt/mayor/knowledge/capture.sh <type> "<title>" "<body>" "<source>"
# Types: pattern, anti-pattern, decision, operations, product
```

### Log a changelog event
```bash
bash ~/gt/mayor/changelog/append.sh <type> "<rigs>" "<title>" "<body>"
# Types: decision, deploy, fix, incident, milestone, infra
```

### Health check
```bash
bash ~/gt/mayor/knowledge/health-check.sh
```
