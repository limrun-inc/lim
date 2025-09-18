# `lim` CLI

`lim` lets you run all AI sandbox services from Limrun.

## Installation

### Homebrew

```bash
# Available for both macOS and Linux.
brew install limrun-inc/tap/lim
```

### Linux

Install dependencies:
```bash
sudo apt install adb scrcpy
```

Install `lim`:

```bash
VERSION=v0.1.0
ARCH=amd64
curl -Lo lim https://github.com/limrun-inc/lim/releases/download/${VERSION}/lim-linux-${ARCH}
chmod +x lim
sudo mv lim /usr/local/bin/
```

### Windows

Download the ZIP archive from [releases](https://github.com/limrun-inc/lim/releases) and unpack it.

## Quick Start


Run the following to get an Android sandbox created and streaming!
```bash
lim run android
```

For iOS, run the following:
```bash
lim run ios
```

