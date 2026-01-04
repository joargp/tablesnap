# tablesnap 📊

Convert markdown tables to PNG images. Fast, CLI-first, dark mode by default.

![Example](https://github.com/joargp/tablesnap/raw/master/examples/demo.png)

## Install

```bash
go install github.com/joargp/tablesnap/cmd/tablesnap@latest
```

Or download from [Releases](https://github.com/joargp/tablesnap/releases).

## Usage

```bash
# From stdin
echo "| Name | Price |
|------|-------|
| Foo  | $10   |
| Bar  | $20   |" | tablesnap > table.png

# From file
tablesnap -i data.md -o table.png

# With options
tablesnap --theme light --font-size 16 --padding 12 -o table.png
```

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `-i` | stdin | Input file |
| `-o` | stdout | Output file |
| `--theme` | dark | Theme: `dark` or `light` |
| `--font-size` | 14 | Font size in pixels |
| `--padding` | 10 | Cell padding in pixels |

## Themes

**Dark** (default) — perfect for Telegram, Discord, Slack dark mode:
- Background: `#1a1a1a`
- Text: `#e0e0e0`
- Headers: `#4fc3f7`

**Light** — for light mode apps:
- Background: `#ffffff`
- Text: `#333333`
- Headers: `#1a73e8`

## Supported Symbols

The bundled Inter font supports these symbols:

| Use | Symbol |
|-----|--------|
| Yes/check | ✓ |
| No/cross | ✗ |
| Bullet | ● ○ |
| Star | ★ ☆ |
| Arrow | → ← ↑ ↓ |

**Note:** Emoji (✅ ❌ 🔴) are not supported — use text symbols instead.

## Why?

Messaging apps like Telegram don't render markdown tables. This tool converts them to clean PNG images that display correctly everywhere.

## Building from source

```bash
git clone https://github.com/joargp/tablesnap.git
cd tablesnap
go build -o tablesnap ./cmd/tablesnap
```

## License

MIT
