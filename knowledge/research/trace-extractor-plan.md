# Trace Extractor for RL/SFT Corpus — Plan

**Status:** research only. No code. No service. Offline function that parses what GT already persists.

**Audience:** mayor + whoever implements the extractor later.

---

## TL;DR

Claude Code already writes **everything** you need in JSONL session logs under
`~/.claude-accounts/<account>/projects/<project-slug>/<session-id>.jsonl`. Each line is
a typed event (`user`, `assistant`, `system`, `hook_success`, `tool_use`, `tool_result`).
Combined with Dolt (beads, wasteland stamps, MR records), those two sources are 95% of
the corpus. tmux logs are redundant. No new instrumentation is required — the extractor
is a pure function: `jsonl + dolt → episodes`.

---

## 1. Data Source Map

For each field the bead spec asks for, this is where it **already lives**:

| Field | Primary source | Path / table | Notes |
|-------|---------------|--------------|-------|
| `prompt` | Claude Code session JSONL | `~/.claude-accounts/<acct>/projects/<slug>/<sid>.jsonl` — events where `type=user` (content → model input) and `type=system` (hook additionalContext injected into the prompt) | Full prompt = system prompt + CLAUDE.md + injected hook context + tool_results + user text. JSONL preserves every layer. |
| `reasoning` | Same JSONL | `type=assistant`, `message.content[].type=text` blocks *between* `tool_use` blocks in the same assistant turn | Already separated from tool calls by the JSONL structure. |
| tool calls + args | Same JSONL | `type=assistant`, `message.content[].type=tool_use` (name, input) | Direct — no parsing needed. |
| tool results | Same JSONL | `type=user`, `message.content[].type=tool_result` | Pairs to tool_use by `tool_use_id`. |
| `state_change` (bead delta) | Dolt — `<rig>.issues`, `<rig>.issue_history` | e.g. `villa_ai_planogram.issues`, `gt_monitor.issues`. Dolt is git-for-data, so `dolt diff` between commits bounding a turn gives the exact bead state delta for that window. | Key: every `bd create/update/close` is one Dolt commit. Use commit author+timestamp to align to turn. |
| `state_change` (rig/file delta) | git (worktree repos) | Each polecat worktree is a git branch. `git log --since=<turn_start> --until=<turn_end>` bounds per-turn file deltas. | Multi-turn commits are fine — attach diff to the turn that ran `git commit`. |
| `code_diff` | git | `git show <sha>` for commits inside the turn window | Optional; only if turn mutated code. |
| `end_result` | Dolt — `wl_commons.completions`, `wl_commons.stamps`, `gastown.merge_requests`, `<rig>.issues.status` | Episode outcome = (bead closed? + Q/R/C scores + MR merged?) | One row per episode; join on bead id + assignee. |
| Session metadata | JSONL first event | `cwd`, `sessionId`, `gitBranch`, `version`, `entrypoint` | Free. |
| Role / rig | Derived | JSONL `cwd` + `GT_ROLE` env snapshot (in the `additionalContext` of the gastown SessionStart hook) | Gastown hook dumps role into context every session start. |

**tmux logs** (`~/gt/<rig>/tmux-client-*.log`): 6–7KB fragments of scrollback. Not useful
for traces — everything of value is already in JSONL. **Skip them.**

**Witness / refinery events**: currently land in `~/gt/logs/*.log` and in Dolt
(`gastown.merge_requests`, `gastown.convoys`). Refinery-mediated outcomes (merged vs.
rejected MR, verification result) live in Dolt; use Dolt as the authoritative signal.

---

## 2. Extraction Algorithm (Pseudocode)

```
# Inputs: session_id, dolt connection, git repo roots
# Output: Episode = { session_meta, turns[], outcome }

def extract_episode(session_id):
    jsonl = load_jsonl_by_session_id(session_id)
    meta  = first_event(jsonl)                    # cwd, gitBranch, user (→ account)
    rig, role = parse_cwd_and_role(meta)

    turns = []
    for each assistant_event in jsonl where type=assistant:
        turn = {
            "t_start":  prev_user_event.timestamp,
            "t_end":    assistant_event.timestamp,
            "prompt":   reconstruct_prompt(jsonl, up_to=assistant_event),
            "reasoning": extract_text_blocks(assistant_event),
            "tool_calls": extract_tool_uses(assistant_event),
            "tool_results": pair_tool_results(jsonl, assistant_event),
        }
        turn["state_change"] = {
            "beads": dolt_diff_in_window(rig, turn.t_start, turn.t_end),
            "files": git_diff_in_window(meta.cwd, turn.t_start, turn.t_end),
        }
        if turn.state_change.files.non_empty():
            turn["code_diff"] = turn.state_change.files.patches
        turns.append(turn)

    outcome = derive_outcome(jsonl, dolt)          # see §3
    return Episode(meta=meta, role=role, rig=rig, turns=turns, outcome=outcome)

def derive_outcome(jsonl, dolt):
    # Look up by bead id referenced in the hook bead (gt prime output) or
    # by polecat assignee = role, session window.
    beads   = dolt.query("SELECT id,status,closed_at,close_reason FROM <rig>.issues "
                        "WHERE assignee=? AND updated_at BETWEEN ? AND ?", role, t0, t1)
    stamps  = dolt.query("SELECT q,r,c,notes FROM wl_commons.stamps "
                        "WHERE agent=? AND created_at BETWEEN ? AND ?", role, t0, t1)
    mrs     = dolt.query("SELECT status,merged_at FROM gastown.merge_requests "
                        "WHERE branch=? AND created_at BETWEEN ? AND ?", branch, t0, t1)
    return combine(beads, stamps, mrs)
```

**Episode boundaries.** An episode is one session (`sessionId`). For polecats that is
one hook bead end-to-end. Multi-session chains (handoffs) can be stitched later via the
shared hook bead id, but v1 treats each session as an episode.

**Turn boundaries.** An assistant event in JSONL *is* a turn. Each assistant event may
contain multiple tool_use blocks; treat them as one turn with multiple actions.

---

## 3. Reward Signal Design

Reward is computed **per episode**, distributed to turns at training time via return
shaping (discounted or uniform).

```
R_episode = w_q · Q/5 + w_r · R/5 + w_c · C/5
          + w_close   · I[bead_closed_legitimately]       ∈ {0,1}
          + w_merge   · I[MR_merged_by_refinery]          ∈ {0,1}
          - w_reject  · I[MR_rejected | bead_reopened]
          - w_zombie  · I[session_died_without_gt_done]
          - w_escal   · I[agent_escalated_HIGH]           (optional; sometimes correct)
```

Default weights (tune empirically): `w_q=w_r=0.2, w_c=0.1, w_close=0.2, w_merge=0.3,
w_reject=0.4, w_zombie=0.3, w_escal=0.0`.

Q/R/C come from `wl_commons.stamps` (already scored by the MiniMax stamp agent).
`legitimate close` means `close_reason` is not `no-changes:*` AND bead was `in_progress`
before close — filters out zombie patrol resets.

**Caveats.**
- Stamps are noisy today (see Part 3). Treat Q/R/C as a weak signal and weight merge
  outcome (a hard fact) higher until stamp quality is fixed.
- Episodes with zero stamps AND zero MR should be discarded from SFT, not assigned zero
  reward — absence of signal ≠ bad work.

---

## 4. Target Training Use

**SFT corpus v1 (first milestone):** filter episodes to `outcome.merge_status=merged`
AND `min(Q,R) >= 4` AND no zombie flags. Emit one JSONL per good episode:

```json
{ "messages": [
    { "role": "system",    "content": "<CLAUDE.md + role + hook>" },
    { "role": "user",      "content": "<prime + bead spec>" },
    { "role": "assistant", "content": "<reasoning + tool_calls>" },
    { "role": "tool",      "content": "<tool_result>" },
    ...
  ],
  "reward": 1.0, "tags": { "rig": "...", "role": "polecat" }
}
```

Target model: Qwen 20B-class open-source. Goal: **native `gt`/`bd` CLI fluency** —
model emits correct `bd update … --status=in_progress` / `gt done` / `gt nudge` calls
without reading the help pages every session. Secondary goal: convergence on short,
non-spammy output (learn from episodes that closed beads fast).

**RL corpus v2 (later):** all episodes, weighted by `R_episode`, trained with
DPO/GRPO against the SFT base. This is the "refineries + witnesses produce training
data" loop in the user's original idea.

---

## 5. Architecture Diagram

```
┌────────────────────────────────────────────────────────────────────────┐
│                        SOURCES (already persisted)                      │
│                                                                          │
│  ~/.claude-accounts/*/projects/*/<session-id>.jsonl                     │
│    └─ prompts, reasoning, tool_use, tool_result (authoritative)         │
│                                                                          │
│  Dolt @ :3307                                                            │
│    ├─ <rig>.issues, <rig>.issue_history    (bead state, deltas)         │
│    ├─ wl_commons.completions, wl_commons.stamps (Q/R/C, evidence)       │
│    └─ gastown.merge_requests, gastown.convoys   (MR outcome)            │
│                                                                          │
│  Git worktrees (~/gt/<rig>/polecats/<name>/)                             │
│    └─ git log / git show (code diffs per turn window)                   │
└────────────────────────────────────────────────────────────────────────┘
                              │ offline read
                              ▼
           ┌───────────────────────────────────────────┐
           │   extract_episode(session_id) — pure fn   │
           │   1. parse JSONL → turns                   │
           │   2. align timestamps → Dolt/git windows   │
           │   3. derive outcome → reward               │
           └───────────────────────────────────────────┘
                              │
                              ▼
           ┌───────────────────────────────────────────┐
           │   CORPUS  (parquet or JSONL on disk)      │
           │   episodes/<rig>/<date>/<session_id>.json │
           └───────────────────────────────────────────┘
                              │
                              ▼
           ┌──────────────────┴──────────────────┐
           │                                       │
           ▼                                       ▼
    SFT pipeline                             RL pipeline
    (filter reward>θ)                        (DPO / GRPO on R)
           │                                       │
           └──────────────── train ───────────────┘
                            Qwen-20B
```

---

## 6. Non-Goals (Reaffirmed)

- ❌ No new agent, no new service, no new MCP tool.
- ❌ No agentic trace collection — every data point is already on disk.
- ❌ No live instrumentation or streaming pipeline. v1 is a batch job.
- ❌ No schema migration — Dolt + JSONL schemas are sufficient as-is.

---

## 7. Open Questions

1. **Account fan-out.** Three Claude accounts (`pratham`, `pratham_cc1`, `pratham_cc3`) each
   hold session JSONLs. Extractor must walk all three to reconstruct multi-account
   sessions (rare but real during quota cycling).
2. **JSONL retention.** How long does Claude keep session logs on disk? Needs a
   one-liner `du -sh` and a rotation policy decision before we can trust it as a corpus.
3. **Tool input redaction.** Some tool inputs contain secrets (env pulls, tokens in
   error text). Extractor must have a redaction pass before corpus hits disk.
4. **Episode stitching across handoffs.** `gt handoff` starts a new session with a
   different sessionId but same hook bead. v1 = separate episodes; v2 = stitch by
   `hook_bead_id` once hook bead id is logged reliably.
