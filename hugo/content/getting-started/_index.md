---
title: "Getting Started"
weight: 1
---

## Installation

### Using Nix (Recommended)

```bash
nix run github:timlinux/cheetah
```

### From Source

```bash
git clone https://github.com/timlinux/cheetah.git
cd cheetah
make build
./cheetah
```

## Quick Start

1. **Launch Cheetah**: Run `./cheetah` or open the web interface
2. **Load a document**: Drag and drop or use the file picker
3. **Press SPACE** to start reading
4. **Adjust speed**: Use `j`/`k` or `←`/`→` to change WPM
5. **Pause anytime**: Press SPACE again

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| **SPACE** | Play/Pause |
| **j** / **←** | Decrease speed 50 WPM |
| **k** / **→** | Increase speed 50 WPM |
| **h** / **↑** | Previous paragraph |
| **l** / **↓** | Next paragraph |
| **r** | Return to start |
| **b** / **ESC** | Back to document picker |
| **1-9** | Speed presets (200-1000 WPM) |
