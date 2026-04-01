# Joining the Deepwork Wasteland

The Deepwork Wasteland is a private, federated work board for collaborative development. It connects your Gas Town instance with the Deepwork team via DoltHub (for task coordination) and GitHub (for code).

## Architecture

```
Your Gastown ◄──DoltHub──► Deepwork Gastown
     │                          │
     └──── GitHub (masti-ai) ───┘
            (code, PRs, reviews)
```

- **DoltHub** — Wasteland task board (post work, claim, track reputation)
- **GitHub** — Code lives in the `masti-ai` org (clone, branch, PR)
- **No VPN/tunnel needed** — everything is on public internet

## Prerequisites

1. **Install Gas Town**
   ```bash
   go install github.com/steveyegge/gastown/cmd/gt@latest
   ```

2. **Install Dolt**
   ```bash
   curl -L https://github.com/dolthub/dolt/releases/latest/download/install.sh | bash
   ```

3. **Accounts needed:**
   - **DoltHub** — https://www.dolthub.com/ (get API token from Settings > Tokens)
   - **GitHub** — Ask Pratham to add you to the `masti-ai` org

4. **Initialize Gas Town** (if you don't have one)
   ```bash
   mkdir my-town && cd my-town
   gt init
   gt up
   ```

## Step 1: Join the Wasteland

```bash
export DOLTHUB_TOKEN="your-dolthub-api-token"
export DOLTHUB_ORG="your-dolthub-username"

gt wl join deepwork/gt-collab --handle your-name --display-name "Your Name"
```

This forks the shared database to your DoltHub account and registers you.

## Step 2: Clone Repos from GitHub

All code lives in the `masti-ai` GitHub org:

```bash
# Clone whichever project you're working on
git clone https://github.com/masti-ai/ai-planogram.git
git clone https://github.com/masti-ai/OfficeWorld.git
git clone https://github.com/masti-ai/products.git
# ... etc
```

Available repos: ai-planogram, alc-ai-villa, OfficeWorld, products, media-studio, gt-mesh, website

## Step 3: Browse and Claim Work

```bash
gt wl browse                      # See open tasks
gt wl browse --project gastown    # Filter by project
gt wl show w-abc123               # Full details of a task
gt wl claim w-abc123              # Claim it
```

## Step 4: Do the Work

```bash
cd ai-planogram                   # (or whichever repo)
git checkout -b feat/your-feature
# ... code ...
git push origin feat/your-feature
# Create a PR on GitHub
gh pr create --title "Fix auth flow" --body "Resolves w-abc123"
```

## Step 5: Submit Completion

```bash
gt wl done w-abc123 --evidence "https://github.com/masti-ai/ai-planogram/pull/42"
```

## Step 6: Sync Regularly

```bash
gt wl sync    # Pull latest tasks and updates from upstream
```

## Posting New Work

Anyone in the wasteland can post tasks:

```bash
gt wl post \
  --title "Add dark mode to dashboard" \
  --project ai-planogram \
  --type feature \
  --priority 2 \
  --tags "frontend,ui,react" \
  --description "The planogram dashboard needs dark mode support.
Repo: https://github.com/masti-ai/ai-planogram
Relevant files: crew/manager/dashboard/
Acceptance criteria:
- Toggle in settings
- Persists across sessions
- All charts readable in dark mode"
```

**Always include in descriptions:**
- Repo link
- Relevant file paths
- Acceptance criteria
- Any design docs or specs

## Reputation

```bash
gt wl charsheet              # Your character sheet
gt wl stamps your-handle     # Stamps received
gt wl scorekeeper            # Compute tier standings
```

Tiers: newcomer → contributor → trusted → maintainer

## Projects in the Org

| Project | Repo | Description |
|---------|------|-------------|
| ai-planogram | masti-ai/ai-planogram | ML shelf analysis |
| alc-ai-villa | masti-ai/alc-ai-villa | AI alcohol concierge |
| OfficeWorld | masti-ai/OfficeWorld | GBA-style agent visualizer |
| products | masti-ai/products | Product catalog |
| media-studio | masti-ai/media-studio | Media processing |
| gt-mesh | masti-ai/gt-mesh | Gas Town infrastructure |

## Troubleshooting

**"rig has not joined a wasteland"** — Run the `gt wl join` command from Step 1.

**"database not found"** — Run `gt up` first to start the Dolt server.

**Sync failures** — Check `DOLTHUB_TOKEN` is valid: `dolt login`

**GitHub access** — Ask Pratham (@pratham-bhatnagar) to add you to masti-ai org.

## Contributing Back

Found something useful? Add it to the org knowledge base:
1. Clone the config pack: `git clone https://github.com/masti-ai/deepwork-base.git`
2. Add your learnings to `knowledge/` or update docs
3. Submit a PR

Or post it as a wasteland item:
```bash
gt wl post --title "Learning: ..." --type docs --project wasteland
```
