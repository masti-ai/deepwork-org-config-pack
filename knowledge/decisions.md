# Decisions — Why We Chose This

Key architectural decisions with context. ADR-lite format.

### Gitea as sole git platform for agents (2026-03-07)
**Context:** GitHub suspended freebird-ai due to agent API abuse.
**Decision:** All agent git operations use Gitea (port 3300). GitHub is public mirror only.
**Consequences:** Faster local operations, no rate limits, but no GitHub Actions CI.

### Dolt as unified data plane (2026-02)
**Context:** Needed persistent, queryable storage for beads, mail, agent state.
**Decision:** Single Dolt SQL server on port 3307 with per-rig databases.
**Consequences:** SQL for everything, DoltHub sync for mesh, but single point of failure.

### Polecats for code, crew for coordination (2026-03)
**Context:** Claude Code is expensive. Separate planning from execution.
**Decision:** Claude Opus handles coordination (mayor, crew). Disposable polecats handle coding.
**Consequences:** Clean separation. But spawning polecats has overhead (worktree + tmux + claude start).

### Beads over external trackers (2026-02)
**Context:** Considered Linear, GitHub Issues, Jira.
**Decision:** Custom beads system in Dolt. Agents use `bd` CLI.
**Consequences:** Fully integrated, no external API calls, survives GitHub outages. Custom software needs maintenance.

### Tunnel all services for user access (2026-03)
**Context:** Server is remote. User accesses from phone/laptop.
**Decision:** Use cloudflared tunnels. Never show localhost URLs.
**Consequences:** *.trycloudflare.com URLs work anywhere. Tunnels are ephemeral.

### Self-evolving knowledge over static docs (2026-04-01)
**Context:** Agents kept rediscovering the same problems.
**Decision:** Three-layer knowledge system: cron + plugin + handoff hook auto-capture lessons from closed beads.
**Consequences:** Knowledge grows from operational experience. Requires pruning stale entries.
