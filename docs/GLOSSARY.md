# Gas Town Glossary

## A

**Agent** — A Claude Code session with a specific role (mayor, deacon, witness, polecat, crew)

**Anti-Pattern** — A common mistake that causes problems, documented for avoidance

## B

**Balanced Mode** — Default GT mode with normal resource usage (5 agents, 30s scheduler)

**Bead** — A work item tracked in the beads system

**Burn (mol)** — Discard a molecule without preserving history

## C

**Cooldown Gate** — Execution gate that enforces minimum time between runs

**Crew** — Persistent worker agent (vs transient polecat)

**Cron Gate** — Execution gate triggered by cron schedule

## D

**Deacon** — Automated patrol agent for town health monitoring

**Debug Mine** — Diagnostic system for capturing context during issues

**DI (Deepwork Intelligence)** — Structured content generation via MCP

**Directive** — Agent work instruction system

**Dog** — Background task runner spawned by deacon

## E

**Eco Mode** — Low resource GT mode (2 agents, 5m intervals)

**Estop** — Emergency stop that freezes all agent work

## F

**Formula** — Reusable workflow template for common tasks

## G

**Gate** — Execution control mechanism (cron, cooldown, manual)

**GT Mode** — Runtime configuration preset (eco, balanced, turbo, maintenance)

**GT Monitor** — API-based control plane for town operations

## H

**Handoff** — Transferring work between agents

**Hook** — Claude Code settings.json configuration

## M

**Maintenance Mode** — Admin-only GT mode (0 auto-agents, estop active)

**Mayor** — Human/agent coordinator with override capabilities

**Mesh** — Cross-rig communication system

**Molecule** — Agent work unit managed via `mol` commands

## O

**Orphan** — Detached commit or branch without reference

**Order** — Scheduled task with retry and dependency management

## P

**Pack** — Config bundle (formulas, roles, rules, knowledge)

**Pattern** — Proven approach that works reliably

**Polecat** — Transient worker agent

## R

**Refinery** — Merge queue processor for serializing merges to main

**Retry Strategy** — How failed jobs are retried (fixed, exponential, linear)

**Rig** — Git repository under Gas Town management

## S

**Scheduler** — Capacity-controlled dispatch system

**Smart Cron** — Cron with retry, backoff, and dependency awareness

**Squash (mol)** — Compress molecule steps to single digest

**Synthesis** — Work bundle for coordinated multi-agent tasks

## T

**Thaw** — Resume from estop

**Turbo Mode** — High performance GT mode (10 agents, 10s intervals)

## W

**Wasteland** — Public work board for external contributors

**Witness** — Per-rig lifecycle agent

**Worktree** — Git worktree for agent sandboxing
