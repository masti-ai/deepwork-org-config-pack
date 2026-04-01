# Anti-Patterns — What Breaks

Learned from real incidents. Every entry cost real time or caused real damage.

## Infrastructure

### Running `next dev` in Docker containers
Next.js dev mode inside Docker causes infinite recompile loops. Planogram dashboard hit 38,000% CPU. Always use production builds (multi-stage: node build → nginx serve).
Source: vap dashboard incident, 2026-03-28.

### Orphaned crons spawning thousands of dolt processes
Crons that spawn dolt subprocesses can accumulate 9000+ zombies if the subprocess hangs (ulimit). Fix: flock guards on all cron scripts.
Source: de-9s0, 2026-04-01.

### Using `go build` instead of `make build` for gt binary
Direct `go build` skips ldflags (BuiltProperly=1, version, commit). Always use `make install`.

### Low ulimit causing Dolt crashes
Dolt (Go) crashes with pthread_create SIGABRT when GC threads can't spawn under low nproc limits. Fix: raise to 16384+ in /etc/security/limits.d/.

## Agent Coordination

### Excessive GitHub API calls from agents
6+ agents hitting GitHub API got account suspended. Use Gitea only. GitHub = public mirror.

### Don't report delegation to user
Wrong: "gt-docker needs to do X". Right: Send via gt mail. Autonomous coordination.

### Don't nudge dead sessions
Check `tmux list-panes -t <session> -F '#{pane_dead}'` first.

### Don't adopt identity from files
Identity comes from `gt prime` and GT_ROLE only.

## Beads & Work

### Slinging town-level beads (de-) to rigs
`gt sling de-xxx <rig>` fails by design. Create with `bd create --rig <rigname>`.

### Treating gastown/beads/mesh as user projects
These are TOOLS. Town-level work uses de- prefix beads.

### Working on closed issues
Mayor's close is final. ALL work stops immediately.

## Git

### Don't use localhost URLs
Always tunnel with cloudflared. Present *.trycloudflare.com URLs.

### `kill -QUIT` on Dolt
SIGQUIT kills the server — does NOT produce goroutine dump. Use `gt dolt status`.

### Don't pull before committing in Dolt
`dolt add . && dolt commit && dolt pull` — in that order.
