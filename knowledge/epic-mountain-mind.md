# Epic / Mountain / Mind — Canonical Doctrine

> **Every agent reads this at `gt prime`. This is the law.**

All work in Deepwork flows through three integrated layers. Ignore any legacy
pattern that does not fit this shape — wasteland_publish_epic + manual stamps
is the old way. This is the new way.

| Layer       | Owner                | Holds                                    |
|-------------|----------------------|------------------------------------------|
| **Epic**    | Mayor                | Scope, child beads, dependency graph     |
| **Mountain**| DI (`di.epic_mountain` view) | Wave order, progress, completion gate |
| **Mind**    | DI (per-rig)         | AuditResult stamps, memory, provenance   |

---

## Rule 1 — All work is epic-first

No standalone beads for non-trivial work. If an Overseer instruction produces
code that ships, it is an **Epic**. Polecats never create epics; they implement
beads **under** an epic.

```bash
# Mayor (only actor authorized to open epics):
bd create --type=epic --title="..." --label="scope:<s>" --children="..."
```

A one-off typo fix is the only exception. If you find yourself about to
`bd create` without a parent epic, ask: "who is the Overseer for this?" If
nobody, mail the Mayor — do not open the bead yourself.

---

## Rule 2 — Every epic carries a scope label

`scope:` controls which event hooks fire, who sees the work, and what the
audit threshold is. Missing scope label = malformed epic — Mayor rejects.

### `scope:local`
- One rig, one owner. No cross-org announcement.
- `wasteland_forge` is **skipped**.
- AuditResult still produced on close.
- Push target: `masti-ai/<owner-rig>`.

### `scope:shared`
- Multiple known rigs. Announced to `#shared-work`.
- `wasteland_forge` runs; crew from affected rigs claim.
- Audit coverage threshold: **90%**.
- Push target: primary rig repo or `masti-ai/shared-<name>`.

### `scope:town`
- Town-wide. Freezes possible. Announced to `#town-announcements`.
- `wasteland_forge` runs with `town_broadcast=true` (pings every Witness).
- Audit is two-phase: automated **and** human signoff.
- Push target: `masti-ai/platform`.

---

## Rule 3 — Mountain tracks epic completion

The mountain is a materialized view over the epic's child beads. It has no
`advance` command — closing a child bead advances the mountain implicitly.
Wave order comes from the dotted-child topology (see
`knowledge/rules/epic-pattern.md`).

```
Epic open
  └─ wave 1 ready → crew claims → beads close
       └─ wave 2 ready → crew claims → beads close
            └─ ... → last bead close → Epic close
```

Agents do **not** poll the mountain. Polecats see one bead at a time via their
hook; epic-awareness lives in Mayor + DI, not in the worker.

---

## Rule 4 — `wasteland_forge` fires on EpicCreated (shared/town)

When an epic with `scope:shared` or `scope:town` enters `open`, DI invokes:

```python
wasteland_forge(epic_id)
# - materializes the mountain (wave plan)
# - opens child beads
# - announces to the right channel
```

Local epics skip this hook. This is the only mechanism that opens child
beads on shared/town epics — do not open them manually.

---

## Rule 5 — `wasteland_audit` fires on EpicClosed (always)

When the last child bead closes, DI invokes:

```python
wasteland_audit(epic_id) → AuditResult { quality, risk, coverage, notes }
```

The `AuditResult` is stamped on the epic bead itself. Schema lives at
`agents/shared/schemas.py:AuditResult`. Thresholds per scope live at
`orgs/deepwork/config.yaml → epic.scope_policy`.

**Audit fails → epic re-opens with `needs:rework`, no PR pushed.**

---

## Rule 6 — `github_push_epic` fires on audit-pass

After `AuditResult` passes its scope threshold, DI invokes:

```python
github_push_epic(epic_id, target="masti-ai/<repo>")
# - squashes integration branch into one commit
# - opens PR on masti-ai/<repo>
# - closes the epic bead
```

Polecats **never** open PRs on `masti-ai/*` themselves. The epic lifecycle
does it. If you feel the urge to `gh pr create`, stop — you are bypassing
the audit gate.

---

## Implementation Status

| Piece                | Status             | Notes                                    |
|----------------------|--------------------|------------------------------------------|
| Epic creation        | ✅ Implemented     | `epic_create.py` (Mayor-only)            |
| Mountain (wave plan) | ✅ Implemented     | `mountain_stage.py`, `di.epic_mountain`  |
| Scope labels         | ✅ Convention      | Enforced by Mayor intake                 |
| `wasteland_forge`    | 🚧 Planned         | Event-bus driven; spec in DI rig         |
| `wasteland_audit`    | 🚧 Planned         | `AuditResult` schema landed (ab1cc5cb)   |
| `github_push_epic`   | 🚧 Planned         | Replaces manual PR flow                  |
| Event dispatcher     | 🚧 Planned         | `pipeline_events` + `eventd`             |

Planned pieces ship via the flywheel rewrite. Until they land, Mayor runs
the forge/audit/push steps by hand — the doctrine is identical, only the
automation is missing.

---

## FAQ

**Q: Can a polecat create an epic?**
No. Mayor only. If you think you need an epic, mail the Mayor.

**Q: What if I find a bug while working a bead?**
File a bead (`bd create`), do not fix it in your current branch. If the bug
deserves its own epic, mail the Mayor.

**Q: Can I re-scope an epic?**
Mayor can, **before** the first child bead closes. After that the mountain
is materialized and re-scoping would desync triggers.

**Q: Where do I see the mountain?**
`gt-monitor` → Epics tab, or query `di.epic_mountain`.

---

## See Also

- `knowledge/mind-ops.md` — how the Mind layer routes docs + memory
- `knowledge/rules/epic-pattern.md` — dotted-child topology for mountain
- `agents/shared/schemas.py` — `AuditResult` schema
- `orgs/deepwork/config.yaml` — scope thresholds
