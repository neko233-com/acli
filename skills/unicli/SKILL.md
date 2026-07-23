---
name: unicli
description: Use acli as an agent-safe, cross-platform CLI for network, process, document, certificate, and code-search tasks.
---

# acli Agent Skill

Run `acli agent` first for full network/system guide. Use structured output (`--json`) when command provides it.

## Code

- Search: `acli code search '<pattern>' <path> --json --glob '*.go'`
- Files: `acli code files <path>`
- `code search` directly invokes `rg`, never a shell. Prefer it over parsing `grep`, `findstr`, or IDE output.

## Universal Tooling

- Inventory: `acli tool doctor --json`
- Run installed executable without shell expansion: `acli tool run git -- status --short`
- Files: `acli file list|read|write|append|copy|move|remove|hash`
- Git: `acli git status|diff|log|add|commit|fetch|pull|push`
- Project templates: `acli dev test|build|format` detects Go, Node, or Python projects.
- `file write` needs `--force` to replace an existing file. `file remove` needs `--yes`; it refuses filesystem root and current directory.

## Documents

- XLSX: `acli excel create|sheets|read|get|set|add-sheet|delete-sheet`
- DOCX: `acli word create|read|append|replace|delete`
- PDF: `acli pdf create|info|merge|delete-pages|keep-pages|watermark|split`
- Mutations require `--out <file>` or explicit `--in-place`. Prefer `--out` for agent workflows.

## Certificates

`acli acme` integrates `neko233-com/acme-go` (published Go module: `github.com/neko233-com/acme233`). Keep DNS credentials in `config_acme.local.json`, never source control.

1. `acli acme validate --config config_acme.json`
2. `acli acme plan --config config_acme.json`
3. `acli acme issue --config config_acme.json --name example`
4. `acli acme renew --config config_acme.json --name example`

Use `providers`, `info`, `revoke`, `install`, `deploy`, and `watch` for lifecycle operations. Certificate issuance/revocation change external state: validate and plan before action.

## SSH and transfers

- Profiles: `acli ssh add/list/remove/export/import`; list with `--json` for agents.
- Verify: `acli ssh check profile`.
- Execute: `acli ssh profile -- command` or `acli ssh exec --group prod -- command`.
- Transfer: `acli scp source profile:/path` and `acli sync source profile:/path --dry-run`.
- SCP transfers are atomic: temporary output is renamed only after full copy. Use `scp --dry-run` before a material transfer.
