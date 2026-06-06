# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**salah-tui** is a Go terminal UI for tracking daily Muslim prayer times. Built with [bubbletea](https://github.com/charmbracelet/bubbletea) (Elm-like state machine) and [lipgloss](https://github.com/charmbracelet/lipgloss) for styling.

## Commands

```sh
go build ./...          # build
go run main.go          # run dashboard
go run main.go --minimal --watch  # minimal/tmux mode
go test ./...           # run all tests
go test ./internal/...  # run a specific package's tests
golangci-lint run       # lint (if configured)
```

## Architecture

The app follows the bubbletea Model-Update-View pattern. All application state lives in `internal/model/model.go`. The `View()` method delegates to UI renderers in `internal/ui/`. Network I/O and caching happen outside the render path.

**Data flow:**
1. On startup: load config (`internal/config/`) → check cache (`internal/cache/`) → fetch APIs if stale (`internal/api/`) → populate model
2. Every second: tick message updates the countdown in the model
3. `r` keypress: sends a refresh command, re-fetches APIs

**Key packages:**

| Package | Responsibility |
|---|---|
| `internal/api/aladhan.go` | Fetch/parse prayer times from [Aladhan API](https://aladhan.com/prayer-times-api) |
| `internal/api/quran.go` | Fetch/parse ayah from [alquran.cloud API](https://alquran.cloud/api) |
| `internal/cache/` | File-based JSON cache under `~/.cache/salah/`; keyed by date |
| `internal/config/` | TOML config at `~/.config/salah/config.toml`; triggers first-run setup if missing |
| `internal/prayer/` | Pure logic: next prayer calculation, countdown formatting |
| `internal/ui/` | Lipgloss renderers: `dashboard.go` (2-pane layout), `schedule.go` (left pane), `ayah.go` (right pane), `styles.go` (color palette + countdown color thresholds) |
| `internal/model/` | Bubbletea `Model` struct; `Init`, `Update`, `View` |

## Modes

- **Dashboard** (default): split-pane with full prayer schedule (left) + ayah of the day (right)
- **Minimal** (`--minimal`): single-line `"NextPrayer (12:30 PM) in HH:MM:SS"`, suitable for tmux; `--watch` enables auto-refresh every second with midnight reload
- **Quran lookup**: press `/` inside the dashboard to search by `surah:ayah` reference

## Caching

Cache files live at `~/.cache/salah/` and are keyed by date:
- `prayer-times-YYYY-MM-DD.json`
- `ayah-YYYY-MM-DD.json`

Old cache files are deleted on launch. When offline and a cache file exists, the app uses stale data and shows an offline indicator rather than erroring.

## Countdown color thresholds

Defined in `internal/ui/styles.go` (`CountdownStyle`) and applied in `internal/ui/schedule.go`:
- `> 15 min` → White
- `≤ 15 min` → Amber `#FFA726`
- `≤ 5 min` → Red `#EF5350`

## Ayah of the day

The daily ayah is seeded deterministically by `api.DailyIndex(date)`, which maps `dayOfYear` to a global ayah index via `((dayOfYear - 1) % 6236) + 1` (range 1–366, stable within a calendar day).
