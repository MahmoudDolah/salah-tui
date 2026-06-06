package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/MahmoudDolah/salah-tui/internal/api"
	"github.com/MahmoudDolah/salah-tui/internal/cache"
	"github.com/MahmoudDolah/salah-tui/internal/config"
	"github.com/MahmoudDolah/salah-tui/internal/prayer"
	"github.com/MahmoudDolah/salah-tui/internal/ui"
)

// --- messages ---

type tickMsg time.Time

type prayerTimesLoadedMsg struct {
	prayers       []prayer.Prayer
	hijriDate     string
	gregorianDate string
	offline       bool
}

type ayahLoadedMsg struct {
	ayah    *api.Ayah
	offline bool
}

type lookupResultMsg struct {
	ayah *api.Ayah
	err  string
}

type errMsg struct{ err error }

// --- model ---

type Model struct {
	cfg           *config.Config
	loc           *time.Location
	use12h        bool

	prayers       []prayer.Prayer
	hijriDate     string
	gregorianDate string
	ayah          *api.Ayah
	offline       bool

	lookupMode    bool
	lookupQuery   string
	lookupResult  *api.Ayah
	lookupErr     string

	showHelp      bool
	loading       bool
	fatalErr      string

	termWidth     int
	termHeight    int
}

// New constructs the initial model from already-loaded data.
func New(
	cfg *config.Config,
	loc *time.Location,
	prayers []prayer.Prayer,
	hijriDate, gregorianDate string,
	ayah *api.Ayah,
	offline bool,
) Model {
	return Model{
		cfg:           cfg,
		loc:           loc,
		use12h:        cfg.Display.TimeFormat == 12,
		prayers:       prayers,
		hijriDate:     hijriDate,
		gregorianDate: gregorianDate,
		ayah:          ayah,
		offline:       offline,
	}
}

func (m Model) Init() tea.Cmd {
	return tick()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height

	case tickMsg:
		return m, tick()

	case tea.KeyMsg:
		return m.handleKey(msg)

	case prayerTimesLoadedMsg:
		m.prayers = msg.prayers
		m.hijriDate = msg.hijriDate
		m.gregorianDate = msg.gregorianDate
		m.offline = msg.offline
		m.loading = false

	case ayahLoadedMsg:
		if msg.ayah != nil {
			m.ayah = msg.ayah
		}
		if msg.offline {
			m.offline = true
		}

	case lookupResultMsg:
		m.lookupResult = msg.ayah
		m.lookupErr = msg.err

	case errMsg:
		m.fatalErr = msg.err.Error()
		m.loading = false
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if m.lookupMode {
		return m.handleLookupKey(msg)
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "r":
		m.loading = true
		m.offline = false
		return m, tea.Batch(cmdRefreshPrayers(m.cfg, m.loc), cmdRefreshAyah())

	case "/":
		m.lookupMode = true
		m.lookupQuery = ""
		m.lookupResult = nil
		m.lookupErr = ""

	case "?":
		m.showHelp = !m.showHelp

	case "esc":
		m.showHelp = false
	}

	return m, nil
}

func (m Model) handleLookupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.lookupMode = false
		m.lookupQuery = ""
		m.lookupResult = nil
		m.lookupErr = ""

	case "enter":
		return m, cmdLookup(m.lookupQuery)

	case "backspace":
		if len(m.lookupQuery) > 0 {
			m.lookupQuery = m.lookupQuery[:len(m.lookupQuery)-1]
		}

	default:
		// Accept digits and ':'
		for _, r := range msg.String() {
			if (r >= '0' && r <= '9') || r == ':' {
				m.lookupQuery += string(r)
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.fatalErr != "" {
		return ui.StyleError.Render("Error: "+m.fatalErr) + "\n"
	}
	if len(m.prayers) == 0 {
		return "Loading…\n"
	}

	now := time.Now().In(m.loc)

	return ui.RenderDashboard(
		m.prayers,
		m.hijriDate,
		m.gregorianDate,
		m.ayah,
		now,
		m.use12h,
		m.offline,
		m.lookupMode,
		m.lookupQuery,
		m.lookupResult,
		m.lookupErr,
		m.showHelp,
		m.termWidth,
		m.termHeight,
	)
}

// --- commands ---

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func cmdRefreshPrayers(cfg *config.Config, loc *time.Location) tea.Cmd {
	return func() tea.Msg {
		today := time.Now()
		methodCode, ok := config.MethodCodes[cfg.Calculation.Method]
		if !ok {
			methodCode = config.MethodCodes["ISNA"]
		}

		pt, err := api.Fetch(cfg.Location.Latitude, cfg.Location.Longitude, methodCode)
		offline := false
		if err != nil {
			// Try cache fallback
			cached, cerr := cache.ReadPrayerTimes(today)
			if cerr != nil || cached == nil {
				return errMsg{fmt.Errorf("fetch prayer times: %w", err)}
			}
			var cachedPT api.PrayerTimes
			if jerr := json.Unmarshal(cached, &cachedPT); jerr != nil {
				return errMsg{fmt.Errorf("fetch prayer times: %w", err)}
			}
			pt = &cachedPT
			offline = true
		} else {
			// Write fresh cache
			if data, merr := json.Marshal(pt); merr == nil {
				_ = cache.WritePrayerTimes(today, data)
			}
		}

		prayers, perr := prayer.ParseTimes(pt, today, loc, cfg.Display.ShowSunrise)
		if perr != nil {
			return errMsg{perr}
		}

		return prayerTimesLoadedMsg{
			prayers:       prayers,
			hijriDate:     pt.HijriDate,
			gregorianDate: pt.GregorianDate,
			offline:       offline,
		}
	}
}

func cmdRefreshAyah() tea.Cmd {
	return func() tea.Msg {
		today := time.Now()
		index := api.DailyIndex(today)
		ayah, err := api.FetchByIndex(index)
		offline := false
		if err != nil {
			cached, cerr := cache.ReadAyah(today)
			if cerr != nil || cached == nil {
				return ayahLoadedMsg{offline: true}
			}
			var cachedAyah api.Ayah
			if jerr := json.Unmarshal(cached, &cachedAyah); jerr != nil {
				return ayahLoadedMsg{offline: true}
			}
			ayah = &cachedAyah
			offline = true
		} else {
			if data, merr := json.Marshal(ayah); merr == nil {
				_ = cache.WriteAyah(today, data)
			}
		}
		return ayahLoadedMsg{ayah: ayah, offline: offline}
	}
}

func cmdLookup(query string) tea.Cmd {
	return func() tea.Msg {
		parts := strings.SplitN(query, ":", 2)
		if len(parts) != 2 {
			return lookupResultMsg{err: "format: surah:ayah (e.g. 2:255)"}
		}
		surah, err1 := strconv.Atoi(parts[0])
		ayahNum, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || surah < 1 || surah > 114 || ayahNum < 1 {
			return lookupResultMsg{err: "invalid reference"}
		}
		result, err := api.FetchByRef(surah, ayahNum)
		if err != nil {
			return lookupResultMsg{err: err.Error()}
		}
		return lookupResultMsg{ayah: result}
	}
}
