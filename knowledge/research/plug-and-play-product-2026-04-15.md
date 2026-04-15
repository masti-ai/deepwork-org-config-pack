# Plug-and-Play: Gas Town + DeepWorkMind + Wasteland + GT Monitor

**Source:** Overseer directive 2026-04-15 — AWS credit contingent on proving this works on a fresh EC2.

## Vision

One bash command on a vanilla Linux server spins up the entire stack:
- **Gas Town** (mayor + crew + workers + witness + refinery + dogs)
- **DeepWorkMind** (MCP server w/ MiniMax 2.7 local auditor + MCP tools)
- **Wasteland** (shared Dolt-backed audit/task board)
- **GT Monitor** (dashboard + collectors)

No manual config. Works identically on laptop + EC2 + k8s pod (future).

## Success criteria

```bash
curl -fsSL https://masti-ai.github.io/gastown/install.sh | bash
```
…produces a healthy running stack within 15 min, with:
- vLLM or sglang running MiniMax 2.7 on GPU (or fallback to remote OpenAI-compatible)
- Dolt server running (or cloud Dolt)
- GT Monitor dashboard at :3000 serving real data
- DeepWorkMind MCP reachable to any Claude session started from the box
- One sample rig + one sample epic + one sample worker, self-tested

## Part 1 — DeepWorkMind audit + MiniMax 2.7 harness

**Audit task:** inventory all current MCP tools in `deepwork_intelligence/server.py` + what they do + which can be improved with MiniMax 2.7 capabilities (longer context 196k, better structured output, reasoning mode, tool calling).

**New capabilities to harness:**
- **Long-context epic audit**: wasteland_audit can now feed full epic (all beads + diffs + test output) in one shot — was chunked before
- **Structured outputs**: AuditResult schema enforced via JSON schema (MiniMax 2.7 has `response_format: json_schema`)
- **Reasoning mode**: enable for complex audits, disable for fast extraction (cost-control)
- **Tool calls**: MCP tools the auditor can invoke during audit (e.g., `get_bead_history`, `diff_since_merge_base`)

**New functions to add (propose):**
- `epic_propose(user_intent_text)` — Overseer types "do X", mind proposes epic scope + child beads + rig
- `pattern_extract(epic_id)` — after close, extract reusable pattern → feeds pack knowledge
- `rig_health_summary(rig)` — 30s narrative: what's working, what's stuck, what decisions pending
- `worker_skill_profile(worker)` — based on audit trail, what worker is good at

## Part 2 — Generalize everything (product-ready)

All three components must be generalizable via config, not hardcoded:

### Gas Town
- `gastown.toml` at workspace root declares rigs, templates path, model endpoints, mirror targets
- CLI reads config, no hardcoded paths
- Onboarding: `gt init <workspace>` scaffolds config + first rig

### DeepWorkMind
- `deepwork.yaml` config declares: model (minimax-m2.7 default), endpoint (localhost:8080 default), wasteland DB (local or cloud), MCP port
- Self-tests on boot: ping model, ping dolt, register MCP tools
- Add health endpoint for monitor to scrape

### Wasteland
- Schema migration runner (versioned, idempotent)
- Seed data for a new install (one demo epic)
- Cross-org sync pluggable: DoltHub, GitHub Gist, or custom remote

### GT Monitor
- Dashboard reads config for rigs to display + API endpoint (env var)
- Static export works on any host (no hardcoded paths)
- Self-test route: `/health` confirms all data sources reachable

## Part 3 — One-command installer

`install.sh` script:
1. Check prerequisites (docker, git, node, python)
2. Install gastown binary (release download from masti-ai/gastown)
3. Clone deepwork-org-config-pack into `~/gt/.gt-mesh/packs/deepwork-base`
4. Run `gt init` with defaults
5. Pull MiniMax 2.7 weights (or offer remote fallback)
6. Start sglang / vLLM
7. Start Dolt server + run migrations + seed
8. Start DeepWorkMind MCP
9. Build + start GT Monitor dashboard
10. Self-test: ping everything, create sample epic, close sample bead
11. Print: "Dashboard: http://localhost:3000 — MCP: stdio via Claude Desktop config"

Deliver as: public `masti-ai/gastown` repo release + `install.sh` on GitHub Pages.

## Part 4 — AWS EC2 proof test

- Spin a fresh t3.2xlarge (or g5 for GPU) via Terraform
- Run install.sh
- Assert dashboard loads, MCP responds, sample epic audits, PR pushes to GitHub
- Tear down
- One-pager report with screenshots + logs

AWS credit from Overseer available for this test phase.

## Deliverables
1. DeepWorkMind audit doc (`research/deepworkmind-audit-2026-04-15.md`)
2. MiniMax 2.7 capability harness: new MCP tools + AuditResult JSON schema
3. Config schemas (gastown.toml, deepwork.yaml) + loaders
4. install.sh + smoke-test suite
5. Terraform for EC2 test + AWS test run report
