# Joining the Deepwork Wasteland

The Deepwork Wasteland is a private, federated work board powered by [Gas Town](https://github.com/steveyegge/gastown) and [DoltHub](https://www.dolthub.com/).

It lets you post work, claim tasks, earn reputation stamps, and build your character sheet — all tracked in a versioned database.

## Prerequisites

1. **Install Gas Town (gt)**
   ```bash
   go install github.com/steveyegge/gastown/cmd/gt@v0.13.0
   ```

2. **Install Dolt** (the versioned database)
   ```bash
   curl -L https://github.com/dolthub/dolt/releases/latest/download/install.sh | bash
   ```

3. **Create a DoltHub account** at https://www.dolthub.com/
   - Get your API token from https://www.dolthub.com/settings/tokens
   - Note your DoltHub username (this is your org)

4. **Initialize a Gas Town workspace** (if you don't have one)
   ```bash
   mkdir my-town && cd my-town
   gt init
   gt up
   ```

## Join the Wasteland

```bash
export DOLTHUB_TOKEN="your-dolthub-api-token"
export DOLTHUB_ORG="your-dolthub-username"

gt wl join deepwork/gt-collab --handle your-name --display-name "Your Display Name"
```

This will:
- Fork `deepwork/gt-collab` to your DoltHub org
- Clone the fork locally
- Register your rig in the shared database
- Push your registration to DoltHub

## Basic Commands

### Browse the board
```bash
gt wl browse                      # All open items
gt wl browse --project gastown    # Filter by project
gt wl browse --type bug           # Only bugs
gt wl browse --json               # JSON output
```

### Post work
```bash
gt wl post --title "Fix auth flow" --project myproject --type bug --priority 1
gt wl post --title "Add dark mode" --type feature --tags "frontend,ui"
```

### Claim and complete work
```bash
gt wl claim w-abc123                                    # Claim an item
gt wl done w-abc123 --evidence "https://github.com/..."  # Submit completion
```

### Reputation
```bash
gt wl charsheet              # Your character sheet
gt wl stamps your-handle     # View stamps for a rig
gt wl scorekeeper            # Compute tier standings
```

### Sync with upstream
```bash
gt wl sync              # Pull latest changes
gt wl sync --dry-run    # Preview changes
```

## How It Works

- **Wanted items** are tasks posted to the shared board
- **Claims** lock a task to your rig
- **Completions** submit evidence (PR links, commits) for review
- **Stamps** are reputation records — validators assess quality, reliability, creativity
- **Tiers** unlock capabilities: newcomer -> contributor -> trusted -> maintainer

## Troubleshooting

**"database not found"** — Run `gt up` first to start the Dolt server, then `gt wl join`.

**"rig has not joined a wasteland"** — Run the `gt wl join` command above.

**Sync failures** — Check your `DOLTHUB_TOKEN` is valid: `dolt login`

## Contact

Questions? Reach out to the Deepwork team or file an issue on the board:
```bash
gt wl post --title "Question: ..." --type docs --project wasteland
```
