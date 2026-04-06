---
name: meteoswisscli-agent-rules
description: Use when contributing to the meteoswisscli Go project with concurrent AI agents — enforces atomic commits, file safety, release notes, and workspace coordination rules
metadata:
  openclaw:
    requires:
      bins:
        - go
    homepage: https://github.com/salam/swissmeteocli
---

# MeteoSwiss CLI Agent Rules

## Overview

Rules for AI agents working concurrently on a Go CLI (`github.com/salam/swissmeteocli`). The core constraint: **multiple agents edit the repo simultaneously**, so every rule exists to prevent destructive interference.

## Commits

**Atomic only.** Commit exactly the files you touched, by path.

Tracked files:
```bash
git commit -m "<scoped message>" -- path/to/file1 path/to/file2
```

New files:
```bash
git restore --staged :/ && git add "path/to/file1" "path/to/file2" && git commit -m "<scoped message>" -- path/to/file1 path/to/file2
```

- Never `git add -A` or `git add .`
- Never commit secrets or `.env` files

## Release Notes

Update `./RELEASE_NOTES.md` for every new feature. Use short, user-facing bullet points.

New section every 4 hours:

```markdown
## Release 1.2 (Mon, Jan 19 09:39)

* Added wind map ASCII rendering [ma]
```

Author abbreviations: `ma` = matthias/salam.

## File Safety

| Action | Allowed? |
|--------|----------|
| Edit files you're working on | Yes |
| Move/rename files | Yes |
| Delete files to fix lint/type errors | **No** — ask the user first |
| Edit `.env` or env var files | **No** — only the user may |
| Revert another agent's edits | **No** — coordinate first |

## Git Restrictions

**Allowed:** Read-only commands (`git log`, `git diff`, `git status`, `git blame`)

**Forbidden:**
- `git stash` (changes workspace state for other agents)
- `git checkout .` / `git restore .` (discards shared work)
- `git reset --hard` (destroys history)
- Any workspace-altering command that interferes with concurrent agents

## Project Layout

| Path | Contents |
|------|----------|
| `cmd/` | CLI entry points |
| `internal/` | Private packages |
| `pkg/` | Public packages |
| `tools/` | Shell scripts (place new scripts here) |
| `test/` | Test fixtures |

## Quick Reference

1. **Before committing:** list only your files explicitly
2. **Before deleting:** ask the user
3. **Before reverting:** coordinate with other agents
4. **Before any git write:** consider concurrent agents
5. **New feature?** Update `RELEASE_NOTES.md`
6. **New script?** Put it in `./tools/`
