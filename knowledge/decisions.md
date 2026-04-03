# Decisions — Why We Chose This

Key architectural and operational decisions. ADR-lite format: context, decision, consequences.

### Gitea as sole git platform for agents (2026-03-07)
**Context:** GitHub suspended freebird-ai due to agent API abuse.
**Decision:** All agent git operations use Gitea (port 3300). GitHub is public mirror only.
**Consequences:** Faster local operations, no rate limits, but no GitHub Actions CI. Agents must be explicitly told not to use `gh` commands.

### Dolt as unified data plane (2026-02)
**Context:** Needed persistent, queryable storage for beads, mail, agent state, and mesh sync.
**Decision:** Single Dolt SQL server on port 3307 with per-rig databases.
**Consequences:** SQL interface for everything, DoltHub sync for mesh, but single point of failure. Dolt crashes take down all of gt/bd.

### Polecats for code, crew for coordination (2026-03)
**Context:** Claude Code is expensive. Needed to separate planning from execution.
**Decision:** Claude Opus handles coordination (mayor, crew). Disposable polecats (potentially cheaper models) handle coding.
**Consequences:** Clean separation of concerns. Polecats can be swapped to cheaper providers. But spawning polecats has overhead (worktree + tmux + claude start).

### Beads over external trackers (2026-02)
**Context:** Considered Linear, GitHub Issues, Jira for work tracking.
**Decision:** Custom beads system in Dolt. Agents use `bd` CLI.
**Consequences:** Fully integrated with agent workflow, no external API calls, survives GitHub outages. But beads is custom software that needs maintenance.

### One town per machine (2026-02)
**Context:** Could have run multiple towns per machine or distributed across machines.
**Decision:** One town = one machine. Cross-machine coordination via DoltHub mesh.
**Consequences:** Simpler resource management. Process limits and GPU allocation are per-town. Mesh adds latency for cross-town work but keeps each town self-contained.

### Tunnel all services for user access (2026-03)
**Context:** Server is remote. User accesses from phone/laptop.
**Decision:** Use cloudflared tunnels for all user-facing services. Never show localhost URLs.
**Consequences:** User gets *.trycloudflare.com URLs that work anywhere. Tunnels are ephemeral (restart needed after reboot). Tailscale as backup for persistent access.

### gt-monitor = data layer, command-center = UI layer (2026-04-02)
gt-monitor collects, stores, and serves rich observability data (tokens, costs, agent health, wasteland metrics, commits, system vitals). Command center becomes the frontend that consumes gt-monitor's API. They merge over time: gt-monitor provides the data, command center renders it. This avoids building two dashboards — dashdev/dashfull polecats should build API endpoints returning JSON, not a standalone UI. The command center (already on port 3100) becomes the single pane of glass.
Source: mayor-decision-2026-04-02.

### Reputation attribution flows to the dispatcher, not the polecat (2026-04-02)
When a polecat completes work, the wasteland reputation (stamps, completions) should be attributed to the CREW MEMBER or AGENT that dispatched it, not the town. Polecats are anonymous execution units — the intelligence and credit belongs to whoever created and instructed them. The rigs table has parent_rig for this. Refinery must use parent_rig when calling gt wl done. Crew members need to be registered as rigs on the wasteland.
Source: mayor-decision-2026-04-02.

### Watchdog auto-detects active rigs — no hardcoded list (2026-04-02)
The witness/refinery watchdog should NOT have a hardcoded ACTIVE_RIGS list. It should detect which rigs are docked (have open beads with assignees, or have active polecats/crew) and only spin up witnesses + refineries for THOSE rigs. Idle rigs get nothing — saves inference tokens. When a rig gets new work (bead created + slung), the watchdog naturally picks it up next cycle. When a rig goes idle (all beads closed), watchdog stops respawning its agents.
Source: mayor-decision-2026-04-02.
