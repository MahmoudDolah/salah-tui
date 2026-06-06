package prayer

import (
	"testing"
	"time"

	"github.com/MahmoudDolah/salah-tui/internal/api"
)

var testPT = &api.PrayerTimes{
	Fajr:    "05:30",
	Sunrise: "07:00",
	Dhuhr:   "12:30",
	Asr:     "15:45",
	Maghrib: "18:00",
	Isha:    "19:30",
}

var testDate = time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC)

func TestParseTimes_HappyPath(t *testing.T) {
	loc := time.UTC
	prayers, err := ParseTimes(testPT, testDate, loc, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prayers) != 6 {
		t.Fatalf("expected 6 prayers, got %d", len(prayers))
	}
	names := []string{"Fajr", "Sunrise", "Dhuhr", "Asr", "Maghrib", "Isha"}
	for i, want := range names {
		if prayers[i].Name != want {
			t.Errorf("prayers[%d].Name = %q, want %q", i, prayers[i].Name, want)
		}
	}
	if prayers[0].Time.Hour() != 5 || prayers[0].Time.Minute() != 30 {
		t.Errorf("Fajr time = %v, want 05:30", prayers[0].Time)
	}
	if prayers[2].Time.Hour() != 12 || prayers[2].Time.Minute() != 30 {
		t.Errorf("Dhuhr time = %v, want 12:30", prayers[2].Time)
	}
}

func TestParseTimes_NoSunrise(t *testing.T) {
	prayers, err := ParseTimes(testPT, testDate, time.UTC, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prayers) != 5 {
		t.Fatalf("expected 5 prayers (no Sunrise), got %d", len(prayers))
	}
	for _, p := range prayers {
		if p.Name == "Sunrise" {
			t.Error("Sunrise should not be included when showSunrise=false")
		}
	}
}

func TestParseTimes_InvalidTime(t *testing.T) {
	bad := &api.PrayerTimes{
		Fajr:    "not-a-time",
		Sunrise: "07:00",
		Dhuhr:   "12:30",
		Asr:     "15:45",
		Maghrib: "18:00",
		Isha:    "19:30",
	}
	_, err := ParseTimes(bad, testDate, time.UTC, false)
	if err == nil {
		t.Error("expected an error for invalid time string, got nil")
	}
}

func makePrayers(loc *time.Location) []Prayer {
	now := time.Now().In(loc)
	y, m, d := now.Date()
	at := func(h, min int) time.Time {
		return time.Date(y, m, d, h, min, 0, 0, loc)
	}
	return []Prayer{
		{Name: "Fajr", Time: at(5, 30)},
		{Name: "Dhuhr", Time: at(12, 30)},
		{Name: "Asr", Time: at(15, 45)},
		{Name: "Maghrib", Time: at(18, 0)},
		{Name: "Isha", Time: at(19, 30)},
	}
}

func TestNext_BeforeFajr(t *testing.T) {
	prayers := makePrayers(time.UTC)
	now := prayers[0].Time.Add(-1 * time.Hour) // before Fajr
	next := Next(prayers, now)
	if next == nil || next.Name != "Fajr" {
		t.Errorf("expected Fajr, got %v", next)
	}
}

func TestNext_CorrectNextPrayer(t *testing.T) {
	prayers := makePrayers(time.UTC)
	// After Dhuhr, before Asr
	now := prayers[1].Time.Add(30 * time.Minute)
	next := Next(prayers, now)
	if next == nil || next.Name != "Asr" {
		t.Errorf("expected Asr, got %v", next)
	}
}

func TestNext_PostIsha(t *testing.T) {
	prayers := makePrayers(time.UTC)
	now := prayers[4].Time.Add(1 * time.Hour) // after Isha
	next := Next(prayers, now)
	if next != nil {
		t.Errorf("expected nil after Isha, got %v", next)
	}
}

func TestCurrentPrayer_BeforeFajr(t *testing.T) {
	prayers := makePrayers(time.UTC)
	now := prayers[0].Time.Add(-1 * time.Hour)
	cur := CurrentPrayer(prayers, now)
	if cur != nil {
		t.Errorf("expected nil before Fajr, got %v", cur)
	}
}

func TestCurrentPrayer_MidDay(t *testing.T) {
	prayers := makePrayers(time.UTC)
	// After Dhuhr, before Asr
	now := prayers[1].Time.Add(30 * time.Minute)
	cur := CurrentPrayer(prayers, now)
	if cur == nil || cur.Name != "Dhuhr" {
		t.Errorf("expected Dhuhr, got %v", cur)
	}
}

func TestCurrentPrayer_AfterIsha(t *testing.T) {
	prayers := makePrayers(time.UTC)
	now := prayers[4].Time.Add(1 * time.Hour)
	cur := CurrentPrayer(prayers, now)
	if cur == nil || cur.Name != "Isha" {
		t.Errorf("expected Isha, got %v", cur)
	}
}

func TestFormatCountdown(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{1*time.Hour + 30*time.Minute + 45*time.Second, "01:30:45"},
		{0, "00:00:00"},
		{-5 * time.Second, "00:00:00"},
		{2 * time.Hour, "02:00:00"},
		{59*time.Second, "00:00:59"},
	}
	for _, tc := range tests {
		got := FormatCountdown(tc.d)
		if got != tc.want {
			t.Errorf("FormatCountdown(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func TestFormatTime(t *testing.T) {
	loc := time.UTC
	tests := []struct {
		h, min int
		use12h bool
		want   string
	}{
		{15, 4, true, "3:04 PM"},
		{15, 4, false, "15:04"},
		{0, 0, true, "12:00 AM"},
		{12, 0, true, "12:00 PM"},
		{9, 5, true, "9:05 AM"},
		{9, 5, false, "09:05"},
	}
	for _, tc := range tests {
		tm := time.Date(2025, 1, 1, tc.h, tc.min, 0, 0, loc)
		got := FormatTime(tm, tc.use12h)
		if got != tc.want {
			t.Errorf("FormatTime(%02d:%02d, use12h=%v) = %q, want %q", tc.h, tc.min, tc.use12h, got, tc.want)
		}
	}
}

func TestCountdown(t *testing.T) {
	loc := time.UTC
	now := time.Date(2025, 1, 1, 10, 0, 0, 0, loc)
	future := Prayer{Name: "Dhuhr", Time: time.Date(2025, 1, 1, 12, 30, 0, 0, loc)}
	past := Prayer{Name: "Fajr", Time: time.Date(2025, 1, 1, 5, 30, 0, 0, loc)}

	if d := Countdown(future, now); d <= 0 {
		t.Errorf("expected positive duration for future prayer, got %v", d)
	}
	if d := Countdown(past, now); d >= 0 {
		t.Errorf("expected negative duration for past prayer, got %v", d)
	}
}
