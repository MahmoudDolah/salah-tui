# Salah TUI — Claude Code Implementation Spec

## Overview

A terminal user interface for daily Muslim prayer time tracking, built in Go. Opens to a dashboard showing today's full prayer schedule, a live countdown to the next prayer, and an ayah of the day. Designed to live in a tmux pane or be launched on demand.

---

## Tech Stack

- **Language:** Go
- **TUI framework:** [bubbletea](https://github.com/charmbracelet/bubbletea)
- **Styling:** [lipgloss](https://github.com/charmbracelet/lipgloss)
- **Prayer times API:** [Aladhan](https://aladhan.com/prayer-times-api) — free, no auth required
- **Quran API:** [alquran.cloud](https://alquran.cloud/api) — free, no auth required
- **Config format:** TOML (`~/.config/salah/config.toml`)
- **Cache:** JSON file at `~/.cache/salah/`

---

## Features

### 1. Dashboard Layout (default)

Split into two vertical panes:

**Left pane — Prayer Schedule**
- Hijri and Gregorian date header
- Full day's prayer times: Fajr, Sunrise, Dhuhr, Asr, Maghrib, Isha
- Current prayer highlighted (different color)
- Next prayer marked with a live countdown timer
- Countdown color changes: normal → amber (≤15 min) → red (≤5 min)

**Right pane — Ayah of the Day**
- Arabic text of the ayah (rendered right-to-left if terminal supports it; fallback to transliteration)
- English translation below
- Surah name + ayah reference (e.g. Al-Baqarah 2:255)
- Rotates daily (deterministic: seeded by date so it's consistent across opens)

### 2. Minimal Mode (`--minimal` flag)

Single-line or compact view: next prayer name + countdown. No panes. Suitable for embedding in a tmux status bar or small terminal window. Exits after displaying (non-interactive) unless `--watch` is also passed, in which case it refreshes every second.

### 3. Quran Lookup (interactive command)

Press `/` in the dashboard to open a lookup prompt. Input format: `2:255` (surah:ayah). Displays the result in the right pane, replacing the ayah of the day temporarily. Press `Esc` to return to ayah of the day.

### 4. Keybindings

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit |
| `/` | Open Quran lookup |
| `Esc` | Close lookup, return to ayah of the day |
| `r` | Force refresh (re-fetch prayer times + ayah) |
| `?` | Toggle help overlay |

---

## Configuration

File: `~/.config/salah/config.toml`

Generated on first launch with sane defaults if missing.

```toml
[location]
latitude = 40.7128
longitude = -74.0060
city = "New York"
timezone = "America/New_York"

[calculation]
# Options: ISNA, MWL, Egyptian, Karachi, UmmAlQura, Tehran, Shia
method = "ISNA"

[display]
# 12 or 24
time_format = 12
# Show Sunrise in the schedule (not a prayer, but useful)
show_sunrise = true
```

On first launch, if no config exists, prompt the user for city/coordinates (or attempt IP-based geolocation as a fallback) and write the config file.

---

## Caching

### Prayer Times Cache
- Path: `~/.cache/salah/prayer-times-YYYY-MM-DD.json`
- On launch: check if today's cache file exists. If yes, load it. If no, fetch from Aladhan API and write cache.
- Old cache files (not today's) are deleted on launch to avoid accumulation.

### Ayah of the Day
- Path: `~/.cache/salah/ayah-YYYY-MM-DD.json`
- Same pattern: daily cache, deterministic selection seeded by date.
- Total Quran ayah count: 6236. Selection: `dayOfYear % 6236` (or similar stable formula).

---

## API Details

### Aladhan — Prayer Times

```
GET https://api.aladhan.com/v1/timings/{timestamp}?latitude={lat}&longitude={lng}&method={method}
```

Method codes (pass as integer):
- ISNA = 2
- MWL = 3
- Egyptian = 5
- UmmAlQura = 4

Response fields used: `Fajr`, `Sunrise`, `Dhuhr`, `Asr`, `Maghrib`, `Isha` from `data.timings`.

### alquran.cloud — Ayah

```
GET https://api.alquran.cloud/v1/ayah/{surah}:{ayah}/editions/quran-uthmani,en.asad
```

Returns both Arabic (uthmani script) and English translation (Muhammad Asad) in one call.

For ayah of the day, use the calculated index:
```
GET https://api.alquran.cloud/v1/ayah/{index}/editions/quran-uthmani,en.asad
```

---

## Project Structure

```
salah-tui/
├── main.go
├── go.mod
├── go.sum
├── internal/
│   ├── api/
│   │   ├── aladhan.go       # Prayer times fetch + response parsing
│   │   └── quran.go         # Ayah fetch + response parsing
│   ├── cache/
│   │   └── cache.go         # Read/write/invalidate cache files
│   ├── config/
│   │   └── config.go        # TOML load/write, first-run setup
│   ├── model/
│   │   └── model.go         # Bubbletea model: state, Init, Update, View
│   ├── ui/
│   │   ├── dashboard.go     # Dashboard layout composition
│   │   ├── schedule.go      # Left pane: prayer schedule rendering
│   │   ├── ayah.go          # Right pane: ayah rendering
│   │   ├── minimal.go       # Minimal mode rendering
│   │   └── styles.go        # All lipgloss styles and color palette
│   └── prayer/
│       └── prayer.go        # Prayer time logic: next prayer, countdown, highlighting
└── README.md
```

---

## Color Palette (lipgloss)

Use a dark-background-friendly palette:

| Element | Color |
|---------|-------|
| Header / Hijri date | Gold `#C9A84C` |
| Current prayer | Green `#4CAF50` |
| Next prayer label | Cyan `#00BCD4` |
| Countdown — normal | White |
| Countdown — ≤15 min | Amber `#FFA726` |
| Countdown — ≤5 min | Red `#EF5350` |
| Ayah Arabic text | White, slightly larger via padding |
| Ayah translation | Dim white `#AAAAAA` |
| Ayah reference | Gold `#C9A84C` |
| Border | Subtle gray `#444444` |
| Help overlay | Dark background with white text |

---

## Error Handling

- If API call fails and no cache exists: display an error message in the affected pane with a retry hint (`r` to retry). Do not crash.
- If cache exists but API is unreachable: use stale cache silently, show a small "offline" indicator in the header.
- If config is missing or malformed: re-run first-launch setup.
- If terminal is too narrow for dashboard: fall back to minimal layout automatically with a notice.

---

## First-Run Experience

1. Check for `~/.config/salah/config.toml`
2. If missing, print a simple setup prompt (not TUI — plain terminal):
   - Ask for city name or lat/lng
   - Attempt to resolve coordinates via a geocoding API (nominatim.openstreetmap.org — free, no auth) if city name is given
   - Ask for preferred calculation method (show numbered list, default ISNA)
   - Write config file
3. Launch dashboard

---

## README Requirements

Include:
- Install instructions (`go install` + binary release)
- Screenshot placeholder
- Config file reference
- Keybindings table
- Supported calculation methods
- Credits to Aladhan and alquran.cloud APIs

---

## Out of Scope (for now)

- System / OS notifications (future `--daemon` mode)
- Qibla direction
- Hijri calendar beyond the date header
- Multiple location profiles
- Windows support (Linux + macOS only)