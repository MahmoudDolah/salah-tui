package prayer

import (
	"fmt"
	"time"

	"github.com/MahmoudDolah/salah-tui/internal/api"
)

// Prayer represents a single prayer with its name and scheduled time.
type Prayer struct {
	Name string
	Time time.Time
}

// ParseTimes converts PrayerTimes string fields into Prayer structs for the given date.
// timeStr is in "HH:MM" 24h format from the API.
// loc is the timezone location to use.
// If showSunrise is true, Sunrise is included in the returned list.
func ParseTimes(pt *api.PrayerTimes, date time.Time, loc *time.Location, showSunrise bool) ([]Prayer, error) {
	year, month, day := date.In(loc).Date()

	parse := func(name, timeStr string) (Prayer, error) {
		var hour, min int
		if _, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &min); err != nil {
			return Prayer{}, fmt.Errorf("parse %s time %q: %w", name, timeStr, err)
		}
		t := time.Date(year, month, day, hour, min, 0, 0, loc)
		return Prayer{Name: name, Time: t}, nil
	}

	var prayers []Prayer

	fajr, err := parse("Fajr", pt.Fajr)
	if err != nil {
		return nil, err
	}
	prayers = append(prayers, fajr)

	if showSunrise {
		sunrise, err := parse("Sunrise", pt.Sunrise)
		if err != nil {
			return nil, err
		}
		prayers = append(prayers, sunrise)
	}

	dhuhr, err := parse("Dhuhr", pt.Dhuhr)
	if err != nil {
		return nil, err
	}
	prayers = append(prayers, dhuhr)

	asr, err := parse("Asr", pt.Asr)
	if err != nil {
		return nil, err
	}
	prayers = append(prayers, asr)

	maghrib, err := parse("Maghrib", pt.Maghrib)
	if err != nil {
		return nil, err
	}
	prayers = append(prayers, maghrib)

	isha, err := parse("Isha", pt.Isha)
	if err != nil {
		return nil, err
	}
	prayers = append(prayers, isha)

	return prayers, nil
}

// Next returns the next upcoming prayer from a list of today's prayers.
// If all prayers have passed, returns nil (caller should handle midnight edge case).
func Next(prayers []Prayer, now time.Time) *Prayer {
	for i := range prayers {
		if prayers[i].Time.After(now) {
			return &prayers[i]
		}
	}
	return nil
}

// Countdown returns the duration until the given prayer.
func Countdown(prayer Prayer, now time.Time) time.Duration {
	return prayer.Time.Sub(now)
}

// FormatCountdown formats a duration as "HH:MM:SS".
func FormatCountdown(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSeconds := int(d.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// FormatTime formats a time.Time as a 12h or 24h string.
// If use12h is true, returns "3:04 PM" style; otherwise "15:04".
func FormatTime(t time.Time, use12h bool) string {
	if use12h {
		return t.Format("3:04 PM")
	}
	return t.Format("15:04")
}

// CurrentPrayer returns which prayer period we are currently in (the last prayer that has passed).
// Returns nil if before Fajr.
func CurrentPrayer(prayers []Prayer, now time.Time) *Prayer {
	var current *Prayer
	for i := range prayers {
		if !prayers[i].Time.After(now) {
			current = &prayers[i]
		} else {
			break
		}
	}
	return current
}
