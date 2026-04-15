# DeepWorkMind MCP Audit + MiniMax 2.7 Capability Harness

**Date:** 2026-04-15
**Bead:** de-hogl
**Scope:** Inventory MCP tools in `deepwork_intelligence/server.py` (2329 lines, 31 tools), map current behavior, assess fit for MiniMax 2.7 long-context (196k) + structured output (json_schema) + reasoning + tool calls. Propose new functions.

---

## Current Configuration

- **Endpoint:** `http://localhost:8080/v1` (sglang, verified up)
- **Model:** `minimax-m2.7` (max_model_len 196608)
- **Schema handling:** `shared/llm.py::_inline_refs` — inlines Pydantic `$defs` because sglang/vLLM grammar compiler can't follow refs (historical bug)
- **Call helpers:** `generate_text(sys, user, max_tokens)` and `generate_structured(sys, user, Model)` — only 3 callers use structured (all GitHub tools)

---

## Tool Inventory (31 tools)

Legend:
- **LLM?** = uses MiniMax at all (via `generate_text` / `generate_structured` / domain agent)
- **Struct?** = already emits a typed Pydantic result
- **LongCtx?** = would materially benefit from 196k context window
- **Reason?** = would benefit from a reasoning/thinking pass
- **ToolCall?** = could be refactored to use native MCP tool-calling loops

| # | Tool | Line | Current behavior | LLM? | Struct? | LongCtx? | Reason? | ToolCall? | Notes |
|---|------|-----:|------------------|:---:|:---:|:---:|:---:|:---:|------|
| 1 | `wasteland_stamp` | 134 | Scores one completion Q/R/C via `score_completion` agent | ✅ | ✅ `StampResult` | — | ✅ | — | Already structured via ADK agent |
| 2 | `wasteland_map_beads` | 213 | Semantic match beads→wasteland items via `map_beads` agent | ✅ | ✅ `MapResult` | ✅ | ✅ | — | 196k lets us match all open beads in one call |
| 3 | `wasteland_complete_matched` | 267 | Auto-complete matched items; generates evidence text (L372) | ✅ | ❌ text only | — | — | — | Promote evidence to structured schema |
| 4 | `wasteland_status` | 454 | Pure Dolt query; no LLM | ❌ | N/A | — | — | — | Deterministic — keep as-is |
| 5 | `wasteland_review_all` | 479 | Aggregates reviews across wasteland | partial | ❌ | ✅ | — | — | Could summarize review set with long ctx |
| 6 | `wasteland_flywheel` | 504 | Orchestrates map→complete→review; composite | ✅ | ❌ | ✅ | ✅ | ✅ | Prime candidate for tool-call loop |
| 7 | `wasteland_cluster_beads` | 630 | LLM clusters open beads into themes (L669) | ✅ | ❌ text | ✅ | ✅ | — | **Upgrade to structured `ClusterResult`** |
| 8 | `wasteland_publish_epic` | 687 | Generates epic description text (L744) | ✅ | ❌ text | — | ✅ | — | Promote to `EpicProposal` schema |
| 9 | `wasteland_check_published` | 785 | Dolt lookup + evidence text (L852) | partial | ❌ | — | — | — | — |
| 10 | `wasteland_import_epic` | 885 | Import wasteland item → bead tree | ❌ | N/A | — | — | — | Deterministic |
| 11 | `docs_create` | 1018 | Generate new doc from context (L1037, 8k tok) | ✅ | ❌ markdown | ✅ | — | — | Long-ctx lets us pass full rig state |
| 12 | `docs_append` | 1066 | Append new section (L1094) | ✅ | ❌ markdown | ✅ | — | — | |
| 13 | `docs_update` | 1110 | Rewrite section (L1137) | ✅ | ❌ markdown | ✅ | — | — | |
| 14 | `docs_generate` | 1153 | Auto-gen doc type from rig state | ✅ | ❌ markdown | ✅ | ✅ | ✅ | Natural tool-call fit (gather rig data → compose) |
| 15 | `town_daily_report` | 1224 | Daily overseer summary from raw data (L1273) | ✅ | ❌ markdown | ✅ | ✅ | — | 196k = entire day's events in one pass |
| 16 | `docs_index` | 1316 | List docs in rig | ❌ | N/A | — | — | — | Deterministic |
| 17 | `feedback_submit` | 1358 | Record Q/R/C correction | ❌ | N/A | — | — | — | Pure write |
| 18 | `feedback_summary` | 1406 | Analyze feedback trends (L1436) | ✅ | ❌ text | ✅ | ✅ | — | Promote to structured trend schema |
| 19 | `feedback_apply` | 1460 | Apply feedback as prompt edits | ✅ | ❌ | ✅ | ✅ | ✅ | Tool-call fit: read→propose→patch→log |
| 20 | `github_create_issue` | 1533 | Gen Gitea issue body (L1559) | ✅ | ✅ `GitIssue` | — | — | — | Already structured |
| 21 | `github_create_pr` | 1625 | Gen Gitea PR body (L1651) | ✅ | ✅ `GitPR` | — | — | — | Already structured |
| 22 | `github_create_release` | 1697 | Gen Gitea release notes (L1734) | ✅ | ✅ `GitRelease` | — | — | — | Already structured |
| 23 | `github_update_readme` | 1779 | Rewrite README section (L1809, 8k tok) | ✅ | ❌ markdown | ✅ | — | — | |
| 24 | `health` | 1837 | Ping check | ❌ | N/A | — | — | — | Deterministic |
| 25 | `analytics_usage` | 1864 | Aggregate tool_calls.jsonl | ❌ | N/A | — | — | — | Deterministic |
| 26 | `analytics_tool_detail` | 1928 | Per-tool stats | ❌ | N/A | — | — | — | Deterministic |
| 27 | `analytics_agent_report` | 1979 | Per-caller stats | ❌ | N/A | — | — | — | Deterministic |
| 28 | `memory_remember` | 2040 | Create memory bead | ❌ | N/A | — | — | — | Pure write |
| 29 | `memory_recall` | 2128 | Keyword/scope search | ❌ | N/A | — | — | — | Deterministic |
| 30 | `memory_forget` | 2235 | Archive stale memory | ❌ | N/A | — | — | — | Pure write |
| 31 | *(module init via `configure_llm`)* | 54 | Config boot | n/a | n/a | — | — | — | |

**Summary:**
- 18/31 tools call MiniMax (58%).
- 5/31 already use structured output (`StampResult`, `MapResult`, 3× Git*).
- 13 LLM tools emit free-form markdown/text that could be upgraded to `response_format: json_schema`.
- 9 tools would materially benefit from 196k context (currently pass truncated state).
- 7 tools are natural tool-call loop fits (multi-step gather→reason→write).

---

## MiniMax 2.7 Capability Gaps vs Current Usage

| Capability | Supported by sglang @ :8080 | Used today | Gap |
|---|:---:|:---:|---|
| Long context 196k | ✅ `max_model_len=196608` | No (callers truncate inputs) | Promote map/cluster/flywheel/docs/town_daily_report to stream full state |
| JSON schema output | ✅ (via grammar; `_inline_refs` already applied) | 5 tools | Upgrade 13 text-emitting tools to typed results |
| Reasoning / thinking tags | ✅ (model emits `<think>`; `_strip_think_tags` drops them) | Stripped (not surfaced) | Route reasoning to a `thoughts` field for auditability on stamp/cluster/epic |
| Tool calling (function calling) | ✅ sglang supports OpenAI tools API | None | Refactor `wasteland_flywheel`, `docs_generate`, `feedback_apply` as tool-loops |
| Streaming | ✅ | Not exposed to MCP callers | Stream `docs_create`/`docs_update` for UX parity with Claude |

---

## Proposed New Functions

### 1. `epic_propose(rig: str, since_days: int = 7) -> EpicProposal`
Bundles `wasteland_cluster_beads` + `wasteland_publish_epic` into one reasoning pass over 196k of open beads, closed-bead summaries, and recent research.

**Schema:**
```python
class EpicProposal(BaseModel):
    title: str
    rationale: str
    bead_ids: list[str]
    estimated_effort: Literal["S","M","L","XL"]
    risks: list[str]
    confidence: float  # 0-1
    thoughts: str  # from reasoning pass
```

**Model mode:** structured + reasoning. Temperature 0.1. Fallback to text parse if grammar fails.

### 2. `pattern_extract(source_path: str, kind: Literal["anti-pattern","learning","architecture"]) -> list[Pattern]`
Feeds closed-bead descriptions, research markdown, or mayor/knowledge/*.md through MiniMax to emit durable knowledge snippets (supersedes the naive grep in de-9lu8's pack-updater).

**Schema:**
```python
class Pattern(BaseModel):
    sha: str       # SHA256 of body for dedup
    kind: str
    title: str
    body: str      # ≤400 chars
    citations: list[str]  # bead IDs, file paths
    confidence: float
```

**Call path:** direct dependency of DOCP updater v2 (de-1u7j). Enforces research path allowlist.

### 3. `rig_health_summary(rig: str, window_hours: int = 24) -> RigHealth`
Composite summary: bead throughput, polecat idle time, witness/refinery status, merge queue depth, escalations, recent failures. Uses long context to ingest raw logs.

**Schema:**
```python
class RigHealth(BaseModel):
    rig: str
    window: str
    score: int  # 0-100
    green: list[str]    # what's working
    yellow: list[str]   # watchlist
    red: list[str]      # needs action
    recommended_actions: list[str]
    evidence: list[str] # log excerpts / bead IDs
```

**Replaces:** ad-hoc `gt rig status` + manual diagnosis in deacon patrols.

### 4. `worker_skill_profile(caller: str, window_days: int = 30) -> SkillProfile`
Mines `tool_calls.jsonl` + closed-bead stamps to build a per-worker reputation snapshot. Feeds the Crown system (hq-cv-a5jkg).

**Schema:**
```python
class SkillProfile(BaseModel):
    caller: str
    total_calls: int
    stamped_completions: int
    avg_quality: float
    avg_reliability: float
    avg_creativity: float
    domains: dict[str, int]    # "wasteland": 42, "docs": 11, ...
    strengths: list[str]       # LLM-derived from pattern
    development_areas: list[str]
    tier: Literal["apprentice","journeyman","master","grandmaster"]
```

**Model mode:** structured output over analytics data (deterministic aggregation feeds LLM for strengths/tier call).

---

## Upgrade Recommendations (ordered by ROI)

1. **Promote `wasteland_cluster_beads` and `wasteland_publish_epic` to structured output** — they're the highest-value free-form emitters and feed epic workflows directly.
2. **Add `thoughts` field to stamp/cluster/epic schemas** — surface the reasoning tags instead of stripping them, for audit trail.
3. **Build `pattern_extract`** — unblocks de-1u7j (DOCP updater v2).
4. **Refactor `wasteland_flywheel` as tool-call loop** — current composite does sequential awaits; model could gate each step on prior evidence.
5. **Stream `docs_create`/`docs_update`** — UX improvement for long generations (currently blocks up to 8k tokens).

---

## Open Questions

- Does sglang 0.x support native OpenAI `tools=[...]` function-calling today, or do we need LiteLLM as adapter? (LiteLLM proxy is running at :4000 with auth — could wrap.)
- Reasoning field: should we persist the raw `<think>` content or the model's own summary of it?
- `pattern_extract` dedup store: Dolt table vs. content-addressed filesystem?

---

*Generated by deacon patrol; source: `/home/pratham2/gt/deepwork_intelligence/server.py` @ 2026-04-15.*
