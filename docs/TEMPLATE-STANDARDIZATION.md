# Template Standardization Guide

How to create, customize, and share templates across your team.

## Template Types

| Type | Purpose | Location |
|------|---------|----------|
| **DI Templates** | Content generation (docs, PRs) | `~/gt/deepwork_intelligence/templates/` |
| **Hook Templates** | Claude Code settings | `~/gt/.claude/templates/` |
| **Mesh Templates** | Town configuration | `~/gt/.config/templates/` |
| **PR Templates** | GitHub PR descriptions | `.github/pull_request_template.md` |
| **Wasteland Templates** | Work package format | `~/gt/.config/wasteland-templates/` |

## DI Template Structure

### Basic Template

```yaml
# my-template.yaml
template:
  name: team-readme
  version: "1.0.0"
  description: "Standard README for team projects"
  
variables:
  - name: ProjectName
    required: true
    description: "Name of the project"
    
  - name: Team
    required: true
    description: "Team name"
    
  - name: Description
    required: true
    description: "Project description"

sections:
  - name: header
    template: |
      # {{ .ProjectName }}
      
      **Team:** {{ .Team }}
      
      {{ .Description }}
  
  - name: installation
    template: |
      ## Installation
      
      ```bash
      git clone {{ .RepoUrl }}
      cd {{ .ProjectName }}
      {{ .InstallCommand }}
      ```
  
  - name: usage
    template: |
      ## Usage
      
      {{ .UsageDescription }}
      
      ```bash
      {{ .UsageExample }}
      ```
  
  - name: team-info
    template: |
      ## Team
      
      - **Lead:** {{ .TeamLead }}
      - **Slack:** {{ .SlackChannel }}
      - **On-call:** {{ .OnCallRotation }}
```

### Template with Conditionals

```yaml
# feature-pr-template.yaml
template:
  name: feature-pr
  
sections:
  - name: summary
    template: |
      ## Summary
      {{ .Summary }}
  
  - name: changes
    template: |
      ## Changes
      {{ range .Changes }}- {{ . }}
      {{ end }}
  
  - name: breaking
    template: |
      {{ if .Breaking }}
      ## ⚠️ Breaking Changes
      {{ .BreakingDescription }}
      
      ### Migration
      {{ .MigrationSteps }}
      {{ end }}
  
  - name: testing
    template: |
      ## Testing
      {{ if .UnitTests }}- [x] Unit tests{{ else }}- [ ] Unit tests{{ end }}
      {{ if .IntegrationTests }}- [x] Integration tests{{ else }}- [ ] Integration tests{{ end }}
      {{ if .ManualTests }}- [x] Manual testing{{ else }}- [ ] Manual testing{{ end }}
```

### Template with Loops

```yaml
# release-notes-template.yaml
template:
  name: release-notes
  
sections:
  - name: version
    template: |
      # Release {{ .Version }}
      
      **Released:** {{ .ReleaseDate }}
  
  - name: highlights
    template: |
      ## ✨ Highlights
      {{ range .Highlights }}
      - {{ . }}
      {{ end }}
  
  - name: features
    template: |
      ## 🚀 New Features
      {{ range .Features }}
      - **{{ .Title }}**: {{ .Description }} ({{ .Author }})
      {{ end }}
  
  - name: bugfixes
    template: |
      {{ if .BugFixes }}
      ## 🐛 Bug Fixes
      {{ range .BugFixes }}
      - {{ .Description }} (#{{ .Issue }})
      {{ end }}
      {{ end }}
  
  - name: contributors
    template: |
      ## 👏 Contributors
      {{ range .Contributors }}@{{ . }} {{ end }}
```

## Standardization Process

### Step 1: Define Standards

Create a `STANDARDS.md` in your team repo:

```markdown
# Template Standards

## PR Templates

### Required Sections
- Summary (1-2 sentences)
- Changes (bullet list)
- Testing (checkboxes)
- Breaking Changes (if applicable)

### Optional Sections
- Screenshots (for UI changes)
- Performance Impact
- Security Considerations

### Tone
- Professional but friendly
- Clear and concise
- No jargon without explanation

## Documentation Templates

### README Structure
1. Title and badges
2. Quick start
3. Installation
4. Usage
5. API reference
6. Contributing
7. License
```

### Step 2: Create Base Templates

```bash
# Create team templates directory
mkdir -p ~/gt/team-templates

# Copy from pack
cp ~/deepwork-org-config-pack/templates/pr-body.md ~/gt/team-templates/

# Customize
vim ~/gt/team-templates/pr-body.md
```

### Step 3: Share with Team

```bash
# Commit to shared repo
cd ~/gt/team-templates
git init
git add .
git commit -m "chore: add team templates"
git remote add origin https://github.com/masti-ai/team-templates.git
git push -u origin main
```

### Step 4: Install Team Templates

```bash
# Each team member runs:
git clone https://github.com/masti-ai/team-templates.git ~/gt/team-templates

# Link to DI
gt di template add ~/gt/team-templates/

# Verify
gt di template list
```

### Step 5: Enforce Usage

Add to your rig's `AGENTS.md`:

```markdown
## Template Usage

All PRs must use the team PR template:
\`\`\`bash
gt di pr --template team-pr
\`\`\`

All READMEs must be generated with:
\`\`\`bash
gt di generate readme --template team-readme
\`\`\`
```

## Wasteland Template Standardization

### Standard Work Package Format

```yaml
# wasteland-standard.yaml
template:
  name: wasteland-standard
  
sections:
  - name: header
    template: |
      ## Context
      {{ .Context }}
      
      ## Repo
      {{ .RepoUrl }}
      
      {{ if .Screenshot }}
      ## Current Behavior
      {{ .CurrentBehavior }}
      {{ end }}
      
      ## Desired Behavior
      {{ .DesiredBehavior }}
  
  - name: technical
    template: |
      {{ if .FilesToModify }}
      ## Files to Modify
      {{ range .FilesToModify }}- `{{ . }}`
      {{ end }}
      {{ end }}
      
      {{ if .TechnicalNotes }}
      ## Technical Notes
      {{ .TechnicalNotes }}
      {{ end }}
  
  - name: acceptance
    template: |
      ## Acceptance Criteria
      {{ range .AcceptanceCriteria }}- [ ] {{ . }}
      {{ end }}
  
  - name: testing
    template: |
      ## How to Test
      {{ .TestingInstructions }}
      
      ## Expected Result
      {{ .ExpectedResult }}
```

### Usage

```bash
# Generate wasteland item
gt di wasteland \
  --bead GTM-123 \
  --template wasteland-standard \
  --output /tmp/wl.md

# Review and post
cat /tmp/wl.md
gt wl post --from-file /tmp/wl.md
```

## Hook Template Standardization

### Base Hook Structure

```json
{
  "hooks": {
    "SessionStart": {
      "actions": [
        {
          "type": "set_context",
          "message": "You are a {{ .Role }} working on {{ .Rig }}"
        },
        {
          "type": "load_knowledge",
          "source": "~/deepwork-org-config-pack/knowledge/"
        }
      ]
    },
    "SessionStop": {
      "actions": [
        {
          "type": "checkpoint",
          "message": "{{ .CheckpointMessage }}"
        },
        {
          "type": "di_capture",
          "template": "session-summary",
          "output": "~/.gt/sessions/{{ .SessionId }}.md"
        }
      ]
    },
    "PreCompact": {
      "actions": [
        {
          "type": "di_generate",
          "template": "work-summary",
          "output": "{{ .MoleculePath }}/digest.md"
        }
      ]
    }
  }
}
```

### Role-Specific Overlays

**Mayor Overlay:**
```json
{
  "hooks": {
    "SessionStart": {
      "actions": [
        {
          "type": "load_overlay",
          "name": "governance"
        },
        {
          "type": "check_escalations"
        }
      ]
    }
  }
}
```

**Polecat Overlay:**
```json
{
  "hooks": {
    "SessionStart": {
      "actions": [
        {
          "type": "load_overlay",
          "name": "worker"
        },
        {
          "type": "attach_molecule",
          "auto": true
        }
      ]
    }
  }
}
```

## Automation

### Auto-Generate on Bead Create

```yaml
# In pack.yaml
automation:
  on_bead_create:
    - action: generate_readme
      template: team-readme
      condition: "bead.type == 'feature'"
    
    - action: post_wasteland
      template: wasteland-standard
      condition: "bead.priority <= 1"
```

### Auto-Update Templates

```bash
# Add to your cron
gt cron create \
  --id template-sync \
  --schedule "0 9 * * 1" \
  --exec "git -C ~/gt/team-templates pull && gt di template refresh"
```

## Quality Checklist

Before publishing a template:

- [ ] All required variables have defaults or are marked required
- [ ] Template renders without errors
- [ ] Output follows team style guide
- [ ] Conditionals handle missing data gracefully
- [ ] Loops work with empty lists
- [ ] Documentation explains how to use it
- [ ] Example inputs/outputs provided

## Migration Guide

### From Old Templates

```bash
# 1. Identify old templates
gt di template list --legacy

# 2. Migrate one by one
gt di template migrate old-template --to new-template

# 3. Test
gt di generate readme --template new-template

# 4. Deprecate old
gt di template deprecate old-template
```

## Examples

### Complete Team Setup

```bash
#!/bin/bash
# setup-team.sh

# Clone team templates
git clone https://github.com/masti-ai/team-templates.git ~/gt/team-templates

# Install base hooks
gt hooks install deepwork-base

# Add team templates
gt di template add ~/gt/team-templates/

# Set defaults
gt config set di.default_pr_template team-pr
gt config set di.default_readme_template team-readme
gt config set wasteland.template wasteland-standard

# Sync hooks
gt hooks sync

echo "Team templates installed!"
```

### Template Versioning

```yaml
# team-readme.yaml
template:
  name: team-readme
  version: "2.1.0"
  changelog:
    - version: "2.1.0"
      changes:
        - "Added security section"
        - "Updated installation steps"
    - version: "2.0.0"
      changes:
        - "Breaking: Renamed variables"
        - "Added performance section"
```

## See Also

- [How to Work with DI](how-to/work-with-di.md)
- [Pack Templates](../templates/)
- [Knowledge: Patterns](../knowledge/patterns.md)
