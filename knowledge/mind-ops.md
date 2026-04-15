# Mind Ops — Segmented Memory, Docs Routing, Event Hooks

> **Every agent reads this at `gt prime`. Doctrine, not telemetry.**

The **Mind** is the memory + audit layer of DI. It is segmented: each rig
owns its own mind segment. Cross-rig knowledge lives in this pack; per-rig
knowledge lives in the rig's mind.

---

## Rule 1 — Minds are segmented per rig

There is no global "town mind." Each rig has its own:

| Segment        | Contents                                        |
|----------------|-------------------------------------------------|
| Memory         | Per-rig patterns, decisions, recurring findings |
| Docs           | Per-rig `.md` files under `<rig>/docs/`         |
| AuditResult    | Stamps on epics owned by the rig                |
| Tool Library   | Generated agents in `~/.deepwork-mind/library/` |

**Pack doctrine (this file, this repo) is the only cross-rig knowledge.**
If a pattern is useful to all rigs, it belongs in the pack — not in any
one rig's mind. Promote by PR to `masti-ai/deepwork-org-config-pack`.

**When you write to memory**, it is scoped to your rig. Do not write
cross-cutting facts into a single rig's mind — they will be invisible to
every other rig.

---

## Rule 2 — `docs_*` tools route per-rig

DI's docs tools all take a `rig` argument. Writes land in
`$GT_ROOT/<rig>/<doc_path>`, committed from that rig's worktree.

```python
docs_create(rig, title, doc_type, context)        # new file
docs_append(rig, doc_path, section, content)      # add a section
docs_update(rig, doc_path, section, content)      # rewrite a section
docs_generate(rig, doc_type)                      # auto-generate from state
docs_index(rig)                                   # list what exists
```

**Routing rule:** pass your own rig as `rig` for rig-local docs. Never
write into another rig's docs path — the commit will land in the wrong
worktree and confuse the refinery.

**Pack docs exception:** do NOT use `docs_*` tools to edit files in this
pack. The pack is edited via normal `Write`/`Edit` tooling against
`/home/pratham2/gt/deepwork-org-config-pack/`, then committed and pushed
to `masti-ai/deepwork-org-config-pack`. Using DI to edit DI's doctrine
creates a circular dependency.

---

## Rule 3 — Event hooks drive the mind

The Mind reacts to events emitted by the epic lifecycle. Canonical events
(see the flywheel spec in DI for full catalog):

| Event               | Triggers                                             |
|---------------------|------------------------------------------------------|
| `epic.created`      | `wasteland_forge` (shared/town), `mountain_create`   |
| `bead.closed`       | `mountain_update`, `check_convoy_complete`           |
| `convoy.complete`   | emits `epic.completed`                               |
| `epic.completed`    | `wasteland_audit`                                    |
| `epic.audited`      | `crown_refresh`, `agent_notify`, `github_push_epic`  |

Events are append-only. Handlers are idempotent. If a handler fails, the
dispatcher retries; no manual poke required.

**Agents do not call these handlers directly.** You emit events implicitly
by doing normal work (`bd close`, `bd create --type=epic`, etc.). The event
bus fans out to handlers. Calling `wasteland_forge` by hand is only
appropriate during the pre-automation transitional period, and only from
the Mayor.

---

## Rule 4 — Memory writes survive session death; context does not

When you learn something durable — a pattern, a decision, a non-obvious
constraint — write it to memory. Context window evaporates on session
exit; memory persists across every future session in the rig.

```bash
# Persist findings during work (survives session crash):
bd update <issue-id> --notes "Findings: <what you discovered>"
bd update <issue-id> --design "<structured findings>"
```

**Write early, write often.** The #1 data-loss mode is "polecat analyzed
for 20 minutes, session hit context limit, work lost." Persist before you
close any step.

---

## Rule 5 — Cross-segment reads go through the pack

When rig A needs knowledge from rig B:

1. **If it is pack-level** (applies to any rig): find it in this pack, or
   propose a PR to add it.
2. **If it is rig-specific**: read `gt-monitor` or the other rig's docs
   directly. Do not copy it into your own mind — that fragments truth.
3. **If it is an AuditResult**: query DI (`di.epic_mountain`,
   `di.audit_results`). These are town-readable by design.

Never write rig-B-specific knowledge into rig-A's mind. Segments stay
segmented.

---

## Anti-Patterns

### Writing cross-rig knowledge into one rig's mind
**Symptom:** other rigs re-discover the same fact later.
**Fix:** PR the knowledge into this pack instead.

### Calling event handlers directly
**Symptom:** `wasteland_forge` fires twice; audit state desyncs.
**Fix:** emit the triggering event (close the bead, open the epic) and
let the dispatcher fan out.

### Using `docs_*` to edit the pack
**Symptom:** circular dependency; DI edits DI's own doctrine at runtime.
**Fix:** edit the pack as a normal git repo, commit, push.

### Skipping `bd update --notes` during long analysis
**Symptom:** session dies, 20 minutes of findings lost.
**Fix:** persist after each meaningful conclusion, not at the end.

---

## See Also

- `knowledge/epic-mountain-mind.md` — the three-layer doctrine this serves
- `knowledge/rules.md` — hard rules for every GT instance
- DI rig: `docs/epic-mountain-mind-integration.md` — long-form onboarding
