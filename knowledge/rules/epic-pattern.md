# Epic Pattern — Mountain-Compatible Children

**Rule (2026-04-15):** When creating an epic with children, children MUST use dotted-child IDs, not label-linked sibling beads.

## Why

`mountain-eater` walks dotted-child IDs (`epic.N`) to compute wave order. It does **not** follow `epic:<id>` labels or BLOCKS/depends-on edges. Epics with only label-linked children report `no slingable tasks in DAG`.

## Correct pattern

```bash
# 1. Create the epic
bd create "EPIC: My initiative" --type epic --priority P0 --description "..."
# → de-abc1

# 2. Create each child under the epic
bd create "Child task A" --parent de-abc1 --priority P0 --description "..."
# → de-abc1.1

bd create "Child task B" --parent de-abc1 --priority P0 --description "..."
# → de-abc1.2
```

Child IDs come out dotted (`de-abc1.1`, `de-abc1.2`), and `gt mountain stage de-abc1` computes waves correctly.

## Wrong pattern (legacy)

```bash
bd create "EPIC: ..." --type epic    # → de-abc1
bd create "Child A" --label "epic:de-abc1" --deps "blocks:de-abc1"  # → de-xyz9
```

Mountain cannot see `de-xyz9` as a child of `de-abc1`. This is what broke today's 12+ epics.

## Remediation for legacy epics

For each child:

```bash
bd create "Stub→<child-id>: <title>" \
  --parent <epic-id> \
  --type task \
  --priority <match-child> \
  --description "Dotted-child stub for mountain compatibility. Full content in <child-id>." \
  --deps "related:<child-id>"
```

This creates a `<epic-id>.N` stub that points back to the full-content sibling. Mountain sees the stub and walks waves; humans follow the `related:` edge for actual content.

*(Alternative: migrate full content into the dotted-child and close the sibling. Stubs preferred for today's beads to avoid content churn.)*

## Enforcement (future, de-y1y9 follow-up)

A pre-commit `bd` hook should reject `bd create --type epic` followed by children created without `--parent <epic-id>`. Suggested failure message:

```
ERROR: epic de-abc1 has no dotted children. Mountain cannot compute waves.
Use: bd create "<title>" --parent de-abc1 ...
See: knowledge/rules/epic-pattern.md
```

Hook not yet implemented (requires bd extension point); tracked in follow-up.

## Reference

- Working example: `gtm-v8k6` → `gtm-v8k6.3/.4/.5` (mountain computes waves)
- Remediated examples: `de-8415` (13 dotted children), `de-3g4r` (6 dotted children), `de-hbsl` (1 dotted child)
- Bug bead: `de-y1y9`
