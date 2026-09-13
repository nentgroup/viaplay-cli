# Installation

To install vip, you need Go 1.25+ installed on your system.

---

## Install via Homebrew (macOS)

If you use Homebrew on macOS, install `vip` with:

```bash
brew tap nentgroup/viaplay-cli https://github.com/nentgroup/viaplay-cli
brew install --cask nentgroup/viaplay-cli/vip
```

---

## Install on Linux

You can install `vip` on Linux in several ways:

### Option 1: Install via Homebrew (Linuxbrew)

If you have [Homebrew on Linux](https://docs.brew.sh/Homebrew-on-Linux) installed:

```bash
brew tap nentgroup/viaplay-cli https://github.com/nentgroup/viaplay-cli
brew install --cask nentgroup/viaplay-cli/vip
```

### Option 2: Install from a package manager

Download a package built by GoReleaser from the [GitHub Releases page](https://github.com/nentgroup/viaplay-cli/releases):

```bash
# Debian / Ubuntu
sudo dpkg -i vip_*.deb

# RPM-based distros
sudo rpm -i vip_*.rpm

# Alpine / APK-based distros
sudo apk add --allow-untrusted vip_*.apk
```

### Option 3: Install from source

```bash
go install github.com/nentgroup/viaplay-cli/cmd/vip@latest
```

### Option 4: Install a downloaded binary

```bash
chmod +x ./vip
sudo install ./vip /usr/local/bin/vip
```

---

## Install on Windows

You can install `vip` on Windows in several ways:

### Option 1: Install via winget

```powershell
winget install nentgroup.viaplay-cli
```

### Option 2: Download a release binary

Download the latest `vip.exe` from the [GitHub Releases page](https://github.com/nentgroup/viaplay-cli/releases) and place it in a directory already on your `PATH`.

### Option 3: Build from source

```powershell
git clone https://github.com/nentgroup/viaplay-cli.git
cd viaplay-cli
go build -o vip.exe ./cmd/vip
Move-Item .\vip.exe "$env:USERPROFILE\bin\vip.exe"
```

---

## Install with Go (all platforms)

```bash
go install github.com/nentgroup/viaplay-cli/cmd/vip@latest
```

This works on macOS, Linux, and Windows when Go is installed and the Go binary directory is on your `PATH`.

---

## Build from source

When building from source, set the GitHub OAuth client ID used by device auth:

```bash
export GITHUB_CLIENT_ID="your-client-id"
git clone https://github.com/nentgroup/viaplay-cli.git
cd viaplay-cli
go build -o vip ./cmd/vip
```

On PowerShell:

```powershell
$env:GITHUB_CLIENT_ID = "your-client-id"
git clone https://github.com/nentgroup/viaplay-cli.git
cd viaplay-cli
go build -o vip.exe .\cmd\vip
```

Place the resulting binary in your `PATH`. This environment variable is required for `vip auth login` to work unless the client ID is embedded in a release build.

---

## Requirements
- Go 1.25 or newer
- Git (for template cloning)
- macOS <i class="fa-brands fa-apple" style="color:#888;"></i>, Linux <i class="fa-brands fa-linux" style="color:#888;"></i>, or Windows <i class="fa-brands fa-windows" style="color:#888;"></i>

### Optional
- [Nerd Fonts](https://www.nerdfonts.com/) - Required for CLI icons and symbols to display correctly. Without Nerd Fonts, some characters may show as placeholders.

---

## Upgrading
To upgrade to the latest version:

```bash
brew upgrade viaplay-cli
```

or

```bash
go install github.com/nentgroup/viaplay-cli/cmd/vip@latest
```

---

## Verify Installation

```bash
vip --version
```

---

For troubleshooting, see the [FAQ](../faq.md) or open an issue on GitHub.
