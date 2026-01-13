<div align="center">

```
     /\     / ___| |_   / \  |_   _|
    /  \    \___ \  | | / _ \   | |  
   / /\ \    ___) | | |/ ___ \  | |  
  /_/  \_\  |____/  |_/_/   \_\ |_|  
                                     
```

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

Tired of waiting for AWS CLI commands to complete? **astat** caches AWS resources stats locally, providing **instant access** to cloud infrastructure from the command line

```bash
# Traditional AWS CLI (slow, every time)
$ time aws <service> describe-* --query '...'
# ... 2-5 seconds

# astat (instant, after first cache)
$ time astat <service> list
# ... 0.05 seconds ⚡
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
</table>

## 📦 Installation

### Homebrew (Recommended for macOS/Linux)

```bash
brew install sunil-saini/tap/astat
```

### Direct Download

Download the latest binary for your platform:

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/sunil-saini/astat/releases/latest/download/astat_Darwin_arm64.tar.gz
tar -xzf astat_Darwin_arm64.tar.gz
sudo mv astat /usr/local/bin/

# macOS (Intel)
curl -LO https://github.com/sunil-saini/astat/releases/latest/download/astat_Darwin_x86_64.tar.gz
tar -xzf astat_Darwin_x86_64.tar.gz
sudo mv astat /usr/local/bin/

# Linux (amd64)
curl -LO https://github.com/sunil-saini/astat/releases/latest/download/astat_Linux_x86_64.tar.gz
tar -xzf astat_Linux_x86_64.tar.gz
sudo mv astat /usr/local/bin/
```

### Go Install

```bash
go install github.com/sunil-saini/astat@latest
```

### One-Liner Script

```bash
curl -sSL https://raw.githubusercontent.com/sunil-saini/astat/main/install.sh | sh
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

5. **Start querying** (instantly!):
   ```bash
   astat ec2 list
   astat s3 list
   astat lambda list
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
- [ ] Support for more AWS services (ECS, RDS, DynamoDB)
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

**[⬆ back to top](#)**

Made with ❤️ by [Sunil Saini](https://github.com/sunil-saini)

If you find this project useful, please consider giving it a ⭐!

</div>
