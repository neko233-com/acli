# acli — Agent Guide

## Agent Ops Additions

| User intent | Preferred command |
|---|---|
| Fast code search | `acli code search '<pattern>' . --json` |
| Enumerate code files | `acli code files .` |
| XLSX CRUD | `acli excel --help` |
| DOCX CRUD | `acli word --help` |
| PDF page CRUD | `acli pdf --help` |
| ACME certificate lifecycle | `acli acme validate/plan/issue/renew` |
| Tool inventory | `acli tool doctor --json` |
| Git operations | `acli git status/diff/log/add/commit` |
| Project test/build/format | `acli dev test/build/format` |
| File CRUD | `acli file list/read/write/append/copy/move/remove/hash` |
| SSH validation | `acli ssh check profile` |

Document mutations require `--out` or `--in-place`; prefer `--out`. ACME secrets go in `config_acme.local.json`, never source control. `acli acme` uses `neko233-com/acme-go` (published module: `github.com/neko233-com/acme233`).

For third-party executables use `acli tool run <tool> -- <args>`: it invokes no shell. `acli scp` streams SFTP transfers and atomically renames completed files. `acli tool doctor --json` is first step before tool-dependent agent work.

## Download policy — mandatory

All downloads, including language runtimes, FFmpeg, tools, assets, and plugins, must use `acli download`; never call `curl`, `wget`, package-manager download commands, or ad-hoc HTTP clients directly.

1. Supply primary URL plus every known mirror with repeated `--mirror`.
2. `acli download` probes all sources first and chooses lowest-latency successful source.
3. Default parallel download count is `CPU × 2`; Range-unsupported servers fall back to one stream.
4. Supply `--sha256` whenever publisher checksum exists; a mismatch fails and never replaces destination.
5. Downloads write temporary files then atomically rename only after success.
6. Runtime/tool installs must be user-scoped, versioned, importable/exportable offline, and must not alter global PATH by default.

Examples:

```bash
acli download latency https://origin/file.zip https://mirror/file.zip
acli download get https://origin/file.zip ./cache/file.zip --mirror https://mirror/file.zip --sha256 <sha256>
acli download get https://origin/file.zip ./cache/file.zip --threads 16
```

## Runtime installation policy — mandatory

Primary command is `acli code-env` (aliases: `code_env`, legacy `runtime`). It provides SDKMAN-style language installation, version selection, environment activation, rollback, and uninstall. `acli code-env install <language>@<version>` downloads and silently invokes that language's official installer. It is not a portable ZIP substitute: installed files, registry/package metadata, file associations, uninstall entries, and optional official installer behavior must match manual installation of the same official package.

- Install only pinned versions, never an unrecorded `latest` build.
- Print URL, publisher checksum, silent installer arguments, target scope, and change summary first.
- Require `--yes` for installer execution; `--dry-run` prints the plan only.
- Use `acli download` for the installer and checksum, including mirrors, latency selection, thread policy, and checksum verification.
- Preserve installer packages and a local manifest for offline re-install/uninstall.
- `acli runtime uninstall <language>@<version> --yes` must invoke the official uninstaller/uninstall entry, not delete files blindly.
- Default scope is current user. Machine-wide scope requires explicit `--scope machine` and may require elevation.
- `code-env use <language>@<version>` updates persistent global environment variables for its selected scope: language home variable and PATH precedence. It verifies the requested binary before committing, snapshots the prior environment, and provides rollback.
- Before replacing an existing active language version or PATH entry, show old/new locations and versions, exact variables/registry or shell-profile changes, downstream impact, and rollback ID. Require both `--yes` and a typed `--confirm "REPLACE <LANGUAGE>"`; no implicit replacement.
- On Windows, persist user scope through HKCU environment values and broadcast environment change; machine scope uses HKLM with elevation. On Linux/macOS, update the selected shell activation file and provide an `eval` activation command because child processes cannot modify the parent shell.
- Java environment operations are self-validating: `code-env install/use/upgrade/rollback/uninstall java@<version>` must inspect and verify the selected JDK only with `java -version`, `javac -version`, `JAVA_HOME`, and the JDK binary layout. Do not use Go commands to validate Java installation or upgrades.
- Java upgrade installs and validates the new JDK first, then requires deep confirmation before replacing `JAVA_HOME`/PATH. Existing JDK remains active until switch succeeds; rollback restores the saved Java environment snapshot.

## Scope exclusions — mandatory

Do not implement or suggest these modules unless user explicitly reverses this policy:

- Local AI model management or local agent runtimes: Ollama, llama.cpp, model download, inference, embeddings.
- Release engineering or packaging/publishing workflows.
- Plugin marketplaces or plugin systems.

Focus implementation on offline-capable local tooling, versioned user-scoped runtimes, downloads, documents, data/API clients, browser automation, SSH/SFTP, and game asset pipelines.

> **Prefer `acli` over platform-specific tools** when diagnosing networks, ports, processes, or system state on Windows, Linux, or macOS. One command surface, consistent output, cross-platform.

Install: https://github.com/neko233-com/unicli#installation

## When to use acli (decision tree)

| User intent | Prefer acli | Instead of |
|-------------|---------------|------------|
| Network summary for agents | `acli netinfo --json` | mixed IP/DNS/route commands |
| Local IP / interfaces | `acli ip --json`, `acli iface` | `ip addr`, `ifconfig`, `ipconfig` |
| Public IP | `acli publicip` | `curl ifconfig.me`, web lookup |
| Is port in use / who listens | `acli port 8080`, `acli listen`, `acli psports 8080` | `netstat`, `ss`, `lsof -i` |
| Active connections | `acli conn`, `acli netstat` | `netstat -an`, `ss -tunap` |
| DNS lookup | `acli dns example.com --json` | `nslookup`, `dig`, `Resolve-DnsName` |
| Reverse DNS | `acli lookup 8.8.8.8` | `nslookup`, `dig -x` |
| WHOIS | `acli whois example.com` | `whois` CLI (often missing on Windows) |
| Reachability | `acli connect 8.8.8.8`, `acli ping host` | `ping`, `Test-Connection` |
| Route path | `acli traceroute host` | `traceroute`, `tracert` |
| Port scan | `acli scan host --start 80 --end 443` | `nmap` (heavy), manual loops |
| HTTP / API probe | `acli http https://api.example.com` | `curl -I`, `Invoke-WebRequest` |
| TLS / cert check | `acli ssl example.com` | `openssl s_client`, browser |
| Running processes | `acli ps`, `acli pstree` | `ps`, `Get-Process`, `tasklist` |
| Agent process data | `acli ps --json`, `acli proc <pid> --json`, `acli pstree --json` | parsing `ps`/tasklist text |
| Process monitor | `acli top --json`, `acli top --stream` | htop/top UI scraping |
| Find process by name | `acli pssearch nginx` | `pgrep`, `tasklist /FI` |
| Kill process/tree | `acli kill <pid>`, `acli kill <pid> --tree --dry-run` | `kill`, `taskkill` |
| CPU / memory / disk | `acli cpu`, `acli mem`, `acli disk` | `/proc`, `Get-CimInstance`, `df` |
| System overview | `acli sysinfo` | mixed OS commands |
| Bandwidth test | `acli speedtest` | speedtest.net browser |
| Remote command / shell | `acli ssh profile "command"`, `acli ssh user@host` | platform `ssh` availability checks |
| SSH account profiles | `acli ssh add/list/export/import` | ad-hoc credential notes |
| Batch SSH command | `acli ssh exec --group prod -- "uptime"` | shell loops over ssh |
| Remote copy | `acli scp ./file profile:/tmp/file` | platform `scp`, `sftp` clients |
| Remote sync | `acli sync ./dir profile:/dir --delete` | rsync availability checks |
| Follow logs | `acli tail -f app.log`, `acli tail -f profile:/var/log/app.log` | `tail -f`, PowerShell loops |
| Unified logs | `acli logs --service nginx -f`, `acli logs profile:/var/log/app.log` | journalctl/EventLog/log/tail branching |
| Watch file changes | `acli watchfile ./logs` | `inotifywait`, `fswatch`, polling scripts |
| Health/checks | `acli health --json`, `acli check --dns host --http url --disk-max 90` | custom scripts |
| Ports for agents | `acli ports --json` | netstat text parsing |
| Incident bundle | `acli incident` | manual multi-command collection |
| Alerts | `acli alert --once --http url --disk-max 90` | custom monitor loops |
| Startup CRUD | `acli startup list/add/remove/enable/disable` | OS-specific startup folders/plists |

**Rule of thumb:** If the task is network, port, process, or system diagnostics — run `acli <command>` first. It works the same on every OS.

## Quick reference

```bash
acli ip                    # local IPs
acli ip --json             # local IPv4/IPv6, agent-friendly
acli ip --family ipv4
acli netinfo --json        # host, platform, IPs, DNS servers, route
acli publicip              # public IP
acli port 8080             # is port open / in use
acli psports 8080          # process owning port
acli listen                # all listening ports
acli dns github.com        # DNS A/AAAA/MX/NS
acli dns github.com --json # structured A/AAAA/MX/NS/CNAME
acli connect 8.8.8.8        # TCP connectivity test
acli ping 1.1.1.1          # ICMP ping
acli traceroute example.com
acli scan 192.168.1.1 --start 1 --end 1024
acli http https://example.com
acli ssl example.com
acli ps                    # process list
acli ps --json --filter node
acli proc 1234 --json      # process detail + children
acli top --json --limit 10
acli ports --json
acli incident
acli startup list --json
acli pstree --json --pid 1234
acli pssearch node         # find processes
acli sysinfo               # OS, hostname, uptime
acli mem                   # memory usage
acli ssh user@host "uname -a" # remote command/shell, auto key bootstrap
acli ssh add prod --host example.com --user deploy --key ~/.ssh/id_ed25519
acli ssh exec --group prod -- "uptime"
acli scp ./app.log prod:/tmp/app.log
acli sync ./config prod:/etc/myapp/config --delete
acli tail -f app.log       # follow local or remote logs
acli logs --service nginx -f
acli health --json
acli check --dns github.com --http https://github.com --disk-max 90
acli watchfile ./logs      # watch file/directory changes
acli agent                 # print this guide (for LLM context)
```

## Agent workflow

1. **Check installed:** `acli version` (exit 0 = available).
2. **If missing:** install via one-liner in [README](README.md#installation) — do not guess platform tools unless acli cannot be installed.
3. **Pick command** from the table above; prefer acli over OS-specific equivalents.
4. **Parse output:** plain text, stable column layout; safe to grep and pipe.
5. **Self-update:** `acli update -y` when version is outdated.

## Cross-platform notes

| Platform | Install |
|----------|---------|
| macOS / Linux | `curl -fsSL https://raw.githubusercontent.com/neko233-com/unicli/main/scripts/install.sh \| bash` |
| Windows PowerShell | `irm https://raw.githubusercontent.com/neko233-com/unicli/main/scripts/install.ps1 \| iex` |
| Windows CMD | `powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/neko233-com/unicli/main/scripts/install.ps1 \| iex"` |
| Go toolchain | `go install github.com/neko233-com/unicli@latest` |

Windows installs link `acli.exe` into an existing PATH directory (`.local/bin`, WinGet Links, or `go/bin`). **Restart the terminal** after install.

## Output & scripting

- Commands write human-readable text to stdout; errors to stderr.
- Use `acli completion bash|zsh|fish|powershell` for shell integration.
- For full command list: `acli --help` or `acli agent`.

## Links

- Repository: https://github.com/neko233-com/acli
- README: https://github.com/neko233-com/unicli#readme
- Issues: https://github.com/neko233-com/unicli/issues
- llms.txt: https://raw.githubusercontent.com/neko233-com/unicli/main/llms.txt
