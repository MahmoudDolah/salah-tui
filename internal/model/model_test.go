package model

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/MahmoudDolah/salah-tui/internal/api"
	"github.com/MahmoudDolah/salah-tui/internal/config"
	"github.com/MahmoudDolah/salah-tui/internal/prayer"
)

// testModel builds a model with a fixed set of prayers for deterministic tests.
func testModel() Model {
	cfg := &config.Config{
		Location:    config.LocationConfig{Timezone: "UTC"},
		Calculation: config.CalculationConfig{Method: "ISNA"},
		Display:     config.DisplayConfig{TimeFormat: 12, ShowSunrise: false},
	}
	prayers := []prayer.Prayer{
		{Name: "Fajr", Time: time.Date(2025, 6, 6, 5, 30, 0, 0, time.UTC)},
		{Name: "Dhuhr", Time: time.Date(2025, 6, 6, 12, 30, 0, 0, time.UTC)},
		{Name: "Asr", Time: time.Date(2025, 6, 6, 15, 45, 0, 0, time.UTC)},
		{Name: "Maghrib", Time: time.Date(2025, 6, 6, 18, 0, 0, 0, time.UTC)},
		{Name: "Isha", Time: time.Date(2025, 6, 6, 19, 30, 0, 0, time.UTC)},
	}
	return New(cfg, time.UTC, prayers, "06 Rajab 1446", "06 Jun 2025", nil, false)
}

// update is a convenience wrapper that casts the returned tea.Model back to Model.
func update(m Model, msg tea.Msg) (Model, tea.Cmd) {
	updated, cmd := m.Update(msg)
	return updated.(Model), cmd
}

// keyMsg builds a rune-based key message.
func keyMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

// --- New ---

func TestNew_InitialState(t *testing.T) {
	m := testModel()
	if len(m.prayers) != 5 {
		t.Errorf("expected 5 prayers, got %d", len(m.prayers))
	}
	if m.hijriDate != "06 Rajab 1446" {
		t.Errorf("unexpected hijriDate: %q", m.hijriDate)
	}
	if m.lookupMode {
		t.Error("expected lookupMode=false on init")
	}
	if m.showHelp {
		t.Error("expected showHelp=false on init")
	}
	if m.offline {
		t.Error("expected offline=false on init")
	}
}

// --- WindowSizeMsg ---

func TestUpdate_WindowSize(t *testing.T) {
	m, _ := update(testModel(), tea.WindowSizeMsg{Width: 120, Height: 40})
	if m.termWidth != 120 || m.termHeight != 40 {
		t.Errorf("expected 120x40, got %dx%d", m.termWidth, m.termHeight)
	}
}

// --- Tick ---

func TestUpdate_Tick_ReturnsCmdNotNil(t *testing.T) {
	_, cmd := update(testModel(), tickMsg(time.Now()))
	if cmd == nil {
		t.Error("expected non-nil cmd from tick")
	}
}

// --- Normal mode keys ---

func TestUpdate_Key_Q_Quits(t *testing.T) {
	_, cmd := update(testModel(), keyMsg("q"))
	if cmd == nil {
		t.Fatal("expected quit cmd, got nil")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("expected QuitMsg from q key")
	}
}

func TestUpdate_Key_CtrlC_Quits(t *testing.T) {
	_, cmd := update(testModel(), tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit cmd, got nil")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("expected QuitMsg from ctrl+c")
	}
}

func TestUpdate_Key_R_SetsLoading(t *testing.T) {
	m, cmd := update(testModel(), keyMsg("r"))
	if !m.loading {
		t.Error("expected loading=true after r key")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd from r key")
	}
}

func TestUpdate_Key_R_ClearsOffline(t *testing.T) {
	base := testModel()
	base.offline = true
	m, _ := update(base, keyMsg("r"))
	if m.offline {
		t.Error("expected offline to be cleared on refresh")
	}
}

func TestUpdate_Key_Slash_OpensLookup(t *testing.T) {
	m, _ := update(testModel(), keyMsg("/"))
	if !m.lookupMode {
		t.Error("expected lookupMode=true after / key")
	}
	if m.lookupQuery != "" {
		t.Error("expected empty query on open")
	}
}

func TestUpdate_Key_Question_TogglesHelp(t *testing.T) {
	m, _ := update(testModel(), keyMsg("?"))
	if !m.showHelp {
		t.Error("expected showHelp=true after first ?")
	}
	m, _ = update(m, keyMsg("?"))
	if m.showHelp {
		t.Error("expected showHelp=false after second ?")
	}
}

func TestUpdate_Key_Esc_ClearsHelp(t *testing.T) {
	base := testModel()
	base.showHelp = true
	m, _ := update(base, tea.KeyMsg{Type: tea.KeyEsc})
	if m.showHelp {
		t.Error("expected showHelp=false after esc")
	}
}

// --- Lookup mode keys ---

func openLookup(m Model) Model {
	updated, _ := update(m, keyMsg("/"))
	return updated
}

func TestUpdate_Lookup_Esc_ClosesLookup(t *testing.T) {
	m := openLookup(testModel())
	m.lookupQuery = "2:25"
	m.lookupResult = &api.Ayah{}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.lookupMode {
		t.Error("expected lookupMode=false after esc")
	}
	if m.lookupQuery != "" {
		t.Error("expected query cleared after esc")
	}
	if m.lookupResult != nil {
		t.Error("expected result cleared after esc")
	}
}

func TestUpdate_Lookup_DigitsAppended(t *testing.T) {
	m := openLookup(testModel())
	m, _ = update(m, keyMsg("2"))
	m, _ = update(m, keyMsg(":"))
	m, _ = update(m, keyMsg("2"))
	m, _ = update(m, keyMsg("5"))
	m, _ = update(m, keyMsg("5"))
	if m.lookupQuery != "2:255" {
		t.Errorf("expected query=2:255, got %q", m.lookupQuery)
	}
}

func TestUpdate_Lookup_LettersIgnored(t *testing.T) {
	m := openLookup(testModel())
	m, _ = update(m, keyMsg("a"))
	m, _ = update(m, keyMsg("x"))
	if m.lookupQuery != "" {
		t.Errorf("expected letters to be ignored, got query=%q", m.lookupQuery)
	}
}

func TestUpdate_Lookup_Backspace(t *testing.T) {
	m := openLookup(testModel())
	m.lookupQuery = "2:25"
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyBackspace})
	if m.lookupQuery != "2:2" {
		t.Errorf("expected 2:2 after backspace, got %q", m.lookupQuery)
	}
}

func TestUpdate_Lookup_BackspaceOnEmpty(t *testing.T) {
	m := openLookup(testModel())
	// Should not panic
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyBackspace})
	if m.lookupQuery != "" {
		t.Error("expected empty query unchanged")
	}
}

func TestUpdate_Lookup_CtrlC_QuitsEvenInLookupMode(t *testing.T) {
	m := openLookup(testModel())
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit cmd from ctrl+c in lookup mode, got nil")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("expected QuitMsg from ctrl+c in lookup mode")
	}
}

func TestUpdate_Lookup_Enter_ReturnsCmdNotNil(t *testing.T) {
	m := openLookup(testModel())
	m.lookupQuery = "bad"
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("expected non-nil cmd on enter in lookup mode")
	}
}

// --- Incoming messages ---

func TestUpdate_PrayerTimesLoaded(t *testing.T) {
	newPrayers := []prayer.Prayer{
		{Name: "Fajr", Time: time.Date(2025, 6, 7, 5, 30, 0, 0, time.UTC)},
	}
	m, _ := update(testModel(), prayerTimesLoadedMsg{
		prayers:       newPrayers,
		hijriDate:     "07 Rajab 1446",
		gregorianDate: "07 Jun 2025",
		offline:       true,
	})
	if len(m.prayers) != 1 {
		t.Errorf("expected 1 prayer after load, got %d", len(m.prayers))
	}
	if m.hijriDate != "07 Rajab 1446" {
		t.Errorf("unexpected hijriDate: %q", m.hijriDate)
	}
	if !m.offline {
		t.Error("expected offline=true from message")
	}
	if m.loading {
		t.Error("expected loading=false after prayers loaded")
	}
}

func TestUpdate_AyahLoaded_UpdatesAyah(t *testing.T) {
	ayah := &api.Ayah{Arabic: "test", Translation: "test", Reference: "1:1"}
	m, _ := update(testModel(), ayahLoadedMsg{ayah: ayah})
	if m.ayah != ayah {
		t.Error("expected ayah to be updated")
	}
}

func TestUpdate_AyahLoaded_NilDoesNotClearExisting(t *testing.T) {
	existing := &api.Ayah{Arabic: "existing"}
	base := testModel()
	base.ayah = existing
	m, _ := update(base, ayahLoadedMsg{ayah: nil})
	if m.ayah != existing {
		t.Error("nil ayah message should not overwrite existing ayah")
	}
}

func TestUpdate_AyahLoaded_SetsOffline(t *testing.T) {
	m, _ := update(testModel(), ayahLoadedMsg{ayah: nil, offline: true})
	if !m.offline {
		t.Error("expected offline=true from ayah offline message")
	}
}

func TestUpdate_LookupResult_SetsResult(t *testing.T) {
	ayah := &api.Ayah{Reference: "2:255"}
	m, _ := update(testModel(), lookupResultMsg{ayah: ayah})
	if m.lookupResult != ayah {
		t.Error("expected lookupResult to be set")
	}
	if m.lookupErr != "" {
		t.Error("expected no error")
	}
}

func TestUpdate_LookupResult_SetsError(t *testing.T) {
	m, _ := update(testModel(), lookupResultMsg{err: "not found"})
	if m.lookupErr != "not found" {
		t.Errorf("expected lookupErr=%q, got %q", "not found", m.lookupErr)
	}
}

func TestUpdate_ErrMsg_SetsFatalErr(t *testing.T) {
	base := testModel()
	base.loading = true
	m, _ := update(base, errMsg{err: fmt.Errorf("something broke")})
	if m.fatalErr == "" {
		t.Error("expected fatalErr to be set")
	}
	if m.loading {
		t.Error("expected loading=false after error")
	}
}

// --- View ---

func TestView_LoadingWhenNoPrayers(t *testing.T) {
	m := testModel()
	m.prayers = nil
	if !strings.Contains(m.View(), "Loading") {
		t.Error("expected loading indicator when prayers is nil")
	}
}

func TestView_ShowsFatalError(t *testing.T) {
	m := testModel()
	m.fatalErr = "connection refused"
	out := m.View()
	if !strings.Contains(out, "connection refused") {
		t.Errorf("expected fatal error in view, got: %q", out)
	}
}

// --- cmdLookup validation (tested without network) ---

func TestCmdLookup_BadFormat(t *testing.T) {
	cmd := cmdLookup("badformat")
	msg, ok := cmd().(lookupResultMsg)
	if !ok {
		t.Fatal("expected lookupResultMsg")
	}
	if msg.err == "" {
		t.Error("expected error for missing colon")
	}
}

func TestCmdLookup_InvalidSurah_Zero(t *testing.T) {
	cmd := cmdLookup("0:1")
	msg := cmd().(lookupResultMsg)
	if msg.err == "" {
		t.Error("expected error for surah=0")
	}
}

func TestCmdLookup_InvalidSurah_TooHigh(t *testing.T) {
	cmd := cmdLookup("115:1")
	msg := cmd().(lookupResultMsg)
	if msg.err == "" {
		t.Error("expected error for surah=115")
	}
}

func TestCmdLookup_InvalidAyah_Zero(t *testing.T) {
	cmd := cmdLookup("2:0")
	msg := cmd().(lookupResultMsg)
	if msg.err == "" {
		t.Error("expected error for ayah=0")
	}
}

func TestCmdLookup_InvalidNotNumbers(t *testing.T) {
	cmd := cmdLookup("abc:def")
	msg := cmd().(lookupResultMsg)
	if msg.err == "" {
		t.Error("expected error for non-numeric reference")
	}
}
