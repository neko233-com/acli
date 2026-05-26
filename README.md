# unicli

A cross-platform network & process CLI tool built with Go 1.23+.

## Purpose

unicli is designed to reduce learning curve for network diagnostics and provide a unified interface across Windows, Linux, and macOS. Perfect for system administrators, DevOps engineers, and developers who need quick network insights.

## Features

### Network Commands
- **IP Management** - Local IPs, public IP, interface details
- **DNS Tools** - Forward lookup, reverse lookup, WHOIS
- **Connectivity** - Ping, traceroute, port scanning
- **HTTP Tools** - HTTP requests, SSL certificate check
- **Speed Test** - Download bandwidth measurement

### Process Commands
- **Process List** - View all running processes
- **Process Tree** - Hierarchical process view
- **Process Search** - Find processes by name
- **Port Mapper** - Find which process uses a port
- **Process Kill** - Terminate processes by PID

### System Commands
- **System Info** - OS details, hostname, uptime
- **CPU Info** - Model, cores, current usage
- **Memory Info** - Physical and virtual memory
- **Disk Usage** - Mounted partitions space

## Quick Start

```bash
# Show local IPs
unicli ip

# Check public IP
unicli publicip

# Check if port 8080 is in use
unicli port 8080

# List processes
unicli ps

# Find process using port
unicli psports 8080

# Test connectivity to Google DNS
unicli connect 8.8.8.8

# DNS lookup
unicli dns github.com

# Check SSL certificate
unicli ssl github.com

# Scan ports on a host
unicli scan 192.168.1.1 --start 80 --end 443

# Check memory usage
unicli mem
```

## Installation

### One-Click Install (Recommended)

**All Platforms (Linux/macOS/Windows)**
```bash
curl -fsSL https://raw.githubusercontent.com/neko233-com/unicli/main/scripts/install.sh | bash
```

**With specific version**
```bash
curl -fsSL https://raw.githubusercontent.com/neko233-com/unicli/main/scripts/install.sh | bash -s -- v1.0.0
```

### From Source
```bash
go install github.com/neko233-com/unicli@latest
```

### Pre-built Binaries
Download from [GitHub Releases](https://github.com/neko233-com/unicli/releases)

### Package Managers

**macOS**
```bash
brew install neko233-com/unicli/unicli
```

## Cross-Platform Support

| Platform | amd64 | arm64 |
|----------|-------|-------|
| Windows  | ✅    | N/A   |
| Linux    | ✅    | ✅    |
| macOS    | ✅    | ✅    |

## Building from Source

### Prerequisites
- Go 1.23+

### Build
```bash
go build -ldflags="-s -w" -o unicli.exe .
```

## Commands Reference

### Network

| Command | Description |
|---------|-------------|
| `unicli ip` | Show local IP addresses |
| `unicli iface` | Show detailed network interfaces |
| `unicli publicip` | Show public IP address |
| `unicli port <port>` | Check if port is in use |
| `unicli listen` | Show all listening ports |
| `unicli conn` | Show network connections |
| `unicli netstat` | Show network statistics |
| `unicli connect <host>` | Test connectivity to host |
| `unicli dns <domain>` | DNS lookup (A, AAAA, MX, NS) |
| `unicli lookup <ip>` | Reverse DNS lookup |
| `unicli whois <domain>` | WHOIS lookup |
| `unicli ping <host>` | Ping host with ICMP |
| `unicli traceroute <host>` | Trace route to host |
| `unicli scan <host>` | Scan ports on host |
| `unicli http <url>` | Make HTTP request |
| `unicli ssl <host>` | Check SSL certificate |
| `unicli speedtest` | Test download speed |

### Process

| Command | Description |
|---------|-------------|
| `unicli ps` | List running processes |
| `unicli pstree` | Show process tree |
| `unicli pssearch <name>` | Search processes by name |
| `unicli psports <port>` | Show process using port |
| `unicli kill <pid>` | Kill process by PID |

### System

| Command | Description |
|---------|-------------|
| `unicli sysinfo` | Show system information |
| `unicli cpu` | Show CPU info and usage |
| `unicli mem` | Show memory usage |
| `unicli disk` | Show disk usage |

### Utilities

| Command | Description |
|---------|-------------|
| `unicli version` | Show version info |
| `unicli update` | Check for updates |
| `unicli completion <shell>` | Generate completion script |
| `unicli doc` | Open documentation |

## SEO & GEO Keywords

network cli, cross-platform networking, port scanner, process manager, net tools, network diagnostics, system administration, devops tools, network utilities, golang cli, dns lookup, whois, ssl check, http client, speedtest, traceroute, ping

## License

MIT License

## Contributing

Contributions are welcome! Please open an issue or submit a PR.

## Links

- GitHub: https://github.com/neko233-com/unicli
- Issues: https://github.com/neko233-com/unicli/issues