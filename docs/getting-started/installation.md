# Installation

To install viaplay-cli, you need Go 1.20+ installed on your system.

---

## Install via Homebrew (Recommended)

If you are on macOS <i class="fa-brands fa-apple" style="color:#888;"></i> or Linux <i class="fa-brands fa-linux" style="color:#888;"></i> and use Homebrew, you can install viaplay-cli with:

> **Note:** If the repository is private, you must set your GitHub API token for Homebrew to access it:
>
> ```bash
> export HOMEBREW_GITHUB_API_TOKEN=your_github_token
> ```

```bash
# Tap the repository and install
brew tap nentgroup/viaplay-cli https://github.com/nentgroup/viaplay-cli
brew install --cask vip
```

---

## Install with Go

```bash
go install github.com/nentgroup/viaplay-cli/cmd/vip@latest
```

---

## Build from Source

```bash
git clone https://github.com/nentgroup/viaplay-cli.git
cd viaplay-cli/cmd/vip
go build -o vip
```

Place the resulting `vip` binary in your PATH.

---

## Install on Windows

You can install viaplay-cli on Windows <i class="fa-brands fa-windows" style="color:#888;"></i> by downloading a prebuilt binary or building from source.

### Option 1: Download Prebuilt Binary

1. Go to the [GitHub Releases page](https://github.com/nentgroup/viaplay-cli/releases).
2. Download the latest `vip.exe` for Windows.
3. Place `vip.exe` in a directory included in your PATH (e.g., `C:\Tools` or `%USERPROFILE%\bin`).
4. Open a new terminal and run:

```cmd
vip --version
```

### Option 2: Build from Source

If you have Go 1.20+ installed:

```powershell
git clone https://github.com/nentgroup/viaplay-cli.git
cd viaplay-cli/cmd/vip
go build -o vip.exe
```

Move `vip.exe` to a directory in your PATH.

---

## Requirements
- Go 1.20 or newer
- Git (for template cloning)
- macOS <i class="fa-brands fa-apple" style="color:#888;"></i>, Linux <i class="fa-brands fa-linux" style="color:#888;"></i>, or Windows <i class="fa-brands fa-windows" style="color:#888;"></i>

### Optional
- [Nerd Fonts](https://www.nerdfonts.com/) - For an enhanced CLI experience with proper icons and symbols. Without Nerd Fonts, some visual elements may display as placeholder characters.

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
