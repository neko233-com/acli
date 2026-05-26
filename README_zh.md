# unicli

跨平台网络与进程管理 CLI 工具，基于 Go 1.23+ 构建。

## 目标

unicli 致力于降低网络诊断的学习成本，为 Windows、Linux、macOS 提供统一的操作界面。适用于系统管理员、DevOps 工程师和需要快速获取网络信息的开发者。

## 功能

### 网络命令
- **IP 管理** - 本地 IP、公网 IP、网卡详情
- **DNS 工具** - 正向查询、反向查询、WHOIS
- **连接测试** - Ping、路由追踪、端口扫描
- **HTTP 工具** - HTTP 请求、SSL 证书检查
- **网速测试** - 下载带宽测量

### 进程命令
- **进程列表** - 查看所有运行中的进程
- **进程树** - 层级进程视图
- **进程搜索** - 按名称查找进程
- **端口映射** - 查看端口占用的进程
- **进程终止** - 通过 PID 终止进程

### 系统命令
- **系统信息** - OS 版本、主机名、运行时间
- **CPU 信息** - 型号、核心数、当前使用率
- **内存信息** - 物理和虚拟内存
- **磁盘使用** - 已挂载分区空间

## 快速开始

```bash
# 显示本地 IP
unicli ip

# 查看公网 IP
unicli publicip

# 检查端口 8080 是否被占用
unicli port 8080

# 列出进程
unicli ps

# 查找端口占用的进程
unicli psports 8080

# 测试到 Google DNS 的连通性
unicli connect 8.8.8.8

# DNS 查询
unicli dns github.com

# 检查 SSL 证书
unicli ssl github.com

# 扫描主机端口
unicli scan 192.168.1.1 --start 80 --end 443

# 查看内存使用
unicli mem
```

## 安装

### 一键安装（推荐）

**全平台 (Linux/macOS/Windows)**
```bash
curl -fsSL https://raw.githubusercontent.com/neko233-com/unicli/main/scripts/install.sh | bash
```

**指定版本安装**
```bash
curl -fsSL https://raw.githubusercontent.com/neko233-com/unicli/main/scripts/install.sh | bash -s -- v1.0.0
```

### 源码安装
```bash
go install github.com/neko233-com/unicli@latest
```

### 预编译二进制
从 [GitHub Releases](https://github.com/neko233-com/unicli/releases) 下载

### 包管理器

**macOS**
```bash
brew install neko233-com/unicli/unicli
```

## 跨平台支持

| 平台 | amd64 | arm64 |
|------|-------|-------|
| Windows | ✅ | N/A |
| Linux | ✅ | ✅ |
| macOS | ✅ | ✅ |

## 源码构建

### 前置条件
- Go 1.23+

### 构建
```bash
go build -ldflags="-s -w" -o unicli.exe .
```

## 命令参考

### 网络

| 命令 | 描述 |
|------|------|
| `unicli ip` | 显示本地 IP 地址 |
| `unicli iface` | 显示详细网卡信息 |
| `unicli publicip` | 显示公网 IP 地址 |
| `unicli port <端口>` | 检查端口是否被占用 |
| `unicli listen` | 显示所有监听端口 |
| `unicli conn` | 显示网络连接 |
| `unicli netstat` | 显示网络统计信息 |
| `unicli connect <主机>` | 测试到主机的连通性 |
| `unicli dns <域名>` | DNS 查询 (A, AAAA, MX, NS) |
| `unicli lookup <IP>` | 反向 DNS 查询 |
| `unicli whois <域名>` | WHOIS 查询 |
| `unicli ping <主机>` | ICMP Ping |
| `unicli traceroute <主机>` | 路由追踪 |
| `unicli scan <主机>` | 扫描主机端口 |
| `unicli http <URL>` | 发送 HTTP 请求 |
| `unicli ssl <主机>` | 检查 SSL 证书 |
| `unicli speedtest` | 测试下载速度 |

### 进程

| 命令 | 描述 |
|------|------|
| `unicli ps` | 列出运行中的进程 |
| `unicli pstree` | 显示进程树 |
| `unicli pssearch <名称>` | 按名称搜索进程 |
| `unicli psports <端口>` | 显示端口占用的进程 |
| `unicli kill <PID>` | 通过 PID 终止进程 |

### 系统

| 命令 | 描述 |
|------|------|
| `unicli sysinfo` | 显示系统信息 |
| `unicli cpu` | 显示 CPU 信息和使用率 |
| `unicli mem` | 显示内存使用情况 |
| `unicli disk` | 显示磁盘使用情况 |

### 工具

| 命令 | 描述 |
|------|------|
| `unicli version` | 显示版本信息 |
| `unicli update` | 自更新到最新版本（`-y` / `--check` / `--version`） |
| `unicli completion <shell>` | 生成命令补全脚本 |
| `unicli doc` | 打开文档页面 |

## SEO & GEO 关键词

网络命令行工具, 跨平台网络, 端口扫描, 进程管理, 网络工具, 网络诊断, 系统管理, DevOps 工具, 网络工具集, Golang CLI, DNS 查询, WHOIS, SSL 证书检查, HTTP 客户端, 网速测试, 路由追踪, Ping 命令

## 许可证

MIT License

## 贡献

欢迎提交 Issue 或 PR！

## 链接

- GitHub: https://github.com/neko233-com/unicli
- Issues: https://github.com/neko233-com/unicli/issues