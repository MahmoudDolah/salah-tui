package api_test

import (
	"testing"
	"time"

	"github.com/mhdolah/salah-tui/internal/api"
)

func TestDailyIndex(t *testing.T) {
	tests := []struct {
		date time.Time
		want int
	}{
		// Jan 1 = day 1: (1 % 6236) + 1 = 2
		{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), 2},
		// Jan 2 = day 2: (2 % 6236) + 1 = 3
		{time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), 3},
		// Dec 31 non-leap = day 365: (365 % 6236) + 1 = 366
		{time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), 366},
		// Dec 31 leap year = day 366: (366 % 6236) + 1 = 367
		{time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), 367},
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
