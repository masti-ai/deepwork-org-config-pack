# Cleanup Report — di-1qg

**Rule:** `knowledge/*.md` holds KNOWLEDGE (patterns, decisions, anti-patterns, learnings), not TELEMETRY (cost totals, token counts, daily spend tables, per-day stats).

**Scope:** every file in `knowledge/**/*.md` plus the auto-updater scripts under `scripts/knowledge/`.

## One-liner per file

### `knowledge/*.md`

- `account-cycling.md` — clean, no telemetry.
- `anti-patterns.md` — clean, no telemetry (symptom phrasing like "high costs" stays — concept, not number).
- `bug-fixes.md` — clean; `$1` hits are bash positional args.
- `conventions.md` — clean.
- `crew-structure.md` — clean.
- `decisions.md` — clean; mentions of "tokens" / "costs" are architectural reasoning, not metric dumps.
- `formulas-reference.md` — clean.
- `hooks-reference.md` — clean; `gt costs record` is a command reference.
- `mail-routing.md` — clean.
- `offloading.md` — **stripped "Cost Impact" `$200-500/month` / `$0` figures**; replaced with qualitative "Impact" paragraph. Keeps the architectural knowledge (patrol → MiniMax) without the spend number.
- `operations.md` — clean; `$2==1` is an awk field ref.
- `patterns.md` — clean.
- `plugins-reference.md` — clean.
- `products.md` — clean.
- `rules.md` — clean; "tokens" refers to API/auth tokens in a security rule.
- `session-handoffs.md` — clean.
- `shared-knowledge.md` — clean.
- `troubleshooting.md` — clean.
- `worker-sla.md` — clean.

### `knowledge/research/*.md`

- `dashboard-ux-audit-2026-04-15.md` — **stripped inline dashboard telemetry** (`$39.43 · 21 · 1612`, `Budget of $100.00`, `$0.000 used`, `$45.27`, `$10.06`, `678.7k tokens`). The UX lessons — unlabeled numbers, two conflicting cost views, sparkline reconciliation — are preserved without the specific snapshot figures.
- `dashboard-vision-2026-04-15.md` — clean.
- `deepworkmind-audit-2026-04-15.md` — clean; `8k tokens` kept as an API limit / config fact (not usage telemetry).
- `deepworkmind-v2-event-driven-2026-04-15.md` — clean; mentions of tokens are architectural ("event-driven saves tokens vs cron") not counts.
- `gas-town-flywheel-spec-2026-04-15.md` — clean.
- `memory-migration-plan.md` — clean.
- `plug-and-play-product-2026-04-15.md` — **stripped `$1000 AWS credit` dollar figures** (two occurrences); kept the fact that an AWS credit is contingent on delivery.
- `town-governance-epic-first-2026-04-15.md` — clean.
- `trace-extractor-plan.md` — clean; mentions of tokens are security/redaction context.
- `wasteland-reputation-redesign.md` — clean.

## Auto-updater output review

- `scripts/knowledge/evolve.sh` + `cron-evolve.sh` — these scan recently closed beads and append **lessons** (close reasons) to `anti-patterns.md` / `patterns.md`. Not telemetry generators. No changes needed; current appended content in those files is clean.
- `scripts/mayor/log-rotate.sh`, `scripts/mayor/readme-release.sh` — cron-scheduled, touch logs/README only, do not write telemetry into `knowledge/`.

## Net change

Telemetry removed from 3 files. Knowledge preserved in all 29 files.
