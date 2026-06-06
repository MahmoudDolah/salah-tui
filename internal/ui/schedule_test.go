package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/MahmoudDolah/salah-tui/internal/prayer"
)

// testNow is between Sunrise and Dhuhr: current=Sunrise, next=Dhuhr.
var testNow = time.Date(2025, 6, 6, 10, 0, 0, 0, time.UTC)

var testPrayers = []prayer.Prayer{
	{Name: "Fajr", Time: time.Date(2025, 6, 6, 5, 30, 0, 0, time.UTC)},
	{Name: "Sunrise", Time: time.Date(2025, 6, 6, 7, 0, 0, 0, time.UTC)},
	{Name: "Dhuhr", Time: time.Date(2025, 6, 6, 12, 30, 0, 0, time.UTC)},
	{Name: "Asr", Time: time.Date(2025, 6, 6, 15, 45, 0, 0, time.UTC)},
	{Name: "Maghrib", Time: time.Date(2025, 6, 6, 18, 0, 0, 0, time.UTC)},
	{Name: "Isha", Time: time.Date(2025, 6, 6, 19, 30, 0, 0, time.UTC)},
}

func renderScheduleStripped(offline bool) string {
	out := RenderSchedule(testPrayers, "06 Rajab 1446", "06 Jun 2025", testNow, true, offline, "", 40)
	return stripANSI(out)
}

func TestRenderSchedule_DateHeader(t *testing.T) {
	out := renderScheduleStripped(false)
	if !strings.Contains(out, "06 Rajab 1446") {
		t.Error("expected Hijri date in output")
	}
	if !strings.Contains(out, "06 Jun 2025") {
		t.Error("expected Gregorian date in output")
	}
}

func TestRenderSchedule_OfflineIndicator(t *testing.T) {
	withOffline := renderScheduleStripped(true)
	if !strings.Contains(withOffline, "(offline)") {
		t.Error("expected offline indicator when offline=true")
	}
	withoutOffline := renderScheduleStripped(false)
	if strings.Contains(withoutOffline, "(offline)") {
		t.Error("unexpected offline indicator when offline=false")
	}
}

func TestRenderSchedule_NextPrayerMarker(t *testing.T) {
	out := renderScheduleStripped(false)
	// Dhuhr is next at testNow=10:00
	if !strings.Contains(out, "▶") {
		t.Error("expected ▶ marker for next prayer")
	}
	if !strings.Contains(out, "Dhuhr") {
		t.Error("expected Dhuhr in output")
	}
}

func TestRenderSchedule_CountdownPresent(t *testing.T) {
	out := renderScheduleStripped(false)
	// Countdown for Dhuhr (12:30 - 10:00 = 2h30m) should appear as HH:MM:SS
	if !strings.Contains(out, "02:30:00") {
		t.Errorf("expected countdown 02:30:00 in output, got:\n%s", out)
	}
}

func TestRenderSchedule_AllPrayersListed(t *testing.T) {
	out := renderScheduleStripped(false)
	for _, name := range []string{"Fajr", "Sunrise", "Dhuhr", "Asr", "Maghrib", "Isha"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %s in output", name)
		}
	}
}

func TestRenderSchedule_NoMarkerOnPassedPrayers(t *testing.T) {
	out := renderScheduleStripped(false)
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		// Fajr has passed and is not current — must not have ▶
		if strings.Contains(line, "Fajr") && strings.Contains(line, "▶") {
			t.Error("Fajr (past, not current) should not have ▶ marker")
		}
	}
}

func TestRenderSchedule_AllPrayersCompleted(t *testing.T) {
	afterIsha := time.Date(2025, 6, 6, 23, 0, 0, 0, time.UTC)
	out := stripANSI(RenderSchedule(testPrayers, "06 Rajab 1446", "06 Jun 2025", afterIsha, true, false, "", 40))
	// No ▶ marker — all prayers done
	if strings.Contains(out, "▶") {
		t.Error("expected no ▶ marker after all prayers completed")
	}
}

func TestRenderSchedule_BeforeFajr(t *testing.T) {
	beforeFajr := time.Date(2025, 6, 6, 3, 0, 0, 0, time.UTC)
	out := stripANSI(RenderSchedule(testPrayers, "06 Rajab 1446", "06 Jun 2025", beforeFajr, true, false, "", 40))
	// Fajr should be next
	if !strings.Contains(out, "▶") {
		t.Error("expected ▶ marker before Fajr")
	}
	if !strings.Contains(out, "Fajr") {
		t.Error("expected Fajr in output")
	}
}

func TestRenderSchedule_ErrorBanner(t *testing.T) {
	out := stripANSI(RenderSchedule(testPrayers, "06 Rajab 1446", "06 Jun 2025", testNow, true, false, "connection refused", 40))
	if !strings.Contains(out, "connection refused") {
		t.Error("expected error message in schedule pane")
	}
	if !strings.Contains(out, "Press r to retry") {
		t.Error("expected retry hint in schedule pane")
	}
}

func TestRenderSchedule_NoErrorBannerByDefault(t *testing.T) {
	out := renderScheduleStripped(false)
	if strings.Contains(out, "retry") {
		t.Error("unexpected retry hint when no error")
	}
}
