# {{ .Title }}

## Context
{{ .Context }}

## Changes
{{ range .Changes }}- {{ . }}
{{ end }}

## Testing
{{ .Testing }}

## Checklist
- [ ] Tests pass
- [ ] Documentation updated
- [ ] Breaking changes noted
- [ ] Wasteland sync (if external contributor applicable)

---

**Generated with Deepwork Intelligence** | Mode: {{ .Mode }} | Molecule: {{ .MoleculeId }}
