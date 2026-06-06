package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/MahmoudDolah/salah-tui/internal/prayer"
)

// RenderSchedule renders the left pane: date header + prayer list + countdown.
// width is the inner width available (excluding border padding).
func RenderSchedule(
	prayers []prayer.Prayer,
	hijriDate, gregorianDate string,
	now time.Time,
	use12h bool,
	offline bool,
	width int,
) string {
	var b strings.Builder

	// Date header
	dateLine := StyleHeader.Render(hijriDate + "  " + gregorianDate)
	if offline {
		dateLine += "  " + StyleOffline.Render("(offline)")
	}
	b.WriteString(dateLine + "\n\n")

	next := prayer.Next(prayers, now)
	current := prayer.CurrentPrayer(prayers, now)

	for _, p := range prayers {
		timeStr := prayer.FormatTime(p.Time, use12h)
		name := p.Name

		var nameFmt, timeFmt string

		switch {
		case next != nil && p.Name == next.Name:
			// Next prayer: cyan label + live countdown
			d := prayer.Countdown(p, now)
			mins := d.Minutes()
			countdown := prayer.FormatCountdown(d)
			countdownStyled := CountdownStyle(mins).Render(countdown)

			nameFmt = StyleNextPrayerLabel.Render(fmt.Sprintf("▶ %-9s", name))
			timeFmt = StyleNextPrayerLabel.Render(timeStr) + "  " + countdownStyled

		case current != nil && p.Name == current.Name:
			// Current (most recent past) prayer: green
			nameFmt = StyleCurrentPrayer.Render(fmt.Sprintf("  %-9s", name))
			timeFmt = StyleCurrentPrayer.Render(timeStr)

		default:
			nameFmt = StyleDimText.Render(fmt.Sprintf("  %-9s", name))
			timeFmt = StyleDimText.Render(timeStr)
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			nameFmt,
			"  ",
			timeFmt,
		)
		b.WriteString(row + "\n")
	}

	_ = width // available for future alignment
	return b.String()
}
