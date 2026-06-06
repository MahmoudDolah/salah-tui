package api_test

import (
	"testing"
	"time"

	"github.com/MahmoudDolah/salah-tui/internal/api"
)

func TestDailyIndex(t *testing.T) {
	tests := []struct {
		date time.Time
		want int
	}{
		// Jan 1 = day 1: ((1-1) % 6236) + 1 = 1
		{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), 1},
		// Jan 2 = day 2: ((2-1) % 6236) + 1 = 2
		{time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), 2},
		// Dec 31 non-leap = day 365: ((365-1) % 6236) + 1 = 365
		{time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), 365},
		// Dec 31 leap year = day 366: ((366-1) % 6236) + 1 = 366
		{time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), 366},
	}
	for _, tc := range tests {
		got := api.DailyIndex(tc.date)
		if got != tc.want {
			t.Errorf("DailyIndex(%v) = %d, want %d", tc.date.Format("2006-01-02"), got, tc.want)
		}
		if got < 1 || got > 6236 {
			t.Errorf("DailyIndex(%v) = %d out of valid range [1, 6236]", tc.date.Format("2006-01-02"), got)
		}
	}
}
