<a id="top"></a>
<div align="center">

<table>
<tr>
<td align="left">
<pre>
    ___    _____ _______  ___  ______
   /   |  / ___//_  __/ /   |/_  __/
  / /| |  \__ \  / /   / /| | / /   
 / ___ |___/ / / /   / ___ |/ /    
/_/  |_/____/ /_/   /_/  |_/_/     
</pre>
</td>
</tr>
</table>

**⚡ Lightning fast local AWS Stats indexer**

[![Release](https://img.shields.io/github/v/release/sunil-saini/astat?style=flat-square)](https://github.com/sunil-saini/astat/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/sunil-saini/astat/release.yml?style=flat-square)](https://github.com/sunil-saini/astat/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/sunil-saini/astat?style=flat-square)](https://goreportcard.com/report/github.com/sunil-saini/astat)
[![License](https://img.shields.io/github/license/sunil-saini/astat?style=flat-square)](LICENSE)
[![Downloads](https://img.shields.io/github/downloads/sunil-saini/astat/total?style=flat-square)](https://github.com/sunil-saini/astat/releases)

[Features](#-features) • [Installation](#-installation) • [Quick Start](#-quick-start) • [Usage](#-usage) • [Documentation](#-documentation)

</div>

---

## 🎯 Why astat?

Tired of waiting for AWS CLI commands to complete? **astat** caches AWS resources stats locally for **instant access** and provides deep **infrastructure tracing** to visualize exactly how your domain requests flow through AWS

```bash
# Traditional AWS CLI (slow, every time)
$ time aws <service> describe-* --query '...'
# ... 2-5 seconds

# astat (instant, after first cache)
$ time astat <service> list
# ... 0.05 seconds ⚡

# understand exactly how your domain requests flow through AWS
$ astat domain trace myr53.hostedrecord.com/api
```

**40-100x faster** for everyday queries!

## ✨ Features

<table>
<tr>
<td width="50%">

### 🚀 Performance
- **Lightning Fast**: Local caching for instant access
- **Concurrent Refresh**: Refresh all services in parallel
- **Smart Caching**: Configurable TTL and Auto Refresh

</td>
<td width="50%">

### 🎨 User Experience
- **Beautiful CLI**: Clean tabular output
- **Multiple Formats**: Table, JSON
- **Shell Auto Completion**: Bash, Zsh, and Fish support

</td>
</tr>
<tr>
<td width="50%">

### 🔧 Flexibility
- **Multi-Service**: EC2, S3, Lambda, CloudFront, Route53, SSM
- **Easy Config**: Simple YAML configuration
- **Self-Updating**: Built-in upgrade command

</td>
<td width="50%">

### 🛡️ Reliability
- **Auto-Refresh**: Keeps data fresh automatically
- **Error Recovery**: Graceful handling of API failures
- **Offline Mode**: Works with cached data when offline

</td>
</tr>
<tr>
<td width="100%" colspan="2">

### 🔍 Infrastructure Tracing
- **Deep Inspection**: Trace a domain or URI flow from DNS down to EC2 instances
- **Visual Mapping**: Beautiful tree representation of your infrastructure
- **Full Stack**: Support for Route53, CloudFront, ALB/NLB/CLB, and more

</td>
</tr>
</table>

## 📦 Installation


### One Liner Script

```bash
curl -sSL https://raw.githubusercontent.com/sunil-saini/astat/refs/heads/main/install.sh | sh
```

### Homebrew

```bash
brew install sunil-saini/tap/astat
```

### Direct Download

Download the latest binary for your platform:

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/sunil-saini/astat/releases/latest/download/astat_darwin_arm64.tar.gz
tar -xzf astat_darwin_arm64.tar.gz
sudo mv astat /usr/local/bin/

# macOS (Intel)
curl -LO https://github.com/sunil-saini/astat/releases/latest/download/astat_darwin_x86_64.tar.gz
tar -xzf astat_darwin_x86_64.tar.gz
sudo mv astat /usr/local/bin/

# Linux (amd64)
curl -LO https://github.com/sunil-saini/astat/releases/latest/download/astat_linux_x86_64.tar.gz
tar -xzf astat_linux_x86_64.tar.gz
sudo mv astat /usr/local/bin/
```

### Go Install

```bash
go install github.com/sunil-saini/astat@latest
```

## 🚀 Quick Start

1. **Install astat** (see above)

2. **Set up shell completion and PATH**:
   ```bash
   astat install
   ```

3. **Configure AWS credentials** (if not already done):
   ```bash
   aws configure
   # or use environment variables, IAM roles, etc.
   ```

4. **Check status and populate cache**:
   ```bash
   astat status
   # First run will trigger automatic cache refresh
   ```

5. **Start using** (instantly!):
   ```bash
   astat ec2 list
   astat s3 list
   astat lambda list
   astat domain trace myr53.hostedrecord.com/api
   ```

## 💡 Usage

### List Resources

```bash
# EC2 instances
astat ec2 list              # or: astat ec2 ls
astat ec2 list --refresh    # Force refresh from AWS

# S3 buckets
astat s3 list

# Lambda functions
astat lambda list

# CloudFront distributions
astat cloudfront list

# Route53 hosted zones
astat route53 list

# Route53 DNS records
astat route53 records

# SSM parameters
astat ssm list
astat ssm get <parameter-name>
```

### 🔍 Infrastructure Tracing

The flagship feature of **astat**! Trace exactly how a domain or request URI is routed through your AWS infrastructure

```bash
# Trace a domain flow
astat domain trace api.example.com

# Trace a specific URI
astat domain trace api.example.com/v1/health
```

**What it traces:**
- **External DNS**: Current IPs and CNAME chains
- **Route53**: Zone matching, A/AAAA/CNAME/Alias records
- **CloudFront**: Distribution aliases, origins, and cache behaviors
- **ELB (v1 & v2)**: ALB/NLB/CLB listeners, rules, and conditions
- **Targets**: Target Groups, health status, and EC2/Lambda targets

### Refresh Cache

```bash
# Refresh all services concurrently
astat refresh

# Refresh specific service
astat ec2 list --refresh
```

### Check Status

```bash
# View cache status and check for updates
astat status
```

### Configuration

```bash
# View current configuration
astat config list

# Set cache TTL (default: 24h)
astat config set ttl 1h

# Enable/disable auto-refresh (default: enabled)
astat config set auto-refresh true
```

### Upgrade

```bash
# Check for and install updates
astat upgrade

# Check current version
astat version
```

## 📚 Documentation

### Configuration

astat stores configuration in `~/.config/astat/config.yaml` and cache in `~/.cache/astat/`.

**Available Settings:**

| Setting | Default | Description |
|---------|---------|-------------|
| `ttl` | `1h` | Cache time-to-live (e.g., `30m`, `2h`, `1d`) |
| `auto-refresh` | `true` | Automatically refresh stale data  |
| `cache_dir` | `~/.cache/astat` | Custom cache directory (optional) |

### Output Formats

```bash
# Table format (default)
astat ec2 list

# JSON format
astat ec2 list --output json

# Pipe to jq for advanced filtering
astat ec2 list --output json | jq '.[] | select(.State.Name == "running")'
```

### Shell Completion

Run `astat install` to automatically set up shell completion, or manually:

```bash
# Bash
astat completion bash > /etc/bash_completion.d/astat

# Zsh
astat completion zsh > "${fpath[1]}/_astat"

# Fish
astat completion fish > ~/.config/fish/completions/astat.fish
```

## 🛠️ Development

### Building from Source

```bash
git clone https://github.com/sunil-saini/astat.git
cd astat
go build -o astat main.go
```

### Running Tests

```bash
go test ./...
```

## 🗺️ Roadmap

- [ ] Multi Region support
- [ ] Support for more AWS services (RDS, SQS, DynamoDB etc)
- [ ] Export to various formats (CSV, YAML)

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Built with these amazing tools:

- [Cobra](https://github.com/spf13/cobra) - Powerful CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [pterm](https://github.com/pterm/pterm) - Beautiful terminal output
- [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2) - Official AWS SDK

---

<div align="center">

**[⬆ back to top](#top)**

Made with ❤️ by [Sunil Saini](https://github.com/sunil-saini)

If you find this project useful, please consider giving it a ⭐!

</div>
