# GitHub Actions Skill — unicli

## Workflows

| File | Trigger | Purpose |
|------|---------|---------|
| `ci.yml` | push/PR to `main` | Format, vet, build, test, cross-compile matrix |
| `release.yml` | tag `v*` | Same checks + upload binaries + GitHub Release |

## Local pre-push checklist

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./...
```

Format check in CI: `test -z "$(gofmt -l .)"` — no auto-modify in CI.

## Release assets naming

```
unicli-linux-amd64
unicli-linux-arm64
unicli-darwin-amd64
unicli-darwin-arm64
unicli-windows-amd64.exe
```

## Deploy

```cmd
deploy-to-github.cmd
```

Or:

```powershell
.\scripts\deploy.ps1
.\scripts\deploy.ps1 v1.0.5
```

## Debug failed CI

```bash
gh run list --workflow=ci.yml
gh run view <run_id> --log
```

## Common failures

| Error | Fix |
|-------|-----|
| `gofmt -l` non-empty | Run `gofmt -w .` locally |
| `go vet` format %s wrong type | Use `%d`, `%t`, `%v` or correct types |
| IPv6 dial format | Use `net.JoinHostPort(host, port)` |
| Windows `mkdir -p` | Cross-compile from ubuntu only (current setup) |
