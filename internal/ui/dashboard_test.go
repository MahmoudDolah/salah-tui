package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/MahmoudDolah/salah-tui/internal/api"
	"github.com/MahmoudDolah/salah-tui/internal/prayer"
)

var dashPrayers = []prayer.Prayer{
	{Name: "Fajr", Time: time.Date(2025, 6, 6, 5, 30, 0, 0, time.UTC)},
	{Name: "Dhuhr", Time: time.Date(2025, 6, 6, 12, 30, 0, 0, time.UTC)},
	{Name: "Asr", Time: time.Date(2025, 6, 6, 15, 45, 0, 0, time.UTC)},
	{Name: "Maghrib", Time: time.Date(2025, 6, 6, 18, 0, 0, 0, time.UTC)},
	{Name: "Isha", Time: time.Date(2025, 6, 6, 19, 30, 0, 0, time.UTC)},
}

var dashAyah = &api.Ayah{
	Arabic:      "بِسْمِ اللَّهِ",
	Translation: "In the name of God",
	Reference:   "Al-Fatiha 1:1",
}

var dashNow = time.Date(2025, 6, 6, 10, 0, 0, 0, time.UTC)

func renderDashStripped(w, h int, lookupMode, showHelp bool) string {
	out := RenderDashboard(
		dashPrayers, "06 Rajab 1446", "06 Jun 2025",
		dashAyah, dashNow,
		true, false,
		lookupMode, "", nil, "",
		showHelp, "", "", w, h,
	)
	return stripANSI(out)
}

func TestRenderDashboard_NarrowFallback(t *testing.T) {
	out := stripANSI(RenderDashboard(
		dashPrayers, "06 Rajab 1446", "06 Jun 2025",
		dashAyah, dashNow, true, false,
		false, "", nil, "", false, "", "",
		minDashboardWidth-1, 40, // 61 cols — just below the two-pane threshold
	))
	if !strings.Contains(out, "too narrow") {
		t.Errorf("expected narrow-terminal notice, got:\n%s", out)
	}
}

func TestRenderDashboard_NormalContainsPrayers(t *testing.T) {
	out := renderDashStripped(120, 40, false, false)
	for _, name := range []string{"Fajr", "Dhuhr", "Asr", "Maghrib", "Isha"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %s in dashboard output", name)
		}
	}
}

func TestRenderDashboard_NormalContainsAyah(t *testing.T) {
	out := renderDashStripped(120, 40, false, false)
	if !strings.Contains(out, "In the name of God") {
		t.Error("expected ayah translation in dashboard output")
	}
}

func TestRenderDashboard_LookupModeShowsPrompt(t *testing.T) {
	out := renderDashStripped(120, 40, true, false)
	if !strings.Contains(out, "Quran Lookup") {
		t.Error("expected lookup prompt in dashboard when lookupMode=true")
	}
	// Normal ayah pane should not be shown in lookup mode
	if strings.Contains(out, "Ayah of the Day") {
		t.Error("expected ayah header to be hidden in lookup mode")
	}
}

func TestRenderDashboard_HelpOverlay(t *testing.T) {
	out := renderDashStripped(120, 40, false, true)
	if !strings.Contains(out, "Keybindings") {
		t.Error("expected help overlay when showHelp=true")
	}
	if !strings.Contains(out, "ctrl+c") {
		t.Error("expected keybindings listed in help overlay")
	}
}

func TestRenderDashboard_NoHelpByDefault(t *testing.T) {
	out := renderDashStripped(120, 40, false, false)
	if strings.Contains(out, "Keybindings") {
		t.Error("expected no help overlay by default")
	}
}

func TestRenderNarrow_NextPrayerShown(t *testing.T) {
	out := stripANSI(renderNarrow(dashPrayers, dashNow, true, 40))
	if !strings.Contains(out, "Dhuhr") {
		t.Errorf("expected Dhuhr (next prayer) in narrow view, got:\n%s", out)
	}
}

func TestRenderNarrow_AllPrayersCompleted(t *testing.T) {
	afterIsha := time.Date(2025, 6, 6, 23, 0, 0, 0, time.UTC)
	out := stripANSI(renderNarrow(dashPrayers, afterIsha, true, 40))
	if !strings.Contains(out, "All prayers completed") {
		t.Errorf("expected completion message, got:\n%s", out)
	}
}

func TestRenderHelp_ContainsAllKeys(t *testing.T) {
	out := stripANSI(renderHelp())
	for _, key := range []string{"q / ctrl+c", "/", "esc", "r", "?"} {
		if !strings.Contains(out, key) {
			t.Errorf("expected key %q in help output", key)
		}
	}
}
