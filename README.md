# salah-tui

A terminal UI for daily Muslim prayer times. Opens to a split-pane dashboard showing today's full prayer schedule with a live countdown, an ayah of the day, and an interactive Quran lookup. Designed to live in a tmux pane or be launched on demand.

---

## Screenshot

> _Screenshot coming soon._

---

## Features

- **Live dashboard** — split-pane view with prayer schedule (left) and ayah of the day (right)
- **Countdown timer** — highlights the next prayer with a live `HH:MM:SS` countdown; color shifts amber at ≤15 min and red at ≤5 min
- **Current prayer highlight** — the most recently started prayer is shown in green
- **Quran lookup** — press `/` to search any verse by `surah:ayah` reference
- **Minimal mode** — single-line output suitable for a tmux status bar
- **Offline support** — falls back to cached data when the network is unavailable, with an indicator in the header
- **Daily ayah** — deterministically seeded by date so it stays consistent across multiple opens on the same day
- **First-run setup** — interactive prompt on first launch; geocodes your city automatically via Nominatim

---

## Installation

**Requirements:** Go 1.21+

```sh
# Build from source
git clone https://github.com/MahmoudDolah/salah-tui.git
cd salah-tui
go build -o salah-tui .

# Or install directly
go install github.com/MahmoudDolah/salah-tui@latest
```

---

## Usage

```sh
# Open the full dashboard (default)
salah-tui

# Single-line output: next prayer name, scheduled time, and countdown
salah-tui --minimal

# Watch mode: refresh every second (useful for tmux status bars)
salah-tui --minimal --watch
```

### tmux status bar

Add the following to your `~/.tmux.conf` to show the next prayer in your status line:

```
set -g status-right "#(salah-tui --minimal)"
```

Or with auto-refresh:

```
set -g status-right "#(salah-tui --minimal)"
set -g status-interval 1
```

---

## Keybindings

| Key | Action |
|-----|--------|
| `q` / `ctrl+c` | Quit |
| `/` | Open Quran lookup |
| `esc` | Close lookup / dismiss help |
| `r` | Force refresh (re-fetches prayer times and ayah) |
| `?` | Toggle keybindings help |

### Quran lookup

Press `/`, type a reference in `surah:ayah` format (e.g. `2:255`), and press `Enter`. The result replaces the ayah pane until you press `esc`.

---

## Configuration

Config file: `~/.config/salah/config.toml`

Created automatically on first launch. You will be prompted for your city name (coordinates are resolved via Nominatim) and your preferred calculation method.

```toml
[location]
latitude  = 40.7128
longitude = -74.0060
city      = "New York"
timezone  = "America/New_York"

[calculation]
# Options: ISNA, MWL, Egyptian, Karachi, UmmAlQura, Tehran, Shia
method = "ISNA"

[display]
# 12 or 24
time_format  = 12
show_sunrise = true
```

### Calculation methods

| Key | Authority |
|-----|-----------|
| `ISNA` | Islamic Society of North America |
| `MWL` | Muslim World League |
| `Egyptian` | Egyptian General Authority of Survey |
| `UmmAlQura` | Umm Al-Qura University, Makkah |
| `Karachi` | University of Islamic Sciences, Karachi |
| `Tehran` | Institute of Geophysics, University of Tehran |
| `Shia` | Shia Ithna Ashari / Leva Research Institute, Qum |

---

## Caching

Cache files live at `~/.cache/salah/` and are keyed by date:

| File | Contents |
|------|----------|
| `prayer-times-YYYY-MM-DD.json` | Prayer times for that day |
| `ayah-YYYY-MM-DD.json` | Ayah of the day |

Stale files (not from today) are deleted on each launch. When offline and a cache file exists, the app uses stale data and shows an `(offline)` indicator in the header rather than erroring.

---

## Architecture

```
salah-tui/
├── main.go                    # Entry point: config → cache → API → run
├── internal/
│   ├── api/
│   │   ├── aladhan.go         # Fetch and parse prayer times from Aladhan API
│   │   └── quran.go           # Fetch and parse ayah from alquran.cloud API
│   ├── cache/
│   │   └── cache.go           # Read / write / clean up JSON cache files
│   ├── config/
│   │   └── config.go          # TOML config load, first-run setup, geocoding
│   ├── prayer/
│   │   └── prayer.go          # Pure logic: next prayer, countdown, formatting
│   ├── model/
│   │   └── model.go           # Bubbletea model: Init / Update / View
│   └── ui/
│       ├── styles.go          # Lipgloss color palette and shared styles
│       ├── schedule.go        # Left pane: prayer schedule renderer
│       ├── ayah.go            # Right pane: ayah and Quran lookup renderer
│       └── dashboard.go       # Layout: joins panes, narrow fallback, help overlay
```

The app follows the [bubbletea](https://github.com/charmbracelet/bubbletea) Elm-like architecture:

- **`model.Model`** holds all application state (prayers, ayah, lookup state, terminal size, etc.)
- **`Update`** handles messages: key presses, tick events, window resize, and async data-load results
- **`View`** delegates to the `ui` renderers, which are pure functions of state — no side effects
- Network I/O happens in bubbletea commands (async), so the UI never blocks

**Data flow on startup:**

1. Load config (`internal/config`) → run first-run setup if missing
2. Check cache (`internal/cache`) → fetch from APIs if stale (`internal/api`)
3. Parse prayer times into typed structs (`internal/prayer`)
4. Hand off to `tea.NewProgram` (dashboard) or `runMinimal` (minimal mode)

---

## Roadmap

- **Daemon / notification mode** — background process that sends OS notifications at prayer time
- **Qibla direction** — compass bearing to Mecca from your configured location
- **Multiple location profiles** — switch between configured cities
- **Hijri calendar view** — full month view beyond the date header
- **Windows support** — currently Linux and macOS only

---

## Credits

- Prayer times provided by [Aladhan API](https://aladhan.com/prayer-times-api) — free, no authentication required
- Quran text and translations provided by [alquran.cloud API](https://alquran.cloud/api) — free, no authentication required
- Geocoding via [Nominatim / OpenStreetMap](https://nominatim.openstreetmap.org)
- Built with [bubbletea](https://github.com/charmbracelet/bubbletea) and [lipgloss](https://github.com/charmbracelet/lipgloss) by [Charm](https://charm.sh)
