# Installation Guide

Complete installation instructions for ARM on all platforms.

## System Requirements

- **Operating System**: macOS, Linux, or Windows
- **Architecture**: x86_64 or ARM64
- **Dependencies**: Git (for Git-based registries)
- **Network**: Internet connection for remote registries

## Installation Methods

### Option 1: Install Script (Recommended)

#### macOS/Linux
```bash
curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | bash
```

#### Windows (PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.ps1 | iex
```

### Option 2: Manual Download

1. Visit [GitHub Releases](https://github.com/jomadu/ai-rules-manager/releases)
2. Download the appropriate binary for your platform:
   - `arm-linux-amd64` - Linux x86_64
   - `arm-linux-arm64` - Linux ARM64
   - `arm-darwin-amd64` - macOS Intel
   - `arm-darwin-arm64` - macOS Apple Silicon
   - `arm-windows-amd64.exe` - Windows x86_64

3. Make executable and move to PATH:
```bash
# Linux/macOS
chmod +x arm-*
sudo mv arm-* /usr/local/bin/arm

# Windows
# Move arm-windows-amd64.exe to a directory in your PATH
```

### Option 3: Package Managers

#### Homebrew (macOS/Linux)
```bash
brew tap max-dunn/arm
brew install arm
```

#### Chocolatey (Windows)
```powershell
choco install ai-rules-manager
```

### Option 4: Build from Source

#### Prerequisites
- Go 1.23.2+
- Git
- Make

#### Build Steps
```bash
git clone https://github.com/jomadu/ai-rules-manager.git
cd ai-rules-manager
make build
sudo mv bin/arm /usr/local/bin/
```

## Verification

```bash
# Check installation
arm version

# Test basic functionality
arm --help
```

## Post-Installation Setup

### Initialize Configuration
```bash
arm install
```

This creates stub configuration files if they don't exist.

### Set Up Your First Registry
```bash
# Add a Git registry
arm config add registry default https://github.com/your-org/rules --type=git

# Add channels for your AI tools
arm config add channel cursor --directories=.cursor/rules
arm config add channel q --directories=.amazonq/rules
```

## Platform-Specific Notes

### macOS
- ARM may be blocked by Gatekeeper on first run
- Allow in System Preferences > Security & Privacy
- Or use `xattr -d com.apple.quarantine /usr/local/bin/arm`

### Linux
- Ensure `/usr/local/bin` is in your PATH
- Some distributions may require `sudo` for installation

### Windows
- Add ARM directory to your PATH environment variable
- PowerShell execution policy may need adjustment:
  ```powershell
  Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
  ```

## Troubleshooting Installation

### Permission Denied
```bash
# Use user directory instead of system-wide
mkdir -p ~/bin
mv arm ~/bin/
echo 'export PATH=$PATH:~/bin' >> ~/.bashrc
source ~/.bashrc
```

### Command Not Found
```bash
# Check PATH
echo $PATH

# Add to shell profile
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
source ~/.bashrc
```

### Network Issues
```bash
# Download manually if script fails
wget https://github.com/jomadu/ai-rules-manager/releases/latest/download/arm-linux-amd64
chmod +x arm-linux-amd64
sudo mv arm-linux-amd64 /usr/local/bin/arm
```

## Uninstallation

```bash
# Remove binary
sudo rm /usr/local/bin/arm

# Remove configuration (optional)
rm -rf ~/.arm
rm .armrc arm.json arm.lock

# Remove cache (optional)
rm -rf ~/.arm/cache
```

## Next Steps

- [Quick Start Guide](quick-start.md) - Get up and running in 5 minutes
- [Configuration Guide](configuration.md) - Set up registries and channels
- [Usage Guide](usage.md) - Learn all available commands
