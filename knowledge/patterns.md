# Patterns — What Works

Proven approaches from Gas Town operations.

## Process Management

### Kill non-essential agents before heavy work
When near process limits, kill witnesses for rigs without P0 work before sling/spawn. Deacon respawns them later. Frees 10-20 processes per witness.
Source: de-9s0, 2026-04-01.

### Process cleanup before restart
Before restarting Dolt or daemon, kill orphaned bd/gt processes (ppid=1). They hold stale connections.
```bash
ps -u pratham2 --no-headers -o pid,ppid,args | awk '$2==1 && /bd |gt /' | awk '{print $1}' | xargs kill
```

### Raise ulimit in existing sessions
If limits.d is updated but session is old: `ulimit -Su 4096` raises soft to hard limit without re-login.

## Dolt

### Direct Dolt queries bypass gt/bd hangs
When gt/bd timeout, query Dolt directly:
```bash
dolt --host 127.0.0.1 --port 3307 --user root --password "" --no-tls sql -q "QUERY"
```
Bypasses Go binary overhead and circuit-breaker logic.

### Cross-rig P0 bead query
```sql
SELECT CONCAT(db, '-', id) as bead, title FROM (
  SELECT 'of' as db, id, title FROM officeworld.issues WHERE status='open' AND priority=0
  UNION ALL
  SELECT 'ds' as db, id, title FROM deepwork_site.issues WHERE status='open' AND priority=0
  -- add more rigs as needed
) t ORDER BY db, id
```

## Git & Coordination

### Gitea over GitHub for all agent work
Local Gitea (port 3300) is faster, has no rate limits, and keeps agent API noise off GitHub. GitHub = public mirror only.
Source: GitHub suspension, 2026-03-07.

### Beads need rig-scoped creation for sling
`bd create --rig <rigname> "title"` gives rig-prefixed ID. Town-level (de-) beads can't be slung to rigs.

## Knowledge System

### Three-layer automation for knowledge capture
1. Cron (every 6h) — scans closed beads, extracts lessons
2. Plugin (every 12h) — deacon patrol runs knowledge-evolve
3. Session Stop hook — graceful-handoff.sh logs to changelog
Any one layer alone is sufficient. All three provide redundancy.
